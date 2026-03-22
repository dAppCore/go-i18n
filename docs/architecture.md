# Architecture

go-i18n is a grammar engine for Go. It is not a translation file manager. Consumers bring their own translation keys; the library provides grammatical primitives for composing and reversing grammatically correct text across multiple languages.

Module: `dappco.re/go/core/i18n`

---

## Package Overview

| Package | Purpose |
|---------|---------|
| Root (`i18n`) | Forward composition: grammar primitives, T(), handlers, service, loader |
| `reversal/` | Reverse grammar: tokeniser, imprint, reference distributions, anomaly detection, multiplier |
| `locales/` | Grammar tables (JSON) — only `gram.*` data |
| `docs/` | Specifications and design documents |

---

## Forward Composition

The root package composes grammatically correct text from base forms. Every public function follows a three-tier lookup:

1. JSON grammar tables loaded for the current language (`gram.verb`, `gram.noun`)
2. Go built-in irregular maps (`irregularVerbs`, `irregularNouns`)
3. Regular morphological rules (algorithmic)

### Grammar Primitives

**`PastTense(verb string) string`**

Applies the three-tier fallback to produce a simple past form. Irregular forms (e.g. `run` → `ran`, `build` → `built`) are resolved at tier 1 or 2; regular forms apply consonant-doubling, e/y-ending rules at tier 3. Benchmark: 26 ns/op for irregular (map lookup, zero allocs), 49 ns/op for regular (one string allocation).

**`Gerund(verb string) string`**

Produces the present participle. Handles `-ie` → `-ying` (e.g. `die` → `dying`), silent-e drop (e.g. `delete` → `deleting`), and consonant doubling (e.g. `run` → `running`).

**`Pluralize(noun string, count int) string`**

Returns singular when `count == 1`, delegates to `PluralForm()` otherwise.

**`PluralForm(noun string) string`**

Three-tier noun plural lookup. Regular rules handle sibilant (`+es`), consonant+y → ies, f/fe → ves, and default (`+s`).

**`Article(word string) string`**

Returns `"a"` or `"an"` based on phonetic rules. Handles exceptions in both directions: consonant-sound words starting with a vowel letter (e.g. `user` → `"a"`) and vowel-sound words starting with a consonant letter (e.g. `hour` → `"an"`). Implemented as prefix lookup tables for the known exceptions, falling back to first-letter vowel test.

**Composite functions**

`Progress(verb)` → `"Building..."`, `ProgressSubject(verb, subject)` → `"Building project..."`, `ActionResult(verb, subject)` → `"File deleted"`, `ActionFailed(verb, subject)` → `"Failed to delete file"`, `Label(word)` → `"Status:"`.

All composite functions look up language-specific punctuation rules from `gram.punct` (`LabelSuffix`, `ProgressSuffix`) to handle differences such as the French space-before-colon convention.

### T() and Key Handlers

`T(key string, args ...any) string` is the entry point for translation. It passes the key through a chain of `KeyHandler` implementations before falling back to the message store. The `i18n.*` namespace is handled by built-in handlers that auto-compose output:

| Key pattern | Output |
|-------------|--------|
| `i18n.label.<word>` | `Label(word)` |
| `i18n.progress.<verb>` | `Progress(verb)` |
| `i18n.count.<noun>` | `Pluralize(noun, n)` with count arg |
| `i18n.done.<verb>` | `ActionResult(verb, subject)` |
| `i18n.fail.<verb>` | `ActionFailed(verb, subject)` |

### Grammar Data Loading

`GrammarData` holds the parsed grammar tables for one language:

```go
type GrammarData struct {
    Verbs    map[string]VerbForms
    Nouns    map[string]NounForms
    Articles ArticleForms
    Words    map[string]string
    Punct    PunctuationRules
    Signals  SignalData
}
```

`FSLoader` reads `locales/<lang>.json`, calls `flattenWithGrammar()`, which walks the raw JSON tree and extracts `gram.*` blocks into typed Go structs before flattening all other keys into a `map[string]Message`. This is why the `gram.*` structure must remain nested in JSON — the extractor relies on path prefixes (`gram.verb.*`, `gram.noun.*`, etc.) to route objects correctly. Flattening to dot-separated keys bypasses this routing and causes silent data loss.

Grammar data is held in a package-level `grammarCache` (map protected by `sync.RWMutex`). `SetGrammarData(lang, data)` stores a loaded instance; `GetGrammarData(lang)` retrieves it.

---

## Reversal Engine

The reversal package reads grammar tables backwards: given an inflected form, recover the base form and the grammatical role.

### Tokeniser

`Tokeniser` maintains inverse indexes built at construction time:

| Index | Direction | Example |
|-------|-----------|---------|
| `pastToBase` | inflected past → base | `"deleted"` → `"delete"` |
| `gerundToBase` | inflected gerund → base | `"deleting"` → `"delete"` |
| `baseVerbs` | base verb set | `"delete"` → true |
| `pluralToBase` | plural → singular | `"files"` → `"file"` |
| `baseNouns` | base noun set | `"file"` → true |
| `words` | word map (case-insensitive) | `"url"` → `"url"` |
| `dualClass` | words in both verb and noun tables | `"commit"` → true |

`buildVerbIndex()` and `buildNounIndex()` populate these indexes in two sub-tiers: first from the loaded `GrammarData` (JSON tables), then from the exported `IrregularVerbs()` and `IrregularNouns()` maps. JSON takes precedence; existing entries are not overwritten.

**`MatchVerb(word string) (VerbMatch, bool)`**

Three-tier lookup:

1. Is the word in `baseVerbs`? Return tense `"base"`.
2. Is the word in `pastToBase` or `gerundToBase`? Return the appropriate tense.
3. Generate candidates via `reverseRegularPast()` or `reverseRegularGerund()` and round-trip verify each through the forward functions `PastTense()` / `Gerund()`. `bestRoundTrip()` resolves ambiguous candidates by preferring known base verbs, then VCe-ending words (the "magic e" pattern: `delete`, `create`, `use`), then words not ending in `e`.

**`MatchNoun(word string) (NounMatch, bool)`**

Same three-tier structure, using `reverseRegularPlural()` for tier 3.

### Tokenisation Pipeline

`Tokenise(text string) []Token` is a two-pass algorithm:

**Pass 1 — Classify and mark**

Each whitespace-separated token is first stripped of trailing punctuation (which becomes a separate `TokenPunctuation` token). The word portion is then checked in order: article → verb + noun combined check → word map. For dual-class base forms (a word that appears in both `baseVerbs` and `baseNouns` and has no inflection to self-resolve), the token is marked as `tokenAmbiguous` (an internal sentinel) with both `VerbInfo` and `NounInfo` stashed for Pass 2.

Inflected forms self-resolve: `"committed"` always resolves as verb (past tense), `"commits"` as noun (plural), regardless of dual-class membership.

**Pass 2 — Disambiguate**

For each `tokenAmbiguous` token, `scoreAmbiguous()` evaluates seven weighted signals and `resolveToken()` converts scores to a classification with confidence values.

| Signal | Weight | Description |
|--------|--------|-------------|
| `noun_determiner` | 0.35 | Preceding token is in the noun determiner list (articles, possessives, quantifiers) |
| `verb_auxiliary` | 0.25 | Preceding token is a modal, auxiliary, or infinitive marker |
| `following_class` | 0.15 | Next token is article/noun (→ verb signal) or verb (→ noun signal) |
| `sentence_position` | 0.10 | Sentence-initial position signals imperative verb |
| `verb_saturation` | 0.10 | A confident verb already exists in the same clause |
| `inflection_echo` | 0.03 | Another token uses the same base in an inflected form |
| `default_prior` | 0.02 | Always fires as verb signal (tiebreaker) |

When total signal weight is below 0.10 (only the default prior fired), confidence is floored at 0.55/0.45 rather than deriving a misleading 1.0.

The `verb_saturation` signal scans within clause boundaries only. Clause boundaries are defined as punctuation tokens and coordinating/subordinating conjunctions (`and`, `or`, `but`, `because`, `when`, `while`, `if`, `then`, `so`).

Confidence values flow into imprints: dual-class tokens contribute to both verb and noun distributions weighted by `Confidence` and `AltConf`, preserving uncertainty for downstream scoring.

**Token type**

```go
type Token struct {
    Raw        string
    Lower      string
    Type       TokenType        // TokenVerb, TokenNoun, TokenArticle, TokenWord, TokenPunctuation, TokenUnknown
    Confidence float64          // 0.0–1.0
    AltType    TokenType        // Runner-up (dual-class only)
    AltConf    float64
    VerbInfo   VerbMatch
    NounInfo   NounMatch
    WordCat    string
    ArtType    string
    PunctType  string
    Signals    *SignalBreakdown // Non-nil only with WithSignals() option
}
```

**Options**

`WithSignals()` allocates `SignalBreakdown` on ambiguous tokens, providing per-component scoring for diagnostic use. It adds approximately 36% latency and 3x allocations versus plain tokenise; keep it off in production paths.

`WithWeights(map[string]float64)` overrides signal weights without code changes, useful for calibration experiments.

**Benchmark baselines (M3 Ultra, arm64)**

| Operation | ns/op | allocs |
|-----------|-------|--------|
| `Tokenise` (3 words) | 639 | 8 |
| `Tokenise` (12 words) | 2859 | 14 |
| `Tokenise` (dual-class) | 1657 | 9 |
| `Tokenise` + `WithSignals` | 2255 | 28 |
| `NewImprint` | 648 | 10 |
| `Imprint.Similar` | 516 | 0 |
| `Multiplier.Expand` | 3609 | 63 |

Tokenise scales approximately linearly at 200–240 ns/word, giving approximately 350K sentences/second single-threaded.

---

## GrammarImprint

`GrammarImprint` is a low-dimensional grammar feature vector computed from a token slice. It is a lossy projection: content is discarded, grammatical structure is preserved.

```go
type GrammarImprint struct {
    VerbDistribution   map[string]float64 // verb base → normalised frequency
    TenseDistribution  map[string]float64 // "past"/"gerund"/"base" → ratio
    NounDistribution   map[string]float64 // noun base → normalised frequency
    PluralRatio        float64            // proportion of plural nouns
    DomainVocabulary   map[string]int     // gram.word category → hit count
    ArticleUsage       map[string]float64 // "definite"/"indefinite" → ratio
    PunctuationPattern map[string]float64 // "label"/"progress"/"question" → ratio
    TokenCount         int
    UniqueVerbs        int
    UniqueNouns        int
}
```

`NewImprint(tokens []Token) GrammarImprint` accumulates counts weighted by token confidence, then normalises all frequency maps to sum to 1.0 via L1 normalisation.

**`Similar(b GrammarImprint) float64`**

Returns weighted cosine similarity (0.0–1.0) between two imprints:

| Component | Weight |
|-----------|--------|
| `VerbDistribution` | 0.30 |
| `NounDistribution` | 0.25 |
| `TenseDistribution` | 0.20 |
| `ArticleUsage` | 0.15 |
| `PunctuationPattern` | 0.10 |

Components where both maps are empty are excluded from the weighted average (no signal contributed). `Similar` is zero-alloc (516 ns/op on M3 Ultra), making it suitable for high-volume comparison.

---

## Reference Distributions

`BuildReferences(tokeniser, samples) (*ReferenceSet, error)` takes a slice of `ClassifiedText` (text + domain label), tokenises each, builds an imprint, and aggregates by domain.

**Centroid computation**

For each domain, `computeCentroid()` accumulates all imprint map fields using `addMap()`, then:
- Normalises the accumulated maps via `normaliseMap()` (L1 norm, sums to 1.0)
- Averages scalar fields (`PluralRatio`, `TokenCount`, `UniqueVerbs`, `UniqueNouns`)

The result is a single centroid `GrammarImprint` representing the grammatical centre of mass for the domain.

**Variance computation**

`computeVariance()` computes sample variance for each key across all imprints in the domain. Keys are prefixed by component (`"verb:"`, `"tense:"`, `"noun:"`, `"article:"`, `"punct:"`) to form a flat variance map. Requires at least two samples; returns nil otherwise.

**Distance metrics**

`ReferenceSet.Compare(imprint)` computes three distance metrics between an imprint and each domain centroid:

| Metric | Implementation | Notes |
|--------|---------------|-------|
| Cosine similarity | `Similar()` | 0.0–1.0, higher is closer |
| KL divergence | Symmetric (Jensen-Shannon style) | 0.0+, lower is closer; epsilon-smoothed at 1e-10 |
| Mahalanobis | Variance-normalised squared distance | Falls back to Euclidean (unit variance) when variance is unavailable |

The same component weights (verb 0.30, noun 0.25, tense 0.20, article 0.15, punct 0.10) are applied to KL divergence and Mahalanobis computations.

**Classification**

`ReferenceSet.Classify(imprint) ImprintClassification` ranks domains by cosine similarity and returns the best match. Confidence is the margin between the best and second-best similarity scores (0.0 when there is only one domain).

---

## Anomaly Detection

`ReferenceSet.DetectAnomalies(tokeniser, samples) ([]AnomalyResult, *AnomalyStats)` compares domain labels from an external classifier (e.g. a 1B language model) against imprint-based classification.

For each sample:
1. Tokenise the text
2. Build a `GrammarImprint`
3. Classify against reference centroids via `Classify()`
4. Compare the model's domain label against the imprint's domain

A sample is flagged as an anomaly when the two domains disagree. The aggregate `AnomalyStats` reports total count, anomaly count, rate, and a per-pair breakdown (`"technical->creative": 4`).

Anomaly detection serves as a training signal: mislabelled samples from a 1B model can be flagged for human review or 27B verification. Validated behaviour: a creative sentence labelled as technical is correctly identified as an anomaly with a measurable confidence margin.

---

## 1B Pre-Sort Pipeline

`ClassifyCorpus(ctx, model, input, output, opts...) (*ClassifyStats, error)` reads JSONL from `input`, batch-classifies each record through a `go-inference` `TextModel`, and writes JSONL with a `domain_1b` field added to `output`.

Architecture:
- Configurable via `WithBatchSize(n)`, `WithPromptField(field)`, `WithPromptTemplate(tmpl)`
- Single-token generation (`WithMaxTokens(1)`) at temperature 0.05 for classification
- `mapTokenToDomain(token)` maps model output to one of `{technical, creative, ethical, casual, unknown}` via exact match and known BPE fragment prefixes
- Mock-friendly via the `inference.TextModel` interface

Observed throughput: 80 prompts/second on M3 Ultra with Gemma3-1B (4-bit quantised), steady-state approaching 152 prompts/second as batch pipeline warms up.

---

## 1B vs 27B Calibration

`CalibrateDomains(ctx, modelA, modelB, samples, opts...) (*CalibrationStats, error)` classifies a corpus with two models sequentially (A first, then B), then computes agreement and accuracy metrics.

`CalibrationStats` includes:
- Total and agreed counts, agreement rate
- Per-model domain distribution
- Confusion pairs in `"domainA->domainB"` format
- Per-model accuracy against ground-truth labels (when provided)
- Per-model classification duration

The 500-sample integration corpus mixes 220 ground-truth sentences (55 per domain) with 280 unlabelled diverse sentences.

---

## Multiplier

`Multiplier.Expand(text string) []string` generates deterministic grammatical variants for training data augmentation with zero API calls.

For each verb in the tokenised text, it produces past, gerund, and base tense variants. For each noun, it toggles plural/singular. Combinations of verb transform and noun transform are also emitted. All variants are deduplicated. Case preservation (`preserveCase()`) maintains the capitalisation pattern of the original token — all-caps, title-case, and lower-case are all handled.

---

## Multi-Language Support

The grammar engine is language-parametric. Every function that produces or classifies text uses the current language to look up `GrammarData`. Grammar tables must be loaded via `SetGrammarData(lang, data)` before use.

**French grammar tables** (`locales/fr.json`) include:
- 50 verb conjugations with `past` (passé composé participial form) and `gerund` (présent participe)
- 24 gendered nouns with `"m"` or `"f"` gender fields
- Gendered articles: `by_gender: {"m": "le", "f": "la"}`; indefinite: `"un"` (both vowel and default — French does not distinguish)
- Punctuation: `label: " :"` (space before colon, per French typographic convention), `progress: "..."`

**Tier 2 and tier 3 fallbacks are English-only.** The `irregularVerbs` and `irregularNouns` Go maps and the regular morphology rules apply English patterns. Non-English languages must therefore provide comprehensive `gram.verb` and `gram.noun` tables in JSON.

**Disambiguation signal lists** in `gram.signal` are per-language. The tokeniser's `buildSignalIndex()` loads each list independently. If a list is absent or empty, the tokeniser falls back to hardcoded English defaults for that signal only.

**Plural rules** for CLDR plural categories are registered per language code in the `pluralRules` map. Supported: en, de, fr, es, ru, pl, ar, zh, ja, ko (with regional variants).

**RTL detection** is provided for Arabic, Hebrew, Persian, Urdu, and related codes via `rtlLanguages`.

See `docs/grammar-table-spec.md` for the full JSON schema.

---

## Irreversibility and Round-Trip Property

The reversal engine maintains the **round-trip property**: for any base form `b`, if `PastTense(b)` produces inflected form `f`, then `MatchVerb(f)` must recover `b`. This property is enforced by round-trip verification in tier 3 of all `Match*` functions: candidate bases are only accepted if the forward function reproduces the original inflected form. Tests in `reversal/roundtrip_test.go` validate this property for all irregular verbs and a sample of regular patterns.

The imprint is inherently lossy — it discards lexical content and retains only grammatical structure. The round-trip property applies to the tokeniser, not to the imprint.
