# FAQ

## What is `go-config`?

A typed, modular configuration library for Go with explicit pipeline stages (sources, parse, merge, resolve, decode, validate).

## What Go version is required?

See `go.mod` for the canonical minimum. The README badge reflects that version.

## How do I install it?

```bash
go get github.com/ArmanAvanesyan/go-config
```

## Do I need Rust to use this library?

Not always.

- If you consume already-committed WASM artifacts, normal Go usage works without local Rust setup.
- If you modify Rust parsers/validators or regenerate WASM artifacts, use the documented Docker flow (`make wasm-verify-docker`).

## Where is pipeline behavior documented?

- [architecture.md#10-pipeline-reference](./architecture.md#10-pipeline-reference)
- [architecture.md](./architecture.md)

## How do I run the main test suites?

- Unit/race: `go test ./... -race`
- Integration: `go test -tags=integration ./...`
- Coverage helpers: `make test-cover`, `make test-cover-integration`

Full details: [testing.md](./testing.md).

## Why does `make wasm-verify` sometimes fail locally but CI passes?

Host toolchains can produce different `.wasm` bytes than the pinned CI container. Use `make wasm-verify-docker` for parity with CI.

## How do I report bugs or propose features?

Open a GitHub issue with:

- clear repro steps
- expected vs actual behavior
- environment details (`go version`, OS, and relevant commands)

## Is this a drop-in replacement for Viper/Koanf?

No. It is a separate library with different design goals (typed-first API, explicit staged pipeline, and optional WASM-backed parsing/validation features).
