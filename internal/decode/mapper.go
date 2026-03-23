package decode

import (
	"fmt"
	"reflect"
	"strings"
)

func decodeValue(dst reflect.Value, src any, opts Options, path string) error {
	if !dst.CanSet() {
		return fmt.Errorf("decode: cannot set %s", path)
	}

	if src == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	if dst.Kind() == reflect.Interface {
		dst.Set(reflect.ValueOf(src))
		return nil
	}

	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return decodeValue(dst.Elem(), src, opts, path)
	}

	switch dst.Kind() {
	case reflect.Struct:
		m, ok := src.(map[string]any)
		if !ok {
			return assignScalar(dst, src, opts, path)
		}
		return decodeStruct(dst, m, opts, path)
	case reflect.Map:
		return decodeMap(dst, src, opts, path)
	case reflect.Slice:
		return decodeSlice(dst, src, opts, path)
	case reflect.Array:
		return decodeArray(dst, src, opts, path)
	default:
		return assignScalar(dst, src, opts, path)
	}
}

func decodeStruct(dst reflect.Value, src map[string]any, opts Options, path string) error {
	used := map[string]struct{}{}
	fields := cachedFields(dst.Type())
	ci := make(map[string]string, len(src))
	for k := range src {
		ci[strings.ToLower(k)] = k
	}

	for _, fs := range fields {
		key := fs.name
		v, ok := src[key]
		if !ok {
			orig, okCI := ci[fs.keyLower]
			if okCI {
				v, ok = src[orig], true
				key = orig
			}
		}
		if !ok {
			continue
		}
		used[key] = struct{}{}
		if err := decodeValue(dst.Field(fs.index), v, opts, path+"."+fs.fieldName); err != nil {
			return err
		}
	}

	if opts.ErrorUnused {
		for k := range src {
			if _, ok := used[k]; !ok {
				return fmt.Errorf("decode: unknown field %s.%s", path, k)
			}
		}
	}
	return nil
}

func decodeMap(dst reflect.Value, src any, opts Options, path string) error {
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}
	sm, ok := src.(map[string]any)
	if !ok {
		return fmt.Errorf("decode: expected map at %s, got %T", path, src)
	}
	for k, v := range sm {
		key := reflect.New(dst.Type().Key()).Elem()
		if err := assignScalar(key, k, Options{WeaklyTypedInput: true}, path+".<key>"); err != nil {
			return err
		}
		val := reflect.New(dst.Type().Elem()).Elem()
		if err := decodeValue(val, v, opts, path+"."+k); err != nil {
			return err
		}
		dst.SetMapIndex(key, val)
	}
	return nil
}

func decodeSlice(dst reflect.Value, src any, opts Options, path string) error {
	arr, ok := src.([]any)
	if !ok {
		// try []string compatibility
		if sarr, ok2 := src.([]string); ok2 {
			arr = make([]any, len(sarr))
			for i, s := range sarr {
				arr[i] = s
			}
		} else {
			return fmt.Errorf("decode: expected slice at %s, got %T", path, src)
		}
	}
	out := reflect.MakeSlice(dst.Type(), len(arr), len(arr))
	for i, item := range arr {
		if err := decodeValue(out.Index(i), item, opts, fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return err
		}
	}
	dst.Set(out)
	return nil
}

func decodeArray(dst reflect.Value, src any, opts Options, path string) error {
	arr, ok := src.([]any)
	if !ok {
		return fmt.Errorf("decode: expected array input at %s, got %T", path, src)
	}
	n := dst.Len()
	for i := 0; i < n && i < len(arr); i++ {
		if err := decodeValue(dst.Index(i), arr[i], opts, fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}
