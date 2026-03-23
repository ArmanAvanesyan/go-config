import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class RenderPrCommentTests(unittest.TestCase):
    def test_render_pr_comment_contains_core_fields(self):
        repo_root = Path(__file__).resolve().parents[3]
        tmp_dir = Path(tempfile.mkdtemp(prefix="reports-pr-test-"))
        try:
            summary = {
                "metadata": {"commit": "abc123"},
                "benchmarks": {
                    "benchmark_count": 2,
                    "comparison": {
                        "regressions": [{"benchmark": "BenchmarkA", "unit": "time/op", "delta_percent": 5.0}],
                        "improvements": [{"benchmark": "BenchmarkB", "unit": "time/op", "delta_percent": -6.0}],
                    },
                    "yaml_insights": {
                        "parse_only_yaml": {
                            "vs": {
                                "viper": {"ns_delta_percent": 95.0, "bytes_delta_percent": -90.0, "allocs_delta_percent": -75.0}
                            }
                        }
                    },
                },
                "profiles": {"count": 2},
                "coverage": {"summary_percent": 80.1},
            }
            input_file = tmp_dir / "summary.json"
            output_file = tmp_dir / "pr-comment.md"
            input_file.write_text(json.dumps(summary), encoding="utf-8")
            script = repo_root / "tooling/reports/scripts/render_pr_comment.py"
            subprocess.run(
                [sys.executable, str(script), "--input-file", str(input_file), "--output-file", str(output_file)],
                check=True,
            )
            text = output_file.read_text(encoding="utf-8")
            self.assertIn("## Tooling report", text)
            self.assertIn("BenchmarkA", text)
            self.assertIn("BenchmarkB", text)
            self.assertIn("YAML go-config vs peers", text)
        finally:
            shutil.rmtree(tmp_dir)


if __name__ == "__main__":
    unittest.main()
