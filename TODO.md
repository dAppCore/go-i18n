# TODO.md — go-i18n Task Queue

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 1: Harden the Engine

- [x] **Add CLAUDE.md** — Document the grammar engine contract: what it is (grammar primitives + reversal), what it isn't (translation file manager). Include build/test commands, the gram.* sacred rule, and the agent-flattening prohibition. *(d5b3eac)*
- [x] **Ambiguity resolution for dual-class words** — Two-pass probabilistic disambiguation with 7 weighted signals. All 6 dual-class words {commit, run, test, check, file, build} correctly disambiguate in both verb and noun contexts. Confidence scores flow into imprints. *(3848297)*
- [x] **Extend irregular verb coverage** — Added 44 irregular verbs: 17 compound (undo, redo, rerun, rewrite, rebuild, resend, override, rethink, remake, undergo, overcome, withdraw, uphold, withhold, outgrow, outrun, overshoot), 22 simple (become, come, give, fall, understand, arise, bind, spin, quit, cast, broadcast, burst, cost, shed, rid, shrink, shoot, forbid, offset, upset, input, output), 5 CVC doubling overrides (debug, embed, unzip, remap, unpin, unwrap).
- [x] **Add benchmarks** — 8 forward composition + 7 reversal benchmarks. Baseline on M3 Ultra: PastTense 26ns/0alloc, Tokenise(short) 639ns/8alloc, Imprint 648ns/10alloc, Similar 516ns/0alloc.

## Phase 2: Reference Distribution + 1B Classification Pipeline

### 2a: 1B Pre-Classification (NEW — based on benchmark findings)

- [x] **Classification benchmark suite** — 220 domain-tagged sentences, leave-one-out classification via imprint similarity. Grammar engine: technical 78%, creative 82%, ethical 46%, casual 11%. Ethical↔technical and casual↔creative confusion confirms 1B model needed for those domains.
- [ ] **1B pre-sort pipeline tool** ⏳ *Blocked: needs go-ai bindings for MLX/Gemma3-1B inference* — CLI command or Go func that reads a JSONL corpus (Phase 0 seeds), sends each text through LEK-Gemma3-1B domain classification, and writes back JSONL with `domain_1b` field added. Target: ~5K sentences/sec on M3.
- [ ] **1B vs 27B calibration check** ⏳ *Blocked: needs go-ai* — Sample 500 sentences, classify with both 1B and 27B, measure agreement rate. Classification benchmark shows ethical↔technical (both base-form heavy) and casual↔creative (both past-tense heavy) are the confusion axes — 1B needs to resolve these.
- [ ] **Article/irregular validator** ⏳ *Blocked: needs go-ai* — Lightweight Go funcs that use the 1B model's strong article correctness (100%) and irregular base form accuracy (100%) as fast validators.

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
