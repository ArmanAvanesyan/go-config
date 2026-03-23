package memory

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Source provides config from an in-memory map (tree and document name).
type Source struct {
	tree map[string]any
	name string
}

// New creates an in-memory source with a default name.
func New(tree map[string]any) *Source {
	return &Source{
		tree: tree,
		name: "memory",
	}
}

// Named creates an in-memory source with a custom document name.
func Named(name string, tree map[string]any) *Source {
	return &Source{
		tree: tree,
		name: name,
	}
}

func (s *Source) Read(_ context.Context) (any, error) {
	return &config.TreeDocument{
		Name: s.name,
		Tree: s.tree,
	}, nil
}
