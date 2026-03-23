# Go Library Conventions

This document defines the standards applied to every library in this project. These rules exist to keep the codebase predictable, dependency-light, and safe to depend on in production.

---

## Repository Naming

Module paths and repository names follow the pattern `go-<capability>`:

```
go-config
go-httpclient
go-errors
go-observability
go-retry
go-cache
```

The import path becomes `github.com/<org>/go-<capability>`.

**Avoid:** `awesome-config`, `config-lib`, `config-go-lib`, `superconfig`. Too generic, redundant, or unclear.

---

## Package Naming

Inside a repository, package names must be:

- Lowercase
- Short
- No underscores
- No camelCase

**Good:** `config`, `loader`, `validator`, `extensions/schema`, `retry`, `cache`

**Bad:** `config_loader`, `configUtils`, `myConfigPackage`

### No stuttering

If the package is named `config`, do not prefix its exported names with `Config`:

```go
// Bad — stutters on import: config.ConfigLoader
type ConfigLoader struct{}

// Good — reads naturally: config.Loader
type Loader struct{}
```

---

## Public API Naming

Exported names must read naturally at the call site. The package name provides context.

```go
// Bad
config.LoadConfigFileConfiguration()

// Good
config.Load()
config.LoadFile()
config.LoadEnv()
```

### Interface naming

Interfaces describe behavior, not objects. Name them after what they do:

```go
// Good
type Loader interface{}
type Validator interface{}
type Provider interface{}
type Resolver interface{}
type Registry interface{}

// Bad — Java-style
type IConfigLoader interface{}
type ConfigLoaderInterface interface{}
```

---

## Exported Symbols

Every exported type, function, method, and constant must have a godoc comment. The comment must start with the symbol name and end with a period.

```go
// Config holds the resolved application configuration.
type Config struct {
    Port int
}

// Load loads configuration using the provided options.
// It returns an error if any required fields are missing or invalid.
func Load(opts ...Option) (*Config, error) { ... }
```

Unexported symbols that contain non-obvious logic should also have comments, but this is not required.

---

## Dependency Rules

### Prefer the standard library

Every external dependency is a liability. Before adding one, verify the standard library cannot do the job.

### Never depend on frameworks

Libraries must be framework-neutral. Do not import `gin`, `fiber`, `echo`, or any HTTP framework. Consumers choose their own stack.

### Allowed external dependencies

Dependencies should be stable, well-maintained, and widely adopted:

- `google.golang.org/protobuf`
- `go.uber.org/zap`
- `gopkg.in/yaml.v3`
- `github.com/google/uuid`

Experimental or unmaintained repositories must not be added.

### Do not leak internal packages into the public API

```go
// Bad — exposes internal types to consumers
func Load(l internal.Loader) (*Config, error)

// Good — internal stays internal
func Load(opts ...Option) (*Config, error)
```

### Avoid cyclic dependencies

```
config → loader → schema      ✅ clean chain
config → loader → schema → config      ❌ cycle
```

---

## Module Versioning

Versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html):

| Segment | Meaning                             |
| ------- | ----------------------------------- |
| PATCH   | Backward-compatible bug fix         |
| MINOR   | Backward-compatible new feature     |
| MAJOR   | Breaking change to the public API   |

For v2 and beyond, the module path must include the major version suffix:

```
github.com/<org>/go-config/v2
```

The `go.mod` must reflect this:

```
module github.com/<org>/go-config/v2
```

---

## API Stability Policy

Once a library reaches `v1.0.0`, the public API is stable. No breaking changes may be introduced without incrementing the major version. This means:

- Exported function signatures do not change
- Exported types do not remove or reorder fields
- Existing behavior is not removed without a deprecation cycle

Pre-`v1` releases (`v0.x`) may break at any time.

---

## Error Handling

- Always return errors. Never swallow them.
- Wrap errors from external calls with `fmt.Errorf("context: %w", err)` to preserve the chain.
- Use typed errors (`errors.go`) when callers need to inspect the error kind.
- Do not use `panic` as a control-flow mechanism. Panics are reserved for programmer errors that represent unrecoverable states (e.g. nil interface passed where never documented as acceptable).

```go
// Bad
func Load() {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // ❌
    }
}

// Good
func Load() error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("load: read file: %w", err)
    }
    _ = data
    return nil
}
```

---

## Logging Rules

Libraries must never log. Logging is the caller's responsibility.

- Do not call `log.Fatal`, `log.Panic`, or `os.Exit`.
- Do not import `log`, `log/slog`, `go.uber.org/zap`, or any logging framework in library code.
- Return errors and let the application decide what to do.

```go
// Bad
func Connect() {
    if err := dial(); err != nil {
        log.Fatal(err) // ❌ kills the caller's process
    }
}

// Good
func Connect() error {
    if err := dial(); err != nil {
        return fmt.Errorf("connect: %w", err)
    }
    return nil
}
```

---

## Context Rules

Every function that performs I/O, network communication, or any operation that could block must accept a `context.Context` as its first parameter.

```go
// Bad — no cancellation support
func Fetch(url string) ([]byte, error)

// Good
func Fetch(ctx context.Context, url string) ([]byte, error)
```

Context must be propagated, not discarded:

```go
// Bad
req, _ := http.NewRequest("GET", url, nil)

// Good
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

---

## Security Rules

- Do not hardcode secrets, tokens, or credentials anywhere in library code.
- Do not use `unsafe` unless absolutely necessary and clearly documented.
- Do not call `os.Exit` or `syscall.Exit`.
- Do not use `reflect` in ways that bypass type safety without a documented rationale.
- Prefer typed configuration structs over `map[string]interface{}`.
- Validate all inputs at API boundaries; return descriptive errors for invalid inputs.
