package merge_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/merge"
	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
	"github.com/ArmanAvanesyan/go-config/providers/merge/replace"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

// compile-time assertions: concrete strategies satisfy merge.Strategy.
var (
	_ merge.Strategy = deep.New()
	_ merge.Strategy = replace.New()
)

func TestStrategies_Contract_NilSafetyAndDeterminism(t *testing.T) {
	t.Parallel()

	type strategyCase struct {
		name string
		s    merge.Strategy
	}
	strategies := []strategyCase{
		{name: "deep", s: deep.New()},
		{name: "replace", s: replace.New()},
	}

	for _, sc := range strategies {
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()

			// Nil inputs should never error and should return a non-nil map.
			got, err := sc.s.Merge(nil, nil)
			testutil.RequireNoError(t, err)
			if got == nil {
				t.Fatal("expected non-nil map result for nil inputs")
			}

			// Determinism: same inputs should produce equal outputs.
			dst := map[string]any{"a": "one", "nested": map[string]any{"x": 1}}
			src := map[string]any{"a": "two", "nested": map[string]any{"y": 2}}
			out1, err := sc.s.Merge(dst, src)
			testutil.RequireNoError(t, err)
			out2, err := sc.s.Merge(dst, src)
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, out1, out2)
		})
	}
}
