package normalize

import (
	"reflect"
	"testing"
)

func TestKeyAndKeys(t *testing.T) {
	t.Parallel()
	if got := Key("  HeLLo "); got != "hello" {
		t.Fatalf("Key normalization failed: %q", got)
	}
	if got := Keys(nil); !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("Keys(nil) mismatch: %#v", got)
	}
	in := map[string]any{
		" A ": 1,
		"B": map[string]any{
			" C ": 2,
		},
	}
	want := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2},
	}
	if got := Keys(in); !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() mismatch got=%#v want=%#v", got, want)
	}
}
