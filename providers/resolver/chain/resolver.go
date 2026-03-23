package chain

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Resolver runs multiple resolvers in sequence (e.g. env then secrets).
type Resolver struct {
	resolvers []config.Resolver
}

// New creates a resolver that applies the provided resolvers in order.
func New(resolvers ...config.Resolver) *Resolver {
	return &Resolver{resolvers: resolvers}
}

// Resolve applies each resolver in order and returns the final tree.
func (r *Resolver) Resolve(ctx context.Context, tree map[string]any) (map[string]any, error) {
	current := tree
	var err error

	for _, resolver := range r.resolvers {
		current, err = resolver.Resolve(ctx, current)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

var _ config.Resolver = (*Resolver)(nil)
