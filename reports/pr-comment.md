## Tooling report

- Commit: `1ddad5c7cae9af432ec239d99a9c89f79144b3a9`
- Bench rows: `870`
- Profile files: `0`
- Coverage: `n/a%`

No benchstat comparison data was found.

### YAML go-config vs peers

- **All YAML (load+decode)**
  - vs `viper`: ns/op `-96.88%`, B/op `-99.35%`, allocs/op `-98.03%`
  - vs `koanf`: ns/op `-97.38%`, B/op `-99.42%`, allocs/op `-98.59%`
- **Parse-only YAML**
  - vs `viper`: ns/op `-89.91%`, B/op `-90.79%`, allocs/op `-82.46%`
  - vs `koanf`: ns/op `-90.72%`, B/op `-90.53%`, allocs/op `-84.62%`
