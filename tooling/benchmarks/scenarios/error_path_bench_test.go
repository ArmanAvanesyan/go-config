package benchmarks_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	configbytes "github.com/ArmanAvanesyan/go-config/providers/source/bytes"
)

func BenchmarkErrorPath_InvalidPayloads(b *testing.B) {
	cases := []struct {
		name string
		raw  []byte
	}{
		{name: "InvalidJSON", raw: []byte(`{"app":`)},
		{name: "InvalidYAML", raw: []byte("app:\n  name: demo\n  - bad")},
	}
	for _, tc := range cases {
		tc := tc
		b.Run("ErrorPath/"+tc.name, func(b *testing.B) {
			ctx := context.Background()
			loader := newGoConfigJSONLoader(tc.raw)
			b.ReportAllocs()
			var out map[string]any
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := loader.Load(ctx, &out); err == nil {
					b.Fatal("expected error")
				}
			}
		})
	}
}

type strictCfg struct {
	App struct {
		Name string `json:"name" mapstructure:"name"`
	} `json:"app" mapstructure:"app"`
}

func BenchmarkErrorPath_SchemaMismatch_Strict(b *testing.B) {
	ctx := context.Background()
	raw := []byte(`{"app":{"name":"demo"},"unknown":{"x":1}}`)
	loader := config.New(config.WithStrict(true)).
		AddSource(configbytes.New("fixture", "json", raw), formatjson.New())

	b.ReportAllocs()
	var out strictCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err == nil {
			b.Fatal("expected strict decode failure")
		}
	}
}

func BenchmarkErrorPath_MissingRequiredFields(b *testing.B) {
	ctx := context.Background()
	raw := []byte(`{"app":{}}`)
	loader := newGoConfigJSONLoader(raw)

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err != nil {
			b.Fatal(err)
		}
		if out.Server.Host != "" || out.Server.Port != 0 {
			b.Fatal("expected zero values for missing fields")
		}
	}
}

func BenchmarkErrorPath_TypeCoercionFailure(b *testing.B) {
	ctx := context.Background()
	raw := []byte(`{"server":{"host":"localhost","port":"not-an-int"}}`)
	loader := newGoConfigJSONLoader(raw)

	b.ReportAllocs()
	var out benchCfg
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := loader.Load(ctx, &out); err == nil {
			b.Fatal("expected type coercion failure")
		}
	}
}
