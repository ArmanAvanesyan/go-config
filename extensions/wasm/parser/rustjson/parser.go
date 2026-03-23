// Package rustjson provides a JSON parser backed by a Rust WASM module
// compiled from rust/parsers/json-parser using serde_json.
//
// Build the WASM binary with:
//
//	cd rust && make build-json
package rustjson

import (
	"context"
	_ "embed"

	"github.com/ArmanAvanesyan/go-config/config"
	wazeroengine "github.com/ArmanAvanesyan/go-config/extensions/wasm/runtime/wazero"
)

//go:embed json_parser.wasm
var wasmBinary []byte

// Parser parses JSON documents using the Rust serde_json crate via WASM.
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
