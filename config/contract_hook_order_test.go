package config

import (
	"context"
	"reflect"
	"testing"
)

func TestContract_HookOrdering_RegularDecodePath(t *testing.T) {
	steps := []string{}
	loader := New(
		WithDefaultsFunc(func(context.Context, any) error {
			steps = append(steps, "defaults-callback")
			return nil
		}),
		WithValidateFunc(func(context.Context, any) error {
			steps = append(steps, "validate-callback")
			return nil
		}),
		WithValidator(&orderingValidator{order: &steps}),
	).AddSource(&testMemSource{tree: map[string]any{"name": "demo"}})

	out := &orderingConfig{steps: &steps}
	if err := loader.Load(context.Background(), out); err != nil {
		t.Fatalf("contract load failed: %v", err)
	}
	want := []string{"defaults-interface", "defaults-callback", "validate-callback", "validator-interface"}
	if !reflect.DeepEqual(want, steps) {
		t.Fatalf("expected hook order %v, got %v", want, steps)
	}
}

func TestContract_HookOrdering_DirectDecodePath(t *testing.T) {
	steps := []string{}
	loader := New(
		WithDirectDecode(true),
		WithDefaultsFunc(func(context.Context, any) error {
			steps = append(steps, "defaults-callback")
			return nil
		}),
		WithValidateFunc(func(context.Context, any) error {
			steps = append(steps, "validate-callback")
			return nil
		}),
		WithValidator(&orderingValidator{order: &steps}),
	).AddSource(&testDocSource{
		doc: &Document{Name: "test.yaml", Format: "yaml", Raw: []byte("name: demo")},
	}, &testTypedParser{})

	out := &struct {
		Name string `json:"name"`
	}{}
	if err := loader.Load(context.Background(), out); err != nil {
		t.Fatalf("contract direct-decode load failed: %v", err)
	}
	want := []string{"defaults-callback", "validate-callback", "validator-interface"}
	if !reflect.DeepEqual(want, steps) {
		t.Fatalf("expected hook order %v, got %v", want, steps)
	}
}
