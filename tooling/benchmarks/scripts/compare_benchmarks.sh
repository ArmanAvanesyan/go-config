#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
BASE_FILE="${BASE_FILE:-$ROOT_DIR/results/raw/base.txt}"
NEW_FILE="${NEW_FILE:-$ROOT_DIR/results/raw/candidate.txt}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/results/compare}"
OUT_FILE="${OUT_FILE:-$OUT_DIR/benchstat.txt}"

mkdir -p "$OUT_DIR"

benchstat "$BASE_FILE" "$NEW_FILE" | tee "$OUT_FILE"
