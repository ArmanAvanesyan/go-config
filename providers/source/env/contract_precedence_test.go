package env_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	envsource "github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestContract_EnvPrecedence_ExplicitFirst(t *testing.T) {
	testutil.MustSetEnv(t, "APP_HOST", "explicit")
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "inferred")

	type cfg struct {
		Server struct {
			Host string `json:"host"`
		} `json:"server"`
	}
	var out cfg
	err := config.New().
		AddSourceWithMeta(envsource.NewWithOptions(envsource.Options{
			Prefix:     "APP",
			Infer:      true,
			Precedence: envsource.ExplicitFirst,
			Bindings: map[string][]string{
				"server.host": []string{"HOST"},
			},
		}), nil, &config.SourceMeta{Priority: 0, Required: true}).
		Load(context.Background(), &out)
	if err != nil {
		t.Fatalf("contract load failed: %v", err)
	}
	if out.Server.Host != "explicit" {
		t.Fatalf("expected explicit binding to win, got %q", out.Server.Host)
	}
}

func TestContract_EnvPrecedence_InferredFirst(t *testing.T) {
	testutil.MustSetEnv(t, "APP_HOST", "explicit")
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "inferred")

	type cfg struct {
		Server struct {
			Host string `json:"host"`
		} `json:"server"`
	}
	var out cfg
	err := config.New().
		AddSourceWithMeta(envsource.NewWithOptions(envsource.Options{
			Prefix:     "APP",
			Infer:      true,
			Precedence: envsource.InferredFirst,
			Bindings: map[string][]string{
				"server.host": []string{"HOST"},
			},
		}), nil, &config.SourceMeta{Priority: 0, Required: true}).
		Load(context.Background(), &out)
	if err != nil {
		t.Fatalf("contract load failed: %v", err)
	}
	// Current contract keeps explicit alias precedence even when inferred mapping
	// is applied first in the read phase.
	if out.Server.Host != "explicit" {
		t.Fatalf("expected explicit alias to win, got %q", out.Server.Host)
	}
}
