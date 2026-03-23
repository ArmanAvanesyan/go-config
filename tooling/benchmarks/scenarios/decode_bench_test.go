package benchmarks_test

import (
	"context"
	"testing"
)

type decodeSlicesCfg struct {
	Tags []string `json:"tags" mapstructure:"tags"`
}

type decodeMapsCfg struct {
	Meta map[string]string `json:"meta" mapstructure:"meta"`
}

type decodePointerCfg struct {
	Name   *string `json:"name" mapstructure:"name"`
	Nested *struct {
		Host string `json:"host" mapstructure:"host"`
	} `json:"nested" mapstructure:"nested"`
}

type decodeOptionalCfg struct {
	Timeout *int `json:"timeout" mapstructure:"timeout"`
}

type decodeTagHeavyCfg struct {
	Name string `json:"name" mapstructure:"name" toml:"name" yaml:"name"`
}

func BenchmarkDecode_StructShapes_JSON(b *testing.B) {
	raw := mustReadFixture(b, "fixtures/json/medium.json")
	ctx := context.Background()
	loader := newGoConfigJSONLoader(raw)

	b.Run("Decode/FlatStruct", func(b *testing.B) {
		b.ReportAllocs()
		var out map[string]any
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Decode/NestedStruct", func(b *testing.B) {
		b.ReportAllocs()
		var out decodeCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/StructWithSlices", func(b *testing.B) {
		b.ReportAllocs()
		var out decodeSlicesCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/StructWithMaps", func(b *testing.B) {
		b.ReportAllocs()
		var out decodeMapsCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/PointerFields", func(b *testing.B) {
		b.ReportAllocs()
		var out decodePointerCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/OptionalFields", func(b *testing.B) {
		b.ReportAllocs()
		var out decodeOptionalCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/TagHeavy", func(b *testing.B) {
		b.ReportAllocs()
		var out decodeTagHeavyCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := loader.Load(ctx, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
