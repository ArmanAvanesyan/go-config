#!/usr/bin/env python3
"""
Validate benchmark sub-label naming contract used in b.Run("...").

Contract goals:
- stable Category/Subcategory/... grouping for benchstat
- avoid accidental taxonomy drift (e.g., all vs All, Keys10 vs Keys_10)
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

LABEL_RE = re.compile(r'b\.Run\("([^"]+)"')

# Canonical category prefixes currently used by scenarios.
ALLOWED_CATEGORY_PREFIXES = {
    "Compare",
    "SingleSource",
    "MultiSource",
    "Decode",
    "Scale",
    "ParallelLoad",
    "ParallelReadAfterLoad",
    "ErrorPath",
    "Feature",
}

# Known labels with explicit underscore-number shape.
def extract_labels(root: Path) -> list[tuple[Path, str]]:
    labels: list[tuple[Path, str]] = []
    for p in sorted(root.glob("**/*_bench_test.go")):
        text = p.read_text(encoding="utf-8")
        for m in LABEL_RE.finditer(text):
            labels.append((p, m.group(1)))
    return labels


def main() -> int:
    scripts_dir = Path(__file__).resolve().parent
    bench_dir = scripts_dir.parent
    scenarios_dir = bench_dir / "scenarios"

    labels = extract_labels(scenarios_dir)
    errors: list[str] = []

    for path, label in labels:
        parts = label.split("/")
        if not parts or not parts[0]:
            errors.append(f"{path}: empty label: {label!r}")
            continue

        category = parts[0]
        if category not in ALLOWED_CATEGORY_PREFIXES:
            errors.append(f"{path}: unknown category prefix '{category}' in '{label}'")

        if "all/" in label or "/all/" in label or label.endswith("/all"):
            errors.append(f"{path}: use 'All' casing consistently (found '{label}')")

        if re.search(r"Keys_\d+", label) and label not in REQUIRES_UNDERSCORE_NUM:
            errors.append(f"{path}: unexpected Keys_* underscore variant: '{label}'")

        if re.search(r"Depth\d+", label):
            errors.append(f"{path}: depth labels must use underscore form Depth_<n>: '{label}'")

    # Some labels are generated dynamically in code. Validate those from source patterns.
    scale_file = scenarios_dir / "scale_bench_test.go"
    if scale_file.exists():
        scale_text = scale_file.read_text(encoding="utf-8")
        if '"Scale/"+tc.name' not in scale_text:
            errors.append(f"{scale_file}: missing dynamic key scale label pattern")
        for token in ['name: "Keys10"', 'name: "Keys100"', 'name: "Keys1000"']:
            if token not in scale_text:
                errors.append(f"{scale_file}: missing {token}")
        if '"Scale/Depth_"+itoa(d)' not in scale_text:
            errors.append(f"{scale_file}: missing depth label pattern Scale/Depth_")
        for token in ["[]int{2, 5, 10}", "depths := []int{2, 5, 10}"]:
            if token in scale_text:
                break
        else:
            errors.append(f"{scale_file}: missing required depth set 2,5,10")

    if errors:
        print("benchmark label validation FAILED")
        for e in errors:
            print(f"- {e}")
        return 1

    print(f"benchmark label validation passed ({len(labels)} labels)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
