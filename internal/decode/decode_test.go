package decode

import "testing"

func TestDecode_TargetValidation(t *testing.T) {
	t.Parallel()
	if err := Decode(map[string]any{}, nil, Options{}); err == nil {
		t.Fatal("expected nil output error")
	}
	var out struct{}
	if err := Decode(map[string]any{}, out, Options{}); err == nil {
		t.Fatal("expected non-pointer error")
	}
	var p *struct{}
	if err := Decode(map[string]any{}, p, Options{}); err == nil {
		t.Fatal("expected nil pointer error")
	}
}
