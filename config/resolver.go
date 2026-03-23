package config

import "context"

// Resolver resolves placeholders or transforms the merged config tree.
type Resolver interface {
	Resolve(ctx context.Context, tree map[string]any) (map[string]any, error)
}
