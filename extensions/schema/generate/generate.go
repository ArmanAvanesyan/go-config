package generate

import (
	"reflect"
)

// Generate builds a JSON Schema document for the provided Go type and
// returns its JSON encoding.
//
// typ should be the non-pointer struct type that configuration is
// decoded into (e.g. reflect.TypeOf(AppConfig{})).
// For backward compatibility, nil typ returns nil, nil.
func Generate(typ reflect.Type, opts ...Option) ([]byte, error) {
	if typ == nil {
		return nil, nil
	}

	o := ResolveOptions(opts...)
	defNames := make(map[reflect.Type]string)
	rootSchema := typeSchema(typ, defNames, o.UseRefs)

	doc := &Schema{
		Schema:      o.SchemaURL,
		Title:       o.Title,
		Description: o.Description,
		Type:        rootSchema.Type,
		Format:      rootSchema.Format,
		Ref:         rootSchema.Ref,
		Properties:  rootSchema.Properties,
		Required:    rootSchema.Required,
		Items:       rootSchema.Items,
	}

	if o.UseRefs && len(defNames) > 0 {
		defs := make(map[string]*Schema, len(defNames))
		for t, name := range defNames {
			defs[name] = inlineStructSchema(t, defNames, o.UseRefs)
		}
		doc.Defs = defs
	}

	return MarshalSchema(doc)
}

// GenerateFor builds a JSON Schema document for the generic type T and
// returns its JSON encoding.
//
//nolint:revive // keep API name explicit at call sites (generate.GenerateFor[T]).
func GenerateFor[T any](opts ...Option) ([]byte, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	return Generate(typ, opts...)
}
