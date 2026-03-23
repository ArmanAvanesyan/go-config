package generate

import (
	"reflect"
	"strconv"
	"strings"

	schemautil "github.com/ArmanAvanesyan/go-config/internal/schemautil"
)

// typeSchema builds a Schema node for the provided Go type.
func typeSchema(t reflect.Type, defs map[reflect.Type]string, useRefs bool) *Schema {
	if t == nil {
		return &Schema{}
	}

	// Follow pointers.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		// Special case time.Time → string with date-time format.
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return &Schema{
				Type:   "string",
				Format: "date-time",
			}
		}
		return structSchema(t, defs, useRefs)
	case reflect.Slice, reflect.Array:
		itemSchema := typeSchema(t.Elem(), defs, useRefs)
		return &Schema{
			Type:  "array",
			Items: itemSchema,
		}
	case reflect.Map:
		// Only map[string]T is supported; other key types fall back to generic object.
		if schemautil.IsStringMap(t) {
			valueSchema := typeSchema(t.Elem(), defs, useRefs)
			return &Schema{
				Type:                 "object",
				AdditionalProperties: valueSchema,
			}
		}
		return &Schema{Type: "object"}
	default:
		if scalar := schemautil.ClassifyScalar(t); scalar != "" {
			return &Schema{Type: scalar}
		}
		// interface{} / any and other unsupported kinds are left unconstrained.
		return &Schema{}
	}
}

// structSchema builds a Schema for a struct type, optionally reusing
// a definition via $defs and $ref when useRefs is true.
func structSchema(t reflect.Type, defs map[reflect.Type]string, useRefs bool) *Schema {
	if !useRefs {
		return inlineStructSchema(t, defs, useRefs)
	}

	// If we have already assigned a definition name, just reference it.
	if name, ok := defs[t]; ok {
		return &Schema{Ref: "#/$defs/" + name}
	}

	// Assign a definition name with package-aware collision avoidance.
	name := nextDefinitionName(t, defs)
	defs[t] = name

	// Build the inline schema for the struct; the caller is responsible
	// for wiring it into Schema.Defs.
	return inlineStructSchema(t, defs, useRefs)
}

// inlineStructSchema returns an object Schema describing the struct
// without introducing a new definition.
func inlineStructSchema(t reflect.Type, defs map[reflect.Type]string, useRefs bool) *Schema {
	props := make(map[string]*Schema)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip unexported fields.
		if f.PkgPath != "" {
			continue
		}

		tag := parseJSONTag(f.Tag.Get("json"))
		if tag.Skip {
			continue
		}

		// Embedded struct without a name: flatten its fields.
		if f.Anonymous && tag.Name == "" && f.Type.Kind() == reflect.Struct && !isTimeType(f.Type) {
			embedded := inlineStructSchema(f.Type, defs, useRefs)
			for k, v := range embedded.Properties {
				props[k] = v
			}
			required = append(required, embedded.Required...)
			continue
		}

		name := tag.Name
		if name == "" {
			name = f.Name
		}

		propSchema := typeSchema(f.Type, defs, useRefs)
		props[name] = propSchema

		if !tag.OmitEmpty {
			required = append(required, name)
		}
	}

	return &Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}
}

func isTimeType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() == "time" && t.Name() == "Time"
}

func nextDefinitionName(t reflect.Type, defs map[reflect.Type]string) string {
	base := t.Name()
	if base == "" {
		base = "AnonStruct"
	}
	pkg := t.PkgPath()
	if pkg != "" {
		pkg = sanitizePkgPath(pkg)
		base = pkg + "_" + base
	}
	candidate := base
	suffix := 2
	for nameInUse(candidate, defs) {
		candidate = base + "_" + strconvItoa(suffix)
		suffix++
	}
	return candidate
}

func nameInUse(name string, defs map[reflect.Type]string) bool {
	for _, v := range defs {
		if v == name {
			return true
		}
	}
	return false
}

func sanitizePkgPath(pkg string) string {
	repl := strings.NewReplacer("/", "_", ".", "_", "-", "_")
	return repl.Replace(pkg)
}

func strconvItoa(v int) string {
	return strconv.Itoa(v)
}
