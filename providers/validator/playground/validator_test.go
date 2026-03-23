package playground_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/validator/playground"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestFuncValidator(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("validation failed")

	cases := []struct {
		name    string
		fn      playground.Func
		input   any
		wantErr bool
	}{
		{
			name:    "func that always passes",
			fn:      func(_ context.Context, _ any) error { return nil },
			input:   struct{}{},
			wantErr: false,
		},
		{
			name:    "func that returns error",
			fn:      func(_ context.Context, _ any) error { return sentinel },
			input:   struct{}{},
			wantErr: true,
		},
		{
			name:    "nil func is no-op",
			fn:      nil,
			input:   struct{}{},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := playground.New(tc.fn)
			err := v.Validate(context.Background(), tc.input)
			if tc.wantErr {
				testutil.RequireError(t, err)
			} else {
				testutil.RequireNoError(t, err)
			}
		})
	}
}
