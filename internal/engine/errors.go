package engine

import (
	"fmt"
)

func errBindingsMustBeSlice() error {
	return fmt.Errorf("engine: bindings must be slice")
}

func errBindingMissingSourceField() error {
	return fmt.Errorf("engine: binding missing source field")
}

func errBindingSourceType() error {
	return fmt.Errorf("engine: binding source does not implement config.Source")
}

func errBindingParserType() error {
	return fmt.Errorf("engine: binding parser does not implement config.Parser")
}

func errBindingElementType() error {
	return fmt.Errorf("engine: binding element must be struct or pointer to struct")
}
