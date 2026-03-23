package decode

import (
	"fmt"
	"reflect"
)

// Options controls decoder behavior.
type Options struct {
	WeaklyTypedInput bool
	ErrorUnused      bool
}

// Decode decodes a generic map tree into out.
func Decode(input map[string]any, out any, opts Options) error {
	if out == nil {
		return fmt.Errorf("decode: output cannot be nil")
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("decode: output must be non-nil pointer")
	}
	return decodeValue(rv.Elem(), input, opts, "root")
}
