#!/usr/bin/env bash
set -euo pipefail

die() {
  echo "error: $*" >&2
  exit 1
}

repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || die "run this script inside the repository"
out_dir="${1:-$repo_root/pages-export}"

reports_src="$repo_root/tooling/reports/output"
profiles_src="$repo_root/tooling/profiles/raw"
hyperfine_src="$repo_root/tooling/benchmarks/results/raw"
reports_index_src="$repo_root/tooling/benchmarks/pages/reports-index.html"
hyperfine_index_src="$repo_root/tooling/benchmarks/pages/hyperfine-index.html"

[[ -f "$reports_src/summary.json" ]] || die "missing $reports_src/summary.json (run report-local first)"
[[ -f "$reports_src/summary.md" ]] || die "missing $reports_src/summary.md (run report-local first)"
[[ -f "$reports_src/pr-comment.md" ]] || die "missing $reports_src/pr-comment.md (run report-pr-local first)"
[[ -d "$profiles_src" ]] || die "missing $profiles_src (run profile capture first if needed)"
[[ -f "$hyperfine_src/hyperfine.json" ]] || die "missing $hyperfine_src/hyperfine.json (run bench-hyperfine-local first)"
[[ -f "$hyperfine_src/hyperfine.md" ]] || die "missing $hyperfine_src/hyperfine.md (run bench-hyperfine-local first)"
[[ -f "$reports_index_src" ]] || die "missing $reports_index_src"
[[ -f "$hyperfine_index_src" ]] || die "missing $hyperfine_index_src"

mkdir -p "$out_dir/reports" "$out_dir/profiles" "$out_dir/hyperfine"

cp "$reports_src/summary.json" \
  "$reports_src/summary.md" \
  "$reports_src/pr-comment.md" \
  "$out_dir/reports/"
cp "$reports_index_src" "$out_dir/reports/index.html"

shopt -s nullglob
for f in "$profiles_src"/*; do
  base="$(basename "$f")"
  [[ "$base" == ".gitkeep" ]] && continue
  [[ -f "$f" ]] || continue
  cp "$f" "$out_dir/profiles/"
done

cp "$hyperfine_src/hyperfine.json" \
  "$hyperfine_src/hyperfine.md" \
  "$out_dir/hyperfine/"
cp "$hyperfine_index_src" "$out_dir/hyperfine/index.html"

echo "staged pages export at: $out_dir"
