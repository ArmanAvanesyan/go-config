# Diagnostics

This guide consolidates stage error handling and troubleshooting workflows for `go-config`.

## 1. Error model

`go-config` wraps failures with stage-specific sentinel errors while preserving root causes.

Primary stage sentinels:

- `ErrSourceReadFailed`
- `ErrParseFailed`
- `ErrMergeFailed`
- `ErrResolutionFailed`
- `ErrDecodeFailed`
- `ErrValidationFailed`

Why this model:

- callers can branch on stage with `errors.Is`
- low-level context stays available in wrapped chains
- diagnostics remain actionable in logs

Typical handling:

```go
if err := loader.Load(ctx, &cfg); err != nil {
    switch {
    case errors.Is(err, config.ErrParseFailed):
        // format/parser input issue
    case errors.Is(err, config.ErrDecodeFailed):
        // target type mismatch
    case errors.Is(err, config.ErrValidationFailed):
        // post-decode validation rejection
    default:
        // generic handling
    }
}
```

Operational guidance:

- treat parse/decode failures as configuration quality issues
- treat read failures as source availability/permissions issues
- keep wrapped error text in logs to preserve root-cause detail

## 2. Common failure scenarios

### 2.1 CI fails on WASM verification

Symptom: `.wasm` drift in CI or `make wasm-verify`.

Fix:

```bash
make wasm-verify-docker
```

If drift remains, regenerate in Docker and commit updated artifacts.

### 2.2 `wasm-verify-docker` fails with git errors in container workflows

`wasm-verify-docker` is designed to build in Docker and diff on host. Avoid relying on in-container `git diff` behavior against bind mounts for custom flows.

### 2.3 Integration tests fail due to parser/validator setup

Try:

```bash
make wasm-build-docker
go test -tags=integration ./...
```

### 2.4 Race tests flaky on watcher paths

- rerun with `-count=1`
- avoid unnecessary parallelism for watch-sensitive tests
- ensure test paths are regular files, not directories

### 2.5 Coverage commands include unwanted packages

Use project targets:

```bash
make test-cover
make test-cover-integration
```

### 2.6 Reload callback appears to miss errors

`WatchTyped` keeps reload loops alive; callback/reload error handling should be explicit in callback logic and application logging.

### 2.7 Lint mismatch between local and CI

Run the same command as CI:

```bash
golangci-lint run
```

## 3. Escalation checklist

When opening an issue, include:

- exact command(s) run
- full error output
- OS and shell
- `go version`
- whether Docker/WASM targets were used

## 4. Related docs

- [architecture.md#10-pipeline-reference](./architecture.md#10-pipeline-reference)
- [architecture.md#11-runtime-and-reload-reference](./architecture.md#11-runtime-and-reload-reference)
- [configuration.md](./configuration.md)
- [architecture.md#12-extensions-reference](./architecture.md#12-extensions-reference)
- [architecture.md#validation-wasm-abi](./architecture.md#validation-wasm-abi)
