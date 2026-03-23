#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
count="${BENCH_COUNT:-10}"
pattern="${BENCH_PATTERN:-Benchmark}"
mkdir -p "results/raw"
go test ./scenarios -bench="$pattern" -benchmem -count="$count" "$@" | tee "results/raw/bench-$(date +%Y%m%d%H%M%S).txt"
