package benchmarks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	formattoml "github.com/ArmanAvanesyan/go-config/providers/parser/toml"
	formatyaml "github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	configbytes "github.com/ArmanAvanesyan/go-config/providers/source/bytes"
	koanfjson "github.com/knadh/koanf/parsers/json"
	koanftoml "github.com/knadh/koanf/parsers/toml/v2"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/viper"
)

func genFlatJSON(n int) []byte {
	m := make(map[string]string, n)
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("k%d", i)] = fmt.Sprintf("val%d", i)
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

type benchCfg struct {
	App struct {
		Name string `json:"name" mapstructure:"name" toml:"name"`
	} `json:"app" mapstructure:"app" toml:"app"`
	Server struct {
		Host string `json:"host" mapstructure:"host" toml:"host"`
		Port int    `json:"port" mapstructure:"port" toml:"port"`
	} `json:"server" mapstructure:"server" toml:"server"`
}

type multiCfg struct {
	A string `json:"a" mapstructure:"a"`
	B string `json:"b" mapstructure:"b"`
	C string `json:"c" mapstructure:"c"`
	D string `json:"d" mapstructure:"d"`
}

type decodeCfg struct {
	Name    string            `json:"name" mapstructure:"name"`
	Enabled bool              `json:"enabled" mapstructure:"enabled"`
	Tags    []string          `json:"tags" mapstructure:"tags"`
	Meta    map[string]string `json:"meta" mapstructure:"meta"`
	Nested  struct {
		Host string `json:"host" mapstructure:"host"`
		Port int    `json:"port" mapstructure:"port"`
	} `json:"nested" mapstructure:"nested"`
}

func mustReadFixture(b *testing.B, rel string) []byte {
	b.Helper()
	try := []string{filepath.Clean(rel), filepath.Join("..", rel)}
	for _, p := range try {
		raw, err := os.ReadFile(filepath.Clean(p))
		if err == nil {
			return raw
		}
	}
	b.Fatalf("read %s: no such file in local/parent benchmark dirs", rel)
	return nil
}

func newGoConfigJSONLoader(raw []byte) *config.Loader {
	return config.New().AddSource(configbytes.New("fixture", "json", raw), formatjson.New())
}

func newGoConfigYAMLLoader(ctx context.Context, raw []byte) (*config.Loader, func(), error) {
	p, err := formatyaml.New(ctx)
	if err != nil {
		return nil, nil, err
	}
	loader := config.New(config.WithDirectDecode(true)).AddSource(configbytes.New("fixture", "yaml", raw), p)
	cleanup := func() { _ = p.Close(ctx) }
	return loader, cleanup, nil
}

func newGoConfigYAMLSharedLoader(ctx context.Context, raw []byte) (*config.Loader, func(), error) {
	p, err := formatyaml.NewShared(ctx)
	if err != nil {
		return nil, nil, err
	}
	loader := config.New(config.WithDirectDecode(true)).AddSource(configbytes.New("fixture", "yaml", raw), p)
	cleanup := func() { _ = p.Close(ctx) }
	return loader, cleanup, nil
}

func newGoConfigTOMLLoader(ctx context.Context, raw []byte) (*config.Loader, func(), error) {
	p, err := formattoml.New(ctx)
	if err != nil {
		return nil, nil, err
	}
	loader := config.New().AddSource(configbytes.New("fixture", "toml", raw), p)
	cleanup := func() { _ = p.Close(ctx) }
	return loader, cleanup, nil
}

func loadWithViper(raw []byte, kind string, out any) error {
	v := viper.New()
	v.SetConfigType(kind)
	if err := v.ReadConfig(bytes.NewReader(raw)); err != nil {
		return err
	}
	return v.Unmarshal(out)
}

func loadWithKoanf(raw []byte, kind string, out any) error {
	k := koanf.New(".")
	var err error
	switch kind {
	case "json":
		err = k.Load(rawbytes.Provider(raw), koanfjson.Parser())
	case "toml":
		err = k.Load(rawbytes.Provider(raw), koanftoml.Parser())
	case "yaml":
		err = k.Load(rawbytes.Provider(raw), koanfyaml.Parser())
	default:
		return fmt.Errorf("unsupported koanf kind: %s", kind)
	}
	if err != nil {
		return err
	}
	return k.UnmarshalWithConf("", out, koanf.UnmarshalConf{Tag: "json"})
}
