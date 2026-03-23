package tree

import (
	"reflect"
	"testing"
)

func TestCloneMap(t *testing.T) {
	t.Parallel()
	if got := CloneMap(nil); !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("CloneMap(nil) mismatch: %#v", got)
	}
	in := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2},
	}
	cloned := CloneMap(in)
	if !reflect.DeepEqual(cloned, in) {
		t.Fatalf("cloned map mismatch got=%#v want=%#v", cloned, in)
	}
	clonedNested := cloned["b"].(map[string]any)
	clonedNested["c"] = 99
	if in["b"].(map[string]any)["c"] == 99 {
		t.Fatal("clone must be deep for nested maps")
	}
}
