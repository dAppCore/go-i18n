---
title: GrammarImprint
description: Lossy grammar fingerprinting for semantic verification without decryption.
---

# GrammarImprint

GrammarImprint is a **lossy feature vector projection** that converts text into a grammar fingerprint. Content is intentionally discarded -- only grammatical structure is preserved. Two texts with similar grammar produce similar imprints, regardless of subject matter.

This is the foundation for the Poindexter classification pipeline, the LEM scoring system, and RFC-023 reverse steganography (semantic verification without decryption).

## Creating an Imprint

```go
tok := reversal.NewTokeniser()
tokens := tok.Tokenise("Deleted the configuration files successfully")
imp := reversal.NewImprint(tokens)

// Result:
// TokenCount:         5
// UniqueVerbs:        1 (delete)
// UniqueNouns:        2 (configuration, file)
// TenseDistribution:  {"past": 1.0}
// PluralRatio:        0.5
// ArticleUsage:       {"definite": 1.0}
```

## The Imprint Struct

```go
type GrammarImprint struct {
    VerbDistribution   map[string]float64 // verb base -> normalised frequency
    TenseDistribution  map[string]float64 // "past"/"gerund"/"base" -> ratio
    NounDistribution   map[string]float64 // noun base -> normalised frequency
    PluralRatio        float64            // proportion of plural nouns (0.0-1.0)
    DomainVocabulary   map[string]int     // gram.word category -> hit count
    ArticleUsage       map[string]float64 // "definite"/"indefinite" -> ratio
    PunctuationPattern map[string]float64 // "label"/"progress"/"question" -> ratio
    TokenCount         int
    UniqueVerbs        int
    UniqueNouns        int
}
```

All frequency maps (`VerbDistribution`, `TenseDistribution`, `NounDistribution`, `ArticleUsage`, `PunctuationPattern`) are **normalised to sum to 1.0** after token collection. This converts raw counts into probability distributions.

### Dual-Class Confidence

When a token is classified as verb with 0.96 confidence but has noun as the alternative at 0.04, the imprint distributes accordingly:
- 0.96 flows into `VerbDistribution` and `TenseDistribution`
- 0.04 flows into `NounDistribution`

This preserves uncertainty -- similar texts produce similar imprints even when individual token classifications wobble at the boundary.

## Similarity Calculation

`Similar(other)` returns 0.0-1.0 using weighted cosine similarity across five distribution dimensions:

```go
sim := imprintA.Similar(imprintB)
```

| Dimension | Weight | Rationale |
|-----------|--------|-----------|
| Verb distribution | **0.30** | Most domain-specific signal |
| Noun distribution | **0.25** | Entity focus indicates topic area |
| Tense distribution | **0.20** | Temporal patterns distinguish narrative from imperative |
| Article usage | **0.15** | Grammatical style (technical docs use more "the") |
| Punctuation pattern | **0.10** | Minor structural signal |

**Key behaviours:**
- Same text -> similarity = 1.0
- Similar grammar, different content -> similarity 0.5-0.9
- Different grammatical structure -> similarity < 0.3
- Empty imprints -> similarity = 1.0 (no signal = no difference)
- Dimensions with no data in either imprint are skipped (do not dilute score)

### Cosine Similarity

For each distribution pair, the engine computes:

```
dot = sum(a[k] * b[k])  for all keys in union(a, b)
|a| = sqrt(sum(a[k]^2))
|b| = sqrt(sum(b[k]^2))
similarity = dot / (|a| * |b|)
```

## The Lossy Property

GrammarImprint is **intentionally lossy**:

- "Delete the configuration file" and "Remove the deployment artifact" can produce similar imprints (both: imperative verb + definite article + noun)
- The actual words do not matter -- only their grammatical roles
- This is the design goal: grammar structure is a privacy-preserving proxy for semantic similarity

## Reference Distributions

The `ReferenceSet` holds per-domain centroid imprints built from classified samples. Use this for domain classification without an LLM.

### Building References

```go
samples := []reversal.ClassifiedText{
    {Text: "Delete the file from cache", Domain: "technical"},
    {Text: "The sunset painted golden light", Domain: "creative"},
    // ... more samples
}

tok := reversal.NewTokeniser()
refs, err := reversal.BuildReferences(tok, samples)
```

`BuildReferences` tokenises each sample, computes its imprint, groups by domain, then averages each group into a centroid. Per-key variance is also computed for Mahalanobis distance.

### Classifying New Text

```go
tokens := tok.Tokenise("Remove the old configuration")
imp := reversal.NewImprint(tokens)

result := refs.Classify(imp)
// result.Domain     -- best-matching domain (e.g. "technical")
// result.Confidence -- margin between best and second-best similarity
// result.Distances  -- per-domain distance metrics
```

### Distance Metrics

`Compare()` returns three distance measures per domain:

```go
type DistanceMetrics struct {
    CosineSimilarity float64 // 0.0-1.0 (1.0 = identical)
    KLDivergence     float64 // 0.0+ (0.0 = identical), symmetric
    Mahalanobis      float64 // 0.0+ (0.0 = identical), variance-normalised
}
```

- **Cosine similarity** is the primary classification metric (used by `Classify()`)
- **KL divergence** uses weighted symmetric (Jensen-Shannon style) computation across the five distribution dimensions
- **Mahalanobis distance** normalises by per-key variance, falling back to Euclidean when variance data is unavailable

## Anomaly Detection

Compare 1B model domain tags against imprint-based classification to find mismatches:

```go
results, stats := refs.DetectAnomalies(tok, samples)

// stats.Total     -- samples processed
// stats.Anomalies -- count where model and imprint disagree
// stats.Rate      -- anomalies / total
// stats.ByPair    -- "technical->creative": count

for _, r := range results {
    if r.IsAnomaly {
        fmt.Printf("Model says %s, imprint says %s: %s\n",
            r.ModelDomain, r.ImprintDomain, r.Text)
    }
}
```

Anomalies are either misclassified by the model (training signal) or genuinely cross-domain text (flagged for review).

## Pipeline Integration

### Phase 2a: 1B Pre-Classification

The `ClassifyCorpus` function reads JSONL, batch-classifies through a 1B model, and writes JSONL with `domain_1b` field added:

```go
stats, err := i18n.ClassifyCorpus(ctx, model, input, output,
    i18n.WithBatchSize(8),
    i18n.WithPromptField("prompt"),
)
// stats.Total, stats.ByDomain, stats.PromptsPerSec
```

### Phase 2b: Model Calibration

`CalibrateDomains` classifies samples with two models (e.g. 1B vs 27B) and computes agreement:

```go
stats, err := i18n.CalibrateDomains(ctx, smallModel, largeModel, samples)
// stats.AgreementRate, stats.ConfusionPairs, stats.AccuracyA, stats.AccuracyB
```

### Validation

Grammar table entries can be validated against a language model:

```go
result, err := i18n.ValidateArticle(ctx, model, "SSH", "an")
// result.Valid: true (model confirms "an SSH" is correct)

result, err := i18n.ValidateIrregular(ctx, model, "go", "past", "went")
// result.Valid: true
```
