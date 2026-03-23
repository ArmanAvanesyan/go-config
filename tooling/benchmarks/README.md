# Local benchmarks (full suite)

This directory is a nested Go module dedicated to performance work. It keeps benchmarking dependencies out of the root module and benchmarks parsing/merge/decode behavior for go-config, Viper, and Koanf.

## Scenario coverage

- Taxonomy naming follows `Category/Subcategory/...` in sub-benchmark labels for cleaner `benchstat` grouping.
- `BenchmarkCompare_*` keeps baseline library-vs-library comparisons.
- `BenchmarkSingleSource_*` covers JSON/YAML/TOML fixture sizes.
- `BenchmarkMultiSource_*` covers layered merge behavior.
- `BenchmarkDecode_*` covers struct-shape decode costs.
- `BenchmarkScale_*` covers key-count growth.
- `BenchmarkScale_DepthJSON` covers nested-depth growth.
- `BenchmarkParallelLoad_*` covers concurrent load paths.
- `BenchmarkParallelReadAfterLoad_*` covers concurrent read hot-path.
- `BenchmarkErrorPath_*` covers invalid payload overhead.
- `BenchmarkSourceType_*` covers bytes/file/memory/env source overhead.
- `BenchmarkRuntime_*` covers loader reuse vs rebuild and hot/cold access patterns.
- `BenchmarkFeature_*` covers env interpolation, defaults layering, validation overhead, and YAML WASM parser behavior.
- `BenchmarkFeature_YAML_InitVsReuse` isolates YAML parser init cost vs per-benchmark reuse vs shared/global reuse.
- `BenchmarkCompare_YAML_ParseOnly` isolates pure parser cost from load+decode cost for fairer YAML comparisons.

### Taxonomy contract (no ad-hoc renames)

- Canonical sub-benchmark format: `Category/Subcategory[/Detail][/Library]`.
- Keep category casing stable (`Compare/All/...`, not `Compare/all/...`).
- Current canonical scale labels:
  - `Scale/Keys10`, `Scale/Keys100`, `Scale/Keys1000`
  - `Scale/Depth_2`, `Scale/Depth_5`, `Scale/Depth_10`
- Labels are defined in `scenarios/*.go` and must stay synchronized with docs/scripts.
- Validate labels after benchmark edits:

```bash
python3 ./scripts/validate_benchmark_labels.py
```

## Fixture layout

- `fixtures/json/{small,medium,large}.json`
- `fixtures/yaml/{small,medium,large}.yaml`
- `fixtures/toml/{small,medium,large}.toml`
- Existing `testdata/basic.*` remains the baseline compatibility fixture.

## Local run commands

From repo root:

```bash
cd tooling/benchmarks
go test -bench=Benchmark -benchmem -count=10
```

Quick runners:

- Unix: `./run.sh`
- PowerShell: `./run.ps1`

Filter specific benchmark groups:

```bash
BENCH_PATTERN=BenchmarkCompare ./run.sh
BENCH_PATTERN=BenchmarkScale ./run.sh
BENCH_PATTERN=BenchmarkCompare_YAML_ParseOnly ./run.sh
```

## Benchstat workflow

Install:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Capture baseline and candidate:

```bash
BENCH_PATTERN=BenchmarkCompare ./run.sh
cp results/raw/bench-<timestamp>.txt results/raw/base.txt

BENCH_PATTERN=BenchmarkCompare ./run.sh
cp results/raw/bench-<timestamp>.txt results/raw/candidate.txt
```

Compare:

```bash
./scripts/compare_benchmarks.sh
```

PowerShell:

```powershell
.\scripts\compare_benchmarks.ps1
```

YAML-focused before/after capture:

```bash
BENCH_PATTERN='BenchmarkCompare_SingleSource_YAML|BenchmarkCompare_YAML_ParseOnly|BenchmarkFeature_YAML_InitVsReuse' BENCH_COUNT=10 ./run.sh
cp results/raw/bench-<before>.txt results/raw/base.txt

# apply optimization changes, then rerun
BENCH_PATTERN='BenchmarkCompare_SingleSource_YAML|BenchmarkCompare_YAML_ParseOnly|BenchmarkFeature_YAML_InitVsReuse' BENCH_COUNT=10 ./run.sh
cp results/raw/bench-<after>.txt results/raw/candidate.txt

./scripts/compare_benchmarks.sh
```

Important for YAML v2 transport:

- `go-config` YAML benchmarks now require a rebuilt YAML WASM binary that matches runtime ABI v2.
- Rebuild before running YAML benches: `cd ../rust && make build-yaml` (or `make all`).
- With stale artifacts, `go-config` YAML benches fail with `invalid transport prefix`.
- For contributor parity with CI no-drift checks, from repo root run `make wasm-verify`.

## GitHub Pages (gobenchdata + hyperfine + reports + profiles)

Interactive charts, benchmark history, unified tooling reports, and YAML profile text exports are published by the same **Benchmarks** workflow ([`.github/workflows/benchmarks.yml`](../../.github/workflows/benchmarks.yml)) that runs the suite and uploads artifacts (manual trigger only).

### One-time repository setup

1. In **Settings → Pages**, set **Build and deployment** source to **Deploy from a branch**, branch **`gh-pages`**, folder **`/ (root)`** (or match where [gobenchdata](https://github.com/marketplace/actions/continuous-benchmarking-for-go) places `index.html`).
2. After the first successful run, open the site URL shown on the Pages settings page.

### How to publish

1. Go to **Actions → Benchmarks → Run workflow**.
2. Choose **full run** (`-count=10` for gobenchdata and `./run.sh`) only when you want a slower, more stable aggregate; default smoke uses `-count=1`.
3. Optionally change **bench pattern** (same as `go test -bench`).

The job:

- Runs **`./run.sh`** (same as the main **Benchmarks** workflow) so `tooling/benchmarks/results/raw/bench-*.txt` exists for dashboards and for [`tooling/reports/`](../reports/README.md).
- Runs [`scripts/profile_yaml.sh`](./scripts/profile_yaml.sh) and copies `yaml-profile-*-top.txt` into [`tooling/profiles/raw/`](../profiles/README.md) (text only; large `.pprof` files are not uploaded to Pages).
- Builds unified reports: `build_report.py`, `render_markdown.py`, `render_pr_comment.py` → published under **`reports/`** on `gh-pages` (`summary.json`, `summary.md`, `pr-comment.md`, `index.html`).
- Runs [hyperfine](https://github.com/sharkdp/hyperfine) on `go test -bench … -benchmem -count=1 ./scenarios/...` with bounded `RUNS`/`WARMUP` for CI (see workflow).
- Runs [gobenchdata](https://github.com/marketplace/actions/continuous-benchmarking-for-go) in **custom** mode (`go run go.bobheadxi.dev/gobenchdata@v1 action`) with `actions/setup-go` so the Go version matches the root [`go.mod`](../../go.mod) and the nested benchmarks module.
- Pushes gobenchdata’s app and merged `benchmarks.json` to **`gh-pages`**, then adds **`hyperfine/`**, **`reports/`**, and **`profiles/`** on the same branch.

### Optional PAT (`GOBENCHDATA_TOKEN`)

If the site does not update after pushes, GitHub sometimes does not rebuild Pages for commits made with the default `GITHUB_TOKEN`. Create a **personal access token** with `repo` scope, add it as repository secret **`GOBENCHDATA_TOKEN`**. The workflow uses it when set; otherwise it uses `github.token`. See the [Continuous Benchmarking for Go](https://github.com/marketplace/actions/continuous-benchmarking-for-go) notes on PAT vs `GITHUB_TOKEN`.

### Runtime expectations

The full **`./scenarios/...`** suite can take **tens of minutes** even with `-count=1` (similar to the main **Benchmarks** workflow). The workflow job timeout is **180 minutes**. Hyperfine repeats the command several times; CI uses low `RUNS`/`WARMUP` to cap wall time.

## Optional tools

- `hyperfine` command-level orchestration:
  - Unix: `./scripts/run_hyperfine.sh`
  - PowerShell: `./scripts/run_hyperfine.ps1`
- YAML baseline (cold vs warm snapshots):
  - Unix: `./scripts/capture_yaml_baseline.sh`
  - PowerShell: `.\scripts\capture_yaml_baseline.ps1`
  - Writes `results/raw/yaml-baseline-<timestamp>-{cold,warm}.txt`
- JSON summary extraction:
  - `python3 ./scripts/extract_go_bench.py <input.txt> results/dashboards/summary.json`
- YAML hotspot profiling (parse-only):
  - PowerShell: `.\scripts\profile_yaml.ps1`
  - Unix: `./scripts/profile_yaml.sh`
  - Outputs CPU and memory tops under `results/raw/yaml-profile-*.txt` and `.out`.

## YAML V3 checkpoint (latest run)

- Candidate run file: `results/raw/bench-20260323120311.txt`
- Compared against: `results/raw/bench-20260323115337.txt`
- `go-config` improvements observed:
  - `Compare/All/YAML`: allocs/op reduced from ~229 to ~201 (about 12%).
  - `Feature/YAML/ReusePerBenchmark`: allocs/op reduced from ~232 to ~204 (about 12%), with ~10% faster `ns/op`.
- Remaining bottleneck:
  - Parse-only and load+decode still spend ~2.47 MiB/op in `go-config` YAML path.
  - Throughput remains far behind Viper/Koanf, so next work should focus on reducing map/tree materialization cost.

## Makefile shortcuts

- `make bench-go`
- `make bench-go-smoke`
- `make bench-compare BASE_FILE=... NEW_FILE=...`
- `make bench-report INPUT=results/raw/bench-...txt`
- `make bench-hyperfine`
- `make bench-yaml-baseline`

## Fairness notes

- JSON and YAML comparisons use identical payloads and equivalent decode structs.
- YAML `go-config` benchmark depends on a rebuilt Rust/WASM YAML parser binary aligned to ABI v2.
- Multi-source comparisons keep equivalent key overrides and layer counts across libraries.

## Output locations

All generated files are stored under `results/`:

- `results/raw/`
- `results/compare/`
- `results/dashboards/`

To aggregate benchmark outputs with profiles and coverage into unified markdown/JSON reports, use `tooling/reports/` from the repo root:

```bash
make report-local
make report-pr-local
```
