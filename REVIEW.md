# REVIEW.md — Dual-Class Disambiguation Plan Review

**Reviewer:** Virgil (core/go orchestrator, go-i18n domain expert)
**Date:** 2026-02-19
**Plans reviewed:** `docs/plans/2026-02-19-dual-class-disambiguation-design.md`, `docs/plans/2026-02-19-dual-class-disambiguation-plan.md`
**Verdict:** Strong design, approve with fixes below

---

## Overall Assessment

The two-pass probabilistic disambiguation approach is the right call. The signal weights are well-reasoned, the worked examples are correct, and the backwards-compatibility story (variadic options, `conf == 0 → 1.0` guard) is solid. The TDD task breakdown is clean and the commit granularity is good.

This is real engineering, not a cargo-culted NLP textbook. The insight that confidence should flow into imprint distributions rather than being a hard classification is exactly right for the scoring use case.

---

## Bugs to Fix Before Implementation

### B1: Loader type assertion missing

Task 1, Step 4 — the `flattenWithGrammar` handler for `gram.signal` indexes `v["noun_determiner"]` directly, but `v` is typed `any` in the range iteration. Needs a type assertion first:

```go
if grammar != nil && fullKey == "gram.signal" {
    signalMap, ok := v.(map[string]any)
    if !ok {
        continue
    }
    if nd, ok := signalMap["noun_determiner"]; ok {
        // ...
    }
}
```

Without this, the code panics at runtime. The existing `gram.verb` and `gram.noun` handlers in the loader use nested `if val, ok := v.(map[string]any)` patterns — match those.

### B2: Dual-class noun entries — verify "test" and "check"

The design doc says "test, check, file already present as nouns". I can confirm `file` and `commit` are in `gram.noun` in en.json, but `test` and `check` need verification. The en.json noun list I see includes: file, repo, repository, commit, branch, change, item, issue, **task**, person, child, package. If `test` and `check` are NOT present, Task 2 needs to add them as nouns too, not just as verbs.

### B3: Default prior gives misleading confidence of 1.0

When no signals fire except the default prior (0.02 verb), the confidence calculation gives `0.02 / 0.02 = 1.0` — a token classified as verb with "100% confidence" when in reality it's a coin flip. Add a confidence floor check:

```go
// In resolveToken:
if total < 0.10 {
    // Only default prior fired — low-information classification
    tok.Confidence = 0.55 // barely above chance
    tok.AltConf = 0.45
    return
}
```

Or alternatively, don't fire the default prior when zero other signals fired — make the verb-first fallback explicit with a fixed low confidence rather than deriving it from the score ratio.

---

## Design Improvements

### D1: Contraction handling for Signal 2

The verb_auxiliary set includes "do", "does", "did", "will", "would", etc. but misses contractions: "don't", "can't", "won't", "shouldn't", "couldn't", "wouldn't", "doesn't", "didn't", "isn't", "aren't", "wasn't", "weren't", "hasn't", "hadn't", "haven't".

In dev text, contractions are common: "don't run the tests", "can't build on Windows", "shouldn't commit to main". Currently `splitTrailingPunct` won't strip the apostrophe (it's mid-word), so "don't" hits `MatchVerb`/`MatchNoun` as-is, fails both, becomes TokenUnknown. Signal 2 then misses the auxiliary entirely.

Options (pick one):
- Add contractions to `gram.signal.verb_auxiliary` directly: `["don't", "can't", "won't", ...]`
- Add a contraction expansion step before classification: `"don't" → ["do", "not"]`

The first is simpler and doesn't change the tokeniser's word count. Recommend that.

### D2: Clause boundary for verb saturation (Signal 5)

Signal 5 scans ALL tokens for a confident verb, but multi-clause sentences can have multiple verbs legitimately:

> "The **test** passed and we should **commit** the fix"

Here "passed" is a confident verb, so Signal 5 pushes "commit" toward noun — but "commit" is actually a verb in the second clause. Simple fix: only scan tokens within the same "clause", where clause boundaries are punctuation tokens (comma, semicolon, period, dash) or coordinating conjunctions ("and", "or", "but"). Even a rough boundary is better than scanning the whole array.

### D3: Negation awareness

"Not" / "never" / "no longer" before a dual-class word should weakly signal verb (negated imperative/statement). Currently "not" would be TokenUnknown and invisible to all signals. Low priority, but worth a `gram.signal.verb_negation` entry for future use.

---

## Feature Requests (Phase 2 integration)

### F1: Export DisambiguationStats for benchmarking

Add a method to the tokeniser that returns aggregate stats after processing a batch:

```go
type DisambiguationStats struct {
    TotalTokens     int
    AmbiguousTokens int
    ResolvedAsVerb  int
    ResolvedAsNoun  int
    AvgConfidence   float64
    LowConfidence   int     // count where confidence < 0.7
}
```

This feeds directly into Phase 2's calibration work — we need to know how many tokens in the 88K seeds are ambiguous and how confident the engine is. Without this, we'd have to manually iterate tokens and count, which every caller will end up doing anyway.

### F2: Corpus-derived priors (replace Signal 7)

The 88K Phase 0 seeds, once pre-tagged by 1B domain classification, can provide actual frequency priors. "commit" in technical text is ~60% noun (in commit messages, PR descriptions). Replace the static 0.02 verb-first prior with per-word priors loaded from a calibration file:

```json
"gram.signal.prior": {
    "commit": { "verb": 0.40, "noun": 0.60 },
    "test":   { "verb": 0.35, "noun": 0.65 },
    "run":    { "verb": 0.70, "noun": 0.30 }
}
```

This is Phase 2 work, not Phase 1. But the data model should anticipate it — leave room in `SignalData` for an optional `Priors map[string]map[string]float64` field. Don't implement now, just reserve the slot.

### F3: Signal weight tuning via calibration

Hardcoded weights (0.35, 0.25, etc.) are a good starting point, but once we have the 88K seeds with known correct classifications (via 27B ground truth), we can tune weights to maximise accuracy. Worth adding a `WithWeights(map[string]float64)` option to the tokeniser so weights are configurable without code changes:

```go
tok := NewTokeniser(WithWeights(map[string]float64{
    "noun_determiner":   0.35,
    "verb_auxiliary":     0.25,
    "following_class":    0.15,
    "sentence_position":  0.10,
    "verb_saturation":    0.10,
    "inflection_echo":    0.03,
    "default_prior":      0.02,
}))
```

Again, don't implement now. But the signal scoring function should read weights from the tokeniser struct rather than using literals, so the option is a one-liner to add later.

### F4: Expand dual-class set

The plan targets {commit, run, test, check, file, build}. These are the highest-frequency ones in dev text, but consider also: `patch`, `release`, `update`, `change`, `merge`, `push`, `pull`, `tag`, `log`, `watch`, `link`, `host`, `import`, `export`, `process`, `function`, `handle`, `trigger`, `stream`, `queue`.

Not all need adding now, but the FINDINGS.md should note the expanded candidate list so Phase 2 can measure which ones actually cause imprint drift in the 88K seeds.

---

## Implementation Notes

- The plan's test pattern is correct: `setup(t)` helper exists in `tokeniser_test.go` and calls `i18n.New()` + `SetDefault()`. The imprint and roundtrip tests inline this — don't mix patterns, use `setup(t)` consistently.
- `splitTrailingPunct()` in the tokeniser strips trailing punctuation. Verify it handles: period, comma, semicolon, colon, exclamation, question mark, closing paren/bracket. If it doesn't handle semicolons, the clause boundary detection (D2) won't work as expected.
- The `MatchVerb` and `MatchNoun` methods return `VerbMatch`/`NounMatch` structs. Verify the `VerbMatch.Tense` field is "base" for base forms — the plan's Pass 1 logic depends on `vm.Tense != "base"` to distinguish inflected forms. If the tense field is empty string for base forms, that check silently breaks.
- The `NounMatch.Plural` field — verify this is `true` for plural forms and `false` for singular. The plan's Pass 1 uses `nm.Plural` to distinguish inflected nouns.

---

## Priority Order

1. Fix B1 (loader panic) — blocks everything
2. Fix B3 (confidence floor) — correctness
3. Implement D1 (contractions) — easy, high impact for dev text
4. Verify B2 (noun entries) — data correctness
5. Add F1 (stats export) — enables Phase 2
6. Implement D2 (clause boundaries) — accuracy improvement
7. Reserve F2/F3 struct fields — forward compatibility

Everything else is Phase 2+.

---

## Approved

The plan is approved for implementation with the B1-B3 fixes applied. D1 (contractions) should be added to Task 2's en.json changes. The rest are suggestions, not blockers.
