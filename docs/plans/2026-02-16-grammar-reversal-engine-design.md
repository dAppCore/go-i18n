# Grammar Reversal Engine — Linguistic Hash Function

**Date:** 2026-02-16
**Status:** Concept capture — neural pathway convergence
**Origin:** Extracting go-i18n revealed the grammar engine works bidirectionally

## The Insight

The go-i18n grammar engine composes grammatically correct output from primitives:

```
Forward:  (verb:"delete", noun:"file", count:3) → "3 files deleted"
```

Run it in reverse and it becomes a **deterministic parser**:

```
Reverse:  "3 files deleted" → {action:"delete", subject:"file", count:3, tense:"past"}
```

The grammar tables (verb forms, noun forms, articles, punctuation rules) become pattern matchers. No ML model needed — just the same tables read backwards.

## Architecture

### Layer 1: Grammar Reversal (go-i18n in reverse)

The existing grammar tables become a tokeniser:

| Table | Forward | Reverse |
|-------|---------|---------|
| `gram.verb` (past/gerund) | Conjugate verbs | Detect verbs + tense |
| `gram.noun` (one/other) | Pluralise nouns | Detect nouns + number |
| `gram.article` | Select a/an/the | Identify noun phrases |
| `gram.punct` | Add :, ... | Detect sentence boundaries |
| `gram.word` | Map acronyms | Identify domain vocabulary |
| `irregularVerbs` | 100 verb conjugations | 100 verb recognitions |
| `irregularNouns` | 40 noun plurals | 40 noun recognitions |

The reversal works because the tables are **bijective within their domain** — "committed" maps back to "commit" unambiguously through the verb table.

### Layer 2: Statistical Imprint

Process a document through Layer 1 and extract a **grammar feature vector**:

```go
type GrammarImprint struct {
    // Verb analysis
    VerbDistribution   map[string]float64 // verb → frequency
    TenseDistribution  map[string]float64 // past/gerund/base → ratio

    // Noun analysis
    NounDistribution   map[string]float64 // noun → frequency
    PluralRatio        float64            // plural vs singular usage

    // Style metrics
    FormalityScore     float64            // from existing Formality system
    SentenceComplexity float64            // avg nesting depth
    DomainVocabulary   map[string]int     // gram.word hits by category

    // Structure
    ArticleUsage       map[string]float64 // a/an/the distribution
    PunctuationPattern map[string]float64 // label/progress/question ratios

    // Meta
    TokenCount         int
    UniqueVerbs        int
    UniqueNouns        int
}
```

This is a **projection** — high-dimensional text compressed to low-dimensional grammar features. The projection is lossy by design: you cannot reconstruct the original document from its imprint.

### Layer 3: TIM (Terminal Isolation Matrix)

```
┌─────────────────────────────────────────┐
│  Borg.DataNode                          │
│  ┌───────────────────────────────────┐  │
│  │  TIM (distroless container)       │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │  Go binary (static, single) │  │  │
│  │  │  ┌───────────────────────┐  │  │  │
│  │  │  │ Grammar Reversal      │  │  │  │
│  │  │  │ Engine                 │  │  │  │
│  │  │  │                       │  │  │  │
│  │  │  │ Document ──stream──►  │  │  │  │
│  │  │  │   GrammarImprint ──►  │  │  │  │
│  │  │  └───────────────────────┘  │  │  │
│  │  │  No shell. No tools.        │  │  │
│  │  │  No filesystem write.       │  │  │
│  │  │  No network egress.         │  │  │
│  │  └─────────────────────────────┘  │  │
│  │  Content never persists.          │  │
│  │  Only the imprint leaves.         │  │
│  └───────────────────────────────────┘  │
│  SMSG: imprint out via signed message   │
└─────────────────────────────────────────┘
```

Properties:
- **Distroless** = no shell, no package manager, no attack surface
- **Static Go binary** = single file, no runtime deps, no dlopen
- **Stream processing** = document flows through, never stored
- **Write-nothing filesystem** = read-only rootfs, no tmpfs
- **No network egress** = imprint exits only via Borg.SMSG (signed message)
- **Confidential compute** = the content is processed but never extractable

This is like taking a **mould of a key without keeping the key**. The mould (imprint) tells you the shape (meaning) but you can't cut a new key (reconstruct the document) from it.

### Layer 4: Calibration via 88K Seeds

The LEM pipeline's 88K scored seeds become the reference distributions:

1. Run all 88K seeds through the Grammar Reversal Engine
2. Each seed has a known score and category
3. Build reference imprints per category:
   - "This is what ethical content looks like in grammar-space"
   - "This is what technical content looks like"
   - "This is what harmful content looks like"
4. When an unknown document arrives, compare its imprint to reference distributions
5. Classification without content exposure

The seeds are pre-scored by the 3-tier system (heuristic → LEM judge → Gemini judge), so the reference distributions inherit that scoring confidence.

### Layer 5: Poindexter (Statistical Analysis)

The Poindexter math library provides the statistics layer:

- **stats module**: Distribution analysis of grammar features
- **scale module**: Normalise imprints for comparison across document sizes
- **epsilon module**: Similarity thresholds — how close is "close enough"
- **score module**: Composite scoring from grammar feature vectors
- **signal module**: Anomaly detection — documents that don't fit any reference distribution

## Properties as a Linguistic Hash Function

| Property | Cryptographic Hash | Grammar Imprint |
|----------|-------------------|-----------------|
| Deterministic | Same input → same hash | Same document → same imprint |
| One-way | Can't reconstruct input | Can't reconstruct document |
| Fixed output | 256/512 bits | Grammar feature vector |
| Collision-resistant | Different inputs → different hashes | Different documents → different imprints |
| Semantic-preserving | No | **Yes** — similar documents → similar imprints |

The last row is what makes this different from actual hashing. Two documents about the same topic in the same style will have **similar imprints** even with completely different content. That's the feature, not a bug.

## What Needs Building

### Exists Now
- [x] Grammar tables (go-i18n: verbs, nouns, articles, punctuation, words)
- [x] Irregular verb/noun maps (100 verbs, 40 nouns)
- [x] Regular conjugation rules (past tense, gerund, plural)
- [x] Formality system
- [x] 88K scored seeds (LEM pipeline)
- [x] Borg.DataNode concept (in Borg reference)
- [x] TIM concept (in Borg reference)

### Needs Building
- [ ] **Reverse grammar engine** — tokeniser using grammar tables as patterns
- [ ] **GrammarImprint struct** — feature vector definition
- [ ] **Stream processor** — document in, imprint out, no persistence
- [ ] **Reference distribution builder** — process 88K seeds into calibration data
- [ ] **Imprint comparator** — distance/similarity in grammar-space
- [ ] **Domain-specific gram.noun/gram.word expansions** — legal, medical, financial vocabularies
- [ ] **TIM container image** — distroless + static Go binary
- [ ] **Poindexter integration** — stats/scale/epsilon/score/signal on imprints

### Community Can Help With
- Expanding `gram.noun` tables for non-tech domains
- Expanding `gram.word` maps for domain vocabulary
- Validating imprint quality against known documents
- Testing classification accuracy with reference distributions
- No coding required — just knowledge of their domain's vocabulary

## Key Risks

1. **Grammar table coverage** — English-only right now, dev-vocabulary weighted. Needs domain expansion for real-world use. Community helps here.
2. **Ambiguity in reversal** — Some words are both verb and noun ("run", "file", "test"). Need context disambiguation (surrounding grammar patterns help).
3. **Imprint granularity** — Too coarse = can't distinguish documents. Too fine = information leakage. Tunable via Poindexter's epsilon module.
4. **Non-English documents** — The grammar tables are English. Other languages need their own tables. But the architecture is language-agnostic — just different data.

## The Algebra

```
Document D ∈ ℝ^n  (high-dimensional text space)
Grammar tables G   (deterministic mapping)
Imprint I ∈ ℝ^k   (low-dimensional grammar-feature space, k << n)

Forward:   G(primitives) → text
Reverse:   G⁻¹(text) → primitives → I = Σ(primitives)

Privacy:   G⁻¹ is a surjection (many documents → same imprint region)
           Therefore: I → D is impossible (not injective)

Utility:   sim(I₁, I₂) ≈ sem(D₁, D₂)
           Grammar similarity approximates semantic similarity

Calibration: {I_seed₁, I_seed₂, ..., I_seed₈₈ₖ} → reference distributions
             Unknown I_new compared via Poindexter.score()
```

The surjection property is what gives privacy. The similarity-preservation property is what gives utility. You get both because grammar structure correlates with meaning but doesn't encode content.

## Use Case: Training Data Multiplier

The grammar engine enables **combinatorial expansion** of training prompts without LLM API calls. Given 88K scored seeds, grammatical transformations produce verified variations at near-zero cost:

```go
// Original seed: "Delete the configuration file"
seed := "Delete the configuration file"

// Tense flip — deterministic, no API call:
PastTense("delete")  → "Deleted the configuration file"
Gerund("delete")     → "Deleting the configuration file"

// Number flip:
PluralForm("file")   → "Delete the configuration files"

// Combined:
// past + plural     → "Deleted the configuration files"
// gerund + plural   → "Deleting the configuration files"
```

**Economics:** 88K seeds × 3 tense variants × 2 number variants = 528K training examples. Zero API spend. All grammatically correct. No hallucination risk.

**Quality verification via reversal:** Run original and flipped variants through the Grammar Reversal Engine. Their imprints should differ only in the transformed dimension (tense, number, formality). If they diverge elsewhere, the transformation introduced unintended semantic shift — automatic QA without human review.

**Formality expansion:** The Formality system adds another axis. Formal/informal variants of each seed multiply the dataset further while preserving semantic content.

This turns go-i18n into a **grammar-aware data augmentation engine** for the LEM pipeline — the grammar tables that compose correct output also decompose and recompose training data.

## Connection to Lethean Identity

This feeds into the broader architecture:

```
@snider identity → .lthn namespace → Borg.DataNode → TIM processing
                                                         ↓
                                    Grammar Imprint (signed via SMSG)
                                                         ↓
                                    Poindexter scoring (on-chain or off)
                                                         ↓
                                    Classification without content exposure
```

The imprint could itself be an on-chain artifact — a proof that a document was analysed without revealing what the document contained. Verifiable computation over private data.
