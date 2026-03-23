package benchmarks_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	rustyaml "github.com/ArmanAvanesyan/go-config/extensions/wasm/parser/rustyaml"
	formatyaml "github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/viper"
)

func BenchmarkCompare_SingleSource_JSON(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	b.Run("Compare/All/JSON/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/JSON/viper", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithViper(raw, "json", &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/JSON/koanf", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithKoanf(raw, "json", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_SingleSource_YAML(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.yaml")
	ctx := context.Background()

	loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
	skipGoConfig := err != nil
	if err != nil {
		b.Logf("go-config YAML skipped (build yaml_parser.wasm under extensions/wasm/parser/rustyaml): %v", err)
	}
	if cleanup != nil {
		b.Cleanup(cleanup)
	}

	b.Run("Compare/All/YAML/go-config", func(b *testing.B) {
		if skipGoConfig {
			b.Skip("YAML WASM parser not available")
		}
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/YAML/viper", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithViper(raw, "yaml", &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/YAML/koanf", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithKoanf(raw, "yaml", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_GoConfig_vs_Viper_JSON(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	b.Run("Compare/go-config_vs_viper/JSON/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Compare/go-config_vs_viper/JSON/viper", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithViper(raw, "json", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_GoConfig_vs_Koanf_JSON(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	b.Run("Compare/go-config_vs_koanf/JSON/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Compare/go-config_vs_koanf/JSON/koanf", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithKoanf(raw, "json", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_GoConfig_vs_Viper_YAML(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.yaml")
	ctx := context.Background()
	loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
	if err != nil {
		b.Skipf("YAML WASM parser not available: %v", err)
	}
	b.Cleanup(cleanup)

	b.Run("Compare/go-config_vs_viper/YAML/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Compare/go-config_vs_viper/YAML/viper", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithViper(raw, "yaml", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_GoConfig_vs_Koanf_YAML(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.yaml")
	ctx := context.Background()
	loader, cleanup, err := newGoConfigYAMLLoader(ctx, raw)
	if err != nil {
		b.Skipf("YAML WASM parser not available: %v", err)
	}
	b.Cleanup(cleanup)

	b.Run("Compare/go-config_vs_koanf/YAML/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Compare/go-config_vs_koanf/YAML/koanf", func(b *testing.B) {
		b.ReportAllocs()
		var out benchCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loadWithKoanf(raw, "yaml", &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompare_YAML_ParseOnly(b *testing.B) {
	raw := mustReadFixture(b, "testdata/basic.yaml")
	ctx := context.Background()

	b.Run("Compare/ParseOnly/YAML/go-config", func(b *testing.B) {
		yp, err := formatyaml.NewShared(ctx)
		if err != nil {
			b.Skipf("YAML parser unavailable: %v", err)
		}
		b.Cleanup(func() { _ = yp.Close(ctx) })
		doc := &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := yp.Parse(ctx, doc); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/ParseOnly/YAML/viper", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(bytes.NewReader(raw)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/ParseOnly/YAML/koanf", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := koanf.New(".")
			if err := k.Load(rawbytes.Provider(raw), koanfyaml.Parser()); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/ParseOnlyDecomposed/YAML/go-config_parse_transport", func(b *testing.B) {
		rp, err := rustyaml.NewShared(ctx)
		if err != nil {
			b.Skipf("YAML parser unavailable: %v", err)
		}
		b.Cleanup(func() { _ = rustyaml.ReleaseShared(ctx) })
		doc := &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := rp.ParseTransport(ctx, doc); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/ParseOnlyDecomposed/YAML/go-config_decode_transport", func(b *testing.B) {
		rp, err := rustyaml.NewShared(ctx)
		if err != nil {
			b.Skipf("YAML parser unavailable: %v", err)
		}
		b.Cleanup(func() { _ = rustyaml.ReleaseShared(ctx) })
		doc := &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw}
		transport, err := rp.ParseTransport(ctx, doc)
		if err != nil {
			b.Fatalf("prepare transport: %v", err)
		}
		var out map[string]any
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			out = nil
			if err := rustyaml.DecodeTransportInto(transport, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
