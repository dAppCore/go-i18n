---
title: Locale JSON Schema
description: Structure of locale JSON files and the grammar table contract.
---

# Locale JSON Schema

Grammar data lives in `locales/{lang}.json`. This page documents the exact structure the loader expects and the sacred rules that must not be violated.

## Top-Level Structure

```json
{
  "gram": {
    "verb": { ... },
    "noun": { ... },
    "article": { ... },
    "word": { ... },
    "punct": { ... },
    "signal": { ... }
  },
  "time": { ... },
  "prompt": { ... }
}
```

Only the `gram.*` namespace is processed by the grammar engine. Everything outside `gram.*` is flattened into the message lookup table as plain key-value strings.

## gram.verb -- Verb Conjugation Tables

Each entry maps a base verb to its inflected forms:

```json
"verb": {
  "delete": { "base": "delete", "past": "deleted", "gerund": "deleting" },
  "commit": { "base": "commit", "past": "committed", "gerund": "committing" },
  "go":     { "base": "go",     "past": "went",      "gerund": "going" },
  "run":    { "base": "run",    "past": "ran",        "gerund": "running" }
}
```

**Required fields**: At least one of `base`, `past`, `gerund` must be present.

The loader detects verb objects via `isVerbFormObject()` -- any map with `base`, `past`, or `gerund` keys (and NOT a plural object).

**When to add entries**: Only for verbs where the regular morphology engine would produce the wrong result:
- Irregular verbs (go/went/going)
- Consonant doubling verbs (commit/committed/committing)
- Verbs the engine guesses wrong

Regular verbs like add/added/adding do not need entries.

## gram.noun -- Noun Pluralisation Tables

Each entry maps a noun to its singular and plural forms:

```json
"noun": {
  "file":          { "one": "file",          "other": "files" },
  "person":        { "one": "person",        "other": "people" },
  "child":         { "one": "child",         "other": "children" },
  "vulnerability": { "one": "vulnerability", "other": "vulnerabilities" },
  "commit":        { "one": "commit",        "other": "commits" }
}
```

**Required fields**: `one` and `other`.
**Optional field**: `gender` (for gendered languages).

The loader detects noun objects by checking for `one` + `other` keys, or the presence of a `gender` key.

As with verbs, only add entries for irregular plurals or cases where the engine guesses wrong. Regular plurals (server -> servers) do not need entries.

## gram.article -- Article Configuration

```json
"article": {
  "indefinite": {
    "default": "a",
    "vowel": "an"
  },
  "definite": "the"
}
```

Maps to the `ArticleForms` struct. The `Article()` function uses phonetic rules (consonant/vowel sound maps) to choose between `default` and `vowel`.

For gendered languages, add a `by_gender` map:

```json
"article": {
  "definite": "the",
  "by_gender": {
    "masculine": "le",
    "feminine": "la"
  }
}
```

## gram.word -- Domain Vocabulary

Maps lowercase keys to display forms:

```json
"word": {
  "url": "URL",
  "ssh": "SSH",
  "api": "API",
  "id": "ID",
  "ci": "CI",
  "qa": "QA",
  "blocked_by": "blocked by",
  "up_to_date": "up to date"
}
```

These are classified as `TokenWord` by the reversal tokeniser and tracked in `DomainVocabulary` in imprints. Add entries for:
- Acronyms with specific capitalisation (URL, SSH, API)
- Multi-word phrases (`blocked_by` -> "blocked by")
- Domain-specific terms that need consistent display

## gram.punct -- Punctuation Rules

```json
"punct": {
  "label": ":",
  "progress": "..."
}
```

Language-specific punctuation suffixes. French would use `" :"` (space before colon) for the label suffix. The `LabelHandler` and `ProgressHandler` read these values.

## gram.signal -- Disambiguation Signal Words

```json
"signal": {
  "noun_determiner": [
    "the", "a", "an", "this", "that", "these", "those",
    "my", "your", "his", "her", "its", "our", "their",
    "every", "each", "some", "any", "no",
    "many", "few", "several", "all", "both"
  ],
  "verb_auxiliary": [
    "is", "are", "was", "were", "has", "had", "have",
    "do", "does", "did", "will", "would", "could", "should",
    "can", "may", "might", "shall", "must"
  ],
  "verb_infinitive": ["to"]
}
```

Used by the [dual-class disambiguation](reversal.md) system to classify ambiguous words like "commit", "test", "run". Each signal list falls back to hardcoded English defaults when absent from the locale file.

## Translation Messages

Everything outside `gram.*` is flattened into the message lookup table:

```json
{
  "time": {
    "just_now": "just now",
    "ago": {
      "minute": {
        "one": "{{.Count}} minute ago",
        "other": "{{.Count}} minutes ago"
      }
    }
  }
}
```

Nested objects with CLDR plural keys (`zero`, `one`, `two`, `few`, `many`, `other`) are detected and stored as `Message` structs with plural forms. All other strings are flattened to dot-notation keys (e.g. `time.just_now`).

### Template Syntax

Message values support Go `text/template` syntax:

```json
"welcome": "Hello, {{.Subject}}!"
```

## Grammar Table Contract

### Sacred Rules

1. **NEVER flatten `gram.*` keys.** The grammar engine depends on nested structure. Flattening `gram.verb.delete.past` to a flat string key breaks the loader silently. Agents and tooling must preserve the nested JSON objects.

2. **Only `gram.*` data belongs in locale files.** Consumer translations are external -- packages register their own locale files via `RegisterLocales()`.

3. **Irregular forms override regular morphology.** If a verb is in `gram.verb`, its forms take precedence over the rule-based engine.

4. **The `one`/`other` keys overlap with CLDR plural categories.** The loader distinguishes noun objects (under `gram.noun.*`) from plural message objects by checking for nested maps.

5. **Values must be the inflected form, not rules.** Store `"deleted"`, not `"d suffix"`.

6. **Round-trip must hold.** `PastTense(base)` then reverse must recover `base`.

### Adding a New Language

1. Create `locales/{lang}.json` with a `gram` section
2. Populate `gram.verb` with irregular verbs for that language
3. Populate `gram.noun` with irregular nouns
4. Define `gram.article` rules (if the language has articles)
5. Define `gram.punct` (language-specific punctuation)
6. Add `gram.signal` word lists for disambiguation
7. Add CLDR plural rules in `language.go` if not already present
8. Run reversal round-trip tests to verify bijective property

### Loader Behaviour

The `FSLoader` reads all `.json` files from the locales directory. For each file:

1. Parse as JSON
2. Walk the key tree, detecting grammar objects (`gram.*`) and plural objects
3. Grammar data populates the `GrammarData` struct (verbs, nouns, articles, words, punctuation, signals)
4. Everything else is flattened into `map[string]Message` for translation lookup
5. Language tags support both `-` and `_` separators (`en-GB` and `en_GB` both work)
