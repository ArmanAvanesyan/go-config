package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
)

type testSource struct {
	tree map[string]any
	err  error
}

func (s *testSource) Read(context.Context) (any, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &config.TreeDocument{
		Name: "test",
		Tree: s.tree,
	}, nil
}

func TestLoadAndMerge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		bindings []sourceBinding
		want     map[string]any
		wantErr  bool
	}{
		{
			name: "single source",
			bindings: []sourceBinding{
				{source: &testSource{tree: map[string]any{"a": "one"}}},
			},
			want: map[string]any{"a": "one"},
		},
		{
			name: "two sources merged",
			bindings: []sourceBinding{
				{source: &testSource{tree: map[string]any{"a": "one"}}},
				{source: &testSource{tree: map[string]any{"b": "two"}}},
			},
			want: map[string]any{"a": "one", "b": "two"},
		},
		{
			name: "second source overrides first",
			bindings: []sourceBinding{
				{source: &testSource{tree: map[string]any{"a": "old", "b": "keep"}}},
				{source: &testSource{tree: map[string]any{"a": "new"}}},
			},
			want: map[string]any{"a": "new", "b": "keep"},
		},
		{
			name: "error from source propagates",
			bindings: []sourceBinding{
				{source: &testSource{err: errors.New("read error")}},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := LoadAndMerge(context.Background(), tc.bindings, deep.New())
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("want %#v, got %#v", tc.want, got)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("key %q: want %v, got %v", k, v, got[k])
				}
			}
		})
	}
}
