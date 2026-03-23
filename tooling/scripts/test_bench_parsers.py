import unittest
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent))
from bench_parsers import parse_benchstat_rows, parse_go_bench_lines


class BenchParsersTests(unittest.TestCase):
    def test_parse_go_bench_lines(self):
        text = (
            "BenchmarkFoo-8  1000  123.4 ns/op  64 B/op  2 allocs/op\n"
            "BenchmarkBar-8  2000  12.0 ns/op\n"
        )
        rows = parse_go_bench_lines(text)
        self.assertEqual(len(rows), 2)
        self.assertEqual(rows[0]["name"], "BenchmarkFoo")
        self.assertEqual(rows[0]["iterations"], 1000)
        self.assertEqual(rows[1]["bytes_per_op"], None)

    def test_parse_benchstat_rows(self):
        text = (
            "name\told time/op\tnew time/op\tdelta\n"
            "BenchmarkFoo-8\t100ns\t110ns\t+10.0%\n"
            "name\told B/op\tnew B/op\tdelta\n"
            "BenchmarkFoo-8\t10\t9\t-10.0%\n"
        )
        rows = parse_benchstat_rows(text)
        self.assertEqual(len(rows), 2)
        self.assertEqual(rows[0]["unit"], "time/op")
        self.assertEqual(rows[1]["unit"], "B/op")


if __name__ == "__main__":
    unittest.main()
