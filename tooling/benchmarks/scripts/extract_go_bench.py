#!/usr/bin/env python3
import argparse
import json
import sys
from pathlib import Path

SCRIPT_ROOT = Path(__file__).resolve().parents[2]
sys.path.append(str(SCRIPT_ROOT / "scripts"))
from bench_parsers import parse_go_bench_lines  # noqa: E402


def args_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Extract Go benchmark rows into JSON")
    parser.add_argument("input_file", help="Raw go test -bench output")
    parser.add_argument("output_file", help="Destination JSON path")
    return parser


def main():
    args = args_parser().parse_args()
    src = Path(args.input_file)
    dst = Path(args.output_file)
    dst.parent.mkdir(parents=True, exist_ok=True)

    rows = parse_go_bench_lines(src.read_text(encoding="utf-8"))
    dst.write_text(json.dumps({"benchmarks": rows}, indent=2), encoding="utf-8")
    print(f"wrote {dst}")


if __name__ == "__main__":
    main()
