package config

import "github.com/ArmanAvanesyan/go-config/providers/merge"

// Options holds loader configuration (decoder, validator, resolver, merge strategy, strict).
type Options struct {
	Decoder       Decoder
	decoderSet    bool
	Validator     Validator
	Resolver      Resolver
	MergeStrategy merge.Strategy
	Strict        bool
	DirectDecode  bool
}

// Option configures a Loader via New(opts...).
type Option func(*Options)

// WithDecoder sets the decoder used to decode the merged config tree into the target struct.
func WithDecoder(d Decoder) Option {
	return func(o *Options) {
		o.Decoder = d
		o.decoderSet = true
	}
}

// WithValidator sets the validator run after decode.
func WithValidator(v Validator) Option {
	return func(o *Options) {
		o.Validator = v
	}
}

// WithResolver sets the resolver used to expand placeholders (e.g. env vars) in the config tree.
func WithResolver(r Resolver) Option {
	return func(o *Options) {
		o.Resolver = r
	}
}

// WithMergeStrategy sets how multiple sources are merged (e.g. deep override).
func WithMergeStrategy(s merge.Strategy) Option {
	return func(o *Options) {
		o.MergeStrategy = s
	}
}

// WithStrict enables strict decoding (e.g. fail on unknown fields) when true.
func WithStrict(strict bool) Option {
	return func(o *Options) {
		o.Strict = strict
	}
}

// WithDirectDecode enables an optional fast path where parsers implementing
// TypedParser may decode directly into the target when pipeline constraints allow it.
func WithDirectDecode(enabled bool) Option {
	return func(o *Options) {
		o.DirectDecode = enabled
	}
}
