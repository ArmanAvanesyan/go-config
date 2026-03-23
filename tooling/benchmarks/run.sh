#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
count="${BENCH_COUNT:-10}"
pattern="${BENCH_PATTERN:-Benchmark}"
timeout="${BENCH_TIMEOUT:-30m}"
mkdir -p "results/raw"
# Run the benchmark scenario tree recursively so newly added subpackages are included.
go test ./scenarios/... -bench="$pattern" -benchmem -count="$count" -timeout="$timeout" "$@" | tee "results/raw/bench-$(date +%Y%m%d%H%M%S).txt"
