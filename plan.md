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

- [x] Add semver-protected contract tests for:
  - [x] env precedence
  - [x] hook ordering
  - [x] merge behavior
  - [x] decode coercion semantics
- [x] Add release checklist item requiring migration notes for policy behavior changes.

Exit criteria:
- [x] CI contract suite is mandatory for merge/release.

## Phase 4: Consumer Compatibility and Deprecation Completion

- [x] Build a consumer-compatibility harness that reproduces representative downstream config outcomes using `go-config` alone (no app-specific adapter assumptions).
- [x] Run parity scenarios against maintained compatibility fixtures that represent key consumer configuration profiles, and record no-diff results.
- [x] Publish a deprecation/migration policy for compatibility wrappers (support window, expected behavior, upgrade notes), and monitor reported drift during the window.
- [ ] Remove compatibility wrappers only after repeated no-diff parity runs and completion of the published deprecation window.

Exit criteria:
- [ ] Repeated no-diff parity outcomes across maintained compatibility fixtures.
- [ ] Deprecation window completed with documented migration guidance and no unresolved blocker regressions.
- [ ] Wrapper removal approved under the project's semver and release policy.

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

### Checkpoint 1: Contract Harness

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added dedicated `TestContract_*` suites for policy compatibility assertions.
  - Organized contract coverage by behavior domain (env, merge, hooks, decode).
  - Kept tests deterministic and fixture-driven.
- Evidence:
  - Modules: `config/contract_merge_semantics_test.go`, `config/contract_hook_order_test.go`, `config/contract_decode_coercion_test.go`, `providers/source/env/contract_precedence_test.go`.
- Next: checkpoint 2 (env+merge contracts).

### Checkpoint 2: Env and Merge Contracts

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added env precedence contract tests for explicit and inferred precedence modes.
  - Added merge contract tests for priority/stable ordering and policy behavior.
  - Added missing-policy fail contract assertion.
- Evidence:
  - Tests: `TestContract_EnvPrecedence_ExplicitFirst`, `TestContract_EnvPrecedence_InferredFirst`, `TestContract_MergeSemantics_*`.
- Next: checkpoint 3 (hook+decode contracts).

### Checkpoint 3: Hook and Decode Contracts

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added hook-order contracts for regular decode and direct-decode paths.
  - Added strict/permissive decode coercion contract tests.
  - Locked duration coercion expectations in contract-level coverage.
- Evidence:
  - Tests: `TestContract_HookOrdering_*`, `TestContract_DecodeCoercion_*`.
- Next: checkpoint 4 (release governance).

### Checkpoint 4: Release Governance

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added policy contract suite step to pre-release checklist.
  - Added explicit migration-note requirement for policy behavior changes.
- Evidence:
  - Docs: `docs/release.md`.
- Next: checkpoint 5 (CI gate).

### Checkpoint 5: CI Gate

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added dedicated `contract-suite` CI job.
  - Added release workflow contract-suite execution step.
  - Standardized contract command as `go test ./... -run '^TestContract_' -count=1`.
- Evidence:
  - Workflows: `.github/workflows/ci.yml`, `.github/workflows/release.yml`.
- Next: checkpoint 6 (closeout).

### Checkpoint 6: Phase 3 Exit Criteria Closure

- Status: completed
- Date: 2026-04-01
- Changes:
  - Ran contract suite and regression package tests.
  - Marked Phase 3 checklist and exit criteria complete.
  - Recorded checkpoint evidence and Phase 4 handoff.
- Evidence:
  - Tests: `go test ./... -run '^TestContract_' -count=1`; `go test ./config ./providers/source/env ./internal/decode`.
- Next: Phase 4 planning.

### Checkpoint 1: Compatibility Harness Contract

- Status: completed
- Date: 2026-04-01
- Changes:
  - Documented fixture-driven compatibility harness contract (input, execution, comparison).
  - Added architecture guidance for maintained compatibility profiles and provenance checks.
- Evidence:
  - Docs: `docs/architecture.md`.
- Next: checkpoint 2 (fixtures + parity scenarios).

### Checkpoint 2: Compatibility Fixtures and Parity Scenarios

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added representative compatibility fixtures for strict, multi-source, and env-heavy profiles.
  - Implemented `TestContract_CompatParity_Fixtures` with canonical no-diff output assertions.
  - Added trace winner assertions for selected compatibility-sensitive keys.
- Evidence:
  - Fixtures: `testdata/compat/security_strict_profile.json`, `testdata/compat/multi_source_profile.json`, `testdata/compat/env_heavy_override_profile.json`.
  - Tests: `config/contract_compatibility_test.go`.
- Next: checkpoint 3 (CI/release gating).

### Checkpoint 3: CI/Release Compatibility Gate

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added dedicated compatibility contract command in CI and release workflows.
  - Kept general policy contract suite gate while making compatibility parity explicit.
- Evidence:
  - Workflows: `.github/workflows/ci.yml`, `.github/workflows/release.yml`.
- Next: checkpoint 4 (deprecation policy).

### Checkpoint 4: Deprecation and Migration Policy

- Status: completed
- Date: 2026-04-01
- Changes:
  - Added wrapper deprecation support window and migration expectations to release runbook.
  - Added objective wrapper-removal go/no-go criteria and drift-monitoring expectations.
- Evidence:
  - Docs: `docs/release.md`.
- Next: Phase 4 closeout criteria tracking.
