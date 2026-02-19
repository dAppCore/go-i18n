# Dual-Class Word Disambiguation Design

**Date:** 2026-02-19
**Status:** Approved
**Scope:** `reversal/` package, `types.go`, `loader.go`, `locales/en.json`

---

## Problem

The tokeniser classifies words by checking verb before noun (line 500 before 503 in `tokeniser.go`). Words that exist in both `gram.verb` and `gram.noun` — "commit", "run", "test", "check", "file", "build" — always classify as verbs regardless of grammatical context.

This silently corrupts `GrammarImprint` distributions. "The test failed" produces a verb-heavy imprint instead of the noun+verb structure it actually has. For the scoring/comprehension use case (replacing LLM-as-judge), this is a systematic bias that undermines similarity comparisons.

## Approach

Multi-signal probabilistic disambiguation with two-pass tokenisation.

**Why probabilistic:** Hard classification loses information. A "commit" classified as noun with 0.85 confidence should contribute 0.85 to `NounDistribution` and 0.15 to `VerbDistribution` in the imprint. This preserves uncertainty — similar texts produce similar imprints even when individual token classifications wobble at the boundary.

**Why two-pass:** Pass 1 resolves unambiguous tokens (inflected forms self-disambiguate: "committed" is always past-tense verb, "commits" is always plural noun). Pass 2 uses the resolved context to score ambiguous base forms. No circular dependencies.

**Computational budget:** This engine replaces hours of LLM-as-judge computation with milliseconds of deterministic analysis. The two-pass overhead is justified.

---

## Core Model

For each ambiguous base-form token, compute competing scores:

```
verb_score = Σ(weight_i × signal_i)  for verb-indicating signals
noun_score = Σ(weight_i × signal_i)  for noun-indicating signals

confidence = max(verb_score, noun_score) / (verb_score + noun_score)
```

Winner becomes the token's `Type`. Loser becomes `AltType`. Both match structs (`VerbInfo`, `NounInfo`) are always populated for dual-class words.

---

## Signal Dimensions

Seven signal functions, each returning 0.0-1.0:

### Signal 1: Preceding Noun Determiner (weight: 0.35)

Articles, demonstratives, possessives, and quantifiers almost always precede nouns. If token[i-1] is in the noun determiner set, strong noun signal.

```
"the commit was approved"
      ↑ "the" ∈ nounDet → noun_score += 0.35
```

Word list stored in `gram.signal.noun_determiner` in locale JSON:
```
the, a, an, this, that, these, those,
my, your, his, her, its, our, their,
every, each, some, any, no, many, few, several, all, both
```

### Signal 2: Auxiliary/Modal Preceding (weight: 0.25)

Auxiliary verbs, modals, and infinitive markers precede main verbs.

```
"will commit the changes"  → verb_score += 0.25
"to run the tests"         → verb_score += 0.25
```

Word list stored in `gram.signal.verb_auxiliary`:
```
is, are, was, were, has, had, have,
do, does, did, will, would, could, should,
can, may, might, shall, must
```

And `gram.signal.verb_infinitive`: `to`

### Signal 3: Following Token Class (weight: 0.15)

What follows the ambiguous word, using Pass 1 classifications:

- Followed by article/determiner → verb signal ("**run** the tests")
- Followed by noun → verb signal ("**build** artifacts")
- Followed by verb → noun signal ("the **test** failed")

### Signal 4: Sentence Position (weight: 0.10)

Sentence-initial position with no preceding determiner suggests imperative verb.

```
"Build the project" → index 0, no determiner → verb_score += 0.10
```

### Signal 5: Verb Saturation (weight: 0.10)

English clauses typically have one main verb. If Pass 1 already resolved a confident verb in the current clause, another base form is more likely a noun.

```
"The system runs every test"
               ↑ confident verb    ↑ verb already found → noun_score += 0.10
```

### Signal 6: Inflection Echo (weight: 0.03)

If the same base appears elsewhere in the text in an inflected form ("commits" alongside "commit"), the base form is more likely being used in the other grammatical role.

### Signal 7: Default Prior (weight: 0.02)

Tiny verb-first nudge. Preserves current behaviour as tiebreaker when all signals are neutral. Can be replaced with corpus-derived priors from the 88K seeds later.

### Weight Summary

| Signal | Weight | Direction | Looks At |
|--------|--------|-----------|----------|
| Noun determiner | 0.35 | Noun | token[i-1] |
| Auxiliary/modal | 0.25 | Verb | token[i-1] |
| Following class | 0.15 | Mixed | token[i+1] |
| Sentence position | 0.10 | Verb | Token index |
| Verb saturation | 0.10 | Noun | Clause-level |
| Inflection echo | 0.03 | Mixed | Full text |
| Default prior | 0.02 | Verb | None |

Weights sum to 1.0 — raw scores are pre-normalised.

---

## Data Model Changes

### en.json: Signal Tables

New `gram.signal` block alongside existing `gram.verb`, `gram.noun`, etc:

```json
{
  "gram": {
    "signal": {
      "noun_determiner": [
        "the", "a", "an",
        "this", "that", "these", "those",
        "my", "your", "his", "her", "its", "our", "their",
        "every", "each", "some", "any", "no",
        "many", "few", "several", "all", "both"
      ],
      "verb_auxiliary": [
        "is", "are", "was", "were",
        "has", "had", "have",
        "do", "does", "did",
        "will", "would", "could", "should",
        "can", "may", "might", "shall", "must"
      ],
      "verb_infinitive": ["to"]
    }
  }
}
```

### en.json: Dual-Class Word Additions

Words flagged in TODO.md that need entries in both verb and noun tables:

**Add to `gram.verb`:** `test`, `check`, `file` (run and build already present as verbs)
**Add to `gram.noun`:** `run`, `build` (test, check, file already present as nouns)

Dual-class set is derived automatically: any word in both `gram.verb` and `gram.noun`.

### types.go: GrammarData

```go
type SignalData struct {
    NounDeterminers []string
    VerbAuxiliaries []string
    VerbInfinitive  []string
}

type GrammarData struct {
    Verbs    map[string]VerbForms
    Nouns    map[string]NounForms
    Articles ArticleForms
    Words    map[string]string
    Punct    PunctuationRules
    Signals  SignalData           // NEW
}
```

### reversal/tokeniser.go: Token Struct

```go
type Token struct {
    Raw        string
    Lower      string
    Type       TokenType
    Confidence float64         // 0.0-1.0 classification confidence
    AltType    TokenType       // Runner-up classification (dual-class only)
    AltConf    float64         // Runner-up confidence

    VerbInfo   VerbMatch       // Populated when Type OR AltType == TokenVerb
    NounInfo   NounMatch       // Populated when Type OR AltType == TokenNoun
    WordCat    string
    ArtType    string
    PunctType  string

    Signals    *SignalBreakdown // nil unless WithSignals() option set
}

type SignalBreakdown struct {
    VerbScore  float64
    NounScore  float64
    Components []SignalComponent
}

type SignalComponent struct {
    Name    string   // "noun_determiner", "verb_auxiliary", etc.
    Weight  float64
    Value   float64
    Contrib float64  // weight × value
    Reason  string   // "preceded by 'the'"
}
```

### reversal/tokeniser.go: Tokeniser Struct

```go
type TokeniserOption func(*Tokeniser)

func WithSignals() TokeniserOption {
    return func(t *Tokeniser) { t.withSignals = true }
}

type Tokeniser struct {
    // existing
    pastToBase   map[string]string
    gerundToBase map[string]string
    baseVerbs    map[string]bool
    pluralToBase map[string]string
    baseNouns    map[string]bool
    words        map[string]string
    lang         string

    // new
    dualClass    map[string]bool
    nounDet      map[string]bool
    verbAux      map[string]bool
    verbInf      map[string]bool
    withSignals  bool
}
```

`NewTokeniser(opts ...TokeniserOption)` — variadic options, backwards compatible.

---

## Two-Pass Algorithm

### Pass 1: Classify & Mark

For each word in input:

1. Strip trailing punctuation
2. Check article → TokenArticle (confidence 1.0)
3. Check verb:
   - Inflected form (past/gerund from inverse maps or round-trip) → TokenVerb (confidence 1.0)
   - Base form AND `dualClass[base]` → mark as **ambiguous**, stash VerbMatch
   - Base form, not dual-class → TokenVerb (confidence 1.0)
4. Check noun:
   - If already marked ambiguous → stash NounMatch (don't override)
   - Inflected form (plural from inverse maps or round-trip) → TokenNoun (confidence 1.0)
   - Base form AND `dualClass[base]` → mark as ambiguous, stash both
   - Base form, not dual-class → TokenNoun (confidence 1.0)
5. Check word → TokenWord (confidence 1.0)
6. Else → TokenUnknown (confidence 0.0)

### Pass 2: Contextual Scoring

For each ambiguous token at index `i`:

1. Evaluate all 7 signals against surrounding context
2. Sum verb_score and noun_score
3. If both zero: fall through to default prior (verb wins at 0.02)
4. Compute confidence: `winner / (verb_score + noun_score)`
5. Set Type, Confidence, AltType, AltConf
6. If `withSignals`: populate SignalBreakdown

### Edge Cases

**Adjacent ambiguous tokens** ("test run"): Each sees the other as ambiguous in Pass 2. Signal 3 returns 0.0 for both (no information from unresolved neighbours). Other signals still fire. Worst case: both fall to default prior — verb-first, same as current behaviour.

**No gram.signal in JSON**: Hardcoded English fallback for determiners/auxiliaries. Engine works without the JSON block, just with less configurability.

**Zero ambiguous tokens**: Pass 2 is a no-op. Zero overhead for texts without dual-class base forms.

---

## Imprint Integration

The imprint system uses confidence to weight contributions:

```go
case TokenVerb:
    conf := tok.Confidence
    if conf == 0 { conf = 1.0 }
    imp.VerbDistribution[tok.VerbInfo.Base] += conf
    imp.TenseDistribution[tok.VerbInfo.Tense] += conf
    verbBases[tok.VerbInfo.Base] = true

    if tok.AltType == TokenNoun && tok.NounInfo.Base != "" {
        altConf := tok.AltConf
        imp.NounDistribution[tok.NounInfo.Base] += altConf
        totalNouns++
    }
```

Two texts using "commit" in the same grammatical role produce converging imprints. Two texts using "commit" in different roles produce slightly divergent imprints. This is the correct behaviour for comprehension scoring.

---

## Worked Examples

### "Delete the commit" (imperative, commit=noun)

| Pass | Token | Classification | Confidence |
|------|-------|---------------|------------|
| 1 | Delete | TokenVerb (base, not dual-class) | 1.0 |
| 1 | the | TokenArticle | 1.0 |
| 1 | commit | AMBIGUOUS | — |
| 2 | commit | TokenNoun (det="the" 0.35, saturation 0.10 vs prior 0.02) | 0.96 |

### "Commit the changes" (imperative, commit=verb)

| Pass | Token | Classification | Confidence |
|------|-------|---------------|------------|
| 1 | Commit | AMBIGUOUS | — |
| 1 | the | TokenArticle | 1.0 |
| 1 | changes | TokenNoun (plural) | 1.0 |
| 2 | Commit | TokenVerb (following=article 0.15, position 0.10, prior 0.02) | 1.0 |

### "The test failed because the commit introduced a regression"

| Pass | Token | Classification | Confidence |
|------|-------|---------------|------------|
| 1 | The | TokenArticle | 1.0 |
| 1 | test | AMBIGUOUS | — |
| 1 | failed | TokenVerb (past, inflected) | 1.0 |
| 1 | because | TokenUnknown | 0.0 |
| 1 | the | TokenArticle | 1.0 |
| 1 | commit | AMBIGUOUS | — |
| 1 | introduced | TokenVerb (past, inflected) | 1.0 |
| 1 | a | TokenArticle | 1.0 |
| 1 | regression | TokenUnknown | 0.0 |
| 2 | test | TokenNoun (det="The" 0.35, saturation 0.10 vs prior 0.02) | 0.96 |
| 2 | commit | TokenNoun (det="the" 0.35, saturation 0.10 vs prior 0.02) | 0.96 |

### "Run tests" (imperative, run=verb)

| Pass | Token | Classification | Confidence |
|------|-------|---------------|------------|
| 1 | Run | AMBIGUOUS | — |
| 1 | tests | TokenNoun (plural) | 1.0 |
| 2 | Run | TokenVerb (following=noun 0.15, position 0.10, prior 0.02) | 1.0 |

---

## Files Changed

| File | Change |
|------|--------|
| `locales/en.json` | Add `gram.signal` block; add missing verb/noun entries for dual-class words |
| `types.go` | Add `SignalData` struct, add `Signals` field to `GrammarData` |
| `loader.go` | Handle `gram.signal` key prefix in loader |
| `reversal/tokeniser.go` | Add `TokeniserOption`, dual-class index, signal index, two-pass `Tokenise()`, `SignalBreakdown` types |
| `reversal/imprint.go` | Weight contributions by `Confidence` and `AltConf` |
| `reversal/multiplier.go` | Use `Confidence` field (backwards compat: treat 0 as 1.0) |
| `reversal/tokeniser_test.go` | Tests for all dual-class scenarios, signal accuracy |
| `reversal/imprint_test.go` | Tests for confidence-weighted imprints |
| `reversal/roundtrip_test.go` | Dual-class round-trip tests |
