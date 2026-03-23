This directory holds test fixtures used by the library's tests.

Place files here (YAML, JSON, text, binary) and load them from test files using relative paths:

```go
data, err := os.ReadFile("testdata/config.yaml")
```

Go automatically excludes `testdata/` from builds. Any file or subdirectory placed here is safe to reference from tests without affecting the compiled package.

See [`docs/testing.md`](../docs/testing.md) for the full testing guide, including conventions for golden files and fuzz corpus entries.
