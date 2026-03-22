---
title: go-i18n Grammar Engine
description: Grammar-aware internationalisation for Go with forward composition and reverse decomposition.
---

# go-i18n Grammar Engine

`dappco.re/go/core/i18n` is a **grammar engine** for Go. Unlike flat key-value translation systems, it composes grammatically correct output from verbs, nouns, and articles -- and can reverse the process, decomposing inflected text back into base forms with grammatical metadata.

This is the foundation for the Poindexter classification pipeline and the LEM scoring system.

## Architecture

| Layer | Package | Purpose |
|-------|---------|---------|
| Forward | Root (`i18n`) | Compose grammar-aware messages: `T()`, `PastTense()`, `Gerund()`, `Pluralize()`, `Article()` |
| Reverse | `reversal/` | Decompose text back to base forms with tense/number metadata |
| Imprint | `reversal/` | Lossy feature vector projection for grammar fingerprinting |
| Multiply | `reversal/` | Deterministic training data augmentation |
| Classify | Root (`i18n`) | 1B model domain classification pipeline |
| Data | `locales/` | Grammar tables (JSON) -- only `gram.*` data |

## Quick Start

```go
import i18n "dappco.re/go/core/i18n"

// Initialise the default service (uses embedded en.json)
svc, err := i18n.New()
i18n.SetDefault(svc)

// Forward composition
i18n.T("i18n.progress.build")          // "Building..."
i18n.T("i18n.done.delete", "cache")    // "Cache deleted"
i18n.T("i18n.count.file", 5)           // "5 files"
i18n.PastTense("commit")               // "committed"
i18n.Article("SSH")                     // "an"
```

```go
import "dappco.re/go/core/i18n/reversal"

// Reverse decomposition
tok := reversal.NewTokeniser()
tokens := tok.Tokenise("Deleted the configuration files")

// Grammar fingerprinting
imp := reversal.NewImprint(tokens)
sim := imp.Similar(otherImp) // 0.0-1.0

// Training data augmentation
m := reversal.NewMultiplier()
variants := m.Expand("Delete the file") // 4-7 grammatical variants
```

## Documentation

- [Forward API](forward-api.md) -- `T()`, grammar primitives, namespace handlers, Subject builder
- [Reversal Engine](reversal.md) -- 3-tier tokeniser, matching, morphology rules, round-trip verification
- [GrammarImprint](grammar-imprint.md) -- Lossy feature vectors, weighted cosine similarity, reference distributions
- [Locale JSON Schema](locale-schema.md) -- `en.json` structure, grammar table contract, sacred rules
- [Multiplier](multiplier.md) -- Deterministic variant generation, case preservation, round-trip guarantee

## Key Design Decisions

**Grammar engine, not translation file manager.** Consumers bring their own translations. go-i18n provides the grammatical composition and decomposition primitives.

**3-tier lookup.** All grammar lookups follow the same pattern: JSON locale data (tier 1) takes precedence over irregular Go maps (tier 2), which take precedence over regular morphology rules (tier 3). This lets locale files override any built-in rule.

**Round-trip verification.** The reversal engine verifies tier 3 candidates by applying the forward function and checking the result matches the original. This eliminates phantom base forms like "walke" or "processe".

**Lossy imprints.** GrammarImprint intentionally discards content, preserving only grammatical structure. Two texts with similar grammar produce similar imprints regardless of subject matter. This is a privacy-preserving proxy for semantic similarity.

## Running Tests

```bash
go test ./...                    # All tests
go test -v ./reversal/           # Reversal engine tests
go test -bench=. ./...           # Benchmarks
```

## Status

- **Phase 1** (Harden): Dual-class disambiguation -- design approved, implementation in progress
- **Phase 2** (Reference Distributions): 1B pre-classification pipeline + imprint calibration
- **Phase 3** (Multi-Language): French grammar tables
