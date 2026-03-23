package benchmarks_test

import (
	"context"
	"strconv"
	"testing"
)

func BenchmarkParallelLoad_JSON(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	parallelisms := []int{1, 4, 8}
	for _, p := range parallelisms {
		p := p
		b.Run("ParallelLoad/P"+itoa(p), func(b *testing.B) {
			b.SetParallelism(p)
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				var out benchCfg
				for pb.Next() {
					if err := loader.Load(ctx, &out); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkParallelReadAfterLoad(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)
	var out benchCfg
	if err := loader.Load(ctx, &out); err != nil {
		b.Fatal(err)
	}

	parallelisms := []int{4, 16}
	for _, p := range parallelisms {
		p := p
		b.Run("ParallelReadAfterLoad/P"+itoa(p), func(b *testing.B) {
			b.SetParallelism(p)
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = out.App.Name
					_ = out.Server.Host
					_ = out.Server.Port
				}
			})
		})
	}
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
