// Package rusttoml provides a TOML parser backed by a Rust WASM module
// compiled from rust/parsers/toml-parser using the toml-rs crate.
//
// The WASM binary (toml_parser.wasm) must be present in this directory.
// Build it with:
//
//	cd rust && make build-toml
package rusttoml

import (
	"context"
	_ "embed"

	"github.com/ArmanAvanesyan/go-config/config"
	wazeroengine "github.com/ArmanAvanesyan/go-config/extensions/wasm/runtime/wazero"
)

//go:embed toml_parser.wasm
var wasmBinary []byte

// Parser parses TOML documents using the Rust toml-rs crate via WASM.
// It implements config.Parser.
type Parser struct {
	eng *wazeroengine.Engine
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

// Parse implements config.Parser.
func (p *Parser) Parse(ctx context.Context, doc *config.Document) (map[string]any, error) {
	return p.eng.ParseConfig(ctx, doc.Raw)
}

// Close releases the underlying WASM runtime resources.
func (p *Parser) Close(ctx context.Context) error {
	return p.eng.Close(ctx)
}

var _ config.Parser = (*Parser)(nil)
