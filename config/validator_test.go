package config

import (
	"context"
	"testing"
)

type noOpValidator struct{}

func (v *noOpValidator) Validate(context.Context, any) error { return nil }

func TestValidatorInterface(t *testing.T) {
	t.Parallel()
	var _ Validator = (*noOpValidator)(nil)
}
