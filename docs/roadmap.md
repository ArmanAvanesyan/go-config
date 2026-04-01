# Roadmap

This document tracks likely next improvements for `go-config`.

It is intentionally lightweight: priorities can move based on bug reports, CI reliability, and contributor bandwidth.

The root [README.md](../README.md) stays a short overview; **this file is the canonical roadmap** (process priorities, shipped surface, and feature ideas).

## Product surface (shipped)

These capabilities exist today:

- Typed pipeline: explicit sources → parse → merge → resolve → decode → validate, no globals.
- Built-in sources (file, env, memory, bytes, flags), merge strategies, decoders (including strict), resolver packages, and function-based validators.
- YAML/TOML via Rust/WASM; JSON via stdlib (optional Rust JSON parser in extensions).
- Live reload with `WatchTyped`, immutable snapshots, and path-level diffs (`runtime/watch/*`, `runtime/diff`).
- Schema generation/inference (`extensions/schema/*`), WASM policy validation (`extensions/wasm/validator/rustpolicy`), benchmarks and reporting under `tooling/`.

## Feature backlog (not scheduled)

Concrete extensions that would fit the existing `Source` / pipeline model but are **not** on a fixed timeline:

- Remote source: **etcd**
- Remote source: **AWS S3** (or compatible object storage)
- Remote source: **HashiCorp Vault** (secrets/config)
- Remote source: **Consul** KV

If you want one of these, open an issue with the read semantics you need (paths, auth, refresh); scoped PRs are welcome.

## Current focus

- Keep CI fast and deterministic.
- Keep WASM artifact verification reproducible across environments.
- Improve docs clarity and reduce stale guidance.

## Near-term priorities

### Reliability and correctness

- Continue hardening watch/reload behavior across platforms.
- Reduce flaky test surfaces in integration and race runs.
- Keep parser/validator WASM verification stable in CI.

### Developer experience

- Improve docs discoverability and internal cross-links.
- Keep examples and docs in sync with actual APIs and Make targets.
- Clarify troubleshooting around local vs CI WASM builds.

### Performance and observability

- Expand local benchmark reporting under `tooling/benchmarks`.
- Track parser/merge/decode changes for regressions.
- Improve benchmark output summaries for PR discussion.

### Adapter policy parity

- Add a declarative env binding model: explicit `key -> [ENV1, ENV2...]` aliases, precedence controls between explicit and inferred names, and key normalization options.
- Support tag-driven env binding (`env:"A,B"`) with nested path derivation from `mapstructure` tags where applicable.
- Define and document deterministic lifecycle hook ordering for app-level policy stages (`ApplyDefaults`-style interfaces/callbacks and `Validate`-style interfaces/callbacks).
- Formalize optional-file and multi-file merge semantics, including per-file missing policies and explicit "later overrides earlier" guarantees.
- Expand typed decode coercion policy for common config inputs (for example `string -> time.Duration`) with clear strict vs permissive modes.
- Provide a stable high-level app API (`Load(..., Spec{...})`-style) so applications can declare policy once instead of wiring internals repeatedly.
- Improve validation contracts so callback and interface validators share consistent error wrapping and field-path context.
- Add optional explain/trace output that can report source provenance per key (file/env/default/hook) and final precedence decisions.

### Compatibility guarantees for policy behavior

- Add contract tests (and keep them semver-stable) for env precedence, hook ordering, merge behavior, and decode coercion semantics.
- Treat policy behavior changes as compatibility-impacting, requiring explicit release notes and migration guidance.

### Safe migration path for adapter removal

1. Land policy-parity features in `go-config` with focused contract tests.
2. Build a shadow parity suite that mirrors current adapter outcomes using only `go-config`.
3. Run parity scenarios across `security`, `sona`, and `reporter` configurations.
4. Keep thin compatibility wrappers for 1-2 releases while monitoring no-diff behavior.
5. Remove wrappers only after parity results are stable and regression risk is acceptable.

## Medium-term ideas

- Additional typed decoding and strictness controls where practical.
- More guidance and examples for runtime reload workflows.
- Better coverage tooling ergonomics for package-level targets.

## Out of scope for roadmap

- Fixed date-based promises.
- Features not grounded in this repository's current architecture.
- Cross-project ecosystem plans outside `go-config`.

## How to propose roadmap items

- Open a GitHub issue with problem statement, desired behavior, and impact.
- If accepted, keep PR scope focused and include doc/test updates.
- For large changes, add design notes in `docs/` first.

## Notes

- This file is directional, not contractual.
- Release mechanics remain documented in [release.md](./release.md).
