$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$baseFile = if ($env:BASE_FILE) { $env:BASE_FILE } else { Join-Path $root "results/raw/base.txt" }
$newFile = if ($env:NEW_FILE) { $env:NEW_FILE } else { Join-Path $root "results/raw/candidate.txt" }
$outDir = if ($env:OUT_DIR) { $env:OUT_DIR } else { Join-Path $root "results/compare" }
$outFile = if ($env:OUT_FILE) { $env:OUT_FILE } else { Join-Path $outDir "benchstat.txt" }

New-Item -ItemType Directory -Path $outDir -Force | Out-Null
$content = benchstat $baseFile $newFile
$content | Out-Host
$content | Set-Content -Path $outFile -Encoding utf8
