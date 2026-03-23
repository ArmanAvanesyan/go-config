package json

import (
	"context"
	"encoding/json"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Parser parses JSON config documents into a generic map.
type Parser struct{}

// New returns a new JSON Parser.
func New() *Parser {
	return &Parser{}
}

// Parse unmarshals the document's raw bytes as JSON into a map[string]any.
func (p *Parser) Parse(_ context.Context, doc *config.Document) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(doc.Raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
