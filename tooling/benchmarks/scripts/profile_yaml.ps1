$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$stamp = Get-Date -Format "yyyyMMddHHmmss"
$base = Join-Path "results/raw" "yaml-profile-$stamp"
New-Item -ItemType Directory -Path "results/raw" -Force | Out-Null
$profileMode = if ($env:PROFILE_MODE) { $env:PROFILE_MODE } else { "warm" }

Write-Output "=== YAML PROFILE MODE: $profileMode ==="
go test ./scenarios -run "^$" -bench "BenchmarkCompare_YAML_ParseOnly/Compare/ParseOnly/YAML/go-config" -benchmem -count 1 -cpuprofile "$base-cpu.out" -memprofile "$base-mem.out"
$goTestExit = $LASTEXITCODE

if ($goTestExit -eq 0 -and (Test-Path "$base-cpu.out")) {
	go tool pprof -top "$base-cpu.out" | Tee-Object -FilePath "$base-cpu-top.txt"
}

if (Test-Path "$base-mem.out") {
	go tool pprof -top -alloc_space "$base-mem.out" | Tee-Object -FilePath "$base-mem-top.txt"
}

if ($goTestExit -ne 0) {
	Write-Warning "CPU+mem profiling run failed (often SIGPROF/runtime issue on Windows). Retrying with memprofile only..."
	go test ./scenarios -run "^$" -bench "BenchmarkCompare_YAML_ParseOnly/Compare/ParseOnly/YAML/go-config" -benchmem -count 1 -memprofile "$base-mem.out"
	$goTestExit = $LASTEXITCODE
	if ($goTestExit -eq 0 -and (Test-Path "$base-mem.out")) {
		go tool pprof -top -alloc_space "$base-mem.out" | Tee-Object -FilePath "$base-mem-top.txt"
	}
}

Write-Output "Wrote profiles:"
Write-Output "$base-cpu.out"
Write-Output "$base-mem.out"
Write-Output "Profile mode: $profileMode"
exit $goTestExit
