package config

import (
	"github.com/ArmanAvanesyan/go-config/providers/decoder/mapstructure"
	"github.com/ArmanAvanesyan/go-config/providers/decoder/strict"
	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
	"github.com/ArmanAvanesyan/go-config/providers/validator/noop"
)

// New returns a Loader with the given options (decoder, validator, resolver, merge strategy, strict mode).
func New(opts ...Option) *Loader {
	o := Options{
		Decoder:       mapstructure.New(),
		Validator:     noop.New(),
		MergeStrategy: deep.New(),
	}

	for _, opt := range opts {
		opt(&o)
	}
	if o.Strict && !o.decoderSet {
		o.Decoder = strict.New()
	}

	return &Loader{
		options: o,
	}
}
