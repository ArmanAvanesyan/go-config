package benchmarks_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkSingleSource_JSON_Size(b *testing.B) {
	benchSingleSourceBySize(b, "json", []string{
		"fixtures/json/small.json",
		"fixtures/json/medium.json",
		"fixtures/json/large.json",
	})
}

func BenchmarkSingleSource_YAML_Size(b *testing.B) {
	benchSingleSourceBySize(b, "yaml", []string{
		"fixtures/yaml/small.yaml",
		"fixtures/yaml/medium.yaml",
		"fixtures/yaml/large.yaml",
	})
}

func BenchmarkSingleSource_TOML_Size(b *testing.B) {
	paths := []string{
		"fixtures/toml/small.toml",
		"fixtures/toml/medium.toml",
		"fixtures/toml/large.toml",
	}
	for _, p := range paths {
		raw := mustReadFixture(b, p)
		size := sizeFromFixturePath(p)
		b.Run("SingleSource/TOML/"+size+"/go-config", func(b *testing.B) {
			ctx := context.Background()
			loader, cleanup, err := newGoConfigTOMLLoader(ctx, raw)
			if err != nil {
				b.Skipf("TOML WASM parser not available: %v", err)
			}
			b.Cleanup(cleanup)
			b.ReportAllocs()
			var out map[string]any
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loader.Load(ctx, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("SingleSource/TOML/"+size+"/viper", func(b *testing.B) {
			b.ReportAllocs()
			var out map[string]any
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loadWithViper(raw, "toml", &out); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("SingleSource/TOML/"+size+"/koanf", func(b *testing.B) {
			b.ReportAllocs()
			var out map[string]any
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loadWithKoanf(raw, "toml", &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchSingleSourceBySize(b *testing.B, kind string, paths []string) {
	for _, p := range paths {
		raw := mustReadFixture(b, p)
		size := sizeFromFixturePath(p)
		b.Run("SingleSource/"+strings.ToUpper(kind)+"/"+size, func(b *testing.B) {
			ctx := context.Background()
			switch kind {
			case "json":
				loader := newGoConfigJSONLoader(raw)
				b.ReportAllocs()
				var out map[string]any
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := loader.Load(ctx, &out); err != nil {
						b.Fatal(err)
					}
				}
			case "yaml":
				loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
				if err != nil {
					b.Skipf("YAML WASM parser not available: %v", err)
				}
				b.Cleanup(cleanup)
				b.ReportAllocs()
				var out map[string]any
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := loader.Load(ctx, &out); err != nil {
						b.Fatal(err)
					}
				}
			default:
				b.Fatalf("unsupported kind: %s", kind)
			}
		})
	}
}

func sizeFromFixturePath(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "" {
		return "Unknown"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
