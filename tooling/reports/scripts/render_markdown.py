#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def args_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Render full markdown report from summary.json")
    parser.add_argument("--input-file", default="tooling/reports/output/summary.json")
    parser.add_argument("--output-file", default="tooling/reports/output/summary.md")
    return parser


def _fmt_num(value) -> str:
    if value is None:
        return "-"
    if isinstance(value, float):
        return f"{value:.2f}"
    return str(value)


def main() -> None:
    args = args_parser().parse_args()
    input_file = Path(args.input_file).resolve()
    output_file = Path(args.output_file).resolve()

    data = json.loads(input_file.read_text(encoding="utf-8"))
    metadata = data.get("metadata", {})
    bench = data.get("benchmarks", {})
    profiles = data.get("profiles", {})
    coverage = data.get("coverage", {})

    lines = [
        "# Unified report",
        "",
        "## Metadata",
        "",
        f"- Generated at: `{metadata.get('generated_at', 'n/a')}`",
        f"- Commit: `{metadata.get('commit', 'n/a')}`",
        f"- Branch: `{metadata.get('branch', 'n/a')}`",
        "",
        "## Benchmarks",
        "",
        f"- Source file: `{bench.get('source_file') or 'n/a'}`",
        f"- Benchmark rows: `{bench.get('benchmark_count', 0)}`",
        "",
    ]

    benches = bench.get("benchmarks", [])
    if benches:
        lines.extend(
            [
                "### Top benchmarks by time/op",
                "",
                "| Benchmark | ns/op | B/op | allocs/op | iterations |",
                "|---|---:|---:|---:|---:|",
            ]
        )
        top = sorted(benches, key=lambda row: row.get("ns_per_op", 0), reverse=True)[:10]
        for row in top:
            lines.append(
                f"| `{row.get('name', '')}` | {_fmt_num(row.get('ns_per_op'))} | "
                f"{_fmt_num(row.get('bytes_per_op'))} | {_fmt_num(row.get('allocs_per_op'))} | "
                f"{_fmt_num(row.get('iterations'))} |"
            )
        lines.append("")

    comparison = bench.get("comparison")
    if comparison:
        lines.extend(
            [
                "### Benchstat summary",
                "",
                f"- Source file: `{comparison.get('source_file', 'n/a')}`",
                f"- Regressions: `{len(comparison.get('regressions', []))}`",
                f"- Improvements: `{len(comparison.get('improvements', []))}`",
                "",
            ]
        )

    lines.extend(
        [
            "## Profiles",
            "",
            f"- Profile files found: `{profiles.get('count', 0)}`",
            f"- Latest profile file: `{profiles.get('latest_file') or 'n/a'}`",
            "",
            "## Coverage",
            "",
            f"- Present: `{coverage.get('present', False)}`",
            f"- Source file: `{coverage.get('source_file') or 'n/a'}`",
            f"- Profile rows: `{coverage.get('line_count', 0)}`",
            f"- Mode: `{coverage.get('mode') or 'n/a'}`",
            f"- Total line coverage: `{_fmt_num(coverage.get('summary_percent'))}%`",
            "",
        ]
    )

    output_file.parent.mkdir(parents=True, exist_ok=True)
    output_file.write_text("\n".join(lines), encoding="utf-8")
    print(f"wrote {output_file}")


if __name__ == "__main__":
    main()
