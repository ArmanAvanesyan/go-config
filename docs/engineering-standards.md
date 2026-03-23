# Engineering Standards

This document defines the collaboration workflow and Go library standards for **go-config**.
It is the canonical policy for branch naming, pull request workflow, commit conventions, versioning policy, and code/API quality expectations.

## 1. Purpose and scope

Goals:

- Keep `main` stable, releasable, and green in CI.
- Keep development history reviewable through short-lived branches and pull requests.
- Keep the public API predictable, dependency-light, and safe for production use.

This policy applies to all contributors and all packages in this repository.

## 2. Collaboration standards

### 2.1 Branching model

The repository follows trunk-based development:

- `main` is the only long-lived integration branch.
- Daily work happens in short-lived topic branches.
- Changes merge to `main` via pull requests.

### 2.2 `main` branch policy

`main` should always be:

- Passing [CI](../.github/workflows/ci.yml) checks.
- Suitable for release tagging (see [release.md](./release.md)).

Typical branch protection settings in GitHub:

- No direct pushes to `main`.
- Required status checks before merge.
- Required review approval before merge.

### 2.3 Branch types and naming

Use type-prefixed branch names:

| Type        | Purpose |
| ----------- | ------- |
| `feat/`     | New functionality or capability |
| `fix/`      | Bug fixes |
| `refactor/` | Internal changes without intended behavior change |
| `perf/`     | Performance improvements |
| `docs/`     | Documentation-only changes |
| `test/`     | Test additions or test-only refactors |
| `chore/`    | Maintenance not tied to `ci/` |
| `ci/`       | CI/CD workflow or automation changes |

Legacy/alternate prefixes remain valid for compatibility:
`feature/`, `bugfix/`, `release/`, `hotfix/`.

Format:

```text
<type>/<scope>-<short-description>
```

Use kebab-case after the prefix and keep names specific.

Examples:

```text
feat/loader-direct-decode
fix/env-nested-mapping
refactor/pipeline-simplification
docs/readme-architecture
ci/add-benchmark-workflow
feature/loader-direct-decode
bugfix/env-nested-mapping
```

Automation reference: [branch-naming.yml](../.github/workflows/branch-naming.yml).

### 2.4 Pull request workflow

1. Create a branch from `main`.
2. Keep commits focused and coherent.
3. Push and open a PR into `main`.
4. Address review feedback and ensure CI is green.
5. Prefer squash merges to keep `main` linear.

Each PR should document:

- What changed and why.
- Scope of impact (API, behavior, docs, CI).

Checklist:

- [ ] CI passes.
- [ ] Tests are added or updated when behavior changes.
- [ ] Documentation is updated for user-facing or workflow changes.
- [ ] No unrelated changes are mixed into the PR.

## 3. Commit and versioning policy

### 3.1 Conventional commits

Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>(<scope>): <description>
```

Examples:

```text
feat(loader): add direct decode fast path
fix(env): handle empty prefix edge case
refactor(engine): simplify pipeline execution
docs(readme): link standards from index
```

Release tooling groups changelog notes by prefix (`feat`, `fix`, `refactor`, `perf`, `chore`, etc.). Release execution details are defined in [release.md](./release.md).

### 3.2 Versioning and API stability

The project follows [Semantic Versioning](https://semver.org/):

| Change | Typical bump |
| ------ | ------------- |
| New backward-compatible API | Minor |
| Bug fixes | Patch |
| Breaking API changes | Major |

Once the library is `v1.0.0+`, public API changes must remain backward-compatible unless accompanied by a major version bump.
For v2+, module paths must include the version suffix (for example `.../go-config/v2`).
This section defines policy only; release procedure, tagging, and automation are documented in [release.md](./release.md).

## 4. Go library engineering standards

### 4.1 Repository and package naming

- Repository/module naming follows `go-<capability>`.
- Package names must be lowercase, short, and free of underscores or camelCase.
- Avoid stuttering (for example avoid `config.ConfigLoader`; prefer `config.Loader`).

### 4.2 Public API naming and interfaces

- Exported names should read naturally at the call site.
- Interfaces should describe behavior (`Source`, `Parser`, `Validator`, `Resolver`, `ReloadTrigger`, and similar contracts in `config/`).
- Avoid Java-style interface names (`IConfigLoader`, `ConfigLoaderInterface`).

### 4.3 Exported symbol documentation

Every exported type, function, method, and constant must include godoc:

- Starts with the symbol name.
- Ends with a period.

### 4.4 Dependency rules

- Prefer the Go standard library.
- Do not depend on web frameworks in library code (`gin`, `fiber`, `echo`, etc.).
- Use stable, maintained dependencies only.
- Do not leak `internal` package types into public APIs.
- Avoid cyclic package dependencies.

## 5. Runtime and safety standards

### 5.1 Error handling

- Always return errors; never swallow errors.
- Wrap external failures with `fmt.Errorf("context: %w", err)`.
- Use typed errors when callers must branch on error kind.
- Do not use panic for normal control flow.

### 5.2 Logging

Libraries must not log or terminate the host process:

- Do not call `log.Fatal`, `log.Panic`, `os.Exit`, or `syscall.Exit`.
- Return errors and let applications decide logging/exit behavior.

### 5.3 Context propagation

Functions that perform I/O or blocking work must accept `context.Context` as the first parameter.
Always propagate context to downstream operations (for example `http.NewRequestWithContext`).

### 5.4 Security

- Never hardcode credentials or secrets.
- Avoid `unsafe` unless strictly necessary and documented.
- Avoid reflection patterns that bypass type safety without documented rationale.
- Prefer typed configuration structs to `map[string]interface{}` at API boundaries.
- Validate inputs and return descriptive errors.

## 6. Contributor quick checklist

Before opening a PR:

- Branch name follows the branch naming convention.
- Commit messages follow Conventional Commits.
- `main` target branch is used for the PR.
- CI is green and tests/documentation are updated as needed.
- Changes follow API, dependency, and safety rules in this document.

## 7. Related documents

- [CONTRIBUTING.md](../CONTRIBUTING.md)
- [release.md](./release.md)
- [architecture.md](./architecture.md)
- [testing.md](./testing.md)
