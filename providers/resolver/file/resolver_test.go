//go:build !windows

// On Windows, Application Control (or similar) often blocks test executables
// that Go builds in %TEMP%, causing "An Application Control policy has blocked
// this file". Skip this test file on Windows so go test ./... passes; file
// resolver tests still run on Linux and macOS.
package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	fileresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/file"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestFileResolver(t *testing.T) {
	dir := t.TempDir()
	secretPath := filepath.Join(dir, "secret.txt")
	const secretContent = "my-secret-token"
	testutil.RequireNoError(t, os.WriteFile(secretPath, []byte(secretContent+"\n"), 0600))

	relPath := filepath.Join(dir, "rel.txt")
	testutil.RequireNoError(t, os.WriteFile(relPath, []byte("relative"), 0600))
	relBase, _ := filepath.Split(relPath)
	oldWd, err := os.Getwd()
	testutil.RequireNoError(t, err)
	testutil.RequireNoError(t, os.Chdir(relBase))
	defer func() { _ = os.Chdir(oldWd) }()

	cases := []struct {
		name    string
		input   map[string]any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "expands file placeholder with absolute path",
			input: map[string]any{
				"password": "${FILE:" + secretPath + "}",
			},
			want: map[string]any{
				"password": secretContent,
			},
		},
		{
			name: "expands file placeholder with relative path",
			input: map[string]any{
				"data": "${FILE:rel.txt}",
			},
			want: map[string]any{
				"data": "relative",
			},
		},
		{
			name:  "non-pattern string unchanged",
			input: map[string]any{"key": "plain"},
			want:  map[string]any{"key": "plain"},
		},
		{
			name: "nested map",
			input: map[string]any{
				"db": map[string]any{"secret": "${FILE:" + secretPath + "}"},
			},
			want: map[string]any{
				"db": map[string]any{"secret": secretContent},
			},
		},
		{
			name: "missing file returns error",
			input: map[string]any{
				"x": "${FILE:" + filepath.Join(dir, "nonexistent") + "}",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fileresolver.New().Resolve(context.Background(), tc.input)
			if tc.wantErr {
				testutil.RequireError(t, err)
				return
			}
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}
