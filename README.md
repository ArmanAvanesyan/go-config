# go-config

[![Go Reference](https://pkg.go.dev/badge/github.com/ArmanAvanesyan/go-config.svg)](https://pkg.go.dev/github.com/ArmanAvanesyan/go-config)
![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
[![CI](https://github.com/ArmanAvanesyan/go-config/actions/workflows/ci.yml/badge.svg)](https://github.com/ArmanAvanesyan/go-config/actions/workflows/ci.yml)

A typed, **pipeline-based** configuration system for Go with explicit behavior and zero globals.

> **Not another Viper wrapper.** No globals, no magic, no hidden precedence.
> Just an explicit pipeline:
>
> **Sources → Parse → Merge → Resolve → Decode → Validate → your struct**

---

## Why go-config?

Single comparison versus Viper and Koanf (✅ yes · ⚠️ partial or pattern-dependent · ❌ no):

| Capability | Viper | Koanf | go-config |
|---|---|---|---|
| Typed pipeline (decode into structs, not string state) | ❌ | ⚠️ | ✅ |
| No globals | ❌ | ✅ | ✅ |
| Dependency-light core | ❌ | ⚠️ | ✅ (stdlib-first core) |
| Strict decode / unknown keys | ⚠️ | ✅ | ✅ |
| First-class resolver stage (`${ENV:...}`, `${FILE:...}`, `${REF:...}`) | ❌ | ⚠️ | ✅ |
| WASM-backed YAML/TOML parsers | ❌ | ❌ | ✅ |
| Reload snapshots + path-level diff | ❌ | ❌ | ✅ |
| Composable pipeline (explicit stages) | ❌ | ⚠️ | ✅ |

Unlike Viper and Koanf, go-config uses **one deterministic pipeline** instead of ad-hoc key access: **Sources → Parse → Merge → Resolve → Decode → Validate**.

---

## What You Get

- Deterministic configuration loading (no hidden precedence)
- Strong typing end-to-end (no casting, no globals)
- Fully composable pipeline (swap any stage)
- Live reload with structured diff awareness
- WASM-powered parsing and policy validation

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [What You Get](#what-you-get)
- [Core Concepts](#core-concepts)
  - [Mental Model](#mental-model)
  - [Configuration Pipeline (Contracts)](#configuration-pipeline-contracts)
  - [Capability Layers](#capability-layers)
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
  - [Typed load](#typed-load)
  - [Custom validation](#custom-validation)
  - [Live reload](#live-reload)
- [Package Layout](#package-layout)
- [Architecture Docs](#architecture-docs)
- [Dependency Policy](#dependency-policy)
- [Schema generation](#schema-generation)
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

The `config` layer is stdlib-first. YAML/TOML use Rust/WASM via [wazero](https://github.com/tetratelabs/wazero) — see [Rust/WASM Parsers](#rustwasm-parsers) and [docs/wasm-parsers.md](docs/wasm-parsers.md). The core stays stdlib-only for public loader contracts; extensions may add dependencies — see [Dependency Policy](#dependency-policy).

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

### Mental Model

`go-config` treats configuration as a **data pipeline**, not global state.

Each stage is:

- explicit
- replaceable
- composable

You build a configuration system by composing:

- Sources (where data comes from)
- Parsers (how input is interpreted)
- Merge strategy (how conflicts are resolved)
- Resolvers (how references are expanded)
- Decoder (how data becomes typed)
- Validator (how correctness is enforced)

### Configuration Pipeline (Contracts)

At a high level:

```
Sources → Parse → Merge → Resolve → Decode → Validate
```

Each stage is defined by a contract:

- **Source**: returns `Document` or `TreeDocument`
- **Parser**: converts raw `Document` bytes into tree data
- **merge.Strategy**: combines trees deterministically
- **Resolver**: transforms merged trees (for example placeholders or indirection)
- **Decoder**: maps tree data into typed targets
- **Validator**: enforces post-decode correctness rules

### Capability Layers

The system is structured into layered capabilities:

#### Layer 1 — Core Engine

- pipeline orchestration
- typed loading
- deterministic behavior

#### Layer 2 — Providers (Built-ins)

- sources: file, env, flags, memory, bytes
- parsers: JSON, YAML, TOML
- resolvers, decoders, validators, merge strategies

#### Layer 3 — Runtime Capabilities

- live reload
- diff tracking
- immutable old/new snapshots in callbacks

#### Layer 4 — Extensions

- WASM-backed parser/runtime adapters
- policy validation adapters
- schema generation/inference helpers

#### Layer 5 — Tooling

- benchmark suites
- profiling/report scripts
- CI parity and comparison workflows

This layering ensures:

- core stays dependency-light
- optional features remain isolated
- behavior is explicit and composable

### Pipeline

Every call to `Load` executes the same deterministic pipeline:

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

A `Parser` converts raw `*Document` bytes into a structured tree (`map[string]any`).

> Most users do not need to build WASM artifacts — prebuilt binaries are embedded.

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

> **Advanced:** transport/ABI details: [docs/wasm-parsers.md](docs/wasm-parsers.md), [architecture extensions](docs/architecture.md#12-extensions-reference).

YAML uses ABI v2 (`GCFGMP1` + Msgpack). After changes to YAML parser transport or runtime, rebuild the YAML WASM artifact (`cd rust && make build-yaml`) so guest and host stay aligned. Contributors building from source: [docs/wasm-parsers.md](docs/wasm-parsers.md).

> JSON parsing uses the Go standard library with zero additional setup.

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

A `Resolver` transforms the merged tree before decoding (placeholder expansion, indirection).

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

A `Decoder` maps the final merged tree into a typed struct.

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

A `Validator` runs after decoding. Default is no-op.

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

> All examples are deterministic and side-effect free — no global state involved. For loader options and composition, see [docs/configuration.md](docs/configuration.md).

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

YAML can use `${ENV:KEY}`, `${FILE:path}`, and `${REF:key.path}` (see [Resolvers](#resolvers)). Wire a resolver on the loader, for example `chain.New(file, env, ref)` so file and env expand before `${REF:...}` sees the merged tree.

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

Use `WatchTyped` with a `ReloadTrigger` to receive immutable snapshots on each reload:

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

Each reload produces a new immutable snapshot; the callback receives `(old, new)` and `changes` (path-level diff).

Reload guarantees:

- no shared mutable state
- serialized updates (no race conditions)
- diff-aware change handling

Use `runtime/watch/polling` for a polling-based trigger with no file watcher dependency.

---

## Package Layout

Public API: **`config/`**. Everything else is optional providers, runtime helpers, or internal packages.

| Area | Role |
|------|------|
| `config/` | Loader, options, contracts (`Source`, `Parser`, `Decoder`, …) |
| `providers/` | Sources, parsers, merge, decoders, resolvers, validators |
| `runtime/` | Watch triggers (`fsnotify`, `polling`), config diff |
| `extensions/` | WASM parsers/validator, schema generate/infer |
| `internal/` | Pipeline engine and helpers (not a stable API) |
| `examples/`, `rust/` | Runnable samples and WASM crate sources |

For a fuller map of packages and behavior, see [docs/architecture.md](docs/architecture.md).

---

## Architecture Docs

For focused implementation references:

- [Docs index](docs/index.md)
- [Architecture](docs/architecture.md)
- [Configuration guide](docs/configuration.md)
- [Engineering standards](docs/engineering-standards.md)
- [Pipeline](docs/architecture.md#10-pipeline-reference)
- [Runtime](docs/architecture.md#11-runtime-and-reload-reference)
- [Diagnostics](docs/diagnostics.md)
- [Extensions](docs/architecture.md#12-extensions-reference)
- [FAQ](docs/faq.md)
- [Roadmap (docs)](docs/roadmap.md) — process priorities, shipped surface, feature backlog
- [Rust/WASM parsers (docs)](docs/wasm-parsers.md)
- [Testing guide](docs/testing.md)
- [Release process](docs/release.md)

## Dependency Policy

**The core must not depend on optional capabilities.**

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
| `providers/parser/yaml` | `wazero` (WASM runtime; shares transport stack with `extensions/wasm`) |
| `providers/parser/toml` | `wazero` (WASM runtime; shares transport stack with `extensions/wasm`) |
| `extensions/wasm/*` | `wazero` (WASM runtime); `github.com/vmihailenco/msgpack/v5` (parser ABI v2 output decode) |
| `extensions/wasm/validator/*` | `wazero` (WASM runtime) |
| `runtime/watch/fsnotify` | `golang.org/x/sys` (Linux/macOS: OS events; Windows/other: stdlib polling only) |
| `runtime/watch/polling` | none — stdlib only |
| `runtime/diff` | none — stdlib only |

The core `config/` package imports only stdlib and `runtime/diff` (for Watch diff). Optional stacks add third-party modules: `github.com/tetratelabs/wazero` (WASM runtime, no CGO), `github.com/vmihailenco/msgpack/v5` (Msgpack decode of WASM parser transport after the `GCFGMP1` prefix), and `golang.org/x/sys` (optional, for `runtime/watch/fsnotify` OS-event backend on Linux and macOS only; Windows and other platforms use stdlib-only polling).

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

For a best-effort schema from a merged tree, use `extensions/schema/infer` (`infer.GenerateFromTree`). Output targets JSON Schema draft 2020-12 by default.

---

## Rust/WASM Parsers

YAML and TOML are parsed by Rust crates compiled to WebAssembly and run in-process with **wazero**. Prebuilt `.wasm` files are embedded, so **you do not need Rust** to use the module as a dependency.

**Why:** serde-based parsers, one shared WASM runtime, no `gopkg.in/yaml.v3` / `go-toml` dependency for those formats.

**Contributors:** rebuilding parsers or policy WASM, parser ABI, optional Rust JSON parser, and `rustpolicy` usage → **[docs/wasm-parsers.md](docs/wasm-parsers.md)**. Validation ABI: [docs/architecture.md#validation-wasm-abi](docs/architecture.md#validation-wasm-abi).

---

## Roadmap

The core loader is production-ready. **Process priorities, what is already shipped, and the feature backlog** (including remote sources such as etcd, S3, Vault, and Consul) live in **[docs/roadmap.md](docs/roadmap.md)** — that file is the canonical roadmap so this README stays a stable overview.

---

## Benchmarks

Benchmarks are provided to measure and compare:

- pipeline overhead
- parser performance
- multi-source merge behavior

The [`tooling/benchmarks/`](tooling/benchmarks/) module compares go-config with Viper and Koanf (nested `go.mod` keeps optional deps out of the core). Commands and **benchstat** workflow: [`tooling/benchmarks/README.md`](tooling/benchmarks/README.md).

Unified benchmark, profile, and coverage summaries: [`tooling/reports/`](tooling/reports/) (`make report-local`, `make report-pr-local`). Schema contracts and coverage targets live under `tooling/reports/schemas/`.

<!-- BENCHMARK_TABLE:START -->
Representative comparison snapshot (auto-generated from `tooling/reports/output/summary.json`; lower is better for `ns/op`):

| Benchmark point (`ns/op`) | go-config | Viper | Koanf |
| ------ | ------: | ------: | ------: |
| `Compare/All/JSON` | `1.00x` (4851) | `3.44x` (16711) | `4.70x` (22784) |
| `Compare/All/YAML` | `1.00x` (35300) | `1.00x` (35197) | `1.17x` (41474) |
| `Compare/ParseOnly/YAML` | `1.00x` (33325) | `0.51x` (17066) | `0.56x` (18548) |

`go-config` is normalized to `1.00x`; peer values are time ratios (`peer_ns/op / go-config_ns/op`) for the same benchmark point, so values above `1.00x` indicate the peer is slower and values below `1.00x` indicate the peer is faster. For statistical comparisons across runs, use the `benchstat` workflow in [`tooling/benchmarks/README.md`](tooling/benchmarks/README.md).
<!-- BENCHMARK_TABLE:END -->

Benchmark automation reference:

| Target | When | Output |
| ------ | ---- | ------ |
| `make bench-local-smoke` | Local development quick check | `tooling/benchmarks/results/raw/bench-*.txt` |
| `make bench-local` | Local full comparative run | `tooling/benchmarks/results/raw/bench-*.txt` |
| `make report-local` / `make report-pr-local` | After local benchmark/profile/coverage runs | `tooling/reports/output/summary.md`, `tooling/reports/output/pr-comment.md`, `tooling/reports/output/summary.json` |
| `make bench-readme-refresh-local` | After `make report-local` to refresh README benchmark multipliers | Updates benchmark table block in `README.md` from `tooling/reports/output/summary.json` |
| [`.github/workflows/benchmarks.yml`](.github/workflows/benchmarks.yml) | Scheduled weekly or manual dispatch | GitHub artifact `benchmark-tooling-<run_id>` with raw benchmarks, dashboard JSON, and report outputs |

The main PR CI workflow ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) intentionally does not run the full benchmark suite; use the dedicated benchmark workflow for smoke/full benchmark runs and artifacts.

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
