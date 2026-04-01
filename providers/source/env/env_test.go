package env_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	envSource "github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

type taggedEnvConfig struct {
	Server struct {
		Host string `mapstructure:"host" env:"APP_HOST,APP_SERVER_HOST"`
		Port int    `mapstructure:"port" env:"APP_PORT"`
	} `mapstructure:"server"`
}

func TestEnvSource_Prefix(t *testing.T) {
	// No t.Parallel() — uses t.Setenv.
	testutil.MustSetEnv(t, "MYAPP_HOST", "example.com")
	testutil.MustSetEnv(t, "MYAPP_PORT", "9090")
	testutil.MustSetEnv(t, "OTHER_KEY", "ignored")

	src := envSource.New("MYAPP")
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	testutil.RequireEqual(t, "env", doc.Name)

	if doc.Tree["host"] != "example.com" {
		t.Fatalf("expected host=example.com, got %v", doc.Tree["host"])
	}
	if doc.Tree["port"] != "9090" {
		t.Fatalf("expected port=9090, got %v", doc.Tree["port"])
	}
	if _, ok := doc.Tree["other_key"]; ok {
		t.Fatal("OTHER_KEY should have been filtered by prefix")
	}
}

func TestEnvSource_NestedPath(t *testing.T) {
	// No t.Parallel() — uses t.Setenv.
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "nested.com")

	src := envSource.New("APP")
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)

	server, ok := doc.Tree["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected server to be map[string]any, tree=%#v", doc.Tree)
	}
	if server["host"] != "nested.com" {
		t.Fatalf("expected server.host=nested.com, got %v", server["host"])
	}
}

func TestEnvSource_NoPrefix(t *testing.T) {
	// No t.Parallel() — uses t.Setenv.
	testutil.MustSetEnv(t, "NOPFX_UNIQUE_KEY_XYZ", "present")

	src := envSource.New("")
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	if doc.Tree["nopfx_unique_key_xyz"] != "present" {
		t.Fatalf("expected nopfx_unique_key_xyz=present in tree, got %#v", doc.Tree)
	}
}

func TestEnvSource_ExplicitBindingsTakePrecedence(t *testing.T) {
	// No t.Parallel() — uses t.Setenv.
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "inferred")
	testutil.MustSetEnv(t, "APP_HOST", "explicit")

	src := envSource.NewWithOptions(envSource.Options{
		Prefix:     "APP",
		Infer:      true,
		Precedence: envSource.ExplicitFirst,
		Bindings: map[string][]string{
			"server.host": {"HOST"},
		},
	})
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	server := doc.Tree["server"].(map[string]any)
	if server["host"] != "explicit" {
		t.Fatalf("expected explicit value, got %v", server["host"])
	}
}

func TestEnvSource_TagBindingsFromStruct(t *testing.T) {
	// No t.Parallel() — uses t.Setenv.
	testutil.MustSetEnv(t, "APP_HOST", "from-tag")
	testutil.MustSetEnv(t, "APP_PORT", "9191")
	testutil.MustSetEnv(t, "APP_SERVER__HOST", "from-inferred")

	src := envSource.NewWithOptions(envSource.Options{
		Prefix:             "APP",
		Infer:              true,
		Precedence:         envSource.ExplicitFirst,
		UseStructTagEnvFor: taggedEnvConfig{},
	})
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	server, ok := doc.Tree["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected server map, got %#v", doc.Tree)
	}
	if server["host"] != "from-tag" {
		t.Fatalf("expected tag-bound host value, got %v", server["host"])
	}
	if server["port"] != "9191" {
		t.Fatalf("expected tag-bound port value, got %v", server["port"])
	}
}
