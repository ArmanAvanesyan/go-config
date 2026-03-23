package config

import (
	"context"
	"errors"
	"testing"
)

type noOpDecoder struct{}

func (d *noOpDecoder) Decode(map[string]any, any) error { return nil }

type typedParserStub struct {
	parseTypedCalled bool
}

func (p *typedParserStub) Parse(context.Context, *Document) (map[string]any, error) {
	return nil, errors.New("parse should not be called")
}

func (p *typedParserStub) ParseTyped(_ context.Context, _ *Document, out any) error {
	p.parseTypedCalled = true
	if s, ok := out.(*struct {
		V string `json:"v"`
	}); ok {
		s.V = "typed"
	}
	return nil
}

func TestDecoderInterfaces(t *testing.T) {
	t.Parallel()
	var _ Decoder = (*noOpDecoder)(nil)
	var _ TypedParser = (*typedParserStub)(nil)
}
