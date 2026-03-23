$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$outDir = Join-Path $root "results/raw"
$runs = if ($env:RUNS) { $env:RUNS } else { "10" }
$warmup = if ($env:WARMUP) { $env:WARMUP } else { "2" }
$benchPattern = if ($env:BENCH_PATTERN) { $env:BENCH_PATTERN } else { "Benchmark" }

New-Item -ItemType Directory -Path $outDir -Force | Out-Null

hyperfine `
  --warmup $warmup `
  --runs $runs `
  --export-json (Join-Path $outDir "hyperfine.json") `
  --export-markdown (Join-Path $outDir "hyperfine.md") `
  -n "go-bench" `
  "cd `"$root`" ; go test -bench=$benchPattern -benchmem -count=1 ./..."
