# CLAUDE.md

## What This Is

Grammar-aware internationalisation engine for Go. Module: `forge.lthn.ai/core/go-i18n`

This is a **grammar engine** — it provides primitives for composing and reversing grammatically correct text. It is NOT a translation file manager. Consumers bring their own translations.

## Commands

```bash
go test ./...                    # Run all tests
go test -v ./reversal/           # Reversal engine tests
go test -bench=. ./...           # Benchmarks (when added)
```

## Critical Rules

### DO NOT flatten locale JSON files

The grammar engine depends on nested `gram.*` structure:

```json
{
  "gram": {
    "verb": {
      "delete": { "past": "deleted", "gerund": "deleting" }
    }
  }
}
```

If you flatten this to `"gram.verb.delete.past": "deleted"`, the grammar engine breaks silently. **This is the #1 cause of agent-introduced bugs.**

### This library does not manage consumer translations

go-i18n provides grammar primitives. Apps using it (core/cli, etc.) manage their own translation files. Do not add app-specific translation keys to `locales/en.json` — only `gram.*` grammar data belongs there.

## Architecture

| Package | Purpose |
|---------|---------|
| Root | Forward composition: T(), grammar primitives, handlers, service |
| `reversal/` | Reverse grammar: tokeniser, imprint, multiplier |
| `locales/` | Grammar tables (JSON) — only `gram.*` data |
| `docs/plans/` | Design documents |

## Key API

- `T(key, args...)` — Translate with namespace handlers
- `PastTense(verb)`, `Gerund(verb)`, `Pluralize(noun, n)`, `Article(word)` — Grammar primitives
- `reversal.NewTokeniser().Tokenise(text)` — Reverse grammar lookup
- `reversal.NewImprint(tokens)` — Feature vector projection
- `reversal.NewMultiplier().Expand(text)` — Training data augmentation

## Coding Standards

- UK English (colour, organisation, centre)
- `go test ./...` must pass before commit
- Conventional commits: `type(scope): description`
- Co-Author: `Co-Authored-By: Virgil <virgil@lethean.io>`

## Task Queue

See `TODO.md` for dispatched tasks from core/go orchestration.
See `FINDINGS.md` for research notes and architectural decisions.
See the [wiki](https://forge.lthn.ai/core/go-i18n/wiki) for full architecture docs.
