import re
from typing import Any

BENCH_LINE_RE = re.compile(
    r"^(Benchmark[^\s]+)-\d+\s+(\d+)\s+([0-9.]+)\s+ns/op(?:\s+([0-9.]+)\s+B/op\s+([0-9.]+)\s+allocs/op)?"
)
DELTA_RE = re.compile(r"([+\-]?[0-9]+(?:\.[0-9]+)?)%")
BENCH_NAME_RE = re.compile(r"^(Benchmark\S+)")


def parse_go_bench_lines(text: str) -> list[dict[str, Any]]:
    benchmarks: list[dict[str, Any]] = []
    for raw in text.splitlines():
        line = raw.strip()
        match = BENCH_LINE_RE.match(line)
        if not match:
            continue
        benchmarks.append(
            {
                "name": match.group(1),
                "iterations": int(match.group(2)),
                "ns_per_op": float(match.group(3)),
                "bytes_per_op": float(match.group(4)) if match.group(4) else None,
                "allocs_per_op": float(match.group(5)) if match.group(5) else None,
            }
        )
    return benchmarks


def parse_benchstat_rows(text: str) -> list[dict[str, Any]]:
    current_unit = None
    rows: list[dict[str, Any]] = []
    for raw in text.splitlines():
        line = raw.strip()
        if not line:
            continue
        if line.startswith("name\told time/op\tnew time/op\tdelta"):
            current_unit = "time/op"
            continue
        if line.startswith("name\told B/op\tnew B/op\tdelta"):
            current_unit = "B/op"
            continue
        if line.startswith("name\told allocs/op\tnew allocs/op\tdelta"):
            current_unit = "allocs/op"
            continue
        if line.startswith("name\told alloc/op\tnew alloc/op\tdelta"):
            current_unit = "alloc/op"
            continue
        if line.startswith("geomean"):
            continue
        bench_match = BENCH_NAME_RE.match(line)
        delta_match = DELTA_RE.search(line)
        if bench_match and delta_match and current_unit:
            rows.append(
                {
                    "benchmark": bench_match.group(1),
                    "unit": current_unit,
                    "delta_percent": float(delta_match.group(1)),
                }
            )
    return rows
