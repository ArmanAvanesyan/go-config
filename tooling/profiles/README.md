# Profiling artifacts

This directory stores local profiling artifacts used during development.

- `raw/`: generated profiler outputs (`.pprof`, textual dumps, captures).
- `reports/`: derived reports (summaries, rendered outputs).

Policy:

- Treat this tree as development tooling only, not library/runtime code.
- Keep stable placeholders (`.gitkeep`, this README) tracked.
- Ignore generated artifacts via root `.gitignore`.
- Use `tooling/reports/scripts/build_report.py` to include profile captures in the unified report payload.
