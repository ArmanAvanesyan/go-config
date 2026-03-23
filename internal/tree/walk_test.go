package tree

import (
	"reflect"
	"sort"
	"testing"
)

func TestWalkAndJoin(t *testing.T) {
	t.Parallel()
	Walk(nil, func(string, any) {
		t.Fatal("walk on nil tree should not invoke visitor")
	})
	Walk(map[string]any{}, nil)

	in := map[string]any{
		"a": map[string]any{
			"b": 1,
		},
		"c": 2,
	}
	var paths []string
	Walk(in, func(path string, _ any) { paths = append(paths, path) })
	sort.Strings(paths)
	want := []string{"a", "a.b", "c"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("walk paths mismatch got=%v want=%v", paths, want)
	}

	if got := join("", "x"); got != "x" {
		t.Fatalf("join empty prefix failed: %q", got)
	}
	if got := join("a", "b"); got != "a.b" {
		t.Fatalf("join nested failed: %q", got)
	}
}
