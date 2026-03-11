---
title: Reversal Engine
description: Decomposes inflected text back to base forms with grammatical metadata.
---

# Reversal Engine

The reversal engine (`reversal/` package) converts inflected text back to base forms with grammatical metadata. It powers [GrammarImprint](grammar-imprint.md), the [Multiplier](multiplier.md), and the classification pipeline.

The forward engine maps base forms to inflected forms:

```
PastTense("delete") -> "deleted"
Gerund("run")       -> "running"
```

The reversal engine reads those same tables backwards:

```
MatchVerb("deleted")  -> {Base: "delete", Tense: "past"}
MatchVerb("running")  -> {Base: "run",    Tense: "gerund"}
```

## 3-Tier Lookup Strategy

All reverse lookups follow the same three-tier pattern, stopping at the first match:

| Tier | Source | Example |
|------|--------|---------|
| 1 | JSON grammar data (`gram.verb.*`, `gram.noun.*`) | "committed" -> past of "commit" |
| 2 | Irregular verb/noun Go maps (`IrregularVerbs()`) | "went" -> past of "go" |
| 3 | Regular morphology rules + round-trip verification | "processed" -> past of "process" |

**JSON takes precedence** -- if a verb appears in both `en.json` and the irregular Go map, the JSON form wins. This lets locale files override built-in rules.

## Creating a Tokeniser

```go
// English tokeniser (default)
tok := reversal.NewTokeniser()

// Language-specific tokeniser
tok := reversal.NewTokeniserForLang("en")

// With options
tok := reversal.NewTokeniser(
    reversal.WithSignals(),                // Enable detailed SignalBreakdown on ambiguous tokens
    reversal.WithWeights(customWeights),   // Override disambiguation signal weights
)
```

The tokeniser builds six inverse lookup maps at construction time:

| Map | Direction | Example |
|-----|-----------|---------|
| `pastToBase` | "deleted" -> "delete" | Inverse of `gram.verb.*.past` |
| `gerundToBase` | "deleting" -> "delete" | Inverse of `gram.verb.*.gerund` |
| `baseVerbs` | "delete" -> true | All known verb bases |
| `pluralToBase` | "files" -> "file" | Inverse of `gram.noun.*.other` |
| `baseNouns` | "file" -> true | All known noun bases |
| `words` | "url" -> "URL" | Domain vocabulary |

## Token Types

```go
TokenUnknown      // Unrecognised word
TokenVerb         // Matched verb (VerbInfo populated)
TokenNoun         // Matched noun (NounInfo populated)
TokenArticle      // "a", "an", "the"
TokenWord         // Domain word from gram.word map
TokenPunctuation  // "...", "?", "!", ":", ";", ","
```

## Tokenise()

Splits text on whitespace and classifies each word using a two-pass algorithm.

```go
tok := reversal.NewTokeniser()
tokens := tok.Tokenise("Deleted the configuration files successfully")
// tokens[0]: Type=TokenVerb,    VerbInfo={Base:"delete", Tense:"past"}
// tokens[1]: Type=TokenArticle, ArtType="definite"
// tokens[2]: Type=TokenNoun,    NounInfo={Base:"configuration", Plural:false}
// tokens[3]: Type=TokenNoun,    NounInfo={Base:"file", Plural:true}
// tokens[4]: Type=TokenUnknown
```

### Pass 1: Classify and Mark

Each word is checked in this priority order:

1. **Article** -- "a", "an", "the"
2. **Both verb and noun** -- if the word is in both tables and is a known dual-class word, check for self-resolving inflections (inflected verb forms resolve as verbs, plural noun forms resolve as nouns). Base forms are marked as ambiguous for Pass 2.
3. **Verb only** -- base forms, past tense, gerunds
4. **Noun only** -- base forms, plurals
5. **Word** -- domain vocabulary from `gram.word`
6. **Unknown** -- fallback

### Pass 2: Resolve Ambiguous Tokens

Dual-class base forms (words like "commit", "test", "run" that are both verbs and nouns) are resolved using seven weighted disambiguation signals:

| Signal | Weight | Description |
|--------|--------|-------------|
| `noun_determiner` | 0.35 | Preceding token is "the", "a", "my", "this", etc. |
| `verb_auxiliary` | 0.25 | Preceding token is "is", "was", "will", "can", etc. |
| `following_class` | 0.15 | Next token is article/noun (verb signal) or verb (noun signal) |
| `sentence_position` | 0.10 | Sentence-initial position suggests imperative (verb signal) |
| `verb_saturation` | 0.10 | Confident verb already exists in clause (noun signal) |
| `inflection_echo` | 0.03 | Another token shares the same base in inflected form |
| `default_prior` | 0.02 | Always fires as verb signal (tiebreaker) |

The winning classification gets confidence = its score / total score. When total score is below 0.10 (only default prior fired), a low-information confidence floor of 0.55/0.45 is used.

### Token Struct

```go
type Token struct {
    Raw        string          // Original text as it appeared
    Lower      string          // Lowercased form
    Type       TokenType       // Classification
    Confidence float64         // 0.0-1.0 classification confidence
    AltType    TokenType       // Runner-up classification (dual-class only)
    AltConf    float64         // Runner-up confidence
    VerbInfo   VerbMatch       // Populated when Type or AltType is TokenVerb
    NounInfo   NounMatch       // Populated when Type or AltType is TokenNoun
    WordCat    string          // Category key when Type is TokenWord
    ArtType    string          // "definite" or "indefinite" when Type is TokenArticle
    PunctType  string          // "progress", "question", etc. when Type is TokenPunctuation
    Signals    *SignalBreakdown // Non-nil only when WithSignals() option is set
}
```

## Matching Methods

### MatchVerb(word) -> (VerbMatch, bool)

```go
type VerbMatch struct {
    Base  string // "delete"
    Tense string // "past", "gerund", or "base"
    Form  string // Original inflected form
}
```

**Tier 1**: Check `baseVerbs[word]` (is it a known base verb?)
**Tier 2**: Check `pastToBase[word]` and `gerundToBase[word]` (inverse maps)
**Tier 3**: Apply reverse morphology rules, then round-trip verify

### MatchNoun(word) -> (NounMatch, bool)

```go
type NounMatch struct {
    Base   string // Singular form
    Plural bool   // Whether the matched form was plural
    Form   string // Original form
}
```

Same 3-tier pattern with `pluralToBase` inverse map and `reverseRegularPlural()`.

### MatchWord(word) -> (string, bool)

Case-insensitive lookup in the domain vocabulary. Returns the category key.

### MatchArticle(word) -> (string, bool)

Returns `"indefinite"` or `"definite"`.

## Reverse Morphology Rules

When tiers 1 and 2 produce no match, the engine generates candidate base forms by reversing regular English morphology rules. Multiple candidates are generated and then verified by round-tripping through the forward functions.

### Past Tense Reversal

| Pattern | Rule | Example |
|---------|------|---------|
| consonant + `ied` | -> consonant + `y` | copied -> copy |
| doubled consonant + `ed` | -> single consonant | stopped -> stop |
| stem + `d` (stem ends in `e`) | -> stem | created -> create |
| stem + `ed` | -> stem | walked -> walk |

### Gerund Reversal

| Pattern | Rule | Example |
|---------|------|---------|
| `-ying` | -> `-ie` | dying -> die |
| doubled consonant + `ing` | -> single consonant | stopping -> stop |
| direct `-ing` strip | -> stem | walking -> walk |
| add `-e` back | -> stem + `e` | creating -> create |

### Plural Reversal

| Pattern | Rule | Example |
|---------|------|---------|
| consonant + `-ies` | -> consonant + `y` | entries -> entry |
| `-ves` | -> `-f` or `-fe` | wolves -> wolf, knives -> knife |
| sibilant + `-es` | -> sibilant | processes -> process |
| `-s` | -> stem | servers -> server |

## Round-Trip Verification

When tier 3 produces multiple candidate base forms, `bestRoundTrip()` selects the best one by applying the forward function to each candidate and checking if it reproduces the original inflected form. Only verified candidates are accepted.

When multiple candidates pass verification (ambiguity), selection priority is:

1. **Known base verb/noun** -- candidate exists in the grammar index
2. **VCe pattern** -- candidate ends in vowel-consonant-e (the "magic e" pattern found in real English verbs like "delete", "create", "use"). This avoids phantom verbs like "walke" or "processe" which have consonant-consonant-e endings.
3. **No trailing e** -- default morphology path
4. **First match** -- final tiebreaker

## Disambiguation Statistics

```go
stats := reversal.DisambiguationStatsFromTokens(tokens)
// stats.TotalTokens     -- total token count
// stats.AmbiguousTokens -- count of dual-class tokens
// stats.ResolvedAsVerb  -- how many resolved as verb
// stats.ResolvedAsNoun  -- how many resolved as noun
// stats.AvgConfidence   -- average confidence across all classified tokens
// stats.LowConfidence   -- count where confidence < 0.7
```

## Signal Breakdown

Enable `WithSignals()` to get detailed scoring on ambiguous tokens:

```go
tok := reversal.NewTokeniser(reversal.WithSignals())
tokens := tok.Tokenise("the commit failed")

for _, t := range tokens {
    if t.Signals != nil {
        fmt.Printf("verb=%.2f noun=%.2f\n", t.Signals.VerbScore, t.Signals.NounScore)
        for _, c := range t.Signals.Components {
            fmt.Printf("  %s: weight=%.2f value=%.1f contrib=%.3f (%s)\n",
                c.Name, c.Weight, c.Value, c.Contrib, c.Reason)
        }
    }
}
```
