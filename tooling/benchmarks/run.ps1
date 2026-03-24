$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
$count = if ($env:BENCH_COUNT) { $env:BENCH_COUNT } else { "10" }
$pattern = if ($env:BENCH_PATTERN) { $env:BENCH_PATTERN } else { "Benchmark" }
$timeout = if ($env:BENCH_TIMEOUT) { $env:BENCH_TIMEOUT } else { "30m" }
$stamp = Get-Date -Format "yyyyMMddHHmmss"
New-Item -ItemType Directory -Path "results/raw" -Force | Out-Null
$out = Join-Path "results/raw" "bench-$stamp.txt"
$goArgs = @(
	"-bench", $pattern,
	"-benchmem",
	"-count", $count,
	"-timeout", $timeout
) + $args
go test ./scenarios @goArgs | Tee-Object -FilePath $out
if ($LASTEXITCODE -ne 0) {
	exit $LASTEXITCODE
}
