# go-config

[![Go Reference](https://pkg.go.dev/badge/github.com/ArmanAvanesyan/go-config.svg)](https://pkg.go.dev/github.com/ArmanAvanesyan/go-config)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
[![CI](https://github.com/ArmanAvanesyan/go-config/actions/workflows/ci.yml/badge.svg)](https://github.com/ArmanAvanesyan/go-config/actions/workflows/ci.yml)

A typed, modular, dependency-light configuration library for Go.

> **Not another Viper wrapper.** No globals, no magic, no mystery precedence rules. Just an explicit pipeline: sources → merge → resolve → decode → validate → your struct.

---

## Why go-config?

| | Viper | Koanf | **go-config** |
|---|---|---|---|
| Typed-first API | ✗ (string getters) | partial | ✓ |
| No globals | ✗ | ✓ | ✓ |
| Dependency-light core | ✗ | partial | ✓ (stdlib only) |
| Strict unknown-key detection | ✗ | ✗ | ✓ |
| Rust/WASM parsers | ✗ | ✗ | ✓ |
| Immutable reload snapshots | ✗ | ✗ | ✓ |
| Pluggable extension model | partial | ✓ | ✓ |

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Pipeline](#pipeline)
  - [Sources](#sources)
  - [Parsers / Formats](#parsers--formats)
  - [Merge Strategies](#merge-strategies)
  - [Resolvers](#resolvers)
  - [Decoders](#decoders)
  - [Validators](#validators)
- [Usage Examples](#usage-examples)
  - [File + env override](#file--env-override)
  - [Multiple sources with defaults](#multiple-sources-with-defaults)
  - [Strict decoding](#strict-decoding)
  - [Placeholder resolution](#placeholder-resolution)
  - [Custom validation](#custom-validation)
  - [Live reload](#live-reload)
- [Package Layout](#package-layout)
- [Architecture Docs](#architecture-docs)
- [Dependency Policy](#dependency-policy)
- [Rust/WASM Parsers](#rustwasm-parsers)
- [Roadmap](#roadmap)
- [Benchmarks](#benchmarks)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

```bash
go get github.com/ArmanAvanesyan/go-config
```

The core packages depend only on the Go standard library. Format parsers (YAML, TOML) use Rust/WASM via [wazero](https://github.com/tetratelabs/wazero) — see [Rust/WASM Parsers](#rustwasm-parsers).

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ArmanAvanesyan/go-config/config"
    "github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
    "github.com/ArmanAvanesyan/go-config/providers/source/env"
    "github.com/ArmanAvanesyan/go-config/providers/source/file"
)

type AppConfig struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`
    Log struct {
        Level string `json:"level"`
    } `json:"log"`
}

func main() {
    ctx := context.Background()

    // Use NewShared in hot paths to amortize parser init costs.
    yp, err := yaml.NewShared(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer yp.Close(ctx)

    var cfg AppConfig

    err = config.New().
        AddSource(file.New("config.yaml"), yp).  // base config
        AddSource(env.New("APP")).                // APP_SERVER_PORT overrides server.port
        Load(ctx, &cfg)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", cfg)
}
```

---

## Core Concepts

### Pipeline

Every call to `Load` runs the same deterministic pipeline:

```
Sources (in order)
    ↓  Read() → Document or TreeDocument
Parsers
    ↓  Parse() → map[string]any
Merge strategy
    ↓  Merge(dst, src) → merged map[string]any   (each source merged left-to-right)
Resolver (optional)
    ↓  Resolve() → map[string]any                (placeholder expansion)
Decoder
    ↓  Decode() → typed struct
Validator (optional)
    ↓  Validate() → error
```

Sources added first have the **lowest** precedence. Sources added last **win**.

### Sources

A `Source` reads raw configuration material and returns either a `*Document` (raw bytes that need a parser) or a `*TreeDocument` (already-structured map).

| Package | What it reads |
|---------|--------------|
| `providers/source/file` | Files on disk; format inferred from extension (`.yaml`, `.json`, `.toml`) |
| `providers/source/env` | Environment variables with a given prefix |
| `providers/source/memory` | An in-memory `map[string]any` — useful for defaults |
| `providers/source/bytes` | A raw `[]byte` slice |
| `providers/source/flag` | Explicitly set `flag.CommandLine` flags |

```go
// File source — format auto-detected from extension
file.New("config.yaml")
file.WithFormat("config.txt", "yaml")  // explicit format hint

// Env source — APP_SERVER_PORT → server.port
env.New("APP")

// Memory source — useful for defaults
memory.New(map[string]any{
    "server": map[string]any{"port": 8080},
})
```

**Source metadata** — Use `AddSourceWithMeta` to set priority and required/optional behaviour:

- **Priority**: Higher value is merged later (wins). Default 0; equal priority preserves registration order.
- **Required**: If `true` (default), a read failure fails `Load`. If `false`, the source is skipped and an empty tree is used.

```go
loader.AddSourceWithMeta(
    file.New("config.yaml"),
    []Parser{yp},
    &config.SourceMeta{Priority: 10, Required: false},
)
```

### Parsers / Formats

A `Parser` converts raw `*Document` bytes into a `map[string]any` tree.

| Package | Format | Backend |
|---------|--------|---------|
| `providers/parser/json` | JSON | stdlib `encoding/json` |
| `providers/parser/yaml` | YAML | Rust `serde_yaml` via WASM |
| `providers/parser/toml` | TOML | Rust `toml-rs` via WASM |

```go
// Parsers are passed as the second argument to AddSource
yp, _ := yaml.New(ctx)
defer yp.Close(ctx)

loader.AddSource(file.New("config.yaml"), yp)
```

For repeated YAML loads, prefer a shared parser:

```go
yp, _ := yaml.NewShared(ctx)
defer yp.Close(ctx) // decrements shared refcount

loader.AddSource(file.New("config.yaml"), yp)
```

For single-source hot paths where parser and target type are known, you can enable
an experimental direct decode fast path:

```go
loader := config.New(config.WithDirectDecode(true))
```

This lets parsers that implement `config.TypedParser` decode directly into the
target struct, bypassing generic map tree materialization.

When loading YAML bytes repeatedly, use the helper:

```go
cleanup, _ := yaml.AddSharedBytesSource(ctx, loader, "inline", rawYAML)
defer cleanup()
```

YAML transport now uses a strict binary ABI v2 contract (no legacy JSON fallback).
After pulling changes that touch YAML parser transport/runtime, you must rebuild
the YAML WASM artifact (`cd rust && make build-yaml`) so parser and runtime stay aligned.

The YAML and TOML parsers require the Rust/WASM binaries to be built first — see [Rust/WASM Parsers](#rustwasm-parsers). The JSON parser uses the standard library with no extra setup.

### Merge Strategies

When multiple sources are loaded, their trees are merged in order. Two strategies are provided:

| Strategy | Behaviour |
|----------|-----------|
| `deep.New()` | Recursively merges maps; later source wins on conflicts. **Default.** |
| `replace.New()` | Later source completely replaces the earlier result at the top level. |

```go
import (
    "github.com/ArmanAvanesyan/go-config/providers/merge/deep"
)

config.New(
    config.WithMergeStrategy(deep.New()),
)
```

Custom strategies implement `merge.Strategy`:

```go
type Strategy interface {
    Merge(dst, src map[string]any) (map[string]any, error)
}
```

### Resolvers

A `Resolver` transforms the merged tree before decoding — used for placeholder expansion.

| Package | Syntax | Example |
|---------|--------|---------|
| `providers/resolver/env` | `${ENV:KEY}` | `${ENV:HOME}` → `/home/user` |
| `providers/resolver/file` | `${FILE:path}` | `${FILE:./secret.txt}` → file contents |
| `providers/resolver/ref` | `${REF:key.path}` | `${REF:server.host}` → value from config tree |
| `providers/resolver/chain` | — | chains multiple resolvers |

```go
import (
    "github.com/ArmanAvanesyan/go-config/providers/resolver/chain"
    envresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/env"
    fileresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/file"
    refresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/ref"
)

config.New(
    config.WithResolver(envresolver.New()),
    // or chain FILE → ENV → REF (file contents, then env vars, then refs into config):
    // config.WithResolver(chain.New(fileresolver.New(), envresolver.New(), refresolver.New())),
)
```

### Decoders

A `Decoder` maps the final merged tree into the typed target struct.

| Package | Behaviour |
|---------|-----------|
| `providers/decoder/mapstructure` | Default. Uses internal tree-to-struct decode with weakly typed input and struct tags (`json:"..."` or `mapstructure:"..."`). |
| `providers/decoder/strict` | Uses internal strict decode and fails when source keys have no matching struct fields. |

```go
import "github.com/ArmanAvanesyan/go-config/providers/decoder/strict"

config.New(
    config.WithDecoder(strict.New()),
)
```

### Validators

A `Validator` runs after decoding. The default is a no-op.

| Package | Behaviour |
|---------|-----------|
| `providers/validator/noop` | Does nothing. Default. |
| `providers/validator/playground` | Accepts any `func(ctx, v any) error` — wrap any validation library. |
| `extensions/wasm/validator/rustpolicy` | Policy and validation rules run in a WASM module (see [WASM policy/validation](#wasm-policyvalidation)). |

```go
import "github.com/ArmanAvanesyan/go-config/providers/validator/playground"

config.New(
    config.WithValidator(playground.New(func(ctx context.Context, v any) error {
        cfg := v.(*AppConfig)
        if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
            return fmt.Errorf("server.port must be 1–65535")
        }
        return nil
    })),
)
```

---

## Usage Examples

### File + env override

```go
// config.yaml sets the base; APP_* env vars override individual keys.
// APP_SERVER_PORT=9090 overrides server.port.
err = config.New().
    AddSource(file.New("config.yaml"), yp).
    AddSource(env.New("APP")).
    Load(ctx, &cfg)
```

### Multiple sources with defaults

```go
// Precedence: defaults < file < env (lowest → highest)
defaults := memory.New(map[string]any{
    "server": map[string]any{"host": "localhost", "port": 8080},
    "log":    map[string]any{"level": "info"},
})

err = config.New().
    AddSource(defaults).
    AddSource(file.New("config.yaml"), yp).
    AddSource(env.New("APP")).
    Load(ctx, &cfg)
```

### Strict decoding

Strict mode returns an error if the config file contains keys that do not correspond to any field in the target struct. Useful for catching typos.

```go
import "github.com/ArmanAvanesyan/go-config/providers/decoder/strict"

err = config.New(
    config.WithDecoder(strict.New()),
).AddSource(file.New("config.yaml"), yp).
    Load(ctx, &cfg)
// Returns error if config.yaml has: unrecognised_key: value
```

### Placeholder resolution

```yaml
# config.yaml
database:
  password: ${ENV:DB_PASSWORD}
  host: ${ENV:DB_HOST}
  # Or load from file / reference other keys:
  # token: ${FILE:./secrets/token.txt}
  # api_url: https://${REF:server.host}/v1
```

```go
import (
    "github.com/ArmanAvanesyan/go-config/providers/resolver/chain"
    envresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/env"
    fileresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/file"
    refresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/ref"
)

// Single resolver:
err = config.New(
    config.WithResolver(envresolver.New()),
).AddSource(file.New("config.yaml"), yp).
    Load(ctx, &cfg)

// Chain FILE → ENV → REF so file contents and env vars resolve first, then ${REF:...} can reference the merged config.
err = config.New(
    config.WithResolver(chain.New(fileresolver.New(), envresolver.New(), refresolver.New())),
).AddSource(file.New("config.yaml"), yp).
    Load(ctx, &cfg)
// cfg.Database.Password is set from $DB_PASSWORD at load time
```

### Typed load

Use `LoadTyped` to get a typed config value without declaring a variable:

```go
loader := config.New().
    AddSource(file.New("config.yaml"), yp).
    AddSource(env.New("APP"))

cfg, err := config.LoadTyped[AppConfig](ctx, loader)
if err != nil {
    log.Fatal(err)
}
// cfg is AppConfig
```

### Custom validation

```go
import "github.com/ArmanAvanesyan/go-config/providers/validator/playground"

validate := playground.New(func(ctx context.Context, v any) error {
    cfg := v.(*AppConfig)
    if cfg.Server.Port == 0 {
        return errors.New("server.port is required")
    }
    return nil
})

err = config.New(
    config.WithValidator(validate),
).AddSource(file.New("config.yaml"), yp).
    Load(ctx, &cfg)
```

### Live reload

Use `WatchTyped` with a `ReloadTrigger` (e.g. `runtime/watch/fsnotify` or `runtime/watch/polling`) to get immutable snapshots on each reload and an optional config diff:

```go
import (
    "github.com/ArmanAvanesyan/go-config/config"
    "github.com/ArmanAvanesyan/go-config/runtime/diff"
    "github.com/ArmanAvanesyan/go-config/providers/source/file"
    "github.com/ArmanAvanesyan/go-config/runtime/watch/fsnotify"
)

loader := config.New().AddSource(file.New("config.yaml"), yp)
watcher := fsnotify.New("config.yaml")

err := config.WatchTyped[AppConfig](ctx, loader, watcher, func(old, new *AppConfig, changes []diff.Change) {
    if old == nil {
        log.Println("initial load:", new)
        return
    }
    log.Println("reload; changes:", changes)
})
// Blocks until ctx is cancelled; then call watcher.Stop() (WatchTyped does this)
```

Each reload decodes into a new snapshot; the callback receives `(old, new)` and `changes` (path-level diff). Use `runtime/watch/polling` for a polling-based trigger with no file watcher dependency.

---

## Package Layout

```
go-config/
├── config/               Core interfaces and Loader
│   ├── config.go         New() entrypoint
│   ├── loader.go         Loader, AddSource(), Load()
│   ├── watch.go          ReloadTrigger, WatchTyped()
│   ├── options.go        Option functions: WithDecoder, WithValidator, ...
│   ├── source.go         Source interface
│   ├── parser.go         Parser interface
│   ├── decoder.go        Decoder interface
│   ├── validator.go      Validator interface
│   ├── resolver.go       Resolver interface
│   ├── document.go       Document, TreeDocument types
│   └── errors.go         Sentinel errors
│
├── providers/source/
│   ├── file/             File-backed source
│   ├── env/              Environment variable source
│   ├── memory/           In-memory source (defaults)
│   ├── bytes/            Raw bytes source
│   └── flag/             flag.CommandLine source
│
├── providers/parser/
│   ├── json/             JSON parser (stdlib)
│   ├── yaml/             YAML parser (Rust serde_yaml via WASM)
│   └── toml/             TOML parser (Rust toml-rs via WASM)
│
├── providers/merge/
│   ├── merge.go          Strategy interface
│   ├── deep/             DeepOverride strategy package
│   └── replace/          Replace strategy package
│
├── providers/decoder/
│   ├── mapstructure/     Default weakly-typed internal decode
│   └── strict/           Strict decoder (fails on unknown keys)
│
├── extensions/schema/
│   ├── generate/         JSON Schema generation from Go struct types
│   └── infer/            Best-effort JSON Schema inference from trees
│
├── providers/validator/
│   ├── noop/             No-op validator (default)
│   └── playground/       Function-based validator
│
├── providers/resolver/
│   ├── env/              ${ENV:KEY} resolver
│   ├── file/             ${FILE:path} resolver
│   ├── ref/              ${REF:key.path} resolver
│   └── chain/            Chain multiple resolvers
│
├── runtime/diff/                 Config tree diff (path-level changes)
├── runtime/watch/
│   ├── fsnotify/         File-system watcher (ReloadTrigger)
│   └── polling/          Polling watcher (ReloadTrigger)
│
├── extensions/wasm/         Rust/WASM parser infrastructure
│   ├── runtime/wazero/   wazero engine wrapper
│   └── parser/
│       ├── rusttoml/     TOML via Rust
│       ├── rustyaml/     YAML via Rust
│       └── rustjson/     JSON via Rust (opt-in)
│
├── testutil/             Test helpers
├── internal/engine/      Internal pipeline engine (loader/pipeline/context/errors)
├── internal/tree/        Tree path helpers (Get/Walk/CloneMap)
├── internal/normalize/   Key and path normalisation helpers
├── internal/decode/      Internal decode mapper/coercion/tag helpers
├── examples/             Runnable usage examples
└── rust/                 Rust WASM parser source crates
    ├── parsers/toml-parser/
    ├── parsers/yaml-parser/
    ├── parsers/json-parser/
    └── Makefile
```

---

## Architecture Docs

For focused implementation references:

- [Architecture](docs/architecture.md)
- [Pipeline](docs/pipeline.md)
- [Runtime](docs/runtime.md)
- [Extensions](docs/extensions.md)

## Dependency Policy

**The core must never import optional capabilities.**

| Package group | External deps |
|---------------|--------------|
| `config/` | none — stdlib only |
| `providers/source/*` | none — stdlib only |
| `providers/decoder/*` | none — stdlib only |
| `providers/merge/` | none — stdlib only |
| `providers/validator/*` | none — stdlib only |
| `providers/resolver/*` | none — stdlib only |
| `extensions/schema/*` | none — stdlib only |
| `providers/parser/json` | none — stdlib only |
| `providers/parser/yaml` | `wazero` (WASM runtime) |
| `providers/parser/toml` | `wazero` (WASM runtime) |
| `extensions/wasm/*` | `wazero` (WASM runtime) |
| `extensions/wasm/validator/*` | `wazero` (WASM runtime) |
| `runtime/watch/fsnotify` | `golang.org/x/sys` (Linux/macOS: OS events; Windows/other: stdlib polling only) |
| `runtime/watch/polling` | none — stdlib only |
| `runtime/diff` | none — stdlib only |

The core `config/` package imports only stdlib and `runtime/diff` (for Watch diff). Third-party dependencies: `github.com/tetratelabs/wazero` (WASM runtime, no CGO) and `golang.org/x/sys` (optional, for `runtime/watch/fsnotify` OS-event backend on Linux and macOS only; Windows and other platforms use stdlib-only polling).

---

## Schema generation

Use `extensions/schema/generate` to generate JSON Schema for typed config structs.

Generate a schema for a Go struct type:

```go
import (
    "fmt"
    schemagen "github.com/ArmanAvanesyan/go-config/extensions/schema/generate"
)

type AppConfig struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port,omitempty"`
    } `json:"server"`
}

func main() {
    b, err := schemagen.GenerateFor[AppConfig](
        schemagen.WithTitle("App config"),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(string(b))
}
```

You can also infer a best-effort schema from a merged config tree with `extensions/schema/infer`:

```go
import "github.com/ArmanAvanesyan/go-config/extensions/schema/infer"

tree := map[string]any{ /* merged config */ }
b, err := infer.GenerateFromTree(tree)
```

The output targets JSON Schema draft 2020-12 by default.

---

## Rust/WASM Parsers

YAML and TOML parsing is delegated to Rust crates compiled to WebAssembly, running in-process via wazero. This eliminates the `gopkg.in/yaml.v3` and `github.com/pelletier/go-toml/v2` Go dependencies.

**Benefits:**
- Rust's serde ecosystem is among the fastest config parsers available
- Memory-safe parsing with full Rust ownership guarantees
- Single WASM runtime (`wazero`) powers all formats

**Rust crates:**

| Parser | Crate | Location |
|--------|-------|----------|
| TOML | `toml = "0.8"` | `rust/parsers/toml-parser/` |
| YAML | `serde_yaml = "0.9"` | `rust/parsers/yaml-parser/` |
| JSON | `serde_json = "1"` | `rust/parsers/json-parser/` |

### Building the WASM binaries

The compiled `.wasm` files are checked into the repository under `extensions/wasm/parser/rust*/` and embedded into the Go binary via `go:embed`. You only need to rebuild them when modifying the Rust source.

Downstream Go consumers do not need Rust installed when using the repository/module with committed WASM artifacts. Rust is required for contributors who change Rust parser/validator code and for CI artifact verification.

**Prerequisites:**

```bash
# Install Rust (https://rustup.rs)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Add the WASM target
rustup target add wasm32-wasip1
```

**Build:**

```bash
cd rust && make all
```

This compiles all three parsers and copies the `.wasm` binaries to their Go package directories.

From repo root, contributor shortcuts:

```bash
make wasm-build
make wasm-verify
```

`wasm-verify` rebuilds artifacts and fails if committed `.wasm` files are stale.

### WASM ABI

All parser crates expose an identical ABI so the same Go engine code handles all three:

| Export | Signature | Purpose |
|--------|-----------|---------|
| `wasm_alloc` | `(size: u32) → *mut u8` | Allocate input buffer in WASM memory |
| `wasm_dealloc` | `(ptr: *mut u8, size: u32)` | Free input buffer |
| `parse` | `(ptr: *const u8, len: u32) → i32` | Parse bytes; 0 = success |
| `output_ptr` | `() → *const u8` | Pointer to parser output bytes |
| `output_len` | `() → u32` | Length of parser output bytes |

Go writes input bytes into WASM linear memory, calls `parse`, then reads transport bytes back. For YAML, the current contract is ABI v2: `GCFGMP1` prefix + Msgpack payload. The engine (`extensions/wasm/runtime/wazero`) handles the full protocol.

### Using the WASM JSON parser (opt-in)

The `providers/parser/json` package uses stdlib by default. For high-throughput scenarios with large JSON configs, the Rust-backed parser is available:

```go
import rustjson "github.com/ArmanAvanesyan/go-config/extensions/wasm/parser/rustjson"

jp, err := rustjson.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer jp.Close(ctx)

loader.AddSource(file.New("config.json"), jp)
```

### WASM policy/validation

Policy and validation rules can run inside a WebAssembly module, so you can enforce config rules in a sandboxed, portable way. The default embedded policy (from `rust/validators/config-policy`) currently allows all valid JSON; you can supply a custom WASM for stricter rules.

**Usage:**

```go
import "github.com/ArmanAvanesyan/go-config/extensions/wasm/validator/rustpolicy"

validator, err := rustpolicy.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer validator.Close(ctx)

config.New(
    config.WithValidator(validator),
).AddSource(...).Load(ctx, &cfg)
```

**Validation WASM ABI:** Custom policy modules must export the ABI described in [docs/wasm-validation-abi.md](docs/wasm-validation-abi.md) (`wasm_alloc`, `wasm_dealloc`, `validate`, `error_ptr`, `error_len`). Use `rustpolicy.NewFromBytes(ctx, wasmBytes)` to load your own WASM.

**Building the default policy WASM:** From the repo root, run `cd rust && make build-policy` to rebuild the Rust policy crate and copy `policy.wasm` into `extensions/wasm/validator/rustpolicy/`. The committed `policy.wasm` is a minimal allow-all module so the package builds without Rust; replace it with the Rust-built binary for the full implementation.

---

## Roadmap

### Features

| Status | Feature | Description |
|:---:|---|-------------|
| ✅ | ⚙️ **Typed pipeline** | Load config through an explicit pipeline (sources → merge → resolve → decode → validate) with no globals or magic precedence. Decode directly into your structs. |
| ✅ | 🧩 **Pluggable sources and formats** | File, env, memory, bytes, and flag sources; JSON (stdlib), YAML and TOML (Rust/WASM). Deep-merge and replace strategies; optional placeholder resolution (`${ENV:...}`, `${FILE:...}`, `${REF:...}`) and source metadata (priority, required/optional). |
| ✅ | 🔁 **Live reload** | File-system watcher (`runtime/watch/fsnotify`) and polling watcher (`runtime/watch/polling`) implement `ReloadTrigger`. `WatchTyped` delivers immutable snapshots per reload and an optional path-level config diff. |
| ✅ | ✨ **Ergonomic API** | `Load(ctx, &cfg)` for one-shot load; `LoadTyped[AppConfig](ctx, loader)` for typed snapshots; `WatchTyped[T](ctx, loader, trigger, callback)` for reload with (old, new, diff). All behaviour is opt-in and explicit. |
| ✅ | 🧪 **Testability** | No globals; sources, resolvers, decoders, and validators are injectable. Use `providers/source/memory` and a mock `ReloadTrigger` to test load and watch flows without the filesystem. |

### TO-DO

| Status | Feature | Description |
|:---:|---|-------------|
| ✅ | **WASM policy/validation engine** | Policy and validation rules executed via WebAssembly. |
| ✅ | **Schema generation** | Generate JSON Schema from config structs via `extensions/schema/generate` and infer from merged config trees via `extensions/schema/infer`. |
| ⬜ | **Remote sources: etcd** | Source that reads config from etcd. |
| ⬜ | **Remote sources: S3** | Source that reads config from AWS S3 (or compatible). |
| ⬜ | **Remote sources: Vault** | Source that reads secrets/config from HashiCorp Vault. |
| ⬜ | **Remote sources: Consul** | Source that reads config from Consul KV. |

---

## Benchmarks

The [`tooling/benchmarks/`](tooling/benchmarks/) module compares go-config with Viper and Koanf on shared fixtures (JSON, YAML, multi-source merge). It uses a nested `go.mod` so optional dependencies stay out of the core module. See [`tooling/benchmarks/README.md`](tooling/benchmarks/README.md) for commands and how to capture output for [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).

For unified benchmark + profile + coverage summaries, use [`tooling/reports/`](tooling/reports/) and run:

```bash
make report-local
make report-pr-local
```

Reporting notes:
- `tooling/reports/schemas/summary.schema.json` is the canonical top-level report contract.
- Coverage strict-threshold reporting reads targets from `tooling/reports/schemas/coverage-targets.manifest.json`.
- Tree schema inference uses first-element array heuristics for heterogeneous arrays.

---

## Contributing

```bash
git clone https://github.com/ArmanAvanesyan/go-config.git
cd go-config
go mod tidy
go test ./...
```

Run the full test suite with race detection:

```bash
go test -race ./...
```

**Before submitting a PR:**
1. All tests pass: `go test ./...`
2. No vet issues: `go vet ./...`
3. New packages must not import optional capabilities in the core — see [Dependency Policy](#dependency-policy).
4. If Rust parser/validator sources changed, rebuild and commit WASM artifacts: `make wasm-build` (or `make wasm-verify` to enforce no drift).

---

## License

[MIT License](LICENSE)
