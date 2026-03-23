package infer

import (
	"encoding/json"
	"testing"
)

func TestGenerateFromTree_basic(t *testing.T) {
	tree := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
		"enabled": true,
		"tags":    []any{"a", "b"},
	}

	data, err := GenerateFromTree(tree)
	if err != nil {
		t.Fatalf("GenerateFromTree returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected non-empty schema JSON")
	}
}

func TestGenerateFromTree_reflectionFirstShapes(t *testing.T) {
	type item struct {
		Name string
	}
	tree := map[string]any{
		"typedMap": map[string]int{"a": 1},
		"typedArr": []item{{Name: "n"}},
		"typedPtr": &item{Name: "x"},
	}

	data, err := GenerateFromTree(tree)
	if err != nil {
		t.Fatalf("GenerateFromTree returned error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	props := doc["properties"].(map[string]any)
	typedMap := props["typedMap"].(map[string]any)
	if typedMap["type"] != "object" {
		t.Fatalf("typed map should infer object, got %#v", typedMap)
	}
	typedArr := props["typedArr"].(map[string]any)
	if typedArr["type"] != "array" {
		t.Fatalf("typed slice should infer array, got %#v", typedArr)
	}
	typedPtr := props["typedPtr"].(map[string]any)
	if typedPtr["type"] != "object" {
		t.Fatalf("typed pointer should infer object, got %#v", typedPtr)
	}
}
