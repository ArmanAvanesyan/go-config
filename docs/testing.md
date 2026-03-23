# Testing Guide

This document describes the testing strategy, conventions, and performance workflow for the `go-config` library.

---

## Running tests

```bash
# Unit tests
go test ./...

# Race detector
go test -race ./...

# Integration tests
go test -tags=integration ./...
go test -tags=integration -race ./...

# Fuzz (example target)
go test -fuzz=FuzzParseJSON ./providers/parser/json/ -fuzztime=30s

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
go tool cover -html=coverage.out

# Integration coverage
go test -tags=integration ./... -cover -coverprofile=coverage.integration.out
go tool cover -func=coverage.integration.out | grep total

# Project targets
make test-cover
make test-cover-integration
```

## CI parity

To match CI behavior locally, run:

```bash
go mod tidy && git diff --exit-code
make wasm-verify-docker
go build ./...
go test ./... -count=1 -race -short
make test-cover
```

## Test layers

- Unit tests: package-local behavior with deterministic inputs.
- Integration tests: end-to-end pipeline behavior with real parser/decoder/validator implementations.
- Benchmarks: hot paths, allocation profiles, and regression tracking.
- Fuzz tests: parser hardening; targets must never panic.

## Refactor contracts

When refactoring internal packages, treat these contracts as part of the test surface. Contract changes should be intentional behavior changes and should update tests in the same change.

### `internal/decode`

- Weak mode allows string-based numeric and bool parsing when possible.
- Strict mode does not coerce arbitrary strings into non-string scalar targets.
- Signed integer assignment from unsigned inputs must reject values that exceed `math.MaxInt64`.
- Unsigned integer assignment must reject negative values.
- `ErrorUnused` must return an unknown-field error for unmapped input keys.

### `internal/engine`

- Non-slice binding input must return a deterministic error.
- Reflection adapter input must never panic on malformed shapes; it must return errors.
- Source read/parse/merge failures must preserve `config` sentinels via wrapping.

### `internal/normalize`

- `Key()` trims and lowercases keys.
- `Path()` normalizes each `.`-separated segment via `Key()`.
- `EnvToPath()` converts env key separators to path separators before `Path()`.

### `internal/tree`

- `CloneMap()` deep-clones nested `map[string]any` values.
- Non-map nested values are copied by assignment (reference semantics for slices/pointers are preserved).
- `Walk()` is depth-first over nested maps and must no-op on nil tree or nil visitor.

## Conventions

- Prefer table-driven tests with named subtests and `t.Run`.
- Use helpers from `testutil/` (`RequireNoError`, `RequireErrorIs`, `RequireEqual`, etc.).
- Avoid mocks for core pipeline behavior; use real sources/parsers where practical.
- Use `t.Parallel()` for test and subtest scopes unless package-global mutation prevents it.
- Put integration-only tests behind `//go:build integration`.

## Coverage targets

| Package group | Target |
| ------------- | ------ |
| `config/` | >= 80% |
| `providers/merge/` | >= 90% |
| `providers/decoder/` | >= 85% |
| `providers/resolver/` | >= 85% |
| `providers/source/` (excluding file) | >= 80% |
| `providers/parser/json` | >= 85% |
| `providers/validator/` | >= 80% |
| `extensions/wasm/` | integration-only exemption |
| `runtime/watch/` | temporary exemption until implemented |

## Performance and benchmarks

Benchmark tooling and reports live under `tooling/benchmarks`.
CI benchmark automation lives in [`.github/workflows/benchmarks.yml`](../.github/workflows/benchmarks.yml) (scheduled and manual runs with artifact upload).

Quick commands:

```bash
make bench-local
make bench-local-smoke
make bench-report-local
make bench-compare-local
make bench-hyperfine-local
make bench-yaml-baseline-local
```

Raw benchmark fallback:

```bash
go test ./... -run=^$ -bench=. -benchmem
```

Reporting guidance:

- Include exact commands.
- Include environment basics (OS, CPU class, Go version).
- Call out smoke vs full runs.
- Compare multiple runs for noisy paths; do not treat one run as definitive.
