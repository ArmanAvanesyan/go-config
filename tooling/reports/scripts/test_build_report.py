import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class BuildReportTests(unittest.TestCase):
    def test_build_report_outputs_expected_shape(self):
        repo_root = Path(__file__).resolve().parents[3]
        tmp_dir = Path(tempfile.mkdtemp(prefix="reports-test-"))
        try:
            bench_raw = tmp_dir / "bench-raw"
            bench_raw.mkdir(parents=True)
            shutil.copy(repo_root / "tooling/reports/testdata/bench-sample.txt", bench_raw / "bench-20260101.txt")

            bench_compare = tmp_dir / "bench-compare"
            bench_compare.mkdir(parents=True)
            shutil.copy(
                repo_root / "tooling/reports/testdata/benchstat-sample.txt",
                bench_compare / "benchstat.txt",
            )

            profiles_raw = tmp_dir / "profiles-raw"
            profiles_raw.mkdir(parents=True)
            shutil.copy(repo_root / "tooling/reports/testdata/profile-sample.txt", profiles_raw / "profile-1.txt")

            coverage_file = tmp_dir / "coverage.out"
            shutil.copy(repo_root / "tooling/reports/testdata/coverage-sample.out", coverage_file)

            output_file = tmp_dir / "summary.json"
            script = repo_root / "tooling/reports/scripts/build_report.py"
            subprocess.run(
                [
                    sys.executable,
                    str(script),
                    "--repo-root",
                    str(repo_root),
                    "--bench-raw-dir",
                    str(bench_raw),
                    "--bench-compare-file",
                    str(bench_compare / "benchstat.txt"),
                    "--profiles-raw-dir",
                    str(profiles_raw),
                    "--coverage-file",
                    str(coverage_file),
                    "--output-file",
                    str(output_file),
                ],
                check=True,
            )

            report = json.loads(output_file.read_text(encoding="utf-8"))
            self.assertIn("metadata", report)
            self.assertIn("benchmarks", report)
            self.assertIn("profiles", report)
            self.assertIn("coverage", report)
            self.assertEqual(report["benchmarks"]["benchmark_count"], 2)
            self.assertEqual(report["profiles"]["count"], 1)
            self.assertTrue(report["coverage"]["present"])
            self.assertIsNotNone(report["benchmarks"]["comparison"])
            self.assertIn("yaml_insights", report["benchmarks"])
        finally:
            shutil.rmtree(tmp_dir)

    def test_schemas_are_valid_json_files(self):
        repo_root = Path(__file__).resolve().parents[3]
        schema_dir = repo_root / "tooling/reports/schemas"
        for path in schema_dir.glob("*.schema.json"):
            parsed = json.loads(path.read_text(encoding="utf-8"))
            self.assertIn("title", parsed)
            self.assertEqual(parsed.get("type"), "object")


if __name__ == "__main__":
    unittest.main()
