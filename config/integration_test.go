//go:build integration

package config_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	"github.com/ArmanAvanesyan/go-config/providers/resolver/env"
	envSource "github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func testdataPath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..")
	return filepath.Join(root, "testdata", name)
}

type serverCfg struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
}

func TestIntegration_JSONFileLoad(t *testing.T) {
	var cfg serverCfg
	err := config.New().
		AddSource(file.New(testdataPath("basic.json")), formatjson.New()).
		Load(context.Background(), &cfg)
	testutil.RequireNoError(t, err)

	testutil.RequireEqual(t, "localhost", cfg.Server.Host)
	testutil.RequireEqual(t, 8080, cfg.Server.Port)
}

func TestIntegration_EnvOverride(t *testing.T) {
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "override-host")
	testutil.MustSetEnv(t, "APP_SERVER__PORT", "9999")

	type appCfg struct {
		Server struct {
			Host string `json:"host"`
			Port string `json:"port"` // env source returns strings
		} `json:"server"`
	}

	var cfg appCfg
	err := config.New().
		AddSource(memory.New(map[string]any{
			"server": map[string]any{"host": "default-host", "port": "8080"},
		})).
		AddSource(envSource.New("APP")).
		Load(context.Background(), &cfg)
	testutil.RequireNoError(t, err)

	testutil.RequireEqual(t, "override-host", cfg.Server.Host)
	testutil.RequireEqual(t, "9999", cfg.Server.Port)
}

func TestIntegration_MultiSource_Precedence(t *testing.T) {
	testutil.MustSetEnv(t, "APP2_NAME", "from-env")

	type appCfg struct {
		Name  string `json:"name"`
		Extra string `json:"extra"`
	}

	var cfg appCfg
	err := config.New().
		AddSource(memory.New(map[string]any{"name": "default", "extra": "from-memory"})).
		AddSource(memory.New(map[string]any{"name": "from-second-memory"})).
		AddSource(envSource.New("APP2")).
		Load(context.Background(), &cfg)
	testutil.RequireNoError(t, err)

	// env wins for "name"
	testutil.RequireEqual(t, "from-env", cfg.Name)
	// "extra" comes from first memory source, untouched by later sources
	testutil.RequireEqual(t, "from-memory", cfg.Extra)
}

func TestIntegration_StrictDecode_UnknownField(t *testing.T) {
	var cfg serverCfg
	err := config.New().
		AddSource(file.New(testdataPath("basic.json")), formatjson.New()).
		Load(context.Background(), &cfg)
	// basic.json has "app" which is not in serverCfg — with default decoder
	// (mapstructure) this silently ignores it.
	testutil.RequireNoError(t, err)
}

func TestIntegration_EnvResolver(t *testing.T) {
	testutil.MustSetEnv(t, "REAL_HOST", "resolved.example.com")

	type resolvedCfg struct {
		Host string `json:"host"`
	}

	var cfg resolvedCfg
	err := config.New(config.WithResolver(env.New())).
		AddSource(memory.New(map[string]any{"host": "${ENV:REAL_HOST}"})).
		Load(context.Background(), &cfg)
	testutil.RequireNoError(t, err)
	testutil.RequireEqual(t, "resolved.example.com", cfg.Host)
}

func TestIntegration_WithStrict_EnvTypoFailsUnknownKey(t *testing.T) {
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "override-host")
	testutil.MustSetEnv(t, "APP_SERIVCE__HOST", "typo-creates-unknown-key")

	type appCfg struct {
		Server struct {
			Host string `json:"host"`
		} `json:"server"`
	}

	var cfg appCfg
	err := config.New(config.WithStrict(true)).
		AddSource(memory.New(map[string]any{
			"server": map[string]any{"host": "default-host"},
		})).
		AddSource(envSource.New("APP")).
		Load(context.Background(), &cfg)
	testutil.RequireError(t, err)
	testutil.RequireErrorIs(t, err, config.ErrDecodeFailed)
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field decode error, got: %v", err)
	}
	if errors.Is(err, config.ErrValidationFailed) {
		t.Fatalf("unexpected validation error, got: %v", err)
	}
}
