package engine

import "testing"

func TestErrorFactories(t *testing.T) {
	t.Parallel()
	if errBindingsMustBeSlice() == nil ||
		errBindingsMustBeBindingSlice() == nil {
		t.Fatal("expected non-nil error constructors")
	}
}
