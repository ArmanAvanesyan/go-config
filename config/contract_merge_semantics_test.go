package config

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestContract_MergeSemantics_PriorityAndStableOrder(t *testing.T) {
	type cfg struct {
		Value string `json:"value"`
	}
	var out cfg

	err := New().
		AddSourceWithMeta(&testMemSource{tree: map[string]any{"value": "first"}}, nil, &SourceMeta{Priority: 0, Required: true}).
		AddSourceWithMeta(&testMemSource{tree: map[string]any{"value": "second"}}, nil, &SourceMeta{Priority: 0, Required: true}).
		AddSourceWithMeta(&testMemSource{tree: map[string]any{"value": "high"}}, nil, &SourceMeta{Priority: 10, Required: true}).
		Load(context.Background(), &out)
	if err != nil {
		t.Fatalf("contract load failed: %v", err)
	}
	if out.Value != "high" {
		t.Fatalf("expected highest priority to win, got %q", out.Value)
	}
}

func TestContract_MergeSemantics_MissingAndParsePolicies(t *testing.T) {
	type cfg struct {
		Key string `json:"key"`
	}
	var out cfg

	err := New().
		AddSource(&testMemSource{tree: map[string]any{"key": "ok"}}).
		AddSourceWithMeta(&notFoundSource{}, nil, &SourceMeta{
			Required:      true,
			MissingPolicy: MissingPolicyIgnore,
		}).
		AddSourceWithMeta(&testDocSource{doc: &Document{Name: "broken.yaml", Format: "yaml", Raw: []byte(":::")}}, []Parser{&parseErrorParser{}}, &SourceMeta{
			Required:    true,
			ParsePolicy: ParsePolicyIgnore,
		}).
		Load(context.Background(), &out)
	if err != nil {
		t.Fatalf("expected ignored missing/parse policies, got %v", err)
	}
	if out.Key != "ok" {
		t.Fatalf("expected key=ok, got %q", out.Key)
	}
}

func TestContract_MergeSemantics_MissingPolicyFail(t *testing.T) {
	type cfg struct {
		Key string `json:"key"`
	}
	var out cfg

	err := New().
		AddSource(&testMemSource{tree: map[string]any{"key": "ok"}}).
		AddSourceWithMeta(&notFoundSource{}, nil, &SourceMeta{
			Required:      false,
			MissingPolicy: MissingPolicyFail,
		}).
		Load(context.Background(), &out)
	if err == nil {
		t.Fatal("expected missing policy fail to return error")
	}
	if !errors.Is(err, ErrSourceReadFailed) || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected source read + not exists error, got %v", err)
	}
}
