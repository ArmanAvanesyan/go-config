package noop_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/validator/noop"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestNoopValidator_AlwaysNil(t *testing.T) {
	t.Parallel()

	v := noop.New()

	inputs := []any{nil, "string", 42, map[string]any{"k": "v"}, struct{}{}}
	for _, in := range inputs {
		err := v.Validate(context.Background(), in)
		testutil.RequireNoError(t, err)
	}
}
