package config

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContract_DecodeCoercion_PermissiveAllowsDurationAndNumbers(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `json:"timeout"`
		Enabled bool          `json:"enabled"`
		Count   int           `json:"count"`
		Ratio   float64       `json:"ratio"`
	}
	var out cfg
	err := New().
		AddSource(&testMemSource{tree: map[string]any{
			"timeout": "250ms",
			"enabled": "true",
			"count":   "7",
			"ratio":   "1.5",
		}}).
		Load(context.Background(), &out)
	if err != nil {
		t.Fatalf("contract load failed: %v", err)
	}
	if out.Timeout != 250*time.Millisecond || !out.Enabled || out.Count != 7 || out.Ratio != 1.5 {
		t.Fatalf("unexpected permissive decode result: %+v", out)
	}
}

func TestContract_DecodeCoercion_StrictRejectsWeakConversions(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `json:"timeout"`
		Enabled bool          `json:"enabled"`
		Count   int           `json:"count"`
	}
	var out cfg
	err := New(WithStrict(true)).
		AddSource(&testMemSource{tree: map[string]any{
			"timeout": "250ms",
			"enabled": "true",
			"count":   "7",
		}}).
		Load(context.Background(), &out)
	if err == nil {
		t.Fatal("expected strict decode failure")
	}
	if !strings.Contains(err.Error(), "decode failed") {
		t.Fatalf("expected decode failed error, got %v", err)
	}
}
