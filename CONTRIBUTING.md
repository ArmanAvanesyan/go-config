# Contributing

Thank you for considering contributing to this project. We welcome bug reports, feature requests, documentation improvements, and code contributions.

---

## Standards

Before writing any code, read:

- [`docs/conventions.md`](./docs/conventions.md) — naming rules, API design, export requirements, dependency policy, logging rules, context rules, and security expectations.
- [`docs/testing.md`](./docs/testing.md) — table-driven tests, golden tests, integration tests, benchmarks, fuzz tests, and coverage targets.

---

## Development Setup

**Requirements:**

- Go 1.26+
- Git

**Rust / WASM (when changing `rust/` parsers or validators):** CI runs **`make wasm-verify-docker`** — WASM is built in **`rust:1.94-bookworm`**, then **`git diff`** runs on the host (see [`.github/workflows/ci.yml`](./.github/workflows/ci.yml)). **A host `rustc` (including WSL) often produces different `.wasm` bytes** than that image, so `make wasm-verify` may fail locally even when the repo is correct—use the Docker target instead:

```bash
make wasm-verify-docker   # same toolchain as CI; must pass before you push WASM changes
```

Equivalent manual sequence (`make wasm-build` in Docker, then diff on your machine):

```bash
make wasm-build-docker
git diff --exit-code HEAD -- extensions/wasm/parser/rusttoml/toml_parser.wasm \
  extensions/wasm/parser/rustyaml/yaml_parser.wasm \
  extensions/wasm/parser/rustjson/json_parser.wasm \
  extensions/wasm/validator/rustpolicy/policy.wasm
```

After editing Rust sources, `make wasm-build-docker` refreshes checked-in artifacts; then commit if `git status` shows updates under `extensions/wasm/`.

`wasm-verify` only diffs the four copied binaries (`toml_parser.wasm`, `yaml_parser.wasm`, `json_parser.wasm`, `rustpolicy/policy.wasm`) against `HEAD`, so unrelated Go changes in `extensions/wasm/` won’t fail the check.

**Clone and install dependencies:**

```bash
git clone https://github.com/<your-org>/<your-lib>.git
cd <your-lib>
go mod tidy
```

**Available Makefile targets:**

```bash
make tidy              # go mod tidy
make fmt               # go fmt ./...
make lint              # golangci-lint run ./...
make test              # go test ./...
make test-race         # go test ./... -race -cover
make test-integration  # go test -tags=integration ./integration/...
make bench             # benchmarks (time only)
make bench-mem         # benchmarks with allocation stats
make fuzz              # run fuzz tests for 60s
make check             # fmt + vet + lint + test-race (full pre-commit pass)
```

---

## Code Style

- Use `gofmt` formatting (`make fmt`).
- Follow Go naming conventions as defined in [`docs/conventions.md`](./docs/conventions.md).
- All exported symbols must have godoc comments ending with a period.
- Keep dependencies minimal — see the dependency rules in `docs/conventions.md`.

---

## Submitting Changes

1. Fork the repository.
2. Create a branch following the naming convention:
   - `feature/<short-description>` for new features
   - `bugfix/<short-description>` for bug fixes
   - `hotfix/<short-description>` for critical production fixes
   - `release/<version>` for release branches
3. Commit your changes following the [commit message guidelines](#commit-message-guidelines).
4. Push to your fork and open a Pull Request against `main`.

---

## Pull Request Requirements

Before opening a PR:

- [ ] Run `make check` — this runs fmt, vet, lint, and test-race in sequence.
- [ ] Add or update tests to cover your change (see [`docs/testing.md`](./docs/testing.md)).
- [ ] Add or update documentation for any public API changes.
- [ ] Run `make bench-mem` if your change touches a performance-sensitive path.

The CI pipeline runs `go mod tidy`, `go vet`, `golangci-lint run`, and `go test ./... -race -cover`. All checks must pass before merging.

---

## Dependency Hygiene

This project favors:

- Standard library where possible.
- Few, well-justified external dependencies.
- Small, focused libraries over large frameworks.

When proposing a new dependency, explain why the standard library or an existing dependency is insufficient, and consider the impact on downstream users and build times. See the full dependency rules in [`docs/conventions.md`](./docs/conventions.md).

---

## Commit Message Guidelines

Use the conventional commits format:

```
feat: add retry logic for transient errors
fix: resolve nil dereference in error handler
docs: update README usage example
refactor: simplify parser state machine
test: add coverage for edge case in decoder
perf: reduce allocations in hot decode path
```

Subject lines should be 50 characters or fewer. Add a body when additional context is useful.

Prefixes used by the automated release workflow to categorize `CHANGELOG.md` entries: `feat`, `fix`, `refactor`, `perf`, `chore`.

---

## Reporting Bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml) when opening an issue. Include:

- Go version and OS
- Steps to reproduce
- Expected behavior
- Actual behavior and any relevant logs or stack traces

---

## Feature Requests

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.yml). Describe the problem you are trying to solve, your proposed solution, and any alternatives you considered.
