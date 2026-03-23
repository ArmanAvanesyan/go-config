package generate

import (
	"encoding/json"
	"testing"
)

func TestMarshalSchema(t *testing.T) {
	t.Parallel()
	s := &Schema{Type: "object", Title: "x"}
	b, err := MarshalSchema(s)
	if err != nil {
		t.Fatalf("MarshalSchema error: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["type"] != "object" {
		t.Fatalf("type not marshaled: %#v", out)
	}
}
