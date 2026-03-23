package benchmarks_test

import (
	"context"
	"testing"
)

func BenchmarkRuntimeAccess_RepeatedLoads_SameLoader(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRuntimeAccess_LoaderRebuildEveryIteration(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := newGoConfigJSONLoader(raw)
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRuntimeAccess_HotPath_ReadManyTimesAfterLoad(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)
	var out benchCfg
	if err := loader.Load(ctx, &out); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = out.App.Name
		_ = out.Server.Host
		_ = out.Server.Port
	}
}

func BenchmarkRuntimeAccess_ColdStart_OneLoadOnly(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/small.json")
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := newGoConfigJSONLoader(raw)
		var out benchCfg
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}
