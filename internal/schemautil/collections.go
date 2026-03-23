package schemautil

import "reflect"

// IsStringMap reports whether t (after pointer unwrapping) is a map with string keys.
func IsStringMap(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Map && t.Key().Kind() == reflect.String
}

// IsSliceOrArray reports whether t (after pointer unwrapping) is a slice or array.
func IsSliceOrArray(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}
