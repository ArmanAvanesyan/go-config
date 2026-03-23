# Extensions

Extensions add optional capabilities while keeping core pipeline dependencies small.

## Extension areas

- `extensions/schema/*`: schema generation and inference helpers.
- `extensions/wasm/parser/*`: embedded WASM parser adapters.
- `extensions/wasm/runtime/wazero`: shared WASM runtime engine.
- `extensions/wasm/validator/*`: WASM-based validation engine and policies.

## Schema extensions

- `extensions/schema/generate`: produce JSON Schema from Go types.
- `extensions/schema/infer`: infer best-effort schema from merged config trees.

Use generation for typed API contracts and inference for introspecting dynamic trees.

Inference behavior caveats:
- Array inference currently uses first-element heuristics for heterogeneous arrays.
- Unsupported or nil-like dynamic values produce unconstrained schema nodes.

## WASM parser extensions

Parser adapters include:

- `extensions/wasm/parser/rustyaml`
- `extensions/wasm/parser/rusttoml`
- `extensions/wasm/parser/rustjson`

Provider packages call these adapters and expose parser interfaces at `providers/parser/*`.

## WASM runtime engine

`extensions/wasm/runtime/wazero` owns module lifecycle:

- compile/instantiate module
- write input bytes to WASM memory
- call ABI exports (`wasm_alloc`, `wasm_dealloc`, `parse`, `output_ptr`, `output_len`)
- decode transport output

Current parser transport expects `GCFGMP1` prefix followed by Msgpack payload.

## WASM validation extension

- `extensions/wasm/validator/engine` provides validation runtime execution.
- `extensions/wasm/validator/rustpolicy` provides default policy-based validator wrappers.
- Custom WASM policy modules are supported via bytes/module path constructors.

Validation ABI details are documented in `docs/wasm-validation-abi.md`.

## Artifact and lifecycle notes

- `.wasm` parser/policy artifacts are embedded and versioned in extension package directories.
- Parser and validator engines should be closed when no longer needed.
- YAML shared parser flow uses reference-counted reuse for hot-path performance.
