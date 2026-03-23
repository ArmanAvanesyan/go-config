package json_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	formatjson "github.com/ArmanAvanesyan/go-config/providers/parser/json"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func testdataPath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "testdata", name)
}

func TestJSONParser(t *testing.T) {
	t.Parallel()

	basicJSON, err := os.ReadFile(testdataPath("basic.json"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	cases := []struct {
		name    string
		raw     []byte
		check   func(t *testing.T, got map[string]any)
		wantErr bool
	}{
		{
			name: "valid json from testdata",
			raw:  basicJSON,
			check: func(t *testing.T, got map[string]any) {
				app, ok := got["app"].(map[string]any)
				if !ok {
					t.Fatalf("expected app to be a map, got %T", got["app"])
				}
				testutil.RequireEqual(t, "demo", app["name"])

				server, ok := got["server"].(map[string]any)
				if !ok {
					t.Fatalf("expected server to be a map, got %T", got["server"])
				}
				testutil.RequireEqual(t, "localhost", server["host"])
				testutil.RequireEqual(t, float64(8080), server["port"])
			},
		},
		{
			name: "empty object",
			raw:  []byte(`{}`),
			check: func(t *testing.T, got map[string]any) {
				testutil.RequireEqual(t, map[string]any{}, got)
			},
		},
		{
			name:    "invalid json returns error",
			raw:     []byte(`{bad json`),
			wantErr: true,
		},
		{
			name:    "empty bytes returns error",
			raw:     []byte{},
			wantErr: true,
		},
	}

	p := formatjson.New()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := p.Parse(context.Background(), &config.Document{
				Name:   "test",
				Format: "json",
				Raw:    tc.raw,
			})
			if tc.wantErr {
				testutil.RequireError(t, err)
				return
			}
			testutil.RequireNoError(t, err)
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func FuzzParseJSON(f *testing.F) {
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"nested":{"a":1}}`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(``))
	f.Add([]byte(`{bad`))

	p := formatjson.New()

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must never panic regardless of input.
		_, _ = p.Parse(context.Background(), &config.Document{
			Name:   "fuzz",
			Format: "json",
			Raw:    data,
		})
	})
}
