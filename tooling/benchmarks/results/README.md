# Benchmark result layout

This folder stores local benchmark artifacts. Generated outputs are ignored by default.

## Subdirectories

- `raw/`: direct benchmark command output (`bench-<timestamp>.txt`, `base.txt`, `candidate.txt`, `hyperfine.json`, `hyperfine.md`)
- `compare/`: comparison outputs (`benchstat.txt`, optional markdown summaries)
- `dashboards/`: normalized JSON/markdown for charts or external dashboards

## Typical flow

1. Run benchmarks to create files in `raw/`.
2. Copy two files to `raw/base.txt` and `raw/candidate.txt`.
3. Run compare script to write `compare/benchstat.txt`.
4. Optionally extract JSON summaries into `dashboards/`.

## Commands

```bash
cd tooling/benchmarks
./run.sh
./scripts/compare_benchmarks.sh
python3 ./scripts/extract_go_bench.py results/raw/base.txt results/dashboards/base.json
```

PowerShell:

```powershell
cd tooling/benchmarks
.\run.ps1
.\scripts\compare_benchmarks.ps1
python .\scripts\extract_go_bench.py results/raw/base.txt results/dashboards/base.json
```
