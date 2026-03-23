package chain_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/resolver/chain"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

// prefixResolver prepends prefix to every string value (used to verify ordering).
type prefixResolver struct{ prefix string }

func (r prefixResolver) Resolve(_ context.Context, tree map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(tree))
	for k, v := range tree {
		if s, ok := v.(string); ok {
			out[k] = r.prefix + s
		} else {
			out[k] = v
		}
	}
	return out, nil
}

var _ config.Resolver = prefixResolver{}

func TestChainResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		resolvers []config.Resolver
		input     map[string]any
		want      map[string]any
	}{
		{
			name:      "empty chain returns input unchanged",
			resolvers: nil,
			input:     map[string]any{"key": "value"},
			want:      map[string]any{"key": "value"},
		},
		{
			name:      "single resolver applied",
			resolvers: []config.Resolver{prefixResolver{"A-"}},
			input:     map[string]any{"key": "value"},
			want:      map[string]any{"key": "A-value"},
		},
		{
			name:      "two resolvers applied in order",
			resolvers: []config.Resolver{prefixResolver{"A-"}, prefixResolver{"B-"}},
			input:     map[string]any{"key": "value"},
			// first: "A-value", then: "B-A-value"
			want: map[string]any{"key": "B-A-value"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cr := chain.New(tc.resolvers...)
			got, err := cr.Resolve(context.Background(), tc.input)
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}
