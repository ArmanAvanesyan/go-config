# Rust / WASM parsers and policy

YAML and TOML parsing uses Rust crates compiled to WebAssembly, executed in-process via [wazero](https://github.com/tetratelabs/wazero). This avoids common pure-Go YAML/TOML dependencies for those formats.

Downstream users consume **embedded** `.wasm` artifacts checked into the repo; Rust is only required when changing parser or policy Rust sources.

For transport contracts and extension layout, see [Extensions reference](./architecture.md#12-extensions-reference).

## Rust crates

| Parser | Crate | Location |
| ------ | ----- | -------- |
| TOML | `toml = "0.8"` | `rust/parsers/toml-parser/` |
| YAML | `serde_yaml = "0.9"` | `rust/parsers/yaml-parser/` |
| JSON | `serde_json = "1"` | `rust/parsers/json-parser/` |

## Building WASM binaries

Compiled `.wasm` files live under `extensions/wasm/parser/rust*/` and are embedded via `go:embed`. Rebuild when Rust sources change.

**Prerequisites**

```bash
# https://rustup.rs
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup target add wasm32-wasip1
```

**Build**

```bash
cd rust && make all
```

From repo root:

```bash
make wasm-build
make wasm-verify
```

`wasm-verify` rebuilds and fails if committed `.wasm` files drift from sources.

For Docker-based builds aligned with CI, see [configuration.md](./configuration.md#wasm-related-configuration-notes).

## Parser WASM ABI

Parser crates share exports so one Go engine handles all formats:

| Export | Signature | Purpose |
| ------ | --------- | ------- |
| `wasm_alloc` | `(size: u32) → *mut u8` | Allocate input buffer in WASM memory |
| `wasm_dealloc` | `(ptr: *mut u8, size: u32)` | Free input buffer |
| `parse` | `(ptr: *const u8, len: u32) → i32` | Parse bytes; `0` = success |
| `output_ptr` | `() → *const u8` | Pointer to parser output bytes |
| `output_len` | `() → u32` | Length of parser output |

The host writes input into guest memory, calls `parse`, then reads transport output. YAML uses ABI v2: `GCFGMP1` prefix plus Msgpack payload (see [WASM runtime engine](./architecture.md#124-wasm-runtime-engine)).

## Optional Rust JSON parser

`providers/parser/json` uses the stdlib by default. For large JSON configs you can use the Rust-backed adapter:

```go
import rustjson "github.com/ArmanAvanesyan/go-config/extensions/wasm/parser/rustjson"

jp, err := rustjson.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer jp.Close(ctx)

loader.AddSource(file.New("config.json"), jp)
```

## WASM policy validation

Policy modules run in WASM via `extensions/wasm/validator/rustpolicy`. The default embedded policy (`rust/validators/config-policy`) is minimal; supply your own WASM for stricter rules.

```go
import "github.com/ArmanAvanesyan/go-config/extensions/wasm/validator/rustpolicy"

validator, err := rustpolicy.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer validator.Close(ctx)

config.New(
    config.WithValidator(validator),
).AddSource(/* ... */).Load(ctx, &cfg)
```

Custom modules must export the ABI in [Validation WASM ABI](./architecture.md#validation-wasm-abi). Load arbitrary bytes with `rustpolicy.NewFromBytes(ctx, wasmBytes)`.

**Rebuild default policy WASM**

```bash
cd rust && make build-policy
```

This copies `policy.wasm` into `extensions/wasm/validator/rustpolicy/`.
