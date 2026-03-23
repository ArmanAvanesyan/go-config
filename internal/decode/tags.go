package decode

import (
	"reflect"
	"strings"
	"sync"
)

type fieldSpec struct {
	index     int
	name      string
	fieldName string
	keyLower  string
}

//nolint:gochecknoglobals // decode field metadata cache is shared across decode calls.
var fieldCache sync.Map // map[reflect.Type][]fieldSpec

func cachedFields(t reflect.Type) []fieldSpec {
	if v, ok := fieldCache.Load(t); ok {
		return v.([]fieldSpec)
	}
	out := make([]fieldSpec, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key := fieldKey(f)
		if key == "-" || key == "" {
			continue
		}
		out = append(out, fieldSpec{
			index:     i,
			name:      key,
			fieldName: f.Name,
			keyLower:  strings.ToLower(key),
		})
	}
	fieldCache.Store(t, out)
	return out
}

func fieldKey(f reflect.StructField) string {
	if tag, ok := f.Tag.Lookup("mapstructure"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	if tag, ok := f.Tag.Lookup("json"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return f.Name
}
