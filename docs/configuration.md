# Configuration guide

This guide summarizes how to configure `go-config` pipelines and where each option applies.

## Core builder pattern

Most setups start with `config.New()` and chain options and sources:

```go
loader := config.New(
    // options...
).AddSource(sourceA, parserA).
  AddSource(sourceB, parserB)
```

`Load(ctx, &target)` executes the standard staged pipeline. `LoadTyped[T](ctx, loader)` returns a typed value using the same loader configuration.

## Spec-based app API

For app-facing wiring, `config.Spec` can be used to declare sources and policy in one contract:

```go
trace := &config.Trace{}
err := config.LoadWithSpec(ctx, &out, config.Spec{
    Trees: []config.TreeSpec{
        {Tree: map[string]any{"name": "default"}, Priority: 0},
    },
    Strict:     false,
    Trace:      trace,
    DefaultsFn: defaultsFn,
    ValidateFn: validateFn,
})
```

Typed helper:

```go
cfg, err := config.LoadTypedWithSpec[MyConfig](ctx, spec)
```

## Common option areas

The exact option surface evolves, but most configuration falls into:

- merge strategy selection (deep/replace style)
- resolver behavior (enabled/disabled, custom resolver)
- decoder behavior (strictness and mapping preferences)
- lifecycle hooks (`ApplyDefaults`, defaults callback, validate callback, validator)
- direct decode fast path toggles
- runtime watch/reload integration options

For stage behavior details, see [Pipeline Reference](./architecture.md#10-pipeline-reference).

## Source and parser composition

- `Source` provides either raw `Document` or `TreeDocument`.
- Parsers are attached per source where needed.
- `TreeDocument` sources bypass parse.
- Source priority and registration order define merge precedence.
- Per-source policies can control missing-input and parse-error behavior.

### Environment binding options

The env source supports both inferred mapping and explicit bindings:

- explicit key aliases (`key -> [ENV1, ENV2...]`)
- precedence between explicit and inferred env names
- optional struct-tag extraction from `env:"A,B"` tags
- inferred key normalization controls (dot/hyphen to underscore, uppercase inference)

## Explain/trace mode

Trace mode can be enabled with `WithTrace(...)` (loader) or `Spec.Trace` (spec API).
It records:

- final source/value per flattened key
- overridden candidates per key
- lifecycle hook execution order

See [Pipeline Reference](./architecture.md#10-pipeline-reference) for ordering rules.

## Compatibility parity flow (Spec + Trace)

For Phase 4 compatibility checks, use a fixture-driven `Spec` declaration and compare
an expected snapshot against the loaded typed output.

Recommended pattern:

1. Load fixture input (`trees`, `env`, `expected`, `expected_trace`).
2. Build `config.Spec` from fixture trees and env source options.
3. Enable `Spec.Trace` and run `LoadTypedWithSpec`.
4. Canonicalize output to a deterministic snapshot and compare.
5. Assert `Trace.Keys[key].FinalSource` for compatibility-sensitive keys.

Runnable reference:

- `examples/compat_parity/main.go` (user-facing example)
- `config/contract_compatibility_test.go` (contract gate implementation)

## Runtime reload configuration

Reload behavior combines:

- a configured loader
- a `ReloadTrigger` (fsnotify/polling style backends)
- a callback receiving old/new snapshots and diff metadata

See [Runtime and Reload Reference](./architecture.md#11-runtime-and-reload-reference) for lifecycle details.

## WASM-related configuration notes

If you are working with Rust-backed parser/validator artifacts:

- use `make wasm-build-docker` to regenerate
- use `make wasm-verify-docker` to validate against CI parity

This avoids host toolchain drift in `.wasm` outputs.

## Recommended references

- [architecture.md](./architecture.md)
- [architecture.md#10-pipeline-reference](./architecture.md#10-pipeline-reference)
- [architecture.md#11-runtime-and-reload-reference](./architecture.md#11-runtime-and-reload-reference)
- [testing.md](./testing.md)
