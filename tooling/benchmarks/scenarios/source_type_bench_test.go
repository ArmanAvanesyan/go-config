package benchmarks_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	"github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
)

func BenchmarkSourceType_BytesSource_JSON(b *testing.B) {
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

func BenchmarkSourceType_FileSource_JSON(b *testing.B) {
	ctx := context.Background()
	srcPath := filepath.Clean(filepath.Join("..", "fixtures", "json", "small.json"))
	loader := config.New().AddSource(file.New(srcPath), formatjson.New())
	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSourceType_MemorySource_Tree(b *testing.B) {
	ctx := context.Background()
	loader := config.New().AddSource(memory.New(map[string]any{
		"app": map[string]any{"name": "demo"},
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}))
	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSourceType_EnvSource(b *testing.B) {
	ctx := context.Background()
	b.Setenv("BENCH_APP__NAME", "demo-env")
	b.Setenv("BENCH_SERVER__HOST", "localhost")
	b.Setenv("BENCH_SERVER__PORT", "8080")

	loader := config.New().AddSource(env.New("BENCH"))
	b.ReportAllocs()
	var out map[string]any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSourceType_ColdStart_FileSource(b *testing.B) {
	ctx := context.Background()
	srcPath := filepath.Clean(filepath.Join("..", "fixtures", "json", "small.json"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := config.New().AddSource(file.New(srcPath), formatjson.New())
		var out benchCfg
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}
