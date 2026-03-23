# Release Process

This document defines the operational runbook for cutting and publishing releases.
Versioning policy and API stability expectations are defined in [engineering-standards.md](./engineering-standards.md).

## Version and tag format

```
vMAJOR.MINOR.PATCH
```

| Segment | When to increment                  |
| ------- | ---------------------------------- |
| PATCH   | Backward-compatible bug fixes      |
| MINOR   | Backward-compatible new features   |
| MAJOR   | Breaking changes to the public API |

Pre-release tags use `v1.0.0-alpha`, `v1.0.0-beta`, `v1.0.0-rc1`.

---

## Automated Release (recommended)

The release workflow is fully automated. To cut a release:

1. Go to **Actions → Release → Run workflow** on GitHub.
2. Enter the version number (e.g. `1.2.0` — without the `v` prefix).
3. Click **Run workflow**.

The workflow will then:

1. Verify `go.mod` is tidy and all tests pass.
2. Auto-generate a new `CHANGELOG.md` entry from commits since the last tag, grouped by conventional-commit prefix (`feat`, `fix`, `refactor`, `perf`, `chore`).
3. Update the Go version badge in `README.md` to match the `go` directive in `go.mod`.
4. Commit both files back to the default branch with the message `release vX.X.X`.
5. Create and push the `vX.X.X` tag.
6. The tag push automatically triggers GoReleaser, which creates the GitHub Release with release notes.

### Branch protection note

If your default branch requires pull requests, the bot commit in step 4 will fail.
To resolve this, either:
- Exempt `github-actions[bot]` from the "require a pull request" branch protection rule, or
- Commit any pending `CHANGELOG.md` edits manually before triggering the workflow.

---

## Manual Release (fallback)

If you prefer to release without the workflow:

1. Verify dependencies and run tests:

   ```bash
   go mod tidy && go mod verify
   go test ./... -race -cover
   golangci-lint run ./...
   ```

2. Update `CHANGELOG.md` with a new `## [vX.X.X] - YYYY-MM-DD` entry.

3. Update the Go version badge in `README.md` if the `go` directive in `go.mod` changed.

4. Commit and tag:

   ```bash
   git commit -am "release vX.X.X"
   git tag vX.X.X
   git push origin vX.X.X
   ```

   Pushing the tag triggers the `goreleaser` job in `release.yml`, creating the GitHub Release automatically.

---

## Release checklist

### Pre-release validation

Before running the release workflow, validate:

- `go mod tidy && git diff --exit-code`
- `go test ./... -count=1 -race -short`
- `golangci-lint run ./...`
- `make wasm-verify-docker`
- docs and user-facing examples are updated for any behavior/API changes

### Release execution

- Confirm the default branch is green and up to date.
- Run automated release workflow (recommended) or manual fallback.
- Confirm the `vX.X.X` tag exists on remote.
- Confirm GitHub Release notes were generated correctly.

### Post-release verification

- Verify pkg.go.dev reflects the new tag.
- Verify README badges render correctly.
- Run a smoke install/load example against the new version.
- Monitor CI, issues, and regressions for the first 24-48 hours.

---

## Rollback and hotfix lane

For release mistakes or critical regressions:

1. Assess impact and classify severity.
2. If needed, publish a fast patch (`vX.Y.Z+1`) with the fix.
3. Update release notes with issue context and upgrade guidance.
4. Link incident/follow-up issue in the release notes for traceability.

Prefer forward-fix patch releases over rewriting published tags.

---

## Release notes template

```markdown
## go-config vX.X.X

### Highlights
- Key feature or behavior updates

### Added
- ...

### Changed
- ...

### Fixed
- ...

### Notes
- Migration or compatibility notes (if any)
```

---

## Badge maintenance

`README.md` currently ships these badges; keep their URLs and automation aligned with releases:

| Badge        | How it updates                                           |
| ------------ | -------------------------------------------------------- |
| Go Reference | Always live, based on the module path.                   |
| Go Version   | **Auto-updated** by the release workflow from `go.mod`. |
| CI Status    | Always live, points to `.github/workflows/ci.yml`.      |

If you add optional badges later (for example package version, coverage, or benchmark links), document how each is updated in the same table.

If you rename the repository, change the GitHub organization, or rename workflow files, update the badge URLs in `README.md` accordingly.
