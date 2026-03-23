package benchmarks_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/viper"
)

func BenchmarkCompare_MultiSource(b *testing.B) {
	ctx := context.Background()

	s1 := memory.New(map[string]any{"a": "one", "b": "two"})
	s2 := memory.New(map[string]any{"b": "override", "c": "three"})
	s3 := memory.New(map[string]any{"d": "four"})
	loader := config.New().AddSource(s1).AddSource(s2).AddSource(s3)

	m1 := map[string]any{"a": "one", "b": "two"}
	m2 := map[string]any{"b": "override", "c": "three"}
	m3 := map[string]any{"d": "four"}

	b.Run("Compare/All/MultiSource/go-config", func(b *testing.B) {
		b.ReportAllocs()
		var out multiCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/MultiSource/viper", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := viper.New()
			if err := v.MergeConfigMap(m1); err != nil {
				b.Fatal(err)
			}
			if err := v.MergeConfigMap(m2); err != nil {
				b.Fatal(err)
			}
			if err := v.MergeConfigMap(m3); err != nil {
				b.Fatal(err)
			}
			var out multiCfg
			if err := v.Unmarshal(&out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Compare/All/MultiSource/koanf", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k := koanf.New(".")
			if err := k.Load(confmap.Provider(m1, "."), nil); err != nil {
				b.Fatal(err)
			}
			if err := k.Load(confmap.Provider(m2, "."), nil); err != nil {
				b.Fatal(err)
			}
			if err := k.Load(confmap.Provider(m3, "."), nil); err != nil {
				b.Fatal(err)
			}
			var out multiCfg
			if err := k.UnmarshalWithConf("", &out, koanf.UnmarshalConf{Tag: "json"}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMultiSource_Layers(b *testing.B) {
	ctx := context.Background()
	cases := []struct {
		name   string
		layers []map[string]any
	}{
		{name: "2Layers", layers: []map[string]any{{"a": "one"}, {"b": "two"}}},
		{name: "3Layers", layers: []map[string]any{{"a": "one"}, {"b": "two"}, {"c": "three"}}},
		{name: "5Layers", layers: []map[string]any{{"a": "one"}, {"b": "two"}, {"c": "three"}, {"d": "four"}, {"e": "five"}}},
	}

	for _, tc := range cases {
		b.Run("MultiSource/"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				loader := config.New()
				for _, m := range tc.layers {
					loader = loader.AddSource(memory.New(m))
				}
				var out map[string]any
				if err := loader.Load(ctx, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMergeBehavior_LastWriteWins(b *testing.B) {
	ctx := context.Background()
	loader := config.New().
		AddSource(memory.New(map[string]any{"app": map[string]any{"name": "first"}})).
		AddSource(memory.New(map[string]any{"app": map[string]any{"name": "second"}}))

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMergeBehavior_DeepMerge_NestedMaps(b *testing.B) {
	ctx := context.Background()
	loader := config.New().
		AddSource(memory.New(map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"tls":  map[string]any{"enabled": true},
			},
		})).
		AddSource(memory.New(map[string]any{
			"server": map[string]any{
				"port": 8080,
				"tls":  map[string]any{"cert": "server.pem"},
			},
		}))

	b.ReportAllocs()
	var out map[string]any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMergeBehavior_MergeWithMissingKeys(b *testing.B) {
	ctx := context.Background()
	loader := config.New().
		AddSource(memory.New(map[string]any{"app": map[string]any{"name": "demo"}})).
		AddSource(memory.New(map[string]any{"server": map[string]any{"port": 8080}}))

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
	}
}
