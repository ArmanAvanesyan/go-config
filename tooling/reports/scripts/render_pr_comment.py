#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def args_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Render compact PR comment markdown from summary.json")
    parser.add_argument("--input-file", default="tooling/reports/output/summary.json")
    parser.add_argument("--output-file", default="tooling/reports/output/pr-comment.md")
    return parser


def _format_delta(row: dict) -> str:
    delta = row.get("delta_percent")
    if delta is None:
        return "n/a"
    return f"{delta:+.2f}%"


def _fmt_pct(value: float | None) -> str:
    if value is None:
        return "n/a"
    return f"{value:+.2f}%"


def main() -> None:
    args = args_parser().parse_args()
    input_file = Path(args.input_file).resolve()
    output_file = Path(args.output_file).resolve()

    data = json.loads(input_file.read_text(encoding="utf-8"))
    metadata = data.get("metadata", {})
    bench = data.get("benchmarks", {})
    comparison = bench.get("comparison") or {}
    yaml_insights = bench.get("yaml_insights") or {}
    coverage = data.get("coverage", {})
    profiles = data.get("profiles", {})

    regressions = comparison.get("regressions", [])[:5]
    improvements = comparison.get("improvements", [])[:5]

    lines = [
        "## Tooling report",
        "",
        f"- Commit: `{metadata.get('commit', 'n/a')}`",
        f"- Bench rows: `{bench.get('benchmark_count', 0)}`",
        f"- Profile files: `{profiles.get('count', 0)}`",
        f"- Coverage: `{coverage.get('summary_percent') if coverage.get('summary_percent') is not None else 'n/a'}%`",
        "",
    ]

    if regressions:
        lines.extend(["### Top regressions", ""])
        for row in regressions:
            lines.append(f"- `{row.get('benchmark', 'unknown')}` `{row.get('unit', 'n/a')}` `{_format_delta(row)}`")
        lines.append("")

    if improvements:
        lines.extend(["### Top improvements", ""])
        for row in improvements:
            lines.append(f"- `{row.get('benchmark', 'unknown')}` `{row.get('unit', 'n/a')}` `{_format_delta(row)}`")
        lines.append("")

    if not regressions and not improvements:
        lines.extend(["No benchstat comparison data was found.", ""])

    if yaml_insights:
        lines.extend(["### YAML go-config vs peers", ""])
        for group, label in (
            ("all_yaml", "All YAML (load+decode)"),
            ("parse_only_yaml", "Parse-only YAML"),
        ):
            group_data = yaml_insights.get(group) or {}
            vs = group_data.get("vs") or {}
            if not vs:
                continue
            lines.append(f"- **{label}**")
            for peer in ("viper", "koanf"):
                row = vs.get(peer)
                if not row:
                    continue
                lines.append(
                    f"  - vs `{peer}`: ns/op `{_fmt_pct(row.get('ns_delta_percent'))}`, "
                    f"B/op `{_fmt_pct(row.get('bytes_delta_percent'))}`, "
                    f"allocs/op `{_fmt_pct(row.get('allocs_delta_percent'))}`"
                )
        lines.append("")

    output_file.parent.mkdir(parents=True, exist_ok=True)
    output_file.write_text("\n".join(lines), encoding="utf-8")
    print(f"wrote {output_file}")


if __name__ == "__main__":
    main()
