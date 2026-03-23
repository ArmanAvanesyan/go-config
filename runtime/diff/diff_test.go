package diff

import (
	"reflect"
	"sort"
	"testing"
)

func TestChanges_Empty(t *testing.T) {
	t.Parallel()
	old := map[string]any{}
	new := map[string]any{}
	got := Changes(old, new)
	if len(got) != 0 {
		t.Fatalf("expected no changes, got %d: %v", len(got), got)
	}
}

func TestChanges_Add(t *testing.T) {
	t.Parallel()
	old := map[string]any{}
	new := map[string]any{"a": 1, "b": "x"}
	got := Changes(old, new)
	sort.Slice(got, func(i, j int) bool { return got[i].Path < got[j].Path })
	want := []Change{
		{Path: "a", Kind: KindAdd, NewValue: 1},
		{Path: "b", Kind: KindAdd, NewValue: "x"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestChanges_Remove(t *testing.T) {
	t.Parallel()
	old := map[string]any{"a": 1, "b": "x"}
	new := map[string]any{}
	got := Changes(old, new)
	sort.Slice(got, func(i, j int) bool { return got[i].Path < got[j].Path })
	want := []Change{
		{Path: "a", Kind: KindRemove, OldValue: 1},
		{Path: "b", Kind: KindRemove, OldValue: "x"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestChanges_Change(t *testing.T) {
	t.Parallel()
	old := map[string]any{"port": 8080}
	new := map[string]any{"port": 9090}
	got := Changes(old, new)
	want := []Change{
		{Path: "port", Kind: KindChange, OldValue: 8080, NewValue: 9090},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestChanges_NoChange(t *testing.T) {
	t.Parallel()
	old := map[string]any{"port": 8080, "nested": map[string]any{"x": 1}}
	new := map[string]any{"port": 8080, "nested": map[string]any{"x": 1}}
	got := Changes(old, new)
	if len(got) != 0 {
		t.Fatalf("expected no changes, got %v", got)
	}
}

func TestChanges_Nested(t *testing.T) {
	t.Parallel()
	old := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"removed": true,
	}
	new := map[string]any{
		"server": map[string]any{
			"port": 9090,
			"host": "localhost",
		},
		"added": 42,
	}
	got := Changes(old, new)
	// Order may vary; check we have exactly the expected set.
	byPath := make(map[string]Change)
	for _, c := range got {
		byPath[c.Path] = c
	}
	wantPaths := map[string]Change{
		"server.port": {Path: "server.port", Kind: KindChange, OldValue: 8080, NewValue: 9090},
		"removed":     {Path: "removed", Kind: KindRemove, OldValue: true},
		"added":       {Path: "added", Kind: KindAdd, NewValue: 42},
	}
	for path, want := range wantPaths {
		c, ok := byPath[path]
		if !ok {
			t.Fatalf("missing change at %q", path)
		}
		if c.Kind != want.Kind || !reflect.DeepEqual(c.OldValue, want.OldValue) || !reflect.DeepEqual(c.NewValue, want.NewValue) {
			t.Fatalf("at %q: got %+v, want %+v", path, c, want)
		}
	}
	if len(got) != len(wantPaths) {
		t.Fatalf("got %d changes, want %d: %v", len(got), len(wantPaths), got)
	}
}

func TestChanges_NilMaps(t *testing.T) {
	t.Parallel()
	// Treat nil as empty
	got := Changes(nil, map[string]any{"a": 1})
	if len(got) != 1 || got[0].Kind != KindAdd || got[0].Path != "a" {
		t.Fatalf("unexpected: %v", got)
	}
	got = Changes(map[string]any{"a": 1}, nil)
	if len(got) != 1 || got[0].Kind != KindRemove || got[0].Path != "a" {
		t.Fatalf("unexpected: %v", got)
	}
}
