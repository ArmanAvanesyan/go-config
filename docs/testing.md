# Testing Guide

This document describes the testing strategy, conventions, and commands for the `go-config` library.

---

## Running Tests

```bash
# Unit tests (no external dependencies)
go test ./...

# Race detector — required before any merge
go test -race ./...

# Integration tests (require WASM parsers: cd rust && make all)
go test -tags=integration ./...
go test -tags=integration -race ./...

# Benchmarks
go test -bench=. -benchmem ./...
go test -bench=BenchmarkDeepOverride -benchmem ./providers/merge/...

# Fuzz — 30 second run against the JSON parser
go test -fuzz=FuzzParseJSON ./providers/parser/json/ -fuzztime=30s

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
go tool cover -html=coverage.out     # opens browser

# Integration coverage
go test -tags=integration ./... -cover -coverprofile=coverage.integration.out
go tool cover -func=coverage.integration.out | grep total

# Project Make targets (with exclusions for non-runtime/example helper packages)
make test-cover
make test-cover-integration
```

---

## Test Layers

```
┌──────────────────────────────────────────┐
│  E2E / Integration (//go:build integration)│
│  Full pipeline: file → parse → merge →    │
│  resolve → decode → validate              │
├──────────────────────────────────────────┤
│  Unit Tests (go test ./...)               │
│  Per-package, pure logic, no real I/O     │
├──────────────────────────────────────────┤
│  Benchmarks (go test -bench=.)            │
│  merge strategies, decoders, loaders      │
├──────────────────────────────────────────┤
│  Fuzz Tests (go test -fuzz=...)           │
│  providers/parser/json parser                       │
└──────────────────────────────────────────┘
```

Internal refactor invariants are documented in `docs/internal-contracts.md` and should be kept in sync with test expectations.

---

## Conventions

### Table-Driven Tests

Every test function uses a slice of named cases and `t.Run`:

```go
func TestFoo(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"happy path", "input", "expected", false},
        {"error case", "bad", "", true},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            got, err := Foo(tc.input)
            if tc.wantErr {
                testutil.RequireError(t, err)
                return
            }
            testutil.RequireNoError(t, err)
            testutil.RequireEqual(t, tc.want, got)
        })
    }
}
```

### testutil Helpers

Always use the helpers in `testutil/` instead of writing `if err != nil { t.Fatal(...) }` inline:

| Helper | Purpose |
|--------|---------|
| `RequireNoError(t, err)` | Fail if err != nil |
| `RequireError(t, err)` | Fail if err == nil |
| `RequireErrorIs(t, err, target)` | Fail if !errors.Is(err, target) |
| `RequireEqual(t, want, got)` | Fail if !reflect.DeepEqual(want, got) |
| `MustSetEnv(t, key, value)` | Set env var for test duration (auto-restored) |

All helpers call `t.Helper()` so failures point to the call site, not the helper body.

### No Mocks

Use real implementations throughout:

- Config trees → `providers/source/memory`
- Environment variables → `testutil.MustSetEnv` (wraps `t.Setenv`)
- File reads → `testdata/` files with `runtime.Caller(0)` for path resolution
- Parsers → real parser instances

Never create fake implementations that duplicate production logic.

### Parallelism

- Call `t.Parallel()` at the top of every test function and every subtest.
- Exception: tests that mutate package-level state (e.g., `flag.CommandLine`) must not be parallel with others in the same package.

### Build Tags for Integration Tests

Tests that require compiled WASM binaries or significant I/O use `//go:build integration`:

```go
//go:build integration

package foo_test
```

These are excluded from `go test ./...` and run only with `-tags=integration`.

### Benchmarks

```go
func BenchmarkFoo(b *testing.B) {
    // Setup work here
    input := expensiveSetup()

    b.ResetTimer()   // <— always reset after setup
    for i := 0; i < b.N; i++ {
        Foo(input)
    }
}
```

Always pass `-benchmem` to measure allocations. Use `benchstat` for comparing before/after:

```bash
go test -bench=BenchmarkFoo -benchmem -count=5 ./pkg/ | tee old.txt
# ... make changes ...
go test -bench=BenchmarkFoo -benchmem -count=5 ./pkg/ | tee new.txt
benchstat old.txt new.txt
```

### Fuzz Tests

Fuzz corpus seeds live in `testdata/fuzz/<package>/FuzzName/`. The fuzz target must never panic — it may return errors for invalid input.

```go
func FuzzFoo(f *testing.F) {
    f.Add([]byte("seed"))        // corpus seeds
    f.Fuzz(func(t *testing.T, data []byte) {
        _, _ = Foo(data)         // must not panic
    })
}
```

---

## Testdata

Canonical fixture files live in `testdata/` at the repository root:

| File | Purpose |
|------|---------|
| `basic.json` | Standard config in JSON format |
| `basic.yaml` | Same config in YAML format |
| `basic.toml` | Same config in TOML format |

All fixtures share the same logical structure:

```json
{
  "app":    { "name": "demo" },
  "server": { "host": "localhost", "port": 8080 }
}
```

When adding new fixtures, preserve this top-level structure so existing tests continue to work.

Reference testdata from tests using `runtime.Caller(0)` for reliable path resolution regardless of working directory:

```go
func testdataPath(name string) string {
    _, thisFile, _, _ := runtime.Caller(0)
    root := filepath.Join(filepath.Dir(thisFile), "..", "..")  // adjust depth
    return filepath.Join(root, "testdata", name)
}
```

---

## Coverage Targets

| Package group | Target |
|---------------|--------|
| `config/` | ≥ 80% |
| `providers/merge/` | ≥ 90% |
| `providers/decoder/` | ≥ 85% |
| `providers/resolver/` | ≥ 85% |
| `providers/source/` (excl. file) | ≥ 80% |
| `providers/parser/json` | ≥ 85% |
| `providers/validator/` | ≥ 80% |
| `extensions/wasm/` | integration-only (exempt from unit target) |
| `runtime/watch/` | stubs (exempt until implemented) |
