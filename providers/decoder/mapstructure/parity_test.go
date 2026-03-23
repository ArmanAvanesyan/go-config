package mapstructure_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/decoder/mapstructure"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestMapstructureDecoder_WeakTypingAndPointers(t *testing.T) {
	t.Parallel()

	type Item struct {
		Port *int `json:"port"`
	}
	type Cfg struct {
		Enabled bool              `json:"enabled"`
		Tags    []string          `json:"tags"`
		Meta    map[string]string `json:"meta"`
		Item    Item              `json:"item"`
	}

	input := map[string]any{
		"enabled": "true",
		"tags":    []any{"a", "b"},
		"meta": map[string]any{
			"x": "1",
			"y": "2",
		},
		"item": map[string]any{
			"port": "8080",
		},
	}

	var cfg Cfg
	err := mapstructure.New().Decode(input, &cfg)
	testutil.RequireNoError(t, err)
	testutil.RequireEqual(t, true, cfg.Enabled)
	testutil.RequireEqual(t, 2, len(cfg.Tags))
	testutil.RequireEqual(t, "1", cfg.Meta["x"])
	if cfg.Item.Port == nil {
		t.Fatal("expected non-nil pointer field")
	}
	testutil.RequireEqual(t, 8080, *cfg.Item.Port)
}

func TestMapstructureDecoder_MapstructureTagPriority(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Name string `mapstructure:"app_name" json:"ignored"`
	}
	input := map[string]any{"app_name": "demo"}

	var cfg Cfg
	err := mapstructure.New().Decode(input, &cfg)
	testutil.RequireNoError(t, err)
	testutil.RequireEqual(t, "demo", cfg.Name)
}
