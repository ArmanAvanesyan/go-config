package engine

import (
	"reflect"

	"github.com/ArmanAvanesyan/go-config/config"
)

// loadUnknownBindings is a reflection-based adapter that allows the engine
// to work with other internal binding slice types that expose "source" and
// "parser" fields implementing the expected interfaces.
func loadUnknownBindings(pc pipelineContext, bindings any) (map[string]any, error) {
	rv := reflect.ValueOf(bindings)
	if rv.Kind() != reflect.Slice {
		return nil, errBindingsMustBeSlice()
	}

	adapted := make([]config.PipelineBinding, 0, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				return nil, errBindingElementType()
			}
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			return nil, errBindingElementType()
		}

		sourceField := item.FieldByName("source")
		if !sourceField.IsValid() {
			sourceField = item.FieldByName("Source")
		}
		if !sourceField.IsValid() {
			return nil, errBindingMissingSourceField()
		}
		srcAny, ok := safeInterface(sourceField)
		if !ok {
			return nil, errBindingSourceType()
		}
		src, ok := srcAny.(config.Source)
		if !ok {
			return nil, errBindingSourceType()
		}

		var parser config.Parser
		parserField := item.FieldByName("parser")
		if !parserField.IsValid() {
			parserField = item.FieldByName("Parser")
		}
		if parserField.IsValid() && !parserField.IsZero() {
			parserAny, ok := safeInterface(parserField)
			if !ok {
				return nil, errBindingParserType()
			}
			if p, ok := parserAny.(config.Parser); ok {
				parser = p
			} else {
				return nil, errBindingParserType()
			}
		}
		adapted = append(adapted, config.PipelineBinding{Source: src, Parser: parser})
	}
	return config.LoadAndMergeBindings(pc.ctx, adapted, pc.strategy)
}

func safeInterface(v reflect.Value) (out any, ok bool) {
	defer func() {
		if recover() != nil {
			out = nil
			ok = false
		}
	}()
	return v.Interface(), true
}
