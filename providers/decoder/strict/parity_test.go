package strict_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/decoder/strict"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestStrictDecoder_RejectsUnknownDeepField(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Name string `json:"name"`
	}
	type Cfg struct {
		Inner Inner `json:"inner"`
	}

	input := map[string]any{
		"inner": map[string]any{
			"name":    "demo",
			"unknown": "x",
		},
	}

	var cfg Cfg
	err := strict.New().Decode(input, &cfg)
	testutil.RequireError(t, err)
}

func TestStrictDecoder_NoWeakTypeConversion(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Port int `json:"port"`
	}
	input := map[string]any{"port": "8080"}

	var cfg Cfg
	err := strict.New().Decode(input, &cfg)
	testutil.RequireError(t, err)
}
