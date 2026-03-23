// Package rustpolicy provides a config.Validator that runs policy/validation
// rules in a WASM module (e.g. compiled from rust/validators/config-policy).
// See docs/wasm-validation-abi.md for the validation ABI.
package rustpolicy

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/extensions/wasm/validator"
)

//go:embed policy.wasm
var defaultPolicyWasm []byte

// Validator validates config using a WASM-backed policy module.
type Validator struct {
	engine *validator.PolicyEngine
}

// New creates a Validator using the default embedded policy WASM.
// The caller should call Close when the validator is no longer needed.
func New(ctx context.Context) (*Validator, error) {
	engine, err := validator.NewPolicyEngine(ctx, defaultPolicyWasm)
	if err != nil {
		return nil, err
	}
	return &Validator{engine: engine}, nil
}

// NewFromBytes creates a Validator from custom WASM bytes that implement
// the validation ABI. The caller should call Close when done.
func NewFromBytes(ctx context.Context, wasmBytes []byte) (*Validator, error) {
	engine, err := validator.NewPolicyEngine(ctx, wasmBytes)
	if err != nil {
		return nil, err
	}
	return &Validator{engine: engine}, nil
}

// Validate marshals the value to JSON and runs the WASM policy. Returns
// the policy error if validation fails.
func (v *Validator) Validate(ctx context.Context, value any) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return v.engine.Run(ctx, jsonBytes)
}

// Close releases the WASM runtime. Call when the validator is no longer needed.
func (v *Validator) Close(ctx context.Context) error {
	if v.engine != nil {
		return v.engine.Close(ctx)
	}
	return nil
}

var _ config.Validator = (*Validator)(nil)
