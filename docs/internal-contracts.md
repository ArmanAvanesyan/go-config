# Internal Behavior Contracts

This document defines explicit contracts for internal refactor safety.

## `internal/decode`

- Weak mode allows string-based numeric and bool parsing when possible.
- Strict mode does not coerce arbitrary strings into non-string scalar targets.
- Signed integer assignment from unsigned inputs must reject values that exceed `math.MaxInt64`.
- Unsigned integer assignment must reject negative values.
- `ErrorUnused` must return an unknown-field error for unmapped input keys.

## `internal/engine`

- Non-slice binding input must return a deterministic error.
- Reflection adapter input must never panic on malformed shapes; it must return errors.
- Source read/parse/merge failures must preserve `config` sentinels via wrapping.

## `internal/normalize`

- `Key()` trims and lowercases keys.
- `Path()` normalizes each `.`-separated segment via `Key()`.
- `EnvToPath()` converts env key separators to path separators before `Path()`.

## `internal/tree`

- `CloneMap()` deep-clones nested `map[string]any` values.
- Non-map nested values are copied by assignment (reference semantics for slices/pointers are preserved).
- `Walk()` is depth-first over nested maps and must no-op on nil tree or nil visitor.
