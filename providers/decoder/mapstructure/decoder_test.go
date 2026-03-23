package mapstructure_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/decoder/mapstructure"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestMapstructureDecoder(t *testing.T) {
	t.Parallel()

	type Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	type Config struct {
		Server  Server `json:"server"`
		AppName string `json:"app_name"`
	}

	cases := []struct {
		name    string
		input   map[string]any
		check   func(t *testing.T, cfg Config)
		wantErr bool
	}{
		{
			name: "flat decode",
			input: map[string]any{
				"app_name": "myapp",
				"server":   map[string]any{"host": "localhost", "port": float64(8080)},
			},
			check: func(t *testing.T, cfg Config) {
				testutil.RequireEqual(t, "myapp", cfg.AppName)
				testutil.RequireEqual(t, "localhost", cfg.Server.Host)
				testutil.RequireEqual(t, 8080, cfg.Server.Port)
			},
		},
		{
			name: "unknown field is silently ignored",
			input: map[string]any{
				"app_name": "myapp",
				"unknown":  "extra",
			},
			check: func(t *testing.T, cfg Config) {
				testutil.RequireEqual(t, "myapp", cfg.AppName)
			},
		},
		{
			name: "missing field leaves zero value",
			input: map[string]any{
				"server": map[string]any{"host": "example.com"},
			},
			check: func(t *testing.T, cfg Config) {
				testutil.RequireEqual(t, "example.com", cfg.Server.Host)
				testutil.RequireEqual(t, 0, cfg.Server.Port)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cfg Config
			err := mapstructure.New().Decode(tc.input, &cfg)
			if tc.wantErr {
				testutil.RequireError(t, err)
			} else {
				testutil.RequireNoError(t, err)
				if tc.check != nil {
					tc.check(t, cfg)
				}
			}
		})
	}
}
