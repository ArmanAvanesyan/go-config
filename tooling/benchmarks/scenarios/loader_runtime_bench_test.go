package benchmarks_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
)

func BenchmarkRuntime_LoadSingleSource(b *testing.B) {
	ctx := context.Background()
	loader := config.New().AddSource(memory.New(map[string]any{
		"name": "demo",
		"port": float64(8080),
	}))

	var out struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRuntime_LoadMultiSource(b *testing.B) {
	ctx := context.Background()
	loader := config.New().
		AddSource(memory.New(map[string]any{"a": "one", "b": "two"})).
		AddSource(memory.New(map[string]any{"b": "override", "c": "three"})).
		AddSource(memory.New(map[string]any{"d": "four"}))

	var out struct {
		A string `json:"a"`
		B string `json:"b"`
		C string `json:"c"`
		D string `json:"d"`
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}
