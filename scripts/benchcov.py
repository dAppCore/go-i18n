#!/usr/bin/env python3
"""benchcov — cross-reference code coverage with benchmark presence to surface
functions that have no benchmark.

Coverage (`go tool cover -func`) enumerates every non-test function in the
module; this script matches each against the `Benchmark*` functions in the test
files and reports the ones with no benchmark. It is the benchmark-side analogue
of statement coverage: `make cover` shows untested lines, `make benchcov` shows
unbenchmarked functions.

Match rule (lenient, underscore/camel boundary aware): a function `Foo` counts
as benchmarked if a `BenchmarkFoo`, `Benchmark..._Foo`, or `BenchmarkFoo_...`
exists. Run from the module root (the dir holding go.mod).

Usage: python3 scripts/benchcov.py [-v]   (-v lists every unbenched function)
"""
import subprocess, sys, collections, os, re

PROF = "/tmp/benchcov-cov.out"
BENCH_RE = re.compile(r"func Benchmark([A-Za-z0-9_]+)")


def coverage_funcs():
    """Yield (pkg, file, funcname) for every non-test, non-fixture function."""
    subprocess.run(
        ["go", "test", f"-coverprofile={PROF}", "-covermode=atomic", "./..."],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
    )
    if not os.path.exists(PROF):
        print("benchcov: no coverage profile produced", file=sys.stderr)
        sys.exit(1)
    out = subprocess.check_output(["go", "tool", "cover", f"-func={PROF}"], text=True)
    for ln in out.splitlines():
        if ln.startswith("total") or not ln.strip():
            continue
        loc, fn = ln.split()[0], ln.split()[1]
        if "/tests/" in loc:  # skip CLI / test fixtures
            continue
        pkg = loc.rsplit("/", 1)[0].split("/")[-1] if "/" in loc else "."
        yield pkg, loc.split(":")[0].split("/")[-1], fn


def benchmark_bases():
    """Lowercased Benchmark* base names across the module's _test.go files."""
    bases = []
    for root, dirs, files in os.walk("."):
        dirs[:] = [d for d in dirs if d not in ("external", ".git")]
        for name in files:
            if name.endswith("_test.go"):
                with open(os.path.join(root, name), encoding="utf-8") as fh:
                    bases += [m.group(1).lower() for m in BENCH_RE.finditer(fh.read())]
    return bases


def main():
    # The Go module lives in ./go relative to the repo root; run there so
    # `./...` resolves to the i18n module (not the external/ submodules).
    os.chdir(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go"))
    verbose = "-v" in sys.argv
    bl = benchmark_bases()

    def benched(fn):
        f = fn.lower()
        return any(b == f or b.endswith("_" + f) or b.startswith(f + "_") for b in bl)

    per = collections.defaultdict(lambda: [0, 0])
    unb_by_file = collections.defaultdict(list)
    for pkg, fname, fn in coverage_funcs():
        per[pkg][0] += 1
        if benched(fn):
            per[pkg][1] += 1
        else:
            unb_by_file[(pkg, fname)].append(fn)

    tot = sum(v[0] for v in per.values())
    bn = sum(v[1] for v in per.values())
    pct = (100 * bn // tot) if tot else 100
    print(f"=== BENCH COVERAGE: {bn}/{tot} functions benchmarked ({pct}%) ===")
    for pkg, (t, b) in sorted(per.items()):
        p = (100 * b // t) if t else 100
        print(f"  {pkg:12} {b:4}/{t:<4} ({p:3}%)   unbenched: {t - b}")

    print("\n=== files by unbenched function count ===")
    for (pkg, fname), fns in sorted(unb_by_file.items(), key=lambda kv: -len(kv[1])):
        if not fns:
            continue
        print(f"  {len(fns):4}  {pkg}/{fname}")
        if verbose:
            for fn in fns:
                print(f"          {fn}")


if __name__ == "__main__":
    main()
