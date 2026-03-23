package config

import (
	"context"
	"testing"
)

type noOpResolver struct{}

func (r *noOpResolver) Resolve(_ context.Context, tree map[string]any) (map[string]any, error) {
	return tree, nil
}

func TestResolverInterface(t *testing.T) {
	t.Parallel()
	var _ Resolver = (*noOpResolver)(nil)
}
