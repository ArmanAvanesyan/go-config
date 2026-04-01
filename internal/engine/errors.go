package engine

import (
	"fmt"
)

func errBindingsMustBeSlice() error {
	return fmt.Errorf("engine: bindings must be slice")
}

func errBindingsMustBeBindingSlice() error {
	return fmt.Errorf("engine: bindings must be []engine.Binding or []sourceBinding")
}
