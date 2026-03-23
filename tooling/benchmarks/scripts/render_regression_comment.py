#!/usr/bin/env python3
import argparse
import sys
from pathlib import Path

SCRIPT_ROOT = Path(__file__).resolve().parents[2]
sys.path.append(str(SCRIPT_ROOT / "scripts"))
from bench_parsers import parse_benchstat_rows  # noqa: E402

NS_REGRESSION_THRESHOLD = 10.0
ALLOC_REGRESSION_THRESHOLD = 10.0
BYTE_REGRESSION_THRESHOLD = 10.0


def args_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Render benchmark regression comment markdown")
    parser.add_argument("benchstat_file", help="benchstat.txt input path")
    parser.add_argument("output_file", help="comment markdown output path")
    return parser


def main():
    args = args_parser().parse_args()
    benchstat_path = Path(args.benchstat_file)
    comment_path = Path(args.output_file)
    try:
        text = benchstat_path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        text = benchstat_path.read_text(encoding="utf-16")
    findings = parse_benchstat_rows(text)

    regressions = []
    for f in findings:
        d = f["delta_percent"]
        if f["unit"] == "time/op" and d > NS_REGRESSION_THRESHOLD:
            regressions.append(f)
        elif f["unit"] in ("alloc/op", "allocs/op") and d > ALLOC_REGRESSION_THRESHOLD:
            regressions.append(f)
        elif f["unit"] == "B/op" and d > BYTE_REGRESSION_THRESHOLD:
            regressions.append(f)

    lines = [
        "## Benchmark comparison",
        "",
        "```text",
        text.strip(),
        "```",
        "",
    ]
    if regressions:
        lines.append("### Regressions detected")
        lines.append("")
        for r in regressions:
            lines.append(f"- `{r['benchmark']}` `{r['unit']}` regressed by `{r['delta_percent']:.2f}%`")
        comment_path.write_text("\n".join(lines), encoding="utf-8")
        print("Regression detected.")
        sys.exit(2)

    lines.append("### No regressions detected")
    lines.append("")
    lines.append("All monitored deltas stayed within thresholds.")
    comment_path.write_text("\n".join(lines), encoding="utf-8")
    print("No regression detected.")


if __name__ == "__main__":
    main()
