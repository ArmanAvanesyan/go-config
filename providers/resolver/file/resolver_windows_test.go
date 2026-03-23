//go:build windows

package file_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fileresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/file"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestFileResolver_WindowsSafe(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		input       map[string]any
		want        map[string]any
		wantErr     bool
		errContains string
	}{
		{
			name:  "non-pattern string unchanged",
			input: map[string]any{"key": "plain"},
			want:  map[string]any{"key": "plain"},
		},
		{
			name:        "empty file path returns error",
			input:       map[string]any{"x": "${FILE:   }"},
			wantErr:     true,
			errContains: "empty path",
		},
		{
			name:  "unterminated placeholder stays unchanged",
			input: map[string]any{"x": "prefix ${FILE:C:/no/close"},
			want:  map[string]any{"x": "prefix ${FILE:C:/no/close"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := fileresolver.New().Resolve(context.Background(), tc.input)
			if tc.wantErr {
				testutil.RequireError(t, err)
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("expected error to contain %q, got %v", tc.errContains, err)
				}
				return
			}

			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}

func TestFileResolver_Windows_NestedAndSliceExpansion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	secretPath := filepath.Join(dir, "secret.txt")
	testutil.RequireNoError(t, os.WriteFile(secretPath, []byte("token\n"), 0600))

	input := map[string]any{
		"plain": "keep",
		"mix":   "A=${FILE:" + secretPath + "};B=${FILE:" + secretPath + "}",
		"nest": map[string]any{
			"arr": []any{
				"${FILE:" + secretPath + "}",
				map[string]any{"deep": "${FILE:" + secretPath + "}"},
				42,
			},
		},
	}

	got, err := fileresolver.New().Resolve(context.Background(), input)
	testutil.RequireNoError(t, err)
	testutil.RequireEqual(t, "keep", got["plain"])
	testutil.RequireEqual(t, "A=token;B=token", got["mix"])

	nest, ok := got["nest"].(map[string]any)
	if !ok {
		t.Fatalf("expected nest map, got %T", got["nest"])
	}
	arr, ok := nest["arr"].([]any)
	if !ok {
		t.Fatalf("expected arr slice, got %T", nest["arr"])
	}
	testutil.RequireEqual(t, "token", arr[0])
	deep, ok := arr[1].(map[string]any)
	if !ok {
		t.Fatalf("expected deep map, got %T", arr[1])
	}
	testutil.RequireEqual(t, "token", deep["deep"])
	testutil.RequireEqual(t, 42, arr[2])
}
