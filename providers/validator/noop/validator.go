package noop

import "context"

// Validator is a no-op validator that always succeeds.
type Validator struct{}

// New returns a new no-op Validator.
func New() *Validator {
	return &Validator{}
}

// Validate returns nil without checking the value.
func (v *Validator) Validate(_ context.Context, _ any) error {
	return nil
}
