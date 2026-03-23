import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class RenderMarkdownTests(unittest.TestCase):
    def test_render_outputs_stable_sections(self):
        repo_root = Path(__file__).resolve().parents[3]
        tmp_dir = Path(tempfile.mkdtemp(prefix="reports-render-test-"))
        try:
            summary = {
                "metadata": {"generated_at": "2026-03-23T00:00:00Z", "commit": "abc", "branch": "main"},
                "benchmarks": {
                    "source_file": "bench.txt",
                    "benchmark_count": 1,
                    "benchmarks": [
                        {
                            "name": "BenchmarkX",
                            "iterations": 100,
                            "ns_per_op": 10.0,
                            "bytes_per_op": 12.0,
                            "allocs_per_op": 2.0,
                        }
                    ],
                    "comparison": {"source_file": "benchstat.txt", "regressions": [], "improvements": []},
                },
                "profiles": {"count": 1, "latest_file": "profile.txt", "files": ["profile.txt"]},
                "coverage": {
                    "present": True,
                    "source_file": "coverage.out",
                    "line_count": 2,
                    "mode": "set",
                    "summary_percent": 75.5,
                },
            }
            input_file = tmp_dir / "summary.json"
            output_file = tmp_dir / "summary.md"
            input_file.write_text(json.dumps(summary), encoding="utf-8")

            script = repo_root / "tooling/reports/scripts/render_markdown.py"
            subprocess.run(
                [sys.executable, str(script), "--input-file", str(input_file), "--output-file", str(output_file)],
                check=True,
            )
            text = output_file.read_text(encoding="utf-8")
            self.assertIn("# Unified report", text)
            self.assertIn("## Benchmarks", text)
            self.assertIn("## Profiles", text)
            self.assertIn("## Coverage", text)
            self.assertIn("`BenchmarkX`", text)
        finally:
            shutil.rmtree(tmp_dir)


if __name__ == "__main__":
    unittest.main()
