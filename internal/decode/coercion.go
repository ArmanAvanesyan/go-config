package decode

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func assignScalar(dst reflect.Value, src any, opts Options, path string) error {
	sv := reflect.ValueOf(src)
	// Keep weak string conversion predictable; do not rely on arbitrary reflect
	// conversion for non-string inputs (e.g. int -> string rune conversion).
	if dst.Kind() != reflect.String {
		if sv.IsValid() && sv.Type().AssignableTo(dst.Type()) {
			dst.Set(sv)
			return nil
		}
		if sv.IsValid() && sv.Type().ConvertibleTo(dst.Type()) && (sv.Kind() != reflect.String || dst.Kind() == reflect.String) {
			dst.Set(sv.Convert(dst.Type()))
			return nil
		}
	} else if s, ok := src.(string); ok {
		dst.SetString(s)
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		if opts.WeaklyTypedInput {
			dst.SetString(fmt.Sprintf("%v", src))
			return nil
		}
	case reflect.Bool:
		b, ok := toBool(src, opts.WeaklyTypedInput)
		if ok {
			dst.SetBool(b)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dst.Type() == reflect.TypeOf(time.Duration(0)) {
			if d, ok := toDuration(src, opts.WeaklyTypedInput); ok {
				dst.SetInt(int64(d))
				return nil
			}
		}
		i, ok := toInt64(src, opts.WeaklyTypedInput)
		if ok {
			if dst.OverflowInt(i) {
				return fmt.Errorf("decode: overflow assigning %v to %s", src, path)
			}
			dst.SetInt(i)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, ok := toUint64(src, opts.WeaklyTypedInput)
		if ok {
			if dst.OverflowUint(u) {
				return fmt.Errorf("decode: overflow assigning %v to %s", src, path)
			}
			dst.SetUint(u)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		f, ok := toFloat64(src, opts.WeaklyTypedInput)
		if ok {
			if dst.OverflowFloat(f) {
				return fmt.Errorf("decode: overflow assigning %v to %s", src, path)
			}
			dst.SetFloat(f)
			return nil
		}
	}

	return fmt.Errorf("decode: cannot assign %T to %s (%s)", src, path, dst.Type())
}

func toDuration(v any, weak bool) (time.Duration, bool) {
	switch x := v.(type) {
	case time.Duration:
		return x, true
	case string:
		if weak {
			d, err := time.ParseDuration(strings.TrimSpace(x))
			return d, err == nil
		}
	}
	return 0, false
}

func toBool(v any, weak bool) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		if weak {
			b, err := strconv.ParseBool(strings.TrimSpace(x))
			return b, err == nil
		}
	}
	return false, false
}

func toInt64(v any, weak bool) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case uint:
		return int64(x), true
	case uint8:
		return int64(x), true
	case uint16:
		return int64(x), true
	case uint32:
		return int64(x), true
	case uint64:
		if x > math.MaxInt64 {
			return 0, false
		}
		return int64(x), true
	case float32:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		if weak {
			i, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
			return i, err == nil
		}
	}
	return 0, false
}

func toUint64(v any, weak bool) (uint64, bool) {
	switch x := v.(type) {
	case uint:
		return uint64(x), true
	case uint8:
		return uint64(x), true
	case uint16:
		return uint64(x), true
	case uint32:
		return uint64(x), true
	case uint64:
		return x, true
	case int:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case int8:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case int16:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case int32:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case int64:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case float32:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case float64:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case string:
		if weak {
			i, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64)
			return i, err == nil
		}
	}
	return 0, false
}

func toFloat64(v any, weak bool) (float64, bool) {
	switch x := v.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case string:
		if weak {
			f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
			return f, err == nil
		}
	}
	return 0, false
}
