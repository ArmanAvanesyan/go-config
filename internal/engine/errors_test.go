package engine

import "testing"

func TestErrorFactories(t *testing.T) {
	t.Parallel()
	if errBindingsMustBeSlice() == nil ||
		errBindingMissingSourceField() == nil ||
		errBindingSourceType() == nil ||
		errBindingParserType() == nil ||
		errBindingElementType() == nil {
		t.Fatal("expected non-nil error constructors")
	}
}
