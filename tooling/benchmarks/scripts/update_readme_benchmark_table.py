#!/usr/bin/env python3
import argparse
import json
from pathlib import Path

START_MARKER = "<!-- BENCHMARK_TABLE:START -->"
END_MARKER = "<!-- BENCHMARK_TABLE:END -->"


def _format_ns(value: float | int | None) -> str:
    if value is None:
        return "-"
    if float(value).is_integer():
        return str(int(value))
    return f"{float(value):.1f}"


def _build_lookup(rows: list[dict]) -> dict[tuple[str, str], float]:
    lookup: dict[tuple[str, str], float] = {}
    for row in rows:
        name = str(row.get("name", ""))
        ns = row.get("ns_per_op")
        if ns is None:
            continue
        parts = name.split("/")
        if len(parts) < 2:
            continue
        lib = parts[-1]
        key = "/".join(parts[-4:-1]) if len(parts) >= 4 else ""
        if key:
            lookup[(key, lib)] = float(ns)
    return lookup


def _row(point: str, lookup: dict[tuple[str, str], float]) -> str:
    gc = lookup.get((point, "go-config"))
    vp = lookup.get((point, "viper"))
    kf = lookup.get((point, "koanf"))

    def cell(peer: float | None) -> str:
        if gc is None:
            return "-"
        if peer is None:
            return "-"
        # Time ratio relative to go-config:
        # 1.00x = same, >1.00x = slower than go-config, <1.00x = faster.
        mult = peer / gc if gc else 0.0
        return f"`{mult:.2f}x` ({_format_ns(peer)})"

    gc_cell = "-" if gc is None else f"`1.00x` ({_format_ns(gc)})"
    return f"| `{point}` | {gc_cell} | {cell(vp)} | {cell(kf)} |"


def build_section(summary: dict) -> str:
    rows = summary.get("benchmarks", {}).get("benchmarks", [])
    lookup = _build_lookup(rows)
    lines = [
        START_MARKER,
        "Representative comparison snapshot (auto-generated from `tooling/reports/output/summary.json`; lower is better for `ns/op`):",
        "",
        "| Benchmark point (`ns/op`) | go-config | Viper | Koanf |",
        "| ------ | ------: | ------: | ------: |",
        _row("Compare/All/JSON", lookup),
        _row("Compare/All/YAML", lookup),
        _row("Compare/ParseOnly/YAML", lookup),
        "",
        "`go-config` is normalized to `1.00x`; peer values are time ratios (`peer_ns/op / go-config_ns/op`) for the same benchmark point, so values above `1.00x` indicate the peer is slower and values below `1.00x` indicate the peer is faster. For statistical comparisons across runs, use the `benchstat` workflow in [`tooling/benchmarks/README.md`](tooling/benchmarks/README.md).",
        END_MARKER,
    ]
    return "\n".join(lines)


def update_readme(readme_path: Path, section: str) -> None:
    text = readme_path.read_text(encoding="utf-8")
    start = text.find(START_MARKER)
    end = text.find(END_MARKER)
    if start == -1 or end == -1 or end < start:
        raise ValueError(f"README markers not found in {readme_path}")
    end += len(END_MARKER)
    updated = text[:start] + section + text[end:]
    readme_path.write_text(updated, encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="Update README benchmark table from summary.json")
    parser.add_argument("--summary-file", default="tooling/reports/output/summary.json")
    parser.add_argument("--readme-file", default="README.md")
    args = parser.parse_args()

    summary_file = Path(args.summary_file).resolve()
    readme_file = Path(args.readme_file).resolve()
    summary = json.loads(summary_file.read_text(encoding="utf-8"))
    section = build_section(summary)
    update_readme(readme_file, section)
    print(f"updated {readme_file}")


if __name__ == "__main__":
    main()
