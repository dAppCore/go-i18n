# GOAL — bench + coverage + alloc grind (the boring stuff)

**For codex / unattended agents.** This file tells you how to *find* the next
unit of work and *grind* it, with zero human assignment. Snider + Cladius work
the create/improve side; you take the mechanical sweep below to 100%.

Module: `dappco.re/go/i18n` at `go/` (run all `go` commands from `go/`). Pure
Go. Branch `dev`. The **go.work workspace is ACTIVE** — `GOWORK=off` is **not
supported** here; never set it.

## The mission

Every function in this module must have:

1. **A benchmark** — `{file}.go` ships `{file}_bench_test.go` with a
   `Benchmark<Fn>` for *every* function it defines (exported **and**
   unexported). Benchmarks are in-package, so unexported funcs are callable.
2. **≥98% statement coverage** (aim 100%) for `{file}.go`.
3. **Floored allocations** — each function profiled to its real alloc/B-per-op
   floor; every genuine, byte-identical win taken; every *remaining* allocation
   explained (inherent to a returned struct / a dependency / stdlib / map-floor).

…**without changing any return value, function signature, parameter, or struct
definition.** Performance comes from internals only. This is a firm constraint —
a change that alters the public shape is wrong even if it's faster.

## How to find the next unit of work (no human needed)

```bash
make benchcov            # summary: which files have unbenched functions, worst first
make benchcov ARGS=-v    # every unbenched function, grouped by file
```

`make benchcov` cross-references `go tool cover -func` (every non-test function)
against the `Benchmark*` in the suite. **Pick the top file in the priority order
below that `benchcov` still lists as having unbenched functions.** When
`benchcov` reports `0 unbenched`, the bench half is done.

For the coverage half of a file:

```bash
cd go && go test -coverprofile=/tmp/c.out ./<pkg>/ && go tool cover -func=/tmp/c.out | grep <file>.go
```

Any function below 100% is a coverage gap to close.

### Priority order (hot path first)

`GrammarImprint` runs on **any** text — this is no longer just a translation
library, so the reversal/grammar load path is hottest. Grind in this order:

1. `reversal/` — `tokeniser.go`, `multiplier.go`, `reference.go`, `anomaly.go`
   (`imprint.go` is the worked reference — already done).
2. `grammar.go` — the largest coverage gap (`Article`, `DefiniteArticle`,
   `articleFromGrammarForms` and the article/plural helpers are the worst).
3. Root translation-label path — `service.go`, `core_service.go`, `i18n.go`,
   `compose.go`, `context.go`, `handler.go`, `hooks.go`, `language.go`,
   `loader.go`, `localise.go`, `numbers.go`, `state.go`, `time.go`, `types.go`,
   `transform.go`, `debug.go`, `context_map.go`, `default_service.go`.

Three files also miss other companions — add them while you're there:
`context_map.go` + `default_service.go` need `_test.go` **and**
`_example_test.go`; `transform.go` needs `_example_test.go`.

## The per-file recipe (one grind unit)

For the chosen `{file}.go` in package `{pkg}`:

1. **Bench file.** Create `{pkg}/{file}_bench_test.go`. One `Benchmark<Fn>` per
   function in `{file}.go`. `b.ReportAllocs()` + `b.ResetTimer()`; reuse the
   package's setup helper (`benchSetup(b)` in reversal; whatever the root
   package uses). Realistic inputs. **Copy the style of the templates exactly:**
   `go/reversal/imprint_bench_test.go` and `go/reversal/grammardata_bench_test.go`.

2. **0-alloc guards.** For any function whose allocations are (or should be)
   zero, add a suite-level guard with `testing.AllocsPerRun`, modelled on
   `TestImprintZeroAlloc` and `TestGrammarDataLookupZeroAlloc`. The guard runs
   under `go test` (not just `-bench`) and fails on regression. Make it
   **non-vacuous** — assert the input hits the real, populated path, not an
   empty early-return.

3. **Coverage.** Add tests to `{file}_test.go` for every uncovered branch until
   `go tool cover -func` shows `{file}.go` ≥98% (aim 100%).

4. **Floor allocations.** For each function:
   `go test -run='^$' -bench=<B> -benchmem -benchtime=200x -memprofile=/tmp/m.out ./<pkg>/`,
   then `go tool pprof -alloc_objects -list=<fn> /tmp/m.out` — read the **FLAT**
   column (not cum). Read **B/op as hard as allocs/op** (`-alloc_space` for the
   bytes story). Prove escapes with
   `go build -gcflags=-m ./<pkg>/ 2>&1 | grep <file>.go`. Fix each genuine trap.
   Do **not** invent a fix — if the profile bottoms out in an inherent result
   struct / a dependency / stdlib / the 2-alloc map floor, that function is at
   its floor; record it and move on. (Worked example: `NewImprint`'s 10 allocs
   are the six `GrammarImprint` result maps — escaping by design — so it is at
   floor; `verbBases`/`nounBases` are stack-allocated. No win there.)

   Common traps: un-presized slice / `strings.Builder` geometric regrow →
   `make([]T,0,n)` / `b.Grow(n)`; whole-thing clone where a windowed view
   suffices; `%v`/interface boxing in a cold error branch escaping a hot-path
   scratch; a per-call temporary map that can be hoisted or replaced by a
   stack-friendly scan.

5. **Verify + commit.**
   - `cd go && go test ./...` is green and output is **byte-identical**.
   - `make benchcov` shows `{file}.go` with **0 unbenched**.
   - `go tool cover` shows `{file}.go` **≥98%**.
   - Commit to `dev`. Conventional prefixes: `test({pkg}): bench + cover
     {file}.go` for the bench/coverage work; a **separate** `perf({pkg}): reduce
     <fn> allocs N->M/op (AX-11)` commit per real alloc win, with before→after
     numbers in the body. Every commit message ends with exactly:
     `Co-Authored-By: Virgil <virgil@lethean.io>`.

Then re-run `make benchcov`, pick the next file, repeat.

## Hard rules (do not break)

- **No public-shape changes.** No new/changed signatures, return types,
  parameters, or struct fields. Perf via internals only.
- **Byte-identical behaviour.** Never commit a speculative or non-byte-identical
  change. The suite must stay green.
- **Never** set `GOWORK=off`. **Never** `git reset` / `git stash`. **Never**
  push (commit to `dev` only; promotion is handled separately).
- **Never** touch `external/` (vendored submodules) or `tests/` fixtures.
- **Never** flatten the locale JSON under `locales/` — the grammar engine
  depends on the nested `gram.*` structure (see `CLAUDE.md`). This is the #1
  agent-introduced bug.
- UK English in all comments (colour, normalise, behaviour, centre).
- One concern per commit. Don't bundle a perf win with a coverage commit.

## Definition of done

- **Per file:** `make benchcov` = 0 unbenched for it, coverage ≥98%, allocs
  floored (wins committed, remaining allocs explained in the report/commit).
- **Overall:** `make benchcov` reports **100%**, `go tool cover` ≥98% module-wide
  (the `tests/cli` fixture is excluded), `go test ./...` green. The repo then
  carries a self-checking quality ratchet: `make benchcov` + `make cover` fail
  loudly on any new unbenched or uncovered function.

## Progress snapshot (regenerate with `make benchcov`)

At spec time: **16/590 functions benchmarked (2%)**. Worst offenders:
`service.go` 119, `grammar.go` 76, `core_service.go` 62, `tokeniser.go` 59,
`i18n.go` 40, `loader.go` 31. `reversal/imprint.go` is done (the template).
Always trust live `make benchcov` over this line.
