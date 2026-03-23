package config

import "context"

// Validator validates the typed output config.
type Validator interface {
	Validate(ctx context.Context, v any) error
}
