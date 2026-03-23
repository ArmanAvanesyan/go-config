package benchmarks_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	resolverenv "github.com/ArmanAvanesyan/go-config/providers/resolver/env"
	configbytes "github.com/ArmanAvanesyan/go-config/providers/source/bytes"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
	"github.com/ArmanAvanesyan/go-config/providers/validator/noop"
	"github.com/ArmanAvanesyan/go-config/providers/validator/playground"
)

func BenchmarkFeature_Interpolation_EnvExpansion(b *testing.B) {
	ctx := context.Background()
	b.Setenv("BENCH_SERVICE_HOST", "127.0.0.1")
	raw := []byte(`{"server":{"host":"${ENV:BENCH_SERVICE_HOST}","port":8080}}`)
	loader := config.New(config.WithResolver(resolverenv.New())).
		AddSource(configbytes.New("fixture", "json", raw), formatjson.New())

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFeature_Defaults_WithAndWithoutDefaults(b *testing.B) {
	ctx := context.Background()
	withoutDefaults := config.New().
		AddSource(memory.New(map[string]any{"server": map[string]any{"port": 8080}}))
	withDefaults := config.New().
		AddSource(memory.New(map[string]any{"app": map[string]any{"name": "default-name"}, "server": map[string]any{"host": "localhost"}})).
		AddSource(memory.New(map[string]any{"server": map[string]any{"port": 8080}}))

	b.Run("Feature/Defaults/WithoutDefaults", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := withoutDefaults.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Feature/Defaults/WithDefaults", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := withDefaults.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFeature_Validation_EnabledVsDisabled(b *testing.B) {
	ctx := context.Background()
	raw := mustReadFixture(b, "fixtures/json/small.json")
	disabled := config.New(config.WithValidator(noop.New())).
		AddSource(configbytes.New("fixture", "json", raw), formatjson.New())
	enabled := config.New(config.WithValidator(playground.New(func(_ context.Context, _ any) error {
		return nil
	}))).
		AddSource(configbytes.New("fixture", "json", raw), formatjson.New())

	b.Run("Feature/Validation/Disabled", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := disabled.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Feature/Validation/Enabled", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := enabled.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFeature_YAML_WASMAvailableVsUnavailable(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/yaml/small.yaml")
	ctx := context.Background()

	b.Run("Feature/YAML/WASMAvailable", func(b *testing.B) {
		loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
		if err != nil {
			b.Skipf("YAML WASM parser not available: %v", err)
		}
		b.Cleanup(cleanup)
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Feature/YAML/WASMUnavailable", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
			if cleanup != nil {
				cleanup()
			}
			_ = err
		}
	})
}

func BenchmarkFeature_YAML_InitVsReuse(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/yaml/small.yaml")
	ctx := context.Background()

	b.Run("Feature/YAML/InitPerIteration", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
			if err != nil {
				b.Skipf("YAML WASM parser not available: %v", err)
			}
			var out benchCfg
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
			cleanup()
		}
	})

	b.Run("Feature/YAML/ReusePerBenchmark", func(b *testing.B) {
		loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
		if err != nil {
			b.Skipf("YAML WASM parser not available: %v", err)
		}
		b.Cleanup(cleanup)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var out benchCfg
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Feature/YAML/SharedGlobal", func(b *testing.B) {
		loader, cleanup, err := newGoConfigYAMLSharedLoader(ctx, raw)
		if err != nil {
			b.Skipf("YAML shared parser not available: %v", err)
		}
		b.Cleanup(cleanup)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var out benchCfg
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
