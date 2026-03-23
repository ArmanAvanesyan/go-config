package tree_test

import (
	"testing"

	"github.com/ArmanAvanesyan/go-config/internal/tree"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestGet(t *testing.T) {
	tr := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
		"log": map[string]any{
			"level": "info",
		},
	}

	cases := []struct {
		path string
		want any
		ok   bool
	}{
		{"server", tr["server"], true},
		{"server.host", "localhost", true},
		{"server.port", 8080, true},
		{"log.level", "info", true},
		{"missing", nil, false},
		{"server.missing", nil, false},
		{"log.level.extra", nil, false},
		{"", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got, ok := tree.Get(tr, tc.path)
			if ok != tc.ok {
				t.Errorf("Get(%q) ok = %v, want %v", tc.path, ok, tc.ok)
			}
			if tc.ok {
				testutil.RequireEqual(t, tc.want, got)
			}
		})
	}
}

func TestGet_NilTree(t *testing.T) {
	_, ok := tree.Get(nil, "a.b")
	if ok {
		t.Error("Get(nil, ...) should return ok=false")
	}
}
