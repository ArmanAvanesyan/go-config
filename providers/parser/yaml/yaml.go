// Package yaml provides a YAML parser backed by the Rust serde_yaml crate
// running as a WASM module via wazero. This replaces the previous
// gopkg.in/yaml.v3 dependency.
//
// The underlying WASM binary is embedded in extensions/wasm/parser/rustyaml.
// Build it with:
//
//	cd rust && make build-yaml
package yaml

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/extensions/wasm/parser/rustyaml"
	configbytes "github.com/ArmanAvanesyan/go-config/providers/source/bytes"
)

// Parser parses YAML documents. It delegates to the Rust/WASM rustyaml.Parser.
type Parser struct {
	inner  *rustyaml.Parser
	shared bool
}

// New initializes the YAML parser, compiling the WASM module once.
// The caller should call Close on the returned Parser when done.
func New(ctx context.Context) (*Parser, error) {
	p, err := rustyaml.New(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{inner: p}, nil
}

// NewShared returns a process-shared YAML parser backed by a shared Rust/WASM parser.
// Pair this with ReleaseShared when no longer needed.
func NewShared(ctx context.Context) (*Parser, error) {
	p, err := rustyaml.NewShared(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{inner: p, shared: true}, nil
}

// Parse implements config.Parser.
func (p *Parser) Parse(ctx context.Context, doc *config.Document) (map[string]any, error) {
	return p.inner.Parse(ctx, doc)
}

// ParseTyped decodes YAML into out directly from parser transport bytes.
func (p *Parser) ParseTyped(ctx context.Context, doc *config.Document, out any) error {
	return p.inner.ParseTyped(ctx, doc, out)
}

// Close releases WASM runtime resources.
func (p *Parser) Close(ctx context.Context) error {
	if p.shared {
		return rustyaml.ReleaseShared(ctx)
	}
	return p.inner.Close(ctx)
}

// ReleaseShared decrements the shared YAML parser reference count and closes
// shared resources when the count reaches zero.
func ReleaseShared(ctx context.Context) error {
	return rustyaml.ReleaseShared(ctx)
}

// AddSharedBytesSource is a convenience helper for repeated YAML loads.
// It wires a shared YAML parser into the loader and returns a cleanup callback.
func AddSharedBytesSource(ctx context.Context, l *config.Loader, name string, raw []byte) (func(), error) {
	p, err := NewShared(ctx)
	if err != nil {
		return nil, err
	}
	l.AddSource(configbytes.New(name, "yaml", raw), p)
	return func() {
		_ = p.Close(ctx)
	}, nil
}

var _ config.Parser = (*Parser)(nil)
