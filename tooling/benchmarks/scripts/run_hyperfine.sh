#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
OUT_DIR="$ROOT_DIR/results/raw"
RUNS="${RUNS:-10}"
WARMUP="${WARMUP:-2}"
BENCH_PATTERN="${BENCH_PATTERN:-Benchmark}"

mkdir -p "$OUT_DIR"

hyperfine \
  --warmup "$WARMUP" \
  --runs "$RUNS" \
  --export-json "$OUT_DIR/hyperfine.json" \
  --export-markdown "$OUT_DIR/hyperfine.md" \
  -n "go-bench" \
  "cd \"$ROOT_DIR\" && go test -bench=$BENCH_PATTERN -benchmem -count=1 ./scenarios/..."
