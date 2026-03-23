package replace_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/merge/replace"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestReplace(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		dst  map[string]any
		src  map[string]any
		want map[string]any
	}{
		{
			name: "src replaces dst entirely",
			dst:  map[string]any{"a": "old", "b": "keep"},
			src:  map[string]any{"c": "new"},
			want: map[string]any{"c": "new"},
		},
		{
			name: "nil src returns empty map",
			dst:  map[string]any{"a": "old"},
			src:  nil,
			want: map[string]any{},
		},
		{
			name: "both nil",
			dst:  nil,
			src:  nil,
			want: map[string]any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := replace.New().Merge(tc.dst, tc.src)
			testutil.RequireNoError(t, err)
			testutil.RequireEqual(t, tc.want, got)
		})
	}
}
