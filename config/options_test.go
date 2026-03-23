package config

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
)

func TestOptionsApply(t *testing.T) {
	t.Parallel()
	o := ResolveOptionsForTest(
		WithDecoder(&noOpDecoder{}),
		WithValidator(&noOpValidator{}),
		WithResolver(&noOpResolver{}),
		WithMergeStrategy(deep.New()),
		WithStrict(true),
		WithDirectDecode(true),
	)
	if o.Decoder == nil || o.Validator == nil || o.Resolver == nil || o.MergeStrategy == nil || !o.Strict || !o.DirectDecode || !o.decoderSet {
		t.Fatalf("options not applied correctly: %+v", o)
	}
}

// ResolveOptionsForTest mirrors New option application without defaults.
func ResolveOptionsForTest(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
