# Traced Implementation Plan

This file tracks execution of the adapter-removal capability roadmap in a traceable way.

## Scope

- Repository: `go-config`
- Source of requirements: `docs/roadmap.md` (`Adapter policy parity`, `Compatibility guarantees for policy behavior`, `Safe migration path for adapter removal`)
- Goal: make adapter policy behavior first-class in `go-config` before removing wrappers

## Trace Matrix (Requirement -> Workstream -> Verification)

| Requirement | Workstream | Verification |
| --- | --- | --- |
| Declarative env binding (aliases, precedence, normalization, tags) | Env binding subsystem + tag extraction | Contract tests for precedence and alias resolution |
| Deterministic lifecycle hook ordering | Hook pipeline contract | Ordered execution tests and failure-path tests |
| Optional/multi-file merge semantics | File loading policy model | Merge precedence and missing-file policy tests |
| Typed decode coercion policy | Decode mode and coercion rules | Strict/permissive test matrix including `time.Duration` |
| Stable high-level app API (`Spec` style) | Public app-facing facade | Integration tests and docs examples |
| Validation contracts | Unified callback/interface validation flow | Error-shape and field-path assertions |
| Explain/trace provenance | Config explain mode | Source attribution tests and snapshot checks |
| Compatibility/versioning guarantees | Contract test suite + release policy checks | CI contract suite gate on release branches |
| Safe migration path | Shadow parity harness + staged rollout | No-diff outcomes across target scenarios |

## Phased Plan

## Phase 1: Policy Primitives

- [x] Implement env binding model (`key -> [ENV...]`, precedence, normalization).
- [x] Add tag-driven binding support (`env:"A,B"`) with nested path derivation.
- [x] Define and enforce lifecycle hook ordering contract.
- [x] Formalize optional-file and multi-file merge policy behavior.
- [x] Expand decode coercion policy (strict vs permissive).

Exit criteria:
- [x] Unit/contract tests exist for each policy primitive.
- [x] Public behavior is documented in docs.

## Phase 2: App Surface and Observability

- [x] Add stable high-level app API (`Load(..., Spec{...})`-style).
- [x] Unify validation contracts across callbacks/interfaces with consistent error wrapping.
- [x] Add explain/trace mode for key provenance and precedence decisions.

Exit criteria:
- [x] Typed/integration tests cover app API and trace output.
- [x] Docs include usage and expected behavior examples.

## Phase 3: Compatibility Guarantees

- [ ] Add semver-protected contract tests for:
  - [ ] env precedence
  - [ ] hook ordering
  - [ ] merge behavior
  - [ ] decode coercion semantics
- [ ] Add release checklist item requiring migration notes for policy behavior changes.

Exit criteria:
- [ ] CI contract suite is mandatory for merge/release.

## Phase 4: Safe Migration and Adapter Removal

- [ ] Build shadow parity harness that reproduces adapter outcomes using `go-config` only.
- [ ] Run parity scenarios for `security`, `sona`, `reporter`.
- [ ] Keep thin compatibility wrappers for 1-2 releases while monitoring drift.
- [ ] Remove wrappers only after stable no-diff results.

Exit criteria:
- [ ] Repeated no-diff parity runs across scenarios.
- [ ] Explicit go/no-go sign-off recorded.

## Execution Notes

- Keep changes incremental; avoid bundling multiple policy shifts in one PR.
- Prefer one capability per PR with focused contract tests.
- Treat policy behavior deltas as compatibility-sensitive and document them.

## Checkpoint Trace Log

### Checkpoint 1: Declarative Env Binding Core

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added explicit env binding model (`Options.Bindings`) with precedence controls.
  - Added inferred+explicit lookup behavior in env source configuration.
  - Added normalization-aware inferred-name generation for bound keys.
- Evidence:
  - Tests: `TestEnvSource_ExplicitBindingsTakePrecedence`.
  - Modules: `providers/source/env/env.go`, `providers/source/env/env_test.go`.
- Next: checkpoint 2 (tag-driven bindings).

### Checkpoint 2: Tag-Driven Env Binding

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added `BindingsFromStruct` to derive aliases from `env:"A,B"` tags.
  - Added nested path derivation using `mapstructure/json` field keys.
  - Integrated struct-derived aliases via `UseStructTagEnvFor`.
- Evidence:
  - Tests: `TestEnvSource_TagBindingsFromStruct`.
  - Modules: `providers/source/env/env.go`, `providers/source/env/env_test.go`.
- Next: checkpoint 3 (lifecycle ordering).

### Checkpoint 3: Deterministic Lifecycle Hook Ordering

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added post-decode lifecycle chain: `ApplyDefaults` -> defaults callback -> validate callback -> validator interface.
  - Added options `WithDefaultsFunc` and `WithValidateFunc`.
  - Ensured same ordering for direct-decode and regular pipeline.
- Evidence:
  - Tests: `TestLoader_LifecycleHookOrdering`.
  - Modules: `config/options.go`, `config/loader.go`, `config/loader_test.go`, `config/errors.go`.
- Next: checkpoint 4 (optional/multi-file semantics).

### Checkpoint 4: Optional/Multi-File Merge Semantics

- Status: completed
- Date: 2026-04-01
- Changes:
  - Extended `SourceMeta` with `MissingPolicy` and `ParsePolicy`.
  - Added stage-aware error handling for read vs parse failures.
  - Implemented ignore behavior for missing-file and parse-policy cases.
- Evidence:
  - Tests: `TestLoader_SourceMeta_MissingPolicyIgnore`, `TestLoader_SourceMeta_ParsePolicyIgnore`.
  - Modules: `config/types.go`, `config/loader.go`, `config/loader_test.go`.
- Next: checkpoint 5 (decode coercion policy).

### Checkpoint 5: Decode Coercion Policy

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added built-in weak coercion for `string -> time.Duration`.
  - Preserved strict-mode rejection for duration strings.
  - Kept existing bool/int/uint/float coercion behavior.
- Evidence:
  - Tests: `TestAssignScalar_DurationWeakParsing`, `TestAssignScalar_DurationStrictRejectsString`.
  - Modules: `internal/decode/coercion.go`, `internal/decode/coercion_test.go`.
- Next: checkpoint 6 (Phase 1 closeout).

### Checkpoint 6: Phase 1 Exit Criteria Closure

- Status: completed
- Date: 2026-04-01
- Changes:
  - Updated docs to describe lifecycle ordering and env binding capabilities.
  - Verified targeted package test suites for changed areas.
  - Closed Phase 1 checklist and exit criteria.
- Evidence:
  - Tests: `go test ./providers/source/env ./config ./internal/decode`.
  - Docs: `docs/configuration.md`, `docs/architecture.md`.
- Next: Phase 2 planning.

### Checkpoint 1: Spec API Facade

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added high-level `Spec` API with `LoadWithSpec`, `LoadTypedWithSpec`, and `NewFromSpec`.
  - Added source-oriented `SourceSpec` and `TreeSpec` declarations.
  - Kept low-level `Loader` API backward-compatible.
- Evidence:
  - Tests: `TestLoadWithSpec_ParityWithManualLoader`, `TestLoadTypedWithSpec`.
  - Modules: `config/spec.go`, `config/spec_test.go`.
- Next: checkpoint 2 (validation contract).

### Checkpoint 2: Validation Contract Unification

- Status: completed
- Date: 2026-04-01
- Changes:
  - Standardized validation failure messaging with stage markers:
    - `validate-callback`
    - `validate-interface`
  - Preserved sentinel compatibility with `ErrValidationFailed`.
- Evidence:
  - Tests: `TestLoader_ValidationContract_InterfaceStageMarker`, `TestLoadWithSpec_ValidationErrorContract`.
  - Modules: `config/loader.go`, `config/loader_test.go`, `config/spec_test.go`.
- Next: checkpoint 3 (trace core).

### Checkpoint 3: Trace Core

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added trace model (`Trace`, `KeyTrace`, `TraceCandidate`) and collector.
  - Captured per-key merge provenance and final source decisions.
  - Captured lifecycle hook execution order in trace output.
- Evidence:
  - Tests: `TestLoadWithSpec_TraceCapturesSourceAndHooks`.
  - Modules: `config/trace.go`, `config/loader.go`.
- Next: checkpoint 4 (trace surface).

### Checkpoint 4: Trace Surface API

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added `WithTrace` option for loader-based flows.
  - Added `Spec.Trace` wiring for app-level API flows.
  - Added traced merge path via `LoadAndMergeBindingsWithTrace`.
- Evidence:
  - Tests: `go test ./config` (covers loader + spec trace flows).
  - Modules: `config/options.go`, `config/loader.go`, `config/spec.go`.
- Next: checkpoint 5 (docs).

### Checkpoint 5: Docs Update

- Status: completed
- Date: 2026-04-01
- Changes:
  - Documented `Spec` API usage and typed helper in configuration guide.
  - Documented trace mode behavior and outputs.
  - Updated architecture references for spec APIs and trace semantics.
- Evidence:
  - Docs: `docs/configuration.md`, `docs/architecture.md`.
- Next: checkpoint 6 (closeout).

### Checkpoint 6: Phase 2 Exit Criteria Closure

- Status: completed
- Date: 2026-04-01
- Changes:
  - Verified package tests across modified areas.
  - Marked Phase 2 checklist and exit criteria complete.
  - Recorded checkpoint evidence and next-stage handoff.
- Evidence:
  - Tests: `go test ./config ./providers/source/env ./internal/decode`.
- Next: Phase 3 planning.
