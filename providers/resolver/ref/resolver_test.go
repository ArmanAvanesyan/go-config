package ref_test

import (
	"context"
	"errors"
	"testing"

	refresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/ref"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestRefResolver(t *testing.T) {
	cases := []struct {
		name    string
		input   map[string]any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "expands ref to nested string",
			input: map[string]any{
				"server": map[string]any{"host": "localhost", "port": 8080},
				"copy":   "${REF:server.host}",
			},
			want: map[string]any{
				"server": map[string]any{"host": "localhost", "port": 8080},
				"copy":   "localhost",
			},
		},
		{
			name: "expands ref to number via fmt.Sprint",
			input: map[string]any{
				"server": map[string]any{"port": 9090},
				"port":   "${REF:server.port}",
			},
			want: map[string]any{
				"server": map[string]any{"port": 9090},
				"port":   "9090",
			},
		},
		{
			name: "nested map recursion",
			input: map[string]any{
				"a": map[string]any{"b": "value"},
				"x": map[string]any{"y": "${REF:a.b}"},
			},
			want: map[string]any{
				"a": map[string]any{"b": "value"},
				"x": map[string]any{"y": "value"},
			},
		},
		{
			name: "slice element substitution",
			input: map[string]any{
				"host": "api.example.com",
				"urls": []any{"https://${REF:host}/v1", "static"},
			},
			want: map[string]any{
				"host": "api.example.com",
				"urls": []any{"https://api.example.com/v1", "static"},
			},
		},
		{
			name:  "non-pattern string unchanged",
			input: map[string]any{"host": "localhost"},
			want:  map[string]any{"host": "localhost"},
		},
		{
			name: "missing path returns error",
			input: map[string]any{
				"copy": "${REF:server.missing}",
			},
			wantErr: true,
		},
		{
			name: "root missing path returns error",
			input: map[string]any{
				"copy": "${REF:nonexistent}",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := refresolver.New().Resolve(context.Background(), tc.input)
			if tc.wantErr {
				testutil.RequireError(t, err)
				if err != nil && !errors.Is(err, refresolver.ErrRefNotFound) {
					t.Errorf("expected ErrRefNotFound, got %v", err)
				}
				return
			}
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}
