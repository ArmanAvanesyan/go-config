// Package toml provides a TOML parser backed by the Rust toml-rs crate
// running as a WASM module via wazero. This replaces the previous
// github.com/pelletier/go-toml/v2 dependency.
//
// The underlying WASM binary is embedded in extensions/wasm/parser/rusttoml.
// Build it with:
//
//	cd rust && make build-toml
package toml

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/extensions/wasm/parser/rusttoml"
)

// Parser parses TOML documents. It delegates to the Rust/WASM rusttoml.Parser.
type Parser struct {
	inner *rusttoml.Parser
}

// New initializes the TOML parser, compiling the WASM module once.
// The caller should call Close on the returned Parser when done.
func New(ctx context.Context) (*Parser, error) {
	p, err := rusttoml.New(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{inner: p}, nil
}

// Parse implements config.Parser.
func (p *Parser) Parse(ctx context.Context, doc *config.Document) (map[string]any, error) {
	return p.inner.Parse(ctx, doc)
}

// Close releases WASM runtime resources.
func (p *Parser) Close(ctx context.Context) error {
	return p.inner.Close(ctx)
}

var _ config.Parser = (*Parser)(nil)
