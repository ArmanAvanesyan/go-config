package env_test

import (
	"context"
	"testing"

	envresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/env"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestEnvResolver(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T)
		input map[string]any
		want  map[string]any
	}{
		{
			name: "expands env var",
			setup: func(t *testing.T) {
				testutil.MustSetEnv(t, "TEST_HOST", "example.com")
			},
			input: map[string]any{"host": "${ENV:TEST_HOST}"},
			want:  map[string]any{"host": "example.com"},
		},
		{
			name:  "unknown key returns empty string",
			input: map[string]any{"host": "${ENV:DEFINITELY_NOT_SET_12345}"},
			want:  map[string]any{"host": ""},
		},
		{
			name:  "non-pattern string unchanged",
			input: map[string]any{"host": "localhost"},
			want:  map[string]any{"host": "localhost"},
		},
		{
			name: "nested map recursion",
			setup: func(t *testing.T) {
				testutil.MustSetEnv(t, "NESTED_VAL", "resolved")
			},
			input: map[string]any{
				"server": map[string]any{"host": "${ENV:NESTED_VAL}"},
			},
			want: map[string]any{
				"server": map[string]any{"host": "resolved"},
			},
		},
		{
			name: "slice element substitution",
			setup: func(t *testing.T) {
				testutil.MustSetEnv(t, "SLICE_VAL", "from-env")
			},
			input: map[string]any{"items": []any{"${ENV:SLICE_VAL}", "static"}},
			want:  map[string]any{"items": []any{"from-env", "static"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// No t.Parallel() — t.Setenv requires sequential execution.
			if tc.setup != nil {
				tc.setup(t)
			}
			got, err := envresolver.New().Resolve(context.Background(), tc.input)
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}
