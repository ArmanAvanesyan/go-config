package env

import (
	"context"
	"os"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/internal/normalize"
)

// Source provides config from environment variables, optionally filtered by prefix.
type Source struct {
	prefix string
}

// New creates an environment source. When prefix is non-empty, only
// variables starting with PREFIX_ are considered, and the prefix is
// stripped before building the config tree.
func New(prefix string) *Source {
	return &Source{prefix: prefix}
}

func (s *Source) Read(_ context.Context) (any, error) {
	tree := map[string]any{}

	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) == 2 {
			val = parts[1]
		}

		if s.prefix != "" {
			prefix := s.prefix + "_"
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			key = strings.TrimPrefix(key, prefix)
		}

		path := strings.Split(strings.ToLower(key), "__")
		for i := range path {
			path[i] = normalize.Key(path[i])
		}
		insert(tree, path, val)
	}

	return &config.TreeDocument{
		Name: "env",
		Tree: tree,
	}, nil
}

func insert(tree map[string]any, path []string, value string) {
	current := tree
	for i, p := range path {
		if i == len(path)-1 {
			current[p] = value
			return
		}
		next, ok := current[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[p] = next
		}
		current = next
	}
}
