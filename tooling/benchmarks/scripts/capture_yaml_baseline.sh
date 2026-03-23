#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

stamp="$(date +%Y%m%d%H%M%S)"
raw_dir="results/raw"
mkdir -p "${raw_dir}"

pattern="${BENCH_PATTERN:-BenchmarkCompare_SingleSource_YAML|BenchmarkCompare_YAML_ParseOnly}"
count="${BENCH_COUNT:-5}"

cold_out="${raw_dir}/yaml-baseline-${stamp}-cold.txt"
warm_out="${raw_dir}/yaml-baseline-${stamp}-warm.txt"

echo "=== YAML BASELINE COLD START ==="
go test ./scenarios -bench "${pattern}" -benchmem -count "${count}" | tee "${cold_out}"

echo "=== YAML BASELINE WARM START ==="
go test ./scenarios -bench "${pattern}" -benchmem -count "${count}" | tee "${warm_out}"

echo "Wrote baseline files:"
echo "${cold_out}"
echo "${warm_out}"
