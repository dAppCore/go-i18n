# TODO.md — go-i18n Task Queue

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 1: Harden the Engine

- [x] **Add CLAUDE.md** — Document the grammar engine contract: what it is (grammar primitives + reversal), what it isn't (translation file manager). Include build/test commands, the gram.* sacred rule, and the agent-flattening prohibition. *(d5b3eac)*
- [x] **Ambiguity resolution for dual-class words** — Two-pass probabilistic disambiguation with 7 weighted signals. All 6 dual-class words {commit, run, test, check, file, build} correctly disambiguate in both verb and noun contexts. Confidence scores flow into imprints. *(3848297)*
- [x] **Extend irregular verb coverage** — Added 44 irregular verbs: 17 compound (undo, redo, rerun, rewrite, rebuild, resend, override, rethink, remake, undergo, overcome, withdraw, uphold, withhold, outgrow, outrun, overshoot), 22 simple (become, come, give, fall, understand, arise, bind, spin, quit, cast, broadcast, burst, cost, shed, rid, shrink, shoot, forbid, offset, upset, input, output), 5 CVC doubling overrides (debug, embed, unzip, remap, unpin, unwrap).
- [x] **Add benchmarks** — 8 forward composition + 7 reversal benchmarks. Baseline on M3 Ultra: PastTense 26ns/0alloc, Tokenise(short) 639ns/8alloc, Imprint 648ns/10alloc, Similar 516ns/0alloc.

## Phase 2: Reference Distribution + 1B Classification Pipeline

### 2a: 1B Pre-Classification — UNBLOCKED (19 Feb 2026)

go-mlx Phases 2-5 are complete. Gemma3-1B inference validated at 46 tok/s, batch classify at 152 prompts/sec on M3 Ultra. Import go-inference + go-mlx directly — no go-ai needed.

**Setup**: Add to go.mod:
```
require forge.lthn.ai/core/go-inference v0.0.0
require forge.lthn.ai/core/go-mlx v0.0.0

replace forge.lthn.ai/core/go-inference => ../go-inference
replace forge.lthn.ai/core/go-mlx => ../go-mlx
```

**Usage**:
```go
import (
    "forge.lthn.ai/core/go-inference"
    _ "forge.lthn.ai/core/go-mlx"  // registers "metal" backend via init()
)

m, err := inference.LoadModel("/Volumes/Data/lem/LEM-Gemma3-1B-layered-v2")
defer m.Close()

// Option A: Batch classify (prefill-only, 152 prompts/sec) — best for domain sorting
results, err := m.Classify(ctx, prompts, inference.WithMaxTokens(1))

// Option B: Single-token generation (46 tok/s) — for article/irregular validation
for tok := range m.Generate(ctx, prompt, inference.WithMaxTokens(1), inference.WithTemperature(0.05)) {
    fmt.Print(tok.Text)
}

// Model discovery (finds all models under a directory)
models, _ := inference.Discover("/Volumes/Data/lem/")
```

**Key types** (all from `inference` package): `Token`, `Message`, `TextModel`, `Backend`, `GenerateConfig`, `ClassifyResult`, `BatchResult`, `GenerateMetrics`.

---

- [x] **Classification benchmark suite** — 220 domain-tagged sentences, leave-one-out classification via imprint similarity. Grammar engine: technical 78%, creative 82%, ethical 46%, casual 11%. Ethical↔technical and casual↔creative confusion confirms 1B model needed for those domains.
- [x] **1B pre-sort pipeline tool** — `ClassifyCorpus()` in `classify.go`. Streaming JSONL batch classification via `inference.TextModel.Classify()`. Mock-tested (3 test cases) + integration-tested with real Gemma3-1B (80 prompts/sec on 50-prompt run, 100% domain accuracy). Configurable batch size, prompt field, and template.

### Virgil Review: Fix Before Continuing (20 Feb 2026)

**Do these first, in order, before picking up the next Phase 2a task.**

- [x] **Fix go.mod: remove go-mlx from module require** — Removed go-mlx `require` and `replace` from go.mod. Moved integration test to `integration/` sub-module with its own go.mod that depends on go-mlx. Main module now compiles cleanly on all platforms. `go mod tidy` no longer pulls go-mlx.

- [x] **Fix go.mod: go-inference pseudo-version** — `go mod tidy` resolved to the standard replaced-module pseudo-version `v0.0.0-00010101000000-000000000000`. CI-safe.

- [x] **Fix mapTokenToDomain prefix collision** — Replaced `strings.HasPrefix` with exact match + known BPE fragment fallback. Added test cases for "castle", "cascade", "credential", "creature" — all return "unknown".

- [x] **Fix classify_bench_test.go naming** — Added `testing.Short()` skip to `TestClassification_DomainSeparation` and `TestClassification_LeaveOneOut` (the two O(n^2) tests). Verified with `go test -short -v`.

- [x] **Add accuracy assertion to integration test** — Integration test now asserts at least 80% (40/50) of technical prompts classified as "technical". Logs full domain breakdown and misclassified entries on failure. Test moved to `integration/` sub-module.

### Remaining Phase 2a Tasks

- [x] **1B vs 27B calibration check** — `CalibrateDomains()` in `calibrate.go`. Accepts two TextModels + 500 CalibrationSamples (220 ground-truth + 280 unlabelled). Batch-classifies with both models, computes agreement rate, per-domain distribution, confusion pairs, and accuracy vs ground truth. 7 mock tests (race-clean). Integration test at `integration/calibrate_test.go` loads LEM-1B + Gemma3-27B from `/Volumes/Data/lem/`, runs full calibration with detailed reporting. Run with: `cd integration && go test -v -run TestCalibrateDomains_1Bvs27B`
- [x] **Article/irregular validator** — Lightweight Go funcs that use the 1B model's strong article correctness (100%) and irregular base form accuracy (100%) as fast validators. Use `m.Generate()` with `inference.WithMaxTokens(1)` and `inference.WithTemperature(0.05)` for single-token classification.

### 2b: Reference Distributions

- [ ] **Reference distribution builder** — Process the 88K scored seeds from LEM Phase 0 through the tokeniser + imprint pipeline. Pre-sort by `domain_1b` tag from step 2a first. Output: per-category (ethical, technical, creative, casual) reference distributions stored as JSON. This calibrates what "normal" grammar looks like per domain.
- [ ] **Imprint comparator** — Given a new text and reference distributions, compute distance metrics (cosine, KL divergence, Mahalanobis). Return a classification signal with confidence score. This is the Poindexter integration point.
- [ ] **Cross-domain anomaly detection** — Flag texts where 1B domain tag disagrees with imprint-based classification. These are either misclassified by 1B (training signal) or genuinely cross-domain (ethical text using technical jargon). Both are valuable for refining the pipeline.

## Phase 3: Multi-Language

- [x] **Grammar table format spec** — Full JSON schema documented in `docs/grammar-table-spec.md`. Covers all 7 `gram.*` sub-keys, required/optional fields, examples for English and French, 3-tier fallback chain, and new-language checklist.
- [x] **French grammar tables** — 50 verbs, 24 gendered nouns, gendered articles (by_gender), French punctuation spacing, 33 noun determiners, 21 verb auxiliaries. Loader extended to parse `by_gender` article map. Stress test confirms: verb/noun/article/punct/signal all load correctly; elision (l') and plural articles (les/des) need future Article() extension.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
4. New discoveries → add tasks, flag in FINDINGS.md
