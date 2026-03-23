package config

import "testing"

func TestDocumentAndTreeDocumentFields(t *testing.T) {
	t.Parallel()
	d := Document{Name: "n", Format: "json", Raw: []byte("{}"), Metadata: Metadata{"a": "b"}}
	td := TreeDocument{Name: "tn", Tree: map[string]any{"k": "v"}, Metadata: Metadata{"x": "y"}}
	if d.Name == "" || td.Name == "" || d.Metadata["a"] != "b" || td.Metadata["x"] != "y" {
		t.Fatalf("document/tree document fields not set as expected")
	}
}
