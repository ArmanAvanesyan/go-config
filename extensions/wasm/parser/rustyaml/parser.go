// Package rustyaml provides a YAML parser backed by a Rust WASM module
// compiled from rust/parsers/yaml-parser using the serde_yaml crate.
//
// Build the WASM binary with:
//
//	cd rust && make build-yaml
//
// Note: downstream Go consumers do not need Rust installed because committed
// .wasm artifacts are embedded via go:embed. Rust is needed for contributors
// changing Rust parser code and for CI artifact verification.
package rustyaml

import (
	"context"
	_ "embed"
	"sync"

	"github.com/ArmanAvanesyan/go-config/config"
	wazeroengine "github.com/ArmanAvanesyan/go-config/extensions/wasm/runtime/wazero"
)

//go:embed yaml_parser.wasm
var wasmBinary []byte

// Parser parses YAML documents using the Rust serde_yaml crate via WASM.
// It implements config.Parser.
type Parser struct {
	eng *wazeroengine.Engine
}

//nolint:gochecknoglobals // shared parser instance and refcount are process-level by design.
var shared struct {
	mu   sync.Mutex
	p    *Parser
	refs int
}

// New creates a Parser, compiling the embedded WASM module once.
// The caller must call Close when the Parser is no longer needed.
func New(ctx context.Context) (*Parser, error) {
	eng, err := wazeroengine.NewFromBytes(ctx, wasmBinary)
	if err != nil {
		return nil, err
	}
	return &Parser{eng: eng}, nil
}

// NewShared returns a process-shared Parser instance with reference counting.
// Call ReleaseShared when done.
func NewShared(ctx context.Context) (*Parser, error) {
	shared.mu.Lock()
	if shared.p != nil {
		shared.refs++
		p := shared.p
		shared.mu.Unlock()
		return p, nil
	}
	shared.mu.Unlock()

	p, err := New(ctx)
	if err != nil {
		return nil, err
	}

	shared.mu.Lock()
	defer shared.mu.Unlock()
	// Handle races: if another goroutine created one first, prefer it.
	if shared.p != nil {
		shared.refs++
		_ = p.Close(ctx)
		return shared.p, nil
	}
	shared.p = p
	shared.refs = 1
	return shared.p, nil
}

// Parse implements config.Parser.
func (p *Parser) Parse(ctx context.Context, doc *config.Document) (map[string]any, error) {
	return p.eng.ParseConfig(ctx, doc.Raw)
}

// ParseTyped decodes the parser output directly into out.
func (p *Parser) ParseTyped(ctx context.Context, doc *config.Document, out any) error {
	return p.eng.ParseConfigInto(ctx, doc.Raw, out)
}

// ParseTransport returns parser transport bytes (including protocol prefix).
// This is primarily intended for internal decomposition benchmarks.
func (p *Parser) ParseTransport(ctx context.Context, doc *config.Document) ([]byte, error) {
	return p.eng.ParseConfigTransport(ctx, doc.Raw)
}

// DecodeTransportInto decodes parser transport bytes into out.
func DecodeTransportInto(transport []byte, out any) error {
	return wazeroengine.DecodeTransportInto(transport, out)
}

// Close releases the underlying WASM runtime resources.
func (p *Parser) Close(ctx context.Context) error {
	return p.eng.Close(ctx)
}

// ReleaseShared decrements the shared parser refcount and closes it on zero.
func ReleaseShared(ctx context.Context) error {
	shared.mu.Lock()
	if shared.p == nil {
		shared.mu.Unlock()
		return nil
	}
	shared.refs--
	if shared.refs > 0 {
		shared.mu.Unlock()
		return nil
	}
	p := shared.p
	shared.p = nil
	shared.refs = 0
	shared.mu.Unlock()
	return p.Close(ctx)
}

var _ config.Parser = (*Parser)(nil)
