# Documentation index

This folder holds technical documentation for **go-config**.

## Start here

For policies and contribution standards:

- [engineering-standards.md](./engineering-standards.md)
- [CONTRIBUTING.md](../CONTRIBUTING.md)

For system behavior and implementation details:

- [architecture.md](./architecture.md)
- [Pipeline Reference](./architecture.md#10-pipeline-reference)
- [Runtime and Reload Reference](./architecture.md#11-runtime-and-reload-reference)
- [Extensions Reference](./architecture.md#12-extensions-reference)

## Usage overview

For day-to-day usage, most readers only need these three guides:

- [configuration.md](./configuration.md): how to build a loader, attach sources/parsers, and choose options.
- [Runtime and Reload Reference](./architecture.md#11-runtime-and-reload-reference): how to run reload loops with triggers and callback lifecycle expectations.
- [diagnostics.md](./diagnostics.md): how to classify failures by stage and troubleshoot common CI/runtime issues.

Use [architecture.md](./architecture.md) as the single technical specification, with focused sections for pipeline, runtime, and extensions.

## Suggested reading order

1. [engineering-standards.md](./engineering-standards.md)
2. [architecture.md](./architecture.md)
3. [configuration.md](./configuration.md)
4. [Pipeline Reference](./architecture.md#10-pipeline-reference)
5. [Runtime and Reload Reference](./architecture.md#11-runtime-and-reload-reference)
6. [diagnostics.md](./diagnostics.md)
7. [Extensions Reference](./architecture.md#12-extensions-reference)
8. [testing.md](./testing.md)

## Current guides

| Document | Topic |
| -------- | ----- |
| [roadmap.md](./roadmap.md) | Directional priorities, shipped surface, feature backlog |
| [wasm-parsers.md](./wasm-parsers.md) | Building parser/policy WASM, ABI overview, optional Rust JSON parser |
| [engineering-standards.md](./engineering-standards.md) | Branch workflow, commit/release rules, and code/API standards |
| [architecture.md](./architecture.md) | System structure and components |
| [configuration.md](./configuration.md) | Loader options and configuration composition |
| [diagnostics.md](./diagnostics.md) | Error model and troubleshooting workflows |
| [architecture.md#12-extensions-reference](./architecture.md#12-extensions-reference) | Extensions reference (parsers, runtime, validation ABI) |
| [faq.md](./faq.md) | Common usage and workflow questions |
| [architecture.md#10-pipeline-reference](./architecture.md#10-pipeline-reference) | Pipeline reference (stage semantics and precedence) |
| [release.md](./release.md) | Releases and versioning |
| [architecture.md#11-runtime-and-reload-reference](./architecture.md#11-runtime-and-reload-reference) | Reload API and watch backends |
| [testing.md](./testing.md) | How to run tests, benchmark workflows, and maintain internal refactor contracts |

The repository **[README.md](../README.md)** is the primary user-facing overview at the repo root.

---

## Scope: documentation refreshes

When doing **broad doc updates** (structure, accuracy, cross-links), use this split so policy and legal files stay unchanged unless there is a deliberate, separate change.

### In scope

- **[README.md](../README.md)** (root)
- **`docs/*.md`** — technical docs, including (for example):
  - `architecture.md`, `testing.md`, `release.md`
  - New or renamed guides such as **`roadmap.md`** or a root-level **`release.md`** if you introduce it — keep links in this index updated accordingly.

### Out of scope (do not edit as part of a general doc refresh)

Unless a dedicated PR explicitly targets them:

- `LICENSE`
- `CONTRIBUTING.md`
- `CODE_OF_CONDUCT.md`
- `CHANGELOG.md`
- `SUPPORT.md`
- `SECURITY.md`

Those files cover licensing, contribution rules, conduct, release notes, support, and security reporting; treat changes there as **separate, review-heavy PRs**.
