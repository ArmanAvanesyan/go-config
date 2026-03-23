package wazero

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/vmihailenco/msgpack/v5"
)

// Engine wraps a compiled wazero WASM module.
// Call New or NewFromBytes once to compile the module, then call ParseConfig
// for each parse operation. Close the engine when done.
type Engine struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
	mu      sync.Mutex
	inst    api.Module
	mem     api.Memory
	fns     *moduleFns
	// Reused guest input buffer to avoid alloc/dealloc on each parse call.
	inputPtr uint32
	inputCap uint32
}

// New creates an Engine by reading the WASM binary from a file path.
func New(ctx context.Context, wasmPath string) (*Engine, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("wasm engine: read file: %w", err)
	}
	return NewFromBytes(ctx, wasmBytes)
}

// NewFromBytes creates an Engine from a pre-loaded WASM binary (e.g. via go:embed).
// It also instantiates the WASI snapshot_preview1 host module, which is required
// for binaries compiled with the wasm32-wasip1 target.
func NewFromBytes(ctx context.Context, wasmBytes []byte) (*Engine, error) {
	rt := wazero.NewRuntime(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("wasm engine: wasi init: %w", err)
	}

	mod, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("wasm engine: compile module: %w", err)
	}

	e := &Engine{runtime: rt, module: mod}
	if err := e.instantiateLocked(ctx); err != nil {
		_ = rt.Close(ctx)
		return nil, err
	}
	return e, nil
}

// ParseConfig passes input bytes to the WASM parser module and returns the
// parsed configuration as a map.
//
// The WASM module must export the standard go-config parser ABI:
//
//	wasm_alloc(size u32) -> ptr u32
//	wasm_dealloc(ptr u32, size u32)
//	parse(ptr u32, len u32) -> status i32   // 0 = ok, -1 = error
//	output_ptr() -> ptr u32
//	output_len() -> u32
//	output_meta() -> u64 optional packed metadata (low32=ptr, high32=len)
//
// Output is expected to be transport bytes prefixed by transportPrefix.
func (e *Engine) ParseConfig(ctx context.Context, input []byte) (map[string]any, error) {
	var result map[string]any
	if err := e.parseConfigInto(ctx, input, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ParseConfigTransport parses input and returns transport bytes (with protocol
// prefix) copied out of guest memory. This is useful for benchmark decomposition
// where transport decode should be measured separately from parse call overhead.
func (e *Engine) ParseConfigTransport(ctx context.Context, input []byte) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	outBytes, err := e.parseTransportLocked(ctx, input)
	if err != nil {
		return nil, err
	}
	owned := make([]byte, len(outBytes))
	copy(owned, outBytes)
	return owned, nil
}

// ParseConfigInto parses input and decodes transport payload directly into out.
func (e *Engine) ParseConfigInto(ctx context.Context, input []byte, out any) error {
	return e.parseConfigInto(ctx, input, out)
}

func (e *Engine) parseConfigInto(ctx context.Context, input []byte, out any) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	outBytes, err := e.parseTransportLocked(ctx, input)
	if err != nil {
		return err
	}
	return decodeOutputInto(outBytes, out)
}

func (e *Engine) parseTransportLocked(ctx context.Context, input []byte) ([]byte, error) {
	if e.inst == nil || e.mem == nil || e.fns == nil {
		if err := e.instantiateLocked(ctx); err != nil {
			return nil, err
		}
	}

	if err := e.ensureInputBufferLocked(ctx, uint32(len(input))); err != nil {
		return nil, err
	}
	if ok := e.mem.Write(e.inputPtr, input); !ok {
		return nil, fmt.Errorf("wasm engine: write to guest memory failed")
	}

	parseResults, err := e.fns.parse.Call(ctx, uint64(e.inputPtr), uint64(len(input)))
	if err != nil {
		_ = e.reinstantiateLocked(ctx)
		return nil, fmt.Errorf("wasm engine: parse call: %w", err)
	}

	outPtr, outLen, err := e.outputLocationLocked(ctx)
	if err != nil {
		return nil, err
	}

	outBytesView, ok := e.mem.Read(outPtr, outLen)
	if !ok {
		return nil, fmt.Errorf("wasm engine: read output from guest memory failed")
	}
	if status := int32(parseResults[0]); status != 0 {
		outBytes := append([]byte(nil), outBytesView...)
		return nil, fmt.Errorf("wasm engine: parse error: %s", string(outBytes))
	}
	return outBytesView, nil
}

func (e *Engine) ensureInputBufferLocked(ctx context.Context, need uint32) error {
	if need == 0 {
		need = 1
	}
	if e.inputPtr != 0 && e.inputCap >= need {
		return nil
	}
	if e.inputPtr != 0 {
		_, _ = e.fns.dealloc.Call(context.Background(), uint64(e.inputPtr), uint64(e.inputCap))
		e.inputPtr = 0
		e.inputCap = 0
	}
	allocResults, err := e.fns.alloc.Call(ctx, uint64(need))
	if err != nil {
		_ = e.reinstantiateLocked(ctx)
		return fmt.Errorf("wasm engine: alloc: %w", err)
	}
	e.inputPtr = uint32(allocResults[0])
	e.inputCap = need
	return nil
}

func (e *Engine) outputLocationLocked(ctx context.Context) (uint32, uint32, error) {
	if e.fns.outputMeta != nil {
		metaResults, err := e.fns.outputMeta.Call(ctx)
		if err != nil {
			_ = e.reinstantiateLocked(ctx)
			return 0, 0, fmt.Errorf("wasm engine: output_meta: %w", err)
		}
		meta := metaResults[0]
		outPtr := uint32(meta & 0xffffffff)
		outLen := uint32((meta >> 32) & 0xffffffff)
		return outPtr, outLen, nil
	}

	ptrResults, err := e.fns.outputPtr.Call(ctx)
	if err != nil {
		_ = e.reinstantiateLocked(ctx)
		return 0, 0, fmt.Errorf("wasm engine: output_ptr: %w", err)
	}
	lenResults, err := e.fns.outputLen.Call(ctx)
	if err != nil {
		_ = e.reinstantiateLocked(ctx)
		return 0, 0, fmt.Errorf("wasm engine: output_len: %w", err)
	}
	outPtr := uint32(ptrResults[0])
	outLen := uint32(lenResults[0])
	return outPtr, outLen, nil
}

//nolint:gochecknoglobals // transport headers are protocol constants.
var (
	msgpackTransportPrefix = []byte("GCFGMP1")
	jsonTransportPrefix    = []byte("GCFGJS1")
)

func decodeOutput(outBytes []byte) (map[string]any, error) {
	var result map[string]any
	if err := decodeOutputInto(outBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DecodeTransportInto decodes transport bytes (with protocol prefix) into out.
// This is exported for internal benchmark decomposition and transport tests.
func DecodeTransportInto(outBytes []byte, out any) error {
	return decodeOutputInto(outBytes, out)
}

func decodeOutputInto(outBytes []byte, out any) error {
	payload, decodeMode, err := parseTransportPayload(outBytes)
	if err != nil {
		return err
	}
	if err := decodePayloadInto(payload, decodeMode, out); err != nil {
		return err
	}
	return nil
}

func parseTransportPayload(outBytes []byte) ([]byte, string, error) {
	if len(outBytes) > len(msgpackTransportPrefix) && bytes.Equal(outBytes[:len(msgpackTransportPrefix)], msgpackTransportPrefix) {
		return outBytes[len(msgpackTransportPrefix):], "msgpack", nil
	}
	if len(outBytes) > len(jsonTransportPrefix) && bytes.Equal(outBytes[:len(jsonTransportPrefix)], jsonTransportPrefix) {
		return outBytes[len(jsonTransportPrefix):], "json", nil
	}
	return nil, "", fmt.Errorf("wasm engine: invalid transport prefix")
}

type msgpackDecoderHandle struct {
	reader *bytes.Reader
	dec    *msgpack.Decoder
}

//nolint:gochecknoglobals // decoder pool is process-wide and safe for reuse.
var msgpackDecoderPool = sync.Pool{
	New: func() any {
		reader := bytes.NewReader(nil)
		dec := msgpack.NewDecoder(reader)
		return &msgpackDecoderHandle{reader: reader, dec: dec}
	},
}

func decodePayloadInto(payload []byte, decodeMode string, out any) error {
	switch decodeMode {
	case "msgpack":
		handle, _ := msgpackDecoderPool.Get().(*msgpackDecoderHandle)
		if handle == nil {
			return fmt.Errorf("wasm engine: msgpack decoder unavailable")
		}
		defer msgpackDecoderPool.Put(handle)
		handle.reader.Reset(payload)
		handle.dec.Reset(handle.reader)
		if err := handle.dec.Decode(out); err != nil {
			if err == io.EOF {
				return fmt.Errorf("wasm engine: unmarshal msgpack output: empty payload")
			}
			return fmt.Errorf("wasm engine: unmarshal msgpack output: %w", err)
		}
		return nil
	case "json":
		if err := json.Unmarshal(payload, out); err != nil {
			return fmt.Errorf("wasm engine: unmarshal json output: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("wasm engine: invalid transport prefix")
	}
}

type moduleFns struct {
	alloc      api.Function
	dealloc    api.Function
	parse      api.Function
	outputPtr  api.Function
	outputLen  api.Function
	outputMeta api.Function
}

func exportedFns(mod api.Module) (*moduleFns, error) {
	fns := &moduleFns{
		alloc:      mod.ExportedFunction("wasm_alloc"),
		dealloc:    mod.ExportedFunction("wasm_dealloc"),
		parse:      mod.ExportedFunction("parse"),
		outputPtr:  mod.ExportedFunction("output_ptr"),
		outputLen:  mod.ExportedFunction("output_len"),
		outputMeta: mod.ExportedFunction("output_meta"),
	}
	if fns.alloc == nil || fns.dealloc == nil || fns.parse == nil || fns.outputPtr == nil || fns.outputLen == nil {
		return nil, fmt.Errorf("missing required exported symbols")
	}
	return fns, nil
}

func (e *Engine) instantiateLocked(ctx context.Context) error {
	mod, err := e.runtime.InstantiateModule(ctx, e.module, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return fmt.Errorf("wasm engine: instantiate: %w", err)
	}
	mem := mod.Memory()
	fns, err := exportedFns(mod)
	if err != nil || mem == nil {
		_ = mod.Close(ctx)
		if err != nil {
			return fmt.Errorf("wasm engine: missing required exported symbols")
		}
		return fmt.Errorf("wasm engine: missing memory export")
	}
	e.inst = mod
	e.mem = mem
	e.fns = fns
	return nil
}

func (e *Engine) reinstantiateLocked(ctx context.Context) error {
	if e.inputPtr != 0 && e.fns != nil && e.fns.dealloc != nil {
		_, _ = e.fns.dealloc.Call(context.Background(), uint64(e.inputPtr), uint64(e.inputCap))
	}
	e.inputPtr = 0
	e.inputCap = 0
	if e.inst != nil {
		_ = e.inst.Close(ctx)
	}
	e.inst = nil
	e.mem = nil
	e.fns = nil
	return e.instantiateLocked(ctx)
}

// Close shuts down the underlying wazero runtime, releasing all resources.
func (e *Engine) Close(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.inputPtr != 0 && e.fns != nil && e.fns.dealloc != nil {
		_, _ = e.fns.dealloc.Call(context.Background(), uint64(e.inputPtr), uint64(e.inputCap))
	}
	e.inputPtr = 0
	e.inputCap = 0
	if e.inst != nil {
		_ = e.inst.Close(ctx)
		e.inst = nil
		e.mem = nil
		e.fns = nil
	}
	if e.runtime != nil {
		return e.runtime.Close(ctx)
	}
	return nil
}
