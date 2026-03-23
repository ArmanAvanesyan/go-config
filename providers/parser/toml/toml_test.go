//go:build integration

package toml_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formattoml "github.com/ArmanAvanesyan/go-config/providers/parser/toml"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func testdataPath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "testdata", name)
}

func TestTOMLParser_Integration(t *testing.T) {
	ctx := context.Background()

	p, err := formattoml.New(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer p.Close(ctx)

	raw, err := os.ReadFile(testdataPath("basic.toml"))
	testutil.RequireNoError(t, err)

	got, err := p.Parse(ctx, &config.Document{
		Name:   "basic.toml",
		Format: "toml",
		Raw:    raw,
	})
	testutil.RequireNoError(t, err)

	if got["app"] == nil && got["server"] == nil {
		t.Fatalf("expected at least one top-level key, got %#v", got)
	}
}
