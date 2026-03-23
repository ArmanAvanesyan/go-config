#!/usr/bin/env python3
import argparse
import json
import re
import subprocess
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

COVER_TOTAL_RE = re.compile(r"total:.*?([0-9]+(?:\.[0-9]+)?)%")
SCRIPT_ROOT = Path(__file__).resolve().parents[2]
import sys
sys.path.append(str(SCRIPT_ROOT / "scripts"))
from bench_parsers import parse_benchstat_rows, parse_go_bench_lines  # noqa: E402


def latest_file(directory: Path, pattern: str) -> Path | None:
    candidates = list(directory.glob(pattern))
    if not candidates:
        return None
    return max(candidates, key=lambda p: p.stat().st_mtime)


def parse_bench_raw(path: Path) -> list[dict[str, Any]]:
    return parse_go_bench_lines(path.read_text(encoding="utf-8", errors="replace"))


def parse_benchstat(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8", errors="replace")
    rows = parse_benchstat_rows(text)
    regressions = [row for row in rows if row["delta_percent"] > 0]
    improvements = [row for row in rows if row["delta_percent"] < 0]
    regressions.sort(key=lambda row: row["delta_percent"], reverse=True)
    improvements.sort(key=lambda row: row["delta_percent"])
    return {
        "source_file": str(path.as_posix()),
        "all": rows,
        "regressions": regressions[:20],
        "improvements": improvements[:20],
    }


def validate_summary_contract(report: dict[str, Any]) -> None:
    required_top = ("metadata", "benchmarks", "profiles", "coverage")
    for key in required_top:
        if key not in report:
            raise ValueError(f"missing top-level field: {key}")
    bench = report["benchmarks"]
    for key in ("source_file", "benchmark_count", "benchmarks", "comparison"):
        if key not in bench:
            raise ValueError(f"missing benchmarks field: {key}")
    cov = report["coverage"]
    for key in ("present", "source_file", "line_count", "mode", "summary_percent"):
        if key not in cov:
            raise ValueError(f"missing coverage field: {key}")
    prof = report["profiles"]
    for key in ("files", "count", "latest_file"):
        if key not in prof:
            raise ValueError(f"missing profiles field: {key}")


def _lib_from_benchmark_name(name: str) -> str | None:
    for lib in ("go-config", "viper", "koanf"):
        if name.endswith(f"/{lib}"):
            return lib
    return None


def _extract_yaml_group(name: str) -> str | None:
    if "Compare/All/YAML/" in name:
        return "all_yaml"
    if "Compare/ParseOnly/YAML/" in name:
        return "parse_only_yaml"
    return None


def build_yaml_insights(raw_rows: list[dict[str, Any]]) -> dict[str, Any]:
    grouped: dict[str, dict[str, dict[str, Any]]] = {
        "all_yaml": {},
        "parse_only_yaml": {},
    }
    for row in raw_rows:
        name = str(row.get("name", ""))
        group = _extract_yaml_group(name)
        if not group:
            continue
        lib = _lib_from_benchmark_name(name)
        if not lib:
            continue
        grouped[group][lib] = row

    comparisons: dict[str, Any] = {}
    for group, rows in grouped.items():
        base = rows.get("go-config")
        if not base:
            continue
        block: dict[str, Any] = {"go_config": base, "vs": {}}
        for peer in ("viper", "koanf"):
            peer_row = rows.get(peer)
            if not peer_row:
                continue
            # Positive delta means go-config is slower/higher.
            ns_delta_pct = ((base["ns_per_op"] - peer_row["ns_per_op"]) / peer_row["ns_per_op"]) * 100.0
            bytes_delta_pct = ((base["bytes_per_op"] - peer_row["bytes_per_op"]) / peer_row["bytes_per_op"]) * 100.0
            allocs_delta_pct = ((base["allocs_per_op"] - peer_row["allocs_per_op"]) / peer_row["allocs_per_op"]) * 100.0
            block["vs"][peer] = {
                "ns_delta_percent": round(ns_delta_pct, 2),
                "bytes_delta_percent": round(bytes_delta_pct, 2),
                "allocs_delta_percent": round(allocs_delta_pct, 2),
            }
        comparisons[group] = block
    return comparisons


def parse_profiles(profile_raw_dir: Path) -> dict[str, Any]:
    files = sorted(
        [str(path.as_posix()) for path in profile_raw_dir.glob("*") if path.is_file() and path.name != ".gitkeep"]
    )
    latest = files[-1] if files else None
    return {"files": files, "count": len(files), "latest_file": latest}


def parse_coverage(path: Path | None) -> dict[str, Any]:
    if path is None or not path.exists():
        return {"present": False, "source_file": None, "line_count": 0, "mode": None, "summary_percent": None}

    lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
    mode = None
    if lines and lines[0].startswith("mode:"):
        mode = lines[0].split(":", 1)[1].strip()
    body_lines = len(lines[1:]) if lines else 0

    summary_percent = None
    try:
        out = subprocess.run(
            ["go", "tool", "cover", "-func", str(path)],
            check=False,
            capture_output=True,
            text=True,
        )
        if out.returncode == 0:
            for line in out.stdout.splitlines():
                if line.startswith("total:"):
                    match = COVER_TOTAL_RE.search(line)
                    if match:
                        summary_percent = float(match.group(1))
                    break
    except OSError:
        summary_percent = None

    return {
        "present": True,
        "source_file": str(path.as_posix()),
        "line_count": body_lines,
        "mode": mode,
        "summary_percent": summary_percent,
    }


def git_metadata(repo_root: Path) -> dict[str, Any]:
    def run(cmd: list[str]) -> str | None:
        try:
            out = subprocess.run(cmd, cwd=str(repo_root), check=False, capture_output=True, text=True)
        except OSError:
            return None
        if out.returncode != 0:
            return None
        return out.stdout.strip() or None

    return {
        "commit": run(["git", "rev-parse", "HEAD"]),
        "branch": run(["git", "rev-parse", "--abbrev-ref", "HEAD"]),
    }


def args_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Build unified benchmark/profile/coverage report")
    parser.add_argument("--repo-root", default=".", help="Repository root path")
    parser.add_argument("--bench-raw-dir", default="tooling/benchmarks/results/raw")
    parser.add_argument("--bench-compare-file", default="tooling/benchmarks/results/compare/benchstat.txt")
    parser.add_argument("--profiles-raw-dir", default="tooling/profiles/raw")
    parser.add_argument("--coverage-file", default="coverage.out")
    parser.add_argument("--output-file", default="tooling/reports/output/summary.json")
    return parser


def main() -> None:
    parser = args_parser()
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    bench_raw_dir = (repo_root / args.bench_raw_dir).resolve()
    bench_compare_file = (repo_root / args.bench_compare_file).resolve()
    profiles_raw_dir = (repo_root / args.profiles_raw_dir).resolve()
    coverage_file = (repo_root / args.coverage_file).resolve()
    output_file = (repo_root / args.output_file).resolve()

    raw_file = latest_file(bench_raw_dir, "bench-*.txt") if bench_raw_dir.exists() else None
    raw_rows = parse_bench_raw(raw_file) if raw_file else []
    bench_summary = {
        "source_file": str(raw_file.as_posix()) if raw_file else None,
        "benchmark_count": len(raw_rows),
        "benchmarks": raw_rows,
        "comparison": parse_benchstat(bench_compare_file) if bench_compare_file.exists() else None,
        "yaml_insights": build_yaml_insights(raw_rows),
    }

    profile_summary = parse_profiles(profiles_raw_dir) if profiles_raw_dir.exists() else {
        "files": [],
        "count": 0,
        "latest_file": None,
    }
    coverage_summary = parse_coverage(coverage_file if coverage_file.exists() else None)

    report = {
        "metadata": {
            "generated_at": datetime.now(tz=timezone.utc).isoformat(),
            "repository": str(repo_root.as_posix()),
            **git_metadata(repo_root),
            "inputs": {
                "bench_raw_dir": str(bench_raw_dir.as_posix()),
                "bench_compare_file": str(bench_compare_file.as_posix()),
                "profiles_raw_dir": str(profiles_raw_dir.as_posix()),
                "coverage_file": str(coverage_file.as_posix()),
            },
        },
        "benchmarks": bench_summary,
        "profiles": profile_summary,
        "coverage": coverage_summary,
    }
    validate_summary_contract(report)

    output_file.parent.mkdir(parents=True, exist_ok=True)
    output_file.write_text(json.dumps(report, indent=2), encoding="utf-8")
    print(f"wrote {output_file}")


if __name__ == "__main__":
    main()
