package config

import (
	"context"
	"fmt"
)

// SourceSpec allows direct source registration in a high-level Spec.
type SourceSpec struct {
	Source Source
	Parser Parser
	Meta   *SourceMeta
}

// TreeSpec declares an in-memory tree source for Spec loading.
type TreeSpec struct {
	Tree     map[string]any
	Priority int
}

// Spec is a high-level app-facing load contract.
type Spec struct {
	Sources      []SourceSpec
	Trees        []TreeSpec
	Strict       bool
	DirectDecode bool
	Decoder      Decoder
	Resolver     Resolver
	Validator    Validator
	DefaultsFn   func(context.Context, any) error
	ValidateFn   func(context.Context, any) error
	Trace        *Trace
}

// LoadWithSpec builds a loader from Spec and loads output config.
func LoadWithSpec(ctx context.Context, out any, spec Spec) error {
	loader, err := NewFromSpec(spec)
	if err != nil {
		return err
	}
	return loader.Load(ctx, out)
}

// LoadTypedWithSpec builds a loader from Spec and returns typed output.
func LoadTypedWithSpec[T any](ctx context.Context, spec Spec) (T, error) {
	var t T
	err := LoadWithSpec(ctx, &t, spec)
	return t, err
}

// NewFromSpec builds a configured loader from a high-level Spec.
func NewFromSpec(spec Spec) (*Loader, error) {
	opts := []Option{
		WithStrict(spec.Strict),
		WithDirectDecode(spec.DirectDecode),
	}
	if spec.Decoder != nil {
		opts = append(opts, WithDecoder(spec.Decoder))
	}
	if spec.Resolver != nil {
		opts = append(opts, WithResolver(spec.Resolver))
	}
	if spec.Validator != nil {
		opts = append(opts, WithValidator(spec.Validator))
	}
	if spec.DefaultsFn != nil {
		opts = append(opts, WithDefaultsFunc(spec.DefaultsFn))
	}
	if spec.ValidateFn != nil {
		opts = append(opts, WithValidateFunc(spec.ValidateFn))
	}
	if spec.Trace != nil {
		opts = append(opts, WithTrace(spec.Trace))
	}
	l := New(opts...)

	for _, s := range spec.Sources {
		if s.Source == nil {
			return nil, fmt.Errorf("config: spec source must not be nil")
		}
		if s.Parser != nil {
			l.AddSourceWithMeta(s.Source, []Parser{s.Parser}, s.Meta)
		} else {
			l.AddSourceWithMeta(s.Source, nil, s.Meta)
		}
	}

	for _, m := range spec.Trees {
		meta := &SourceMeta{Priority: m.Priority, Required: true}
		l.AddSourceWithMeta(treeSource{tree: m.Tree}, nil, meta)
	}
	return l, nil
}

type treeSource struct {
	tree map[string]any
}

func (s treeSource) Read(context.Context) (any, error) {
	return &TreeDocument{Name: "spec-tree", Tree: s.tree}, nil
}
