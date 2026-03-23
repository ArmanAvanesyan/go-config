package deep_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestDeepOverride(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		dst        map[string]any
		src        map[string]any
		want       map[string]any
		checkClone bool // whether to verify output is a clone of dst
	}{
		{
			name: "nil dst",
			dst:  nil,
			src:  map[string]any{"a": "one"},
			want: map[string]any{"a": "one"},
		},
		{
			name:       "nil src returns cloned dst (no-op)",
			dst:        map[string]any{"a": "one"},
			src:        nil,
			want:       map[string]any{"a": "one"},
			checkClone: true,
		},
		{
			name:       "flat override",
			dst:        map[string]any{"a": "old", "b": "keep"},
			src:        map[string]any{"a": "new"},
			want:       map[string]any{"a": "new", "b": "keep"},
			checkClone: true,
		},
		{
			name:       "nested merge",
			dst:        map[string]any{"server": map[string]any{"host": "localhost", "port": 8080}},
			src:        map[string]any{"server": map[string]any{"port": 9090}},
			want:       map[string]any{"server": map[string]any{"host": "localhost", "port": 9090}},
			checkClone: true,
		},
		{
			name:       "nested override with scalar in dst",
			dst:        map[string]any{"server": "old-scalar"},
			src:        map[string]any{"server": map[string]any{"host": "localhost"}},
			want:       map[string]any{"server": map[string]any{"host": "localhost"}},
			checkClone: true,
		},
		{
			name: "deep nested merge",
			dst: map[string]any{
				"a": map[string]any{
					"b": map[string]any{"x": 1, "y": 2},
				},
			},
			src: map[string]any{
				"a": map[string]any{
					"b": map[string]any{"y": 99},
				},
			},
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{"x": 1, "y": 99},
				},
			},
			checkClone: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Keep a reference to the original dst map (by value copy of the map reference).
			origDst := tc.dst

			got, err := deep.New().Merge(tc.dst, tc.src)
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)

			// When dst is non-nil, the result must be a new map so modifying
			// it does not affect the original dst.
			if tc.checkClone && origDst != nil {
				got["__sentinel__"] = true
				if _, ok := origDst["__sentinel__"]; ok {
					t.Fatal("Merge returned a reference to dst (not a clone)")
				}
			}
		})
	}
}

func TestDeepOverride_DoesNotAliasSrcNestedMap(t *testing.T) {
	t.Parallel()

	src := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 9090,
		},
	}
	got, err := deep.New().Merge(map[string]any{}, src)
	testutil.RequireNoError(t, err)

	gotServer, ok := got["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected result server value to be map, got %T", got["server"])
	}
	gotServer["host"] = "changed"

	srcServer := src["server"].(map[string]any)
	if srcServer["host"] == "changed" {
		t.Fatal("result must not alias src nested maps")
	}
}

func TestDeepOverride_DoesNotAliasSrcNestedMapOnScalarToMapReplace(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"server": "old-scalar"}
	src := map[string]any{
		"server": map[string]any{
			"host": "localhost",
		},
	}

	got, err := deep.New().Merge(dst, src)
	testutil.RequireNoError(t, err)

	gotServer := got["server"].(map[string]any)
	gotServer["host"] = "changed"

	srcServer := src["server"].(map[string]any)
	if srcServer["host"] == "changed" {
		t.Fatal("result must not alias src nested maps when replacing scalar dst value")
	}
}

func BenchmarkDeepOverride_Flat(b *testing.B) {
	dst := map[string]any{"a": "one", "b": "two", "c": "three"}
	src := map[string]any{"a": "override", "d": "new"}
	strategy := deep.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = strategy.Merge(dst, src)
	}
}

func BenchmarkDeepOverride_DeepNested(b *testing.B) {
	dst := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{"key": "value"},
			},
		},
	}
	src := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{"key": "overridden", "extra": "new"},
			},
		},
	}
	strategy := deep.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = strategy.Merge(dst, src)
	}
}
