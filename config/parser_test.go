package config

import (
	"context"
	"testing"
)

type noOpParser struct{}

func (p *noOpParser) Parse(context.Context, *Document) (map[string]any, error) {
	return map[string]any{}, nil
}

func TestParserInterfaces(t *testing.T) {
	t.Parallel()
	var _ Parser = (*noOpParser)(nil)
	var _ TypedParser = (*typedParserStub)(nil)
}
