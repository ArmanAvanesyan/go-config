package rustpolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/extensions/wasm/validator/rustpolicy"
	"github.com/ArmanAvanesyan/go-config/providers/decoder/mapstructure"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
)

func TestValidator_Validate_success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v, err := rustpolicy.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = v.Close(ctx) }()

	cfg := struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}{Host: "localhost", Port: 8080}
	if err := v.Validate(ctx, &cfg); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestValidator_Validate_marshalError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v, err := rustpolicy.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = v.Close(ctx) }()

	// Value that cannot be JSON-marshaled (e.g. channel)
	ch := make(chan int)
	err = v.Validate(ctx, ch)
	if err == nil {
		t.Fatal("expected error for unmarshallable value")
	}
}

func TestValidator_Close(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v, err := rustpolicy.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(ctx); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestValidator_LoadWithWASMPolicy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	validator, err := rustpolicy.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = validator.Close(ctx) }()

	type appCfg struct {
		Name string `json:"name"`
	}
	var cfg appCfg
	err = config.New(
		config.WithDecoder(mapstructure.New()),
		config.WithValidator(validator),
	).
		AddSource(memory.New(map[string]any{"name": "test"})).
		Load(ctx, &cfg)
	if err != nil {
		t.Fatalf("Load with WASM validator: %v", err)
	}
	if cfg.Name != "test" {
		t.Errorf("cfg.Name = %q, want test", cfg.Name)
	}
}

func TestValidator_LoadValidationFailed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Use a custom WASM that rejects (minimal reject wasm from engine_test)
	rejectWasm := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
		0x01, 0x19, 0x05, 0x60, 0x01, 0x7f, 0x01, 0x7f,
		0x60, 0x02, 0x7f, 0x7f, 0x00, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f,
		0x60, 0x00, 0x01, 0x7f, 0x60, 0x00, 0x01, 0x7f,
		0x03, 0x06, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04,
		0x05, 0x03, 0x01, 0x00, 0x01,
		0x07, 0x40, 0x05,
		0x0a, 'w', 'a', 's', 'm', '_', 'a', 'l', 'l', 'o', 'c', 0x00, 0x00,
		0x0c, 'w', 'a', 's', 'm', '_', 'd', 'e', 'a', 'l', 'l', 'o', 'c', 0x00, 0x01,
		0x08, 'v', 'a', 'l', 'i', 'd', 'a', 't', 'e', 0x00, 0x02,
		0x09, 'e', 'r', 'r', 'o', 'r', '_', 'p', 't', 'r', 0x00, 0x03,
		0x09, 'e', 'r', 'r', 'o', 'r', '_', 'l', 'e', 'n', 0x00, 0x04,
		0x0a, 0x1d, 0x05,
		0x05, 0x00, 0x41, 0x00, 0x0f, 0x0b, 0x03, 0x00, 0x0f, 0x0b,
		0x05, 0x00, 0x41, 0x7f, 0x0f, 0x0b, 0x05, 0x00, 0x41, 0x00, 0x0f, 0x0b,
		0x05, 0x00, 0x41, 0x00, 0x0f, 0x0b,
	}
	v, err := rustpolicy.NewFromBytes(ctx, rejectWasm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = v.Close(ctx) }()

	type appCfg struct {
		X int `json:"x"`
	}
	var cfg appCfg
	err = config.New(
		config.WithDecoder(mapstructure.New()),
		config.WithValidator(v),
	).
		AddSource(memory.New(map[string]any{"x": 1})).
		Load(ctx, &cfg)
	if err == nil {
		t.Fatal("expected validation failed error")
	}
	if !errors.Is(err, config.ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed: %v", err)
	}
}
