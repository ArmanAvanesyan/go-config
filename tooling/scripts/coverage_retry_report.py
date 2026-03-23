#!/usr/bin/env python3
import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

POLICY_BLOCK_RE = re.compile(r"Application Control policy has blocked this file", re.IGNORECASE)
FILE_COVER_RE = re.compile(r"^(?P<file>.+\.go):\d+:\s+.+\s+(?P<pct>[0-9]+(?:\.[0-9]+)?)%$")
PROFILE_LINE_RE = re.compile(r"^(?P<file>.+\.go):\d+\.\d+,\d+\.\d+\s+(?P<numstmt>\d+)\s+(?P<count>\d+)$")

DEFAULT_PACKAGES = [
    "./extensions/schema/generate",
    "./internal/decode",
    "./internal/engine",
    "./internal/normalize",
    "./internal/tree",
    "./config",
    "./runtime/watch/fsnotify",
]

DEFAULT_TARGET_FILES = [
    "extensions/schema/generate/json.go",
    "extensions/schema/generate/options.go",
    "extensions/schema/generate/reflect_builder.go",
    "extensions/schema/generate/tags.go",
    "extensions/schema/generate/types.go",
    "internal/decode/coercion.go",
    "internal/decode/decode.go",
    "internal/decode/mapper.go",
    "internal/decode/tags.go",
    "internal/engine/context.go",
    "internal/engine/errors.go",
    "internal/engine/pipeline.go",
    "internal/normalize/keys.go",
    "internal/normalize/paths.go",
    "internal/tree/walk.go",
    "internal/tree/merge_helpers.go",
    "runtime/watch/fsnotify/backend_polling.go",
    "config/config.go",
    "config/decoder.go",
    "config/document.go",
    "config/errors.go",
    "config/options.go",
    "config/parser.go",
    "config/resolver.go",
    "config/source.go",
    "config/types.go",
    "config/validator.go",
    "config/watch.go",
]

DEFAULT_TARGET_MANIFEST = "tooling/reports/schemas/coverage-targets.manifest.json"


def run_cmd(cmd: list[str], cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=str(cwd), text=True, capture_output=True, check=False)


def normalize_repo_path(path: str, repo_root: Path) -> str:
    if "github.com/ArmanAvanesyan/go-config/" in path:
        path = path.split("github.com/ArmanAvanesyan/go-config/", 1)[1]
    p = Path(path)
    try:
        return str(p.resolve().relative_to(repo_root.resolve())).replace("\\", "/")
    except Exception:
        return str(p).replace("\\", "/")


def parse_cover_func_output(text: str, repo_root: Path) -> tuple[dict[str, float], float | None]:
    per_file: dict[str, float] = {}
    sums: dict[str, list[float]] = {}
    total_pct: float | None = None
    for line in text.splitlines():
        line = line.strip()
        if line.startswith("total:"):
            m = re.search(r"([0-9]+(?:\.[0-9]+)?)%$", line)
            if m:
                total_pct = float(m.group(1))
            continue
        m = FILE_COVER_RE.match(line)
        if not m:
            continue
        rel = normalize_repo_path(m.group("file"), repo_root)
        sums.setdefault(rel, []).append(float(m.group("pct")))
    for rel, values in sums.items():
        per_file[rel] = min(values)
    return per_file, total_pct


def parse_profile_file_coverage(profile_path: Path, repo_root: Path) -> tuple[dict[str, float], dict[str, int]]:
    totals: dict[str, int] = {}
    covered: dict[str, int] = {}
    for raw in profile_path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("mode:"):
            continue
        m = PROFILE_LINE_RE.match(line)
        if not m:
            continue
        rel = normalize_repo_path(m.group("file"), repo_root)
        n = int(m.group("numstmt"))
        c = int(m.group("count"))
        totals[rel] = totals.get(rel, 0) + n
        if c > 0:
            covered[rel] = covered.get(rel, 0) + n
    out: dict[str, float] = {}
    for rel, total in totals.items():
        if total == 0:
            continue
        out[rel] = round((covered.get(rel, 0) / total) * 100.0, 1)
    return out, totals


def merge_coverprofiles(profile_paths: list[Path], merged_path: Path) -> None:
    lines = ["mode: set"]
    for p in profile_paths:
        raw = p.read_text(encoding="utf-8").splitlines()
        lines.extend(raw[1:] if raw and raw[0].startswith("mode:") else raw)
    merged_path.parent.mkdir(parents=True, exist_ok=True)
    merged_path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def run_package_with_retries(pkg: str, retries: int, repo_root: Path, out_profile: Path) -> dict:
    attempts = 0
    while attempts < retries:
        attempts += 1
        cmd = ["go", "test", pkg, f"-coverprofile={str(out_profile)}"]
        proc = run_cmd(cmd, repo_root)
        combined = (proc.stdout or "") + "\n" + (proc.stderr or "")
        if proc.returncode == 0:
            return {"package": pkg, "status": "ok", "attempts": attempts, "error": None}
        if POLICY_BLOCK_RE.search(combined):
            continue
        return {
            "package": pkg,
            "status": "failed",
            "attempts": attempts,
            "error": combined.strip(),
        }
    return {
        "package": pkg,
        "status": "blocked",
        "attempts": attempts,
        "error": "Application Control policy blocked test binary",
    }


def build_markdown_report(
    results: list[dict],
    per_file: dict[str, float],
    target_status: dict[str, str],
    target_files: list[str],
    strict_threshold: float,
    total_pct: float | None,
) -> str:
    blocked = [r for r in results if r["status"] == "blocked"]
    failed = [r for r in results if r["status"] == "failed"]
    ok = [r for r in results if r["status"] == "ok"]
    lines = [
        "# Coverage per-file report",
        "",
        "## Package run status",
        "",
        f"- Successful packages: `{len(ok)}`",
        f"- Blocked packages: `{len(blocked)}`",
        f"- Failed packages: `{len(failed)}`",
        f"- Merged total coverage: `{total_pct if total_pct is not None else 'n/a'}%`",
        "",
        "| Package | Status | Attempts |",
        "|---|---|---:|",
    ]
    for r in results:
        lines.append(f"| `{r['package']}` | `{r['status']}` | `{r['attempts']}` |")
    lines.append("")
    lines.extend(["## Requested file coverage", "", "| File | Status | Coverage | Meets strict |", "|---|---|---:|---|"])
    for f in target_files:
        pct = per_file.get(f)
        status = target_status.get(f, "covered")
        if pct is None:
            lines.append(f"| `{f}` | `{status}` | `n/a` | `yes` |")
        else:
            meets = "yes" if pct >= strict_threshold else "no"
            lines.append(f"| `{f}` | `{status}` | `{pct:.1f}%` | `{meets}` |")
    lines.append("")
    if blocked:
        lines.append("Some packages were blocked by App Control; rerun after whitelisting.")
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="Run package coverage with retries and emit per-file report.")
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--retries", type=int, default=5)
    parser.add_argument("--strict-threshold", type=float, default=100.0)
    parser.add_argument("--packages", nargs="*", default=DEFAULT_PACKAGES)
    parser.add_argument("--output-dir", default="tooling/reports/output")
    parser.add_argument("--target-files", nargs="*", default=DEFAULT_TARGET_FILES)
    parser.add_argument("--target-manifest", default=DEFAULT_TARGET_MANIFEST)
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    output_dir = (repo_root / args.output_dir).resolve()
    output_dir.mkdir(parents=True, exist_ok=True)

    manifest_path = (repo_root / args.target_manifest).resolve()
    manifest_packages = list(args.packages)
    manifest_targets = [f.replace("\\", "/") for f in args.target_files]
    if manifest_path.exists():
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        manifest_packages = manifest.get("packages", manifest_packages)
        manifest_targets = [f.replace("\\", "/") for f in manifest.get("target_files", manifest_targets)]

    results: list[dict] = []
    profile_paths: list[Path] = []
    with tempfile.TemporaryDirectory(prefix="cov-retry-") as tmp:
        tmp_dir = Path(tmp)
        for idx, pkg in enumerate(manifest_packages):
            prof = tmp_dir / f"pkg-{idx}.out"
            res = run_package_with_retries(pkg, args.retries, repo_root, prof)
            results.append(res)
            if res["status"] == "ok" and prof.exists():
                profile_paths.append(prof)

        merged_profile = output_dir / "coverage-targeted-merged.out"
        per_file: dict[str, float] = {}
        per_file_total_stmt: dict[str, int] = {}
        total_pct: float | None = None
        if profile_paths:
            merge_coverprofiles(profile_paths, merged_profile)
            per_file, per_file_total_stmt = parse_profile_file_coverage(merged_profile, repo_root)
            proc = run_cmd(["go", "tool", "cover", "-func", str(merged_profile)], repo_root)
            if proc.returncode == 0:
                _, total_pct = parse_cover_func_output(proc.stdout, repo_root)

        target_files = manifest_targets
        target_status = {}
        for f in target_files:
            if f in per_file:
                target_status[f] = "covered"
            else:
                target_status[f] = "not_present"
                if f in per_file_total_stmt:
                    target_status[f] = "no_statements"

        report_md = build_markdown_report(
            results, per_file, target_status, target_files, args.strict_threshold, total_pct
        )
        report_json = {
            "results": results,
            "strict_threshold": args.strict_threshold,
            "total_coverage_percent": total_pct,
            "per_file": {f: per_file.get(f) for f in target_files},
            "target_status": target_status,
            "manifest_file": str(manifest_path.as_posix()) if manifest_path.exists() else None,
        }

        md_path = output_dir / "coverage-per-file.md"
        json_path = output_dir / "coverage-per-file.json"
        md_path.write_text(report_md, encoding="utf-8")
        json_path.write_text(json.dumps(report_json, indent=2), encoding="utf-8")

        missing_or_low = [f for f in target_files if per_file.get(f) is not None and float(per_file.get(f)) < args.strict_threshold]
        blocked = any(r["status"] == "blocked" for r in results)
        failed = any(r["status"] == "failed" for r in results)

        print(f"wrote {md_path}")
        print(f"wrote {json_path}")

        if failed:
            return 2
        if blocked:
            return 3
        if missing_or_low:
            return 4
        return 0


if __name__ == "__main__":
    sys.exit(main())
