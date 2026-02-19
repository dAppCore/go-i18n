# Grammar Table Format Specification

JSON schema for `gram.*` keys in go-i18n locale files. This document defines the contract new languages must implement to work with the grammar engine.

## File Location

Locale files live in `locales/<lang>.json` and are embedded at compile time. The language code follows BCP 47 (e.g. `en`, `fr`, `de`, `es`, `zh`).

## Top-Level Structure

```json
{
  "gram": {
    "verb":    { ... },
    "noun":    { ... },
    "article": { ... },
    "word":    { ... },
    "punct":   { ... },
    "signal":  { ... },
    "number":  { ... }
  },
  "prompt": { ... },
  "time":   { ... },
  "lang":   { ... }
}
```

Only the `gram` block is documented here. The `prompt`, `time`, and `lang` blocks are standard translation keys that consumers manage.

**Critical**: The `gram.*` structure MUST remain nested. The loader extracts grammar data before flattening. If you flatten `gram.verb.delete.past` to a dot-separated key, the grammar engine breaks silently.

---

## `gram.verb` — Verb Conjugations

Maps verb base forms to their inflected forms. Used by `PastTense()`, `Gerund()`, and the reversal tokeniser.

```json
"verb": {
  "<base>": {
    "base":   "<string>",
    "past":   "<string>",
    "gerund": "<string>"
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `base` | Optional | The infinitive form. Defaults to the key name. |
| `past` | Required | Simple past tense form. |
| `gerund` | Required | Present participle (-ing form). |

**Example (English)**:
```json
"delete": { "base": "delete", "past": "deleted", "gerund": "deleting" },
"build":  { "base": "build",  "past": "built",   "gerund": "building" }
```

**Example (French)** — would extend to include mood/person:
```json
"supprimer": { "base": "supprimer", "past": "supprimé", "gerund": "supprimant" }
```

**Detection**: The loader identifies verb objects by the presence of `"past"` or `"gerund"` keys (`isVerbFormObject()`).

**Fallback chain**: JSON `gram.verb` → Go `irregularVerbs` map → regular morphology rules.

---

## `gram.noun` — Noun Plural Forms

Maps noun base forms to their singular/plural forms. Used by `Pluralize()`, `PluralForm()`, and the reversal tokeniser.

```json
"noun": {
  "<base>": {
    "one":    "<string>",
    "other":  "<string>",
    "gender": "<string>"
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `one` | Required | Singular form. |
| `other` | Required | Plural form. |
| `gender` | Optional | Grammatical gender (`"m"`, `"f"`, `"n"`). Unused in English. |

**Example (English)**:
```json
"file":   { "one": "file",   "other": "files" },
"person": { "one": "person", "other": "people" }
```

**Example (French)** — gender required for article agreement:
```json
"fichier": { "one": "fichier", "other": "fichiers", "gender": "m" },
"branche": { "one": "branche", "other": "branches", "gender": "f" }
```

**Detection**: The loader identifies noun objects by the presence of `"one"` and `"other"` keys.

**Fallback chain**: JSON `gram.noun` → Go `irregularNouns` map → regular pluralisation rules.

---

## `gram.article` — Article Rules

Defines article forms. Used by `Article()` and the reversal tokeniser.

```json
"article": {
  "indefinite": {
    "default": "<string>",
    "vowel":   "<string>"
  },
  "definite": "<string>",
  "by_gender": {
    "<gender>": "<string>"
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `indefinite.default` | Required | Default indefinite article ("a" in English). |
| `indefinite.vowel` | Required | Vowel-sound indefinite article ("an" in English). |
| `definite` | Required | Definite article ("the" in English). |
| `by_gender` | Optional | Gender-specific articles (French: `{"m": "le", "f": "la"}`). |

**Example (English)**:
```json
"article": {
  "indefinite": { "default": "a", "vowel": "an" },
  "definite": "the"
}
```

**Example (French)**:
```json
"article": {
  "indefinite": { "default": "un", "vowel": "un" },
  "definite": "le",
  "by_gender": { "m": "le", "f": "la" }
}
```

---

## `gram.word` — Domain Vocabulary

Maps lookup keys to their display forms. Used for acronyms, proper nouns, and multi-word terms that need exact casing or spacing. The reversal tokeniser classifies these as `TokenWord`.

```json
"word": {
  "<key>": "<display_form>"
}
```

Keys are lowercase with underscores for multi-word terms. Values are the exact display string.

**Example**:
```json
"url": "URL",
"api": "API",
"go_mod": "go.mod",
"dry_run": "dry run",
"up_to_date": "up to date"
```

**Note**: Do not add app-specific terms here. Only grammar-level vocabulary (acronyms, proper nouns) belongs in `gram.word`. Consumer apps manage their own translation keys outside `gram.*`.

---

## `gram.punct` — Punctuation Rules

Language-specific punctuation patterns used by `Label()` and `Progress()`.

```json
"punct": {
  "label":    "<string>",
  "progress": "<string>"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `label` | Required | Suffix for label formatting. English: `":"`, French: `" :"`. |
| `progress` | Required | Suffix for progress indicators. Typically `"..."`. |

**Example (English)**:
```json
"punct": { "label": ":", "progress": "..." }
```

**Example (French)** — space before colon:
```json
"punct": { "label": " :", "progress": "..." }
```

---

## `gram.signal` — Disambiguation Signals

Word lists used by the reversal tokeniser's two-pass disambiguation algorithm. These resolve dual-class words (e.g. "commit" as verb vs noun) based on surrounding context.

```json
"signal": {
  "noun_determiner": ["<string>", ...],
  "verb_auxiliary":   ["<string>", ...],
  "verb_infinitive":  ["<string>", ...]
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `noun_determiner` | Recommended | Words that signal a following noun (articles, possessives, quantifiers). |
| `verb_auxiliary` | Recommended | Words that signal a following verb (modals, auxiliaries, contractions). |
| `verb_infinitive` | Recommended | Infinitive markers ("to" in English). |

**Example (English)**:
```json
"signal": {
  "noun_determiner": ["the", "a", "an", "this", "that", "my", "your", ...],
  "verb_auxiliary": ["is", "are", "will", "can", "must", "don't", "can't", ...],
  "verb_infinitive": ["to"]
}
```

**Fallback**: If a signal list is missing or empty, the tokeniser uses hardcoded English defaults. Each signal list falls back independently — partial locale data is handled gracefully.

---

## `gram.number` — Number Formatting

Locale-specific number formatting rules.

```json
"number": {
  "thousands": "<string>",
  "decimal":   "<string>",
  "percent":   "<string>"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `thousands` | Required | Thousands separator. English: `","`, German: `"."`. |
| `decimal` | Required | Decimal separator. English: `"."`, German: `","`. |
| `percent` | Required | Printf format string. Use `%s%%` for "42%". |

**Example (English)**:
```json
"number": { "thousands": ",", "decimal": ".", "percent": "%s%%" }
```

**Example (German)**:
```json
"number": { "thousands": ".", "decimal": ",", "percent": "%s %%" }
```

---

## Adding a New Language

1. Create `locales/<lang>.json` with the full `gram` block.
2. All `gram.verb` entries need `past` and `gerund` fields at minimum.
3. All `gram.noun` entries need `one` and `other` fields. Add `gender` if the language has grammatical gender.
4. Set `gram.article` — use `by_gender` for gendered article systems.
5. Set `gram.punct` — pay attention to spacing conventions (French colon spacing, etc.).
6. Populate `gram.signal` lists for the tokeniser. Without these, disambiguation falls back to English defaults.
7. Add `prompt`, `time`, and `lang` blocks as needed.

The grammar engine uses a 3-tier lookup for verbs and nouns:
1. **JSON grammar tables** (`gram.verb`, `gram.noun`) — checked first
2. **Go irregular maps** (`irregularVerbs`, `irregularNouns`) — English-only fallback
3. **Regular morphology rules** — algorithmic, currently English-only

For non-English languages, the JSON tables must be comprehensive since tiers 2 and 3 only cover English.

---

## Dual-Class Words

Words that function as both verbs and nouns (e.g. "commit", "build", "test") must appear in BOTH `gram.verb` and `gram.noun`. The tokeniser's disambiguation algorithm resolves which role a word plays based on context signals.

See FINDINGS.md "Dual-Class Word Disambiguation" for the algorithm details.

---

## Validation

There is no explicit schema validation. The loader uses safe type assertions and silently skips malformed entries. To verify a new locale file:

```bash
go test -v -run "TestGrammarData" ./...
```

The grammar tests check that loaded data has expected field counts and known values.
