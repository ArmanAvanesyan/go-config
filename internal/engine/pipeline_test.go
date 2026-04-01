package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
)

type docSource struct {
	v   any
	err error
}

func (s *docSource) Read(context.Context) (any, error) {
	return s.v, s.err
}

func TestLoadAndMerge_UnknownBindingTypeRejected(t *testing.T) {
	t.Parallel()
	type altBinding struct {
		source config.Source
	}
	bindings := []altBinding{
		{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"a": "1"}}}},
		{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"b": "2"}}}},
	}
	if _, err := LoadAndMerge(context.Background(), bindings, deep.New()); err == nil {
		t.Fatalf("expected unknown binding type to be rejected")
	}
}

func TestLoadAndMergeBindingType(t *testing.T) {
	t.Parallel()
	bindings := []Binding{{Source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"k": "v"}}}}}
	if got, err := LoadAndMerge(context.Background(), bindings, deep.New()); err != nil || got["k"] != "v" {
		t.Fatalf("binding path failed got=%v err=%v", got, err)
	}
}

func TestLoadAndMergePropagatesReadErrors(t *testing.T) {
	t.Parallel()
	bindings := []Binding{{Source: &docSource{err: errors.New("boom")}}}
	if _, err := LoadAndMerge(context.Background(), bindings, deep.New()); !errors.Is(err, config.ErrSourceReadFailed) {
		t.Fatalf("expected wrapped source read error, got %v", err)
	}
}
