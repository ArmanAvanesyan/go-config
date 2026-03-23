package config

import "testing"

func TestTypesContracts(t *testing.T) {
	t.Parallel()
	m := Metadata{"k": "v"}
	if m["k"] != "v" {
		t.Fatalf("metadata map behavior unexpected")
	}

	sm := SourceMeta{Priority: 7, Required: true}
	if sm.Priority != 7 || !sm.Required {
		t.Fatalf("source meta fields unexpected: %+v", sm)
	}
}
