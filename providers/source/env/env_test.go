package env_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	envSource "github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

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
