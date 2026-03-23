package generate

import (
	"encoding/json"
	"testing"
)

func TestSchema_JSONTagsShape(t *testing.T) {
	t.Parallel()

	s := &Schema{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		ID:          "id://schema",
		Title:       "title",
		Description: "desc",
		Type:        "object",
		Ref:         "#/$defs/X",
		Properties: map[string]*Schema{
			"x": {Type: "string"},
		},
		Required: []string{"x"},
		AdditionalProperties: &Schema{
			Type: "integer",
		},
		Items: &Schema{Type: "string"},
		Defs: map[string]*Schema{
			"X": {Type: "object"},
		},
	}

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	for _, k := range []string{
		"$schema",
		"$id",
		"title",
		"description",
		"type",
		"$ref",
		"properties",
		"required",
		"additionalProperties",
		"items",
		"$defs",
	} {
		if _, ok := out[k]; !ok {
			t.Fatalf("expected key %q in marshaled schema", k)
		}
	}
}
