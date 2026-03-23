# Release Process

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Version Format

```
vMAJOR.MINOR.PATCH
```

| Segment | When to increment                  |
| ------- | ---------------------------------- |
| PATCH   | Backward-compatible bug fixes      |
| MINOR   | Backward-compatible new features   |
| MAJOR   | Breaking changes to the public API |

Pre-release versions use the format `v1.0.0-alpha`, `v1.0.0-beta`, `v1.0.0-rc1`.

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

## Badge Maintenance

The `README.md` includes badges that stay in sync with each release:

| Badge           | How it updates                                                      |
| --------------- | ------------------------------------------------------------------- |
| Go Reference    | Always live, based on the module path.                              |
| Go Version      | **Auto-updated** by the release workflow from `go.mod`.            |
| Version         | Auto-refreshes when a new tag is pushed.                           |
| CI Status       | Always live, points to `.github/workflows/ci.yml`.                 |
| Coverage        | Always live, provided by Codecov/Coveralls (configure separately). |
| Benchmarks      | Static link to the benchmark file; update manually if renamed.     |

If you rename the repository, change the GitHub organization, or rename workflow files, update the badge URLs in `README.md` accordingly.
