---
title: Multiplier
description: Deterministic grammatical variant generation for training data augmentation.
---

# Multiplier

The Multiplier generates grammatical variants of input text without any LLM calls. It produces deterministic, round-trip verified output suitable for training data augmentation and grammar coverage testing.

## Quick Start

```go
m := reversal.NewMultiplier()
variants := m.Expand("Delete the configuration file")
```

**Output (deterministic order):**

1. `Delete the configuration file` (original)
2. `Deleted the configuration file` (past tense)
3. `Deleting the configuration file` (gerund)
4. `Delete the configuration files` (plural noun)
5. `Deleted the configuration files` (past + plural)
6. `Deleting the configuration files` (gerund + plural)

## Creating a Multiplier

```go
// English (default)
m := reversal.NewMultiplier()

// Language-specific
m := reversal.NewMultiplierForLang("en")
```

The Multiplier wraps a `Tokeniser` internally, using it to classify input tokens before applying transforms.

## Variant Generation Algorithm

`Expand()` applies transforms in a fixed order:

### Step 1: Tokenise Input

The tokeniser classifies each word as verb, noun, article, word, punctuation, or unknown.

### Step 2: Single Verb Transforms

For each verb in the token list, generate three variants:
- **Past tense**: apply `PastTense(base)`
- **Gerund**: apply `Gerund(base)`
- **Base form**: keep as-is (deduplicates with original if already in base form)

### Step 3: Single Noun Transforms

For each noun, toggle singular/plural:
- Singular -> `PluralForm(base)`
- Plural -> base form (singular)

### Step 4: Combined Transforms

For each (verb, noun) pair, generate combined variants:
- past + noun toggle
- gerund + noun toggle
- base + noun toggle

### Deduplication

A `map[string]bool` tracks seen outputs. Identical strings are skipped. This handles cases like the base form duplicating the original.

## Case Preservation

Transforms preserve the original capitalisation pattern:

| Original Case | Transform | Result |
|--------------|-----------|--------|
| `"DELETE"` (all caps) | past tense | `"DELETED"` |
| `"Delete"` (title case) | gerund | `"Deleting"` |
| `"delete"` (lower case) | past tense | `"deleted"` |

The `preserveCase()` function detects three patterns: all-uppercase, title-case (first letter upper), and lowercase.

## Reconstruction

Tokens are rejoined with spaces, except punctuation tokens which attach to the preceding token:

```
["Delete", "file", "..."] -> "Delete file..."
["The", "tests", ",", "passed"] -> "The tests, passed"
```

## Round-Trip Guarantee

Every variant can be tokenised back through the reversal engine:

```go
original := "Delete the configuration file"
variants := m.Expand(original)

tok := reversal.NewTokeniser()
origImp := reversal.NewImprint(tok.Tokenise(original))

for _, v := range variants {
    varImp := reversal.NewImprint(tok.Tokenise(v))
    sim := origImp.Similar(varImp)
    // sim >= 0.2 -- variants share grammatical structure with original
}
```

Variants that fail round-trip tokenisation are a signal that the grammar tables need expanding.

## Typical Output Counts

The number of variants depends on the verb and noun count in the input:

| Input | Verbs | Nouns | Variants |
|-------|-------|-------|----------|
| `"Delete the file"` | 1 | 1 | 5-7 |
| `"Build and test the project"` | 2 | 1 | 8-12 |
| `"Run"` | 1 | 0 | 3 |
| `"The configuration file"` | 0 | 1 | 2 |
| `"Hello"` | 0 | 0 | 1 |

## Use Cases

### LEM Training Pipeline

Expand scored seed data to increase training set size without changing semantic content. A corpus of 88K seeds can be multiplied to 400K-600K training samples.

### Grammar Coverage Testing

Generate edge cases to test tokeniser robustness. If a variant fails to tokenise or round-trip correctly, it reveals gaps in the grammar tables.

### Imprint Calibration

Variants of the same text should produce converging imprints. If variants of "Delete the file" produce wildly different imprints, the imprint weights need adjustment. This provides an automated self-test for the [GrammarImprint](grammar-imprint.md) system.
