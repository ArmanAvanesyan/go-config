#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

stamp="$(date +%Y%m%d%H%M%S)"
base="results/raw/yaml-profile-${stamp}"
mkdir -p results/raw
profile_mode="${PROFILE_MODE:-warm}"

echo "=== YAML PROFILE MODE: ${profile_mode} ==="
go test ./scenarios -run '^$' -bench 'BenchmarkCompare_YAML_ParseOnly/Compare/ParseOnly/YAML/go-config' -benchmem -count=1 -cpuprofile "${base}-cpu.out" -memprofile "${base}-mem.out"
go tool pprof -top "${base}-cpu.out" | tee "${base}-cpu-top.txt"
go tool pprof -top -alloc_space "${base}-mem.out" | tee "${base}-mem-top.txt"

echo "Wrote profiles:"
echo "${base}-cpu.out"
echo "${base}-mem.out"
echo "Profile mode: ${profile_mode}"
