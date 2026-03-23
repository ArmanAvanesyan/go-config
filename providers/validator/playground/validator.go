package playground

import "context"

// Func is a validation function that can be passed to New.
type Func func(ctx context.Context, v any) error

// Validator runs a custom validation function against the loaded config.
// It implements config.Validator.
//
// For struct-tag-based validation via go-playground/validator, wrap the
// playground validator in a Func and pass it here, or add a dedicated
// adapter in a future providers/validator/playground/adapter.go.
type Validator struct {
	fn Func
}

// New returns a Validator that runs the given function on Validate.
func New(fn Func) *Validator {
	return &Validator{fn: fn}
}

// Validate calls the configured validation function with the value.
func (v *Validator) Validate(ctx context.Context, value any) error {
	if v.fn == nil {
		return nil
	}
	return v.fn(ctx, value)
}
