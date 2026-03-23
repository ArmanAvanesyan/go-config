package strict_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/decoder/strict"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestStrictDecoder(t *testing.T) {
	t.Parallel()

	type Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	type Config struct {
		Server Server `json:"server"`
	}

	cases := []struct {
		name    string
		input   map[string]any
		wantErr bool
	}{
		{
			name: "known fields decode correctly",
			input: map[string]any{
				"server": map[string]any{"host": "localhost", "port": float64(8080)},
			},
			wantErr: false,
		},
		{
			name: "unknown field returns error",
			input: map[string]any{
				"server":  map[string]any{"host": "localhost", "port": float64(8080)},
				"unknown": "extra",
			},
			wantErr: true,
		},
		{
			name: "unknown nested field returns error",
			input: map[string]any{
				"server": map[string]any{"host": "localhost", "port": float64(8080), "extra": "bad"},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cfg Config
			err := strict.New().Decode(tc.input, &cfg)
			if tc.wantErr {
				testutil.RequireError(t, err)
			} else {
				testutil.RequireNoError(t, err)
				testutil.RequireEqual(t, "localhost", cfg.Server.Host)
				testutil.RequireEqual(t, 8080, cfg.Server.Port)
			}
		})
	}
}
