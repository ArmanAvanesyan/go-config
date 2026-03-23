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

type docParser struct {
	tree map[string]any
	err  error
}

func (p *docParser) Parse(context.Context, *config.Document) (map[string]any, error) {
	return p.tree, p.err
}

func TestLoadUnknownBindings_Branches(t *testing.T) {
	t.Parallel()
	pc := newPipelineContext(context.Background(), deep.New())

	if _, err := loadUnknownBindings(pc, 123); err == nil {
		t.Fatal("expected non-slice error")
	}

	type noSource struct{}
	if _, err := loadUnknownBindings(pc, []noSource{{}}); err == nil {
		t.Fatal("expected missing source field error")
	}

	type badSource struct {
		source int
	}
	if _, err := loadUnknownBindings(pc, []badSource{{source: 1}}); err == nil {
		t.Fatal("expected source type error")
	}

	type good struct {
		source config.Source
		parser any
	}
	if _, err := loadUnknownBindings(pc, []good{{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"a": 1}}}, parser: "x"}}); err == nil {
		t.Fatal("expected parser type error")
	}

	type elemNotStruct []int
	if _, err := loadUnknownBindings(pc, elemNotStruct{1}); err == nil {
		t.Fatal("expected element type error")
	}

	type ptrElem struct{}
	var nilPtr *ptrElem
	if _, err := loadUnknownBindings(pc, []*ptrElem{nilPtr}); err == nil {
		t.Fatal("expected nil pointer element error")
	}

	type goodParserField struct {
		source config.Source
	}
	if _, err := loadUnknownBindings(pc, []goodParserField{{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"ok": 1}}}}}); err == nil {
		t.Fatalf("expected deterministic adapter error for non-engine unexported fields")
	}

	type exportedBinding struct {
		Source config.Source
		Parser config.Parser
	}
	if got, err := loadUnknownBindings(pc, []exportedBinding{{Source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"sb": 1}}}}}); err != nil || got["sb"] != 1 {
		t.Fatalf("expected Source/Parser adapter success, got=%v err=%v", got, err)
	}

	if got, err := loadUnknownBindings(pc, []exportedBinding{{
		Source: &docSource{v: &config.Document{}},
		Parser: &docParser{tree: map[string]any{"doc": 1}},
	}}); err != nil || got["doc"] != 1 {
		t.Fatalf("expected parser path success, got=%v err=%v", got, err)
	}
}

func TestLoadAndMerge_UnknownBindingAdapter(t *testing.T) {
	t.Parallel()
	type altBinding struct {
		source config.Source
	}
	bindings := []altBinding{
		{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"a": "1"}}}},
		{source: &docSource{v: &config.TreeDocument{Tree: map[string]any{"b": "2"}}}},
	}
	if _, err := LoadAndMerge(context.Background(), bindings, deep.New()); err == nil {
		t.Fatalf("expected deterministic adapter error for unexported reflection fields")
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
