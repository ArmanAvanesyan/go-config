// Package validator provides the PolicyEngine for running WASM policy/validation
// modules. See docs/architecture.md#validation-wasm-abi for the validation ABI.
package validator

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// PolicyEngine runs a WASM module that implements the validation ABI.
// Each Run creates a new module instance; Run is safe for concurrent use.
type PolicyEngine struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
}

// NewPolicyEngine compiles the given WASM bytes (validation ABI) and returns
// an engine. The module must export wasm_alloc, wasm_dealloc, validate,
// error_ptr, and error_len. See docs/architecture.md#validation-wasm-abi.
func NewPolicyEngine(ctx context.Context, wasmBytes []byte) (*PolicyEngine, error) {
	rt := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("policy engine: wasi init: %w", err)
	}
	mod, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("policy engine: compile module: %w", err)
	}
	return &PolicyEngine{runtime: rt, module: mod}, nil
}

// Run runs the policy/validation on configJSON. Returns nil if the guest
// returns 0 from validate; otherwise returns an error (including the guest's
// error message when available).
func (e *PolicyEngine) Run(ctx context.Context, configJSON []byte) error {
	mod, err := e.runtime.InstantiateModule(ctx, e.module, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return fmt.Errorf("policy engine: instantiate: %w", err)
	}
	defer func() { _ = mod.Close(ctx) }()

	mem := mod.Memory()
	allocFn := mod.ExportedFunction("wasm_alloc")
	deallocFn := mod.ExportedFunction("wasm_dealloc")
	validateFn := mod.ExportedFunction("validate")
	errorPtrFn := mod.ExportedFunction("error_ptr")
	errorLenFn := mod.ExportedFunction("error_len")

	if allocFn == nil || deallocFn == nil || validateFn == nil || errorPtrFn == nil || errorLenFn == nil {
		return errors.New("policy engine: module missing required exports (wasm_alloc, wasm_dealloc, validate, error_ptr, error_len)")
	}

	allocResults, err := allocFn.Call(ctx, uint64(len(configJSON)))
	if err != nil {
		return fmt.Errorf("policy engine: wasm_alloc: %w", err)
	}
	inputPtr := uint32(allocResults[0])

	if ok := mem.Write(inputPtr, configJSON); !ok {
		_, _ = deallocFn.Call(ctx, uint64(inputPtr), uint64(len(configJSON)))
		return errors.New("policy engine: write to guest memory failed")
	}

	validateResults, err := validateFn.Call(ctx, uint64(inputPtr), uint64(len(configJSON)))
	if err != nil {
		_, _ = deallocFn.Call(ctx, uint64(inputPtr), uint64(len(configJSON)))
		return fmt.Errorf("policy engine: validate call: %w", err)
	}

	if _, err := deallocFn.Call(ctx, uint64(inputPtr), uint64(len(configJSON))); err != nil {
		return fmt.Errorf("policy engine: wasm_dealloc: %w", err)
	}

	status := int32(validateResults[0])
	if status == 0 {
		return nil
	}

	ptrResults, err := errorPtrFn.Call(ctx)
	if err != nil {
		return fmt.Errorf("policy engine: validation failed (error_ptr: %w)", err)
	}
	lenResults, err := errorLenFn.Call(ctx)
	if err != nil {
		return fmt.Errorf("policy engine: validation failed (error_len: %w)", err)
	}
	errPtr := uint32(ptrResults[0])
	errLen := uint32(lenResults[0])
	if errLen == 0 {
		return errors.New("policy engine: validation failed")
	}
	errBytes, ok := mem.Read(errPtr, errLen)
	if !ok {
		return errors.New("policy engine: validation failed (could not read error message)")
	}
	return fmt.Errorf("policy engine: validation failed: %s", string(errBytes))
}

// Close shuts down the runtime and releases resources.
func (e *PolicyEngine) Close(ctx context.Context) error {
	if e.runtime != nil {
		return e.runtime.Close(ctx)
	}
	return nil
}
