//go:build integration

package yaml_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatyaml "github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func testdataPath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "testdata", name)
}

func TestYAMLParser_Integration(t *testing.T) {
	ctx := context.Background()

	p, err := formatyaml.New(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer p.Close(ctx)

	raw, err := os.ReadFile(testdataPath("basic.yaml"))
	testutil.RequireNoError(t, err)

	got, err := p.Parse(ctx, &config.Document{
		Name:   "basic.yaml",
		Format: "yaml",
		Raw:    raw,
	})
	testutil.RequireNoError(t, err)

	app, ok := got["app"].(map[string]any)
	if !ok {
		t.Fatalf("expected app map, got %T", got["app"])
	}
	if app["name"] == "" {
		t.Fatal("expected non-empty app.name")
	}
}

func TestYAMLParser_Shared_Integration(t *testing.T) {
	ctx := context.Background()
	raw, err := os.ReadFile(testdataPath("basic.yaml"))
	testutil.RequireNoError(t, err)

	p1, err := formatyaml.NewShared(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer p1.Close(ctx)

	p2, err := formatyaml.NewShared(ctx)
	testutil.RequireNoError(t, err)
	defer p2.Close(ctx)

	_, err = p1.Parse(ctx, &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw})
	testutil.RequireNoError(t, err)
	_, err = p2.Parse(ctx, &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw})
	testutil.RequireNoError(t, err)
}

func TestYAMLParser_Shared_Parallel_Integration(t *testing.T) {
	ctx := context.Background()
	raw, err := os.ReadFile(testdataPath("basic.yaml"))
	testutil.RequireNoError(t, err)

	p, err := formatyaml.NewShared(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer p.Close(ctx)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				_, e := p.Parse(ctx, &config.Document{Name: "basic.yaml", Format: "yaml", Raw: raw})
				if e != nil {
					t.Errorf("parse failed: %v", e)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestYAMLParser_ParseTyped_Integration(t *testing.T) {
	ctx := context.Background()

	p, err := formatyaml.New(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer p.Close(ctx)

	raw, err := os.ReadFile(testdataPath("basic.yaml"))
	testutil.RequireNoError(t, err)

	var out map[string]any
	err = p.ParseTyped(ctx, &config.Document{
		Name:   "basic.yaml",
		Format: "yaml",
		Raw:    raw,
	}, &out)
	testutil.RequireNoError(t, err)
	app, ok := out["app"].(map[string]any)
	if !ok {
		t.Fatalf("expected app map, got %T", out["app"])
	}
	testutil.RequireEqual(t, "demo", app["name"])
}

func TestYAMLAddSharedBytesSource_Integration(t *testing.T) {
	ctx := context.Background()
	raw, err := os.ReadFile(testdataPath("basic.yaml"))
	testutil.RequireNoError(t, err)

	loader := config.New()
	cleanup, err := formatyaml.AddSharedBytesSource(ctx, loader, "basic.yaml", raw)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer cleanup()

	var out struct {
		App struct {
			Name string `json:"name"`
		} `json:"app"`
	}
	// The helper should be usable in a normal loader pipeline.
	err = loader.Load(ctx, &out)
	testutil.RequireNoError(t, err)
	testutil.RequireEqual(t, "demo", out.App.Name)
}
