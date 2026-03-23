package infer

import (
	"reflect"

	"github.com/ArmanAvanesyan/go-config/extensions/schema/generate"
	schemautil "github.com/ArmanAvanesyan/go-config/internal/schemautil"
)

// Option configures schema generation.
type Option = generate.Option

// Schema is the JSON Schema model used by generation and inference.
type Schema = generate.Schema

// WithTitle sets the root schema title.
func WithTitle(title string) Option {
	return generate.WithTitle(title)
}

// WithDescription sets the root schema description.
func WithDescription(description string) Option {
	return generate.WithDescription(description)
}

// WithSchemaURL sets the $schema meta field (e.g. draft 2020-12).
func WithSchemaURL(url string) Option {
	return generate.WithSchemaURL(url)
}

// GenerateFromTree builds a JSON Schema document inferred from a generic
// configuration tree (map[string]any) and returns its JSON encoding.
//
// This is a best-effort inference intended for dynamic configs; when a
// Go struct type is available, use extensions/schema/generate as the source of truth.
func GenerateFromTree(tree map[string]any, opts ...Option) ([]byte, error) {
	o := generate.ResolveOptions(opts...)

	root := inferSchemaFromValue(tree)
	doc := &Schema{
		Schema:      o.SchemaURL,
		Title:       o.Title,
		Description: o.Description,
		Type:        root.Type,
		Properties:  root.Properties,
		Items:       root.Items,
	}

	return generate.MarshalSchema(doc)
}

func inferSchemaFromValue(v any) *Schema {
	if v == nil {
		return &Schema{}
	}

	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	for rt.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return &Schema{}
		}
		rv = rv.Elem()
		rt = rv.Type()
	}

	if scalar := schemautil.ClassifyScalar(rt); scalar != "" {
		return &Schema{Type: scalar}
	}

	if rt.Kind() == reflect.Map {
		if rt.Key().Kind() != reflect.String {
			return &Schema{Type: "object"}
		}
		props := make(map[string]*Schema, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			props[iter.Key().String()] = inferSchemaFromValue(iter.Value().Interface())
		}
		return &Schema{
			Type:       "object",
			Properties: props,
		}
	}

	if rt.Kind() == reflect.Struct {
		props := make(map[string]*Schema, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" {
				continue
			}
			props[f.Name] = inferSchemaFromValue(reflect.New(f.Type).Elem().Interface())
		}
		return &Schema{
			Type:       "object",
			Properties: props,
		}
	}

	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		// Array policy: first-element heuristic for heterogeneous input.
		// This keeps inference deterministic while avoiding expensive union synthesis.
		itemSchema := &Schema{}
		if rv.Len() > 0 {
			itemSchema = inferSchemaFromValue(rv.Index(0).Interface())
		}
		return &Schema{
			Type:  "array",
			Items: itemSchema,
		}
	}

	return &Schema{}
}
