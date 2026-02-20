# TODO.md — go-i18n Task Queue

Dispatched from core/go orchestration. All phases complete as of 20 Feb 2026.

---

## Phase 1: Harden the Engine — COMPLETE

- [x] **Add CLAUDE.md** *(d5b3eac)*
- [x] **Ambiguity resolution for dual-class words** — Two-pass probabilistic disambiguation, 7 weighted signals, 6 dual-class words *(3848297)*
- [x] **Extend irregular verb coverage** — 44 new irregular verbs (17 compound, 22 simple, 5 CVC overrides), ~140 total
- [x] **Add benchmarks** — 8 forward + 7 reversal. Baselines: PastTense 26ns, Tokenise 639ns, Imprint 648ns, Similar 516ns/0alloc

## Phase 2a: 1B Pre-Classification — COMPLETE

- [x] **Classification benchmark suite** — 220 domain-tagged sentences, grammar engine: 54.1% overall (tech 78%, creative 82%, ethical 46%, casual 11%)
- [x] **1B pre-sort pipeline** — `ClassifyCorpus()` in `classify.go`, 80 prompts/sec on M3 Ultra, mock + integration tested
- [x] **Virgil review fixes** — go.mod cleanup, prefix collision fix, short-mode skip, accuracy assertion (5 items)
- [x] **1B vs 27B calibration** — `CalibrateDomains()` in `calibrate.go`, 500-sample corpus (220 ground-truth + 280 unlabelled), 7 mock tests *(3b7ef9d)*
- [x] **Article/irregular validator** — Single-token classification via `m.Generate()` with temp=0.05

## Phase 2b: Reference Distributions — COMPLETE

- [x] **Reference distribution builder** — `BuildReferences()` in `reversal/reference.go`, per-domain centroid + variance *(c3e9153)*
- [x] **Imprint comparator** — `Compare()` + `Classify()`, cosine/KL/Mahalanobis distance metrics *(c3e9153)*
- [x] **Cross-domain anomaly detection** — `DetectAnomalies()` in `reversal/anomaly.go`, flags model vs imprint disagreements *(c3e9153)*

## Phase 3: Multi-Language — COMPLETE

- [x] **Grammar table format spec** — JSON schema in `docs/grammar-table-spec.md`
- [x] **French grammar tables** — 50 verbs, 24 gendered nouns, gendered articles, punctuation spacing

---

## Integration Tests (require real models on `/Volumes/Data/lem/`)

```bash
# 1B classification pipeline (50 prompts, ~1s)
cd integration && go test -v -run TestClassifyCorpus_Integration

# 1B vs 27B calibration (500 sentences, ~2-5min with 27B)
cd integration && go test -v -run TestCalibrateDomains_1Bvs27B
```

## Future Work (not yet tasked)

- **Expanded dual-class words** — 20 candidates: patch, release, update, change, merge, push, pull, tag, log, watch, link, host, import, export, process, function, handle, trigger, stream, queue. Measure imprint drift in 88K seeds first.
- **French reversal** — Elision (l') and plural articles (les/des) need `Article()` extension.
- **88K seed corpus processing** — Run full pre-sort + reference distribution build against LEM Phase 0 seeds.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
4. New discoveries → add tasks, flag in FINDINGS.md
