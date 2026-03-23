package benchmarks_test

import (
	"context"
	"encoding/json"
	"testing"
)

func BenchmarkScale_FlatJSON(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{name: "Keys10", n: 10},
		{name: "Keys100", n: 100},
		{name: "Keys1000", n: 1000},
	}
	for _, tc := range sizes {
		tc := tc
		b.Run("Scale/"+tc.name, func(b *testing.B) {
			raw := genFlatJSON(tc.n)
			ctx := context.Background()
			loader := newGoConfigJSONLoader(raw)
			b.ReportAllocs()
			var out map[string]string
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loader.Load(ctx, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkScale_DepthJSON(b *testing.B) {
	depths := []int{2, 5, 10}
	for _, d := range depths {
		d := d
		b.Run("Scale/Depth_"+itoa(d), func(b *testing.B) {
			raw := genNestedJSON(d)
			ctx := context.Background()
			loader := newGoConfigJSONLoader(raw)
			b.ReportAllocs()
			var out map[string]any
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loader.Load(ctx, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func genNestedJSON(depth int) []byte {
	tree := map[string]any{"leaf": "value"}
	for i := 0; i < depth; i++ {
		tree = map[string]any{"lvl": tree}
	}
	b, err := json.Marshal(tree)
	if err != nil {
		panic(err)
	}
	return b
}
