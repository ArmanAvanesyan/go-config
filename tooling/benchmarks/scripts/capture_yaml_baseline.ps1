$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$stamp = Get-Date -Format "yyyyMMddHHmmss"
$rawDir = "results/raw"
New-Item -ItemType Directory -Path $rawDir -Force | Out-Null

$pattern = if ($env:BENCH_PATTERN) { $env:BENCH_PATTERN } else { "BenchmarkCompare_SingleSource_YAML|BenchmarkCompare_YAML_ParseOnly" }
$count = if ($env:BENCH_COUNT) { $env:BENCH_COUNT } else { "5" }

$coldOut = Join-Path $rawDir "yaml-baseline-$stamp-cold.txt"
$warmOut = Join-Path $rawDir "yaml-baseline-$stamp-warm.txt"

Write-Output "=== YAML BASELINE COLD START ==="
go test ./scenarios -bench "$pattern" -benchmem -count $count | Tee-Object -FilePath $coldOut
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Output "=== YAML BASELINE WARM START ==="
go test ./scenarios -bench "$pattern" -benchmem -count $count | Tee-Object -FilePath $warmOut
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Output "Wrote baseline files:"
Write-Output $coldOut
Write-Output $warmOut
