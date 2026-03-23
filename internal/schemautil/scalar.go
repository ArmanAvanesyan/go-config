package schemautil

import "reflect"

// ClassifyScalar maps Go scalar kinds to JSON Schema primitive types.
// It returns empty string for non-scalar or unsupported kinds.
func ClassifyScalar(t reflect.Type) string {
	if t == nil {
		return ""
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	default:
		return ""
	}
}
