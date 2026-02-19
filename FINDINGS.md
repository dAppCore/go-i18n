# FINDINGS.md — go-i18n Research & Discovery

Record findings, gaps, and architectural decisions here as work progresses.

---

## 2026-02-19: Initial Assessment (Virgil)

### Current State

- 5,800 lines across 32 files (14 test files)
- All tests pass
- Only dependency: `golang.org/x/text`
- Grammar engine is solid: forward composition + reversal + imprint + multiplier

### Architecture

go-i18n is a **grammar engine**, not a translation file manager. Consumers bring their own translations. The library provides:

1. **Forward composition** — Grammar primitives that compose grammatically correct text
2. **Reverse grammar** — Tokeniser reads grammar tables backwards to extract structure
3. **GrammarImprint** — Lossy feature vector projection (content to grammar fingerprint)
4. **Multiplier** — Deterministic training data augmentation (no LLM calls)

### Key Gaps Identified

| Gap | Impact | Notes |
|-----|--------|-------|
| No CLAUDE.md | High | Agents don't know the rules, will flatten locale files |
| Dual-class word ambiguity | Medium | "file" as verb vs noun, "run" as verb vs noun |
| No benchmarks | Medium | No perf baselines for hot-path usage (TIM, Poindexter) |
| No reference distributions | High | Can't calibrate imprints without scored seed data |
| Only English grammar tables | Medium | Reversal only works with loaded GrammarData |

### Sacred Rules

- `gram.*` keys in locale JSON MUST remain nested — flattening breaks the grammar engine
- Irregular forms in grammar tables override regular morphological rules
- Round-trip property must hold: forward(base) then reverse must recover base
- go-i18n does NOT ship or manage consumer translation files

---

## 2026-02-19: LEK-Gemma3-1B-v2 Benchmark (Virgil)

Tested the fine-tuned 1B model (`/Volumes/Data/lem/LEM-Gemma3-1B-layered-v2`) across three progressively tighter evaluation rounds to find where it provides real value for the grammar pipeline.

### Round 1: Practical Dev Tasks (5 tasks, temp=0.3, max_tokens=512)

Open-ended dev work — bug spotting, commit messages, Go functions, grammar tables, code review. Results: mostly misses. The model hallucinates APIs, generates pad-token degeneration on longer output, and can't reliably write code. Not useful for generative tasks.

### Round 2: Narrow Constrained Tasks (24 tasks, temp=0.1, avg 0.19s/task)

Tighter format — one-word answers, forced categories, fill-in-the-blank.

| Category | Score | Notes |
|----------|-------|-------|
| Domain classification | 3/4 (75%) | Promising — called technical, creative, ethical, casual correctly |
| Article selection | 2/3 (67%) | Got "an API" and "an SSH" right, missed "a URL" |
| Tense detection | 2/4 (50%) | Weak on gerund vs base form |
| Plural detection | 2/3 (67%) | Got "matrices" and "datum", confused on "sheep" |
| Conjugation | Mixed | Some correct, many hallucinated forms |

### Round 3: Tightest Format (27 tasks, temp=0.05, avg 0.18s/task)

Binary T/F, forced A/B choice, single-word domain/tone classification.

| Category | Score | Notes |
|----------|-------|-------|
| Domain classification | 6/8 (75%) | Consistent with Round 2 — this is the sweet spot |
| Article correctness T/F | 3/3 (100%) | Perfect on "an SSH", "a API"→false, "an URL"→false |
| Tone/sentiment | 2/3 (67%) | Got positive + negative, neutral confused |
| Irregular base forms A/B | 2/2 (100%) | "went"→go, "mice"→mouse — strong |
| True/False grammar | 4/8 (50%) | Strong false-bias — says "false" when unsure |
| Pattern fill | 0/4 (0%) | Echoes prompt or hallucinates — dead zone |

### Key Finding: Domain Classification at Scale

**Domain classification at 75% accuracy in 0.17s is genuinely useful.** At that speed, one M3 can pre-sort ~5,000 sentences/second across {technical, creative, ethical, casual}. For the 88K Phase 0 seed corpus, that's ~18 seconds to pre-tag everything.

The technical↔creative confusion (calls some technical text "creative") is the main error pattern — likely fixable with targeted fine-tuning examples showing code/CLI commands vs literary prose.

### Implications for Phase 2

1. **Pre-sort pipeline**: Run 1B domain classification before heavier GrammarImprint analysis. Pre-tagged seeds reduce imprint compute by letting us batch by domain.
2. **Calibration target**: 1B classifications can be spot-checked against Gemma3-27B classifications to measure drift.
3. **Article/irregular strength**: The 100% article correctness and irregular base form accuracy suggest these grammar features are well-learned. Worth testing as lightweight validators in the forward composition path.
4. **Dead zones to avoid**: Don't use 1B for pattern fill, tense detection, or generative tasks. These need the full 27B or rule-based approaches.

---

## 2026-02-19: Dual-Class Word Disambiguation (Implementation)

### What Was Built

Two-pass probabilistic disambiguation for words that exist as both verbs and nouns in the grammar tables. Replaces the previous verb-first hard classification with context-aware scoring.

### Dual-Class Set

{commit, run, test, check, file, build} — all 6 words now appear in both `gram.verb` and `gram.noun` in en.json.

### Algorithm: Two-Pass Tokenise

**Pass 1** classifies unambiguous tokens. Inflected forms self-resolve (e.g. "committed" → verb, "commits" → noun). Base forms of dual-class words are marked as ambiguous.

**Pass 2** evaluates 7 weighted signals to resolve ambiguous tokens:

| Signal | Weight | Description |
|--------|--------|-------------|
| noun_determiner | 0.35 | Preceded by "the", "a", "my", etc. → noun |
| verb_auxiliary | 0.25 | Preceded by "will", "can", "don't", etc. → verb |
| following_class | 0.15 | Followed by noun → verb; followed by verb → noun |
| sentence_position | 0.10 | Sentence-initial → verb (imperative) |
| verb_saturation | 0.10 | Clause already has a confident verb → noun |
| inflection_echo | 0.03 | Inflected form of same word elsewhere → other role |
| default_prior | 0.02 | Verb-first tiebreaker |

### Key Design Decisions

- **Confidence scores** flow into imprints: dual-class tokens contribute to both verb and noun distributions weighted by Confidence and AltConf. This preserves uncertainty for scoring rather than forcing a hard classification.
- **Clause boundaries** for verb saturation: scans only within clause (delimited by punctuation and coordinating conjunctions "and", "or", "but"). Prevents multi-clause sentences from incorrectly pushing second verbs toward noun.
- **Confidence floor** (B3): when only the default prior fires (total < 0.10), confidence is capped at 0.55/0.45 rather than deriving a misleading 1.0 from `0.02/0.02`.
- **Contractions** (D1): 15 contractions added to verb_auxiliary signal list (don't, can't, won't, etc.).
- **Configurable weights** (F3): `WithWeights()` option allows overriding signal weights without code changes.
- **DisambiguationStats** (F1): `DisambiguationStatsFromTokens()` provides aggregate stats for Phase 2 calibration.
- **SignalBreakdown** opt-in: `WithSignals()` populates detailed per-token signal diagnostics.

### Test Coverage

- 9 disambiguation scenario tests (noun after determiner, verb imperative, verb saturation, clause boundary, contraction aux, etc.)
- 12 dual-class round-trip tests covering all 6 words in both roles
- Imprint convergence test (same-role similar, different-role divergent)
- DisambiguationStats tests (ambiguous and non-ambiguous inputs)
- Race detector: clean

### Expanded Dual-Class Candidates (Phase 2)

Per REVIEW.md F4, additional candidates for future expansion: patch, release, update, change, merge, push, pull, tag, log, watch, link, host, import, export, process, function, handle, trigger, stream, queue. Measure which ones cause imprint drift in the 88K seeds before adding.
