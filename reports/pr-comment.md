## Tooling report

- Commit: `ad7643a6ab86e805e5350e5d38ea3c3f54688b68`
- Bench rows: `870`
- Profile files: `0`
- Coverage: `n/a%`

No benchstat comparison data was found.

### YAML go-config vs peers

- **All YAML (load+decode)**
  - vs `viper`: ns/op `-97.27%`, B/op `-99.35%`, allocs/op `-98.03%`
  - vs `koanf`: ns/op `-97.80%`, B/op `-99.42%`, allocs/op `-98.59%`
- **Parse-only YAML**
  - vs `viper`: ns/op `-90.37%`, B/op `-90.79%`, allocs/op `-82.46%`
  - vs `koanf`: ns/op `-90.79%`, B/op `-90.53%`, allocs/op `-84.62%`
