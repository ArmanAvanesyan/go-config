# Unified reports tooling

This directory builds normalized reports from local and CI artifacts produced by:

- `tooling/benchmarks/results/*`
- `tooling/profiles/raw/*`
- `coverage.out` (when present)

The goal is to provide one machine-readable summary plus human-readable markdown outputs.

## Structure

- `scripts/` report builders and renderers
- `schemas/` JSON schema contracts for report sections
- `schemas/summary.schema.json` canonical top-level report contract
- `schemas/coverage-targets.manifest.json` versioned coverage target manifest
- `testdata/` fixture inputs for script tests
- `output/` generated artifacts (ignored in git except `.gitkeep`)

## Outputs

Generated under `tooling/reports/output/`:

- `summary.json` normalized report payload
- `summary.md` full markdown report
- `pr-comment.md` compact PR/CI markdown

## Usage

From repository root:

```bash
python3 tooling/reports/scripts/build_report.py
python3 tooling/reports/scripts/render_markdown.py
python3 tooling/reports/scripts/render_pr_comment.py
```

Each command supports `--help` for custom input/output paths.

## Contract notes

- `build_report.py` enforces required top-level and section fields before writing output.
- `summary.json` is designed to validate against `schemas/summary.schema.json`.
- Coverage strict-file reporting uses `schemas/coverage-targets.manifest.json` by default, with optional CLI override.
