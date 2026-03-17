# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Grammar-aware internationalisation engine for Go. Module: `forge.lthn.ai/core/go-i18n`

This is a **grammar engine** — it provides primitives for composing and reversing grammatically correct text. It is NOT a translation file manager. Consumers bring their own translations.

## Commands

```bash
go test ./...                        # Run all tests
go test -v ./reversal/               # Reversal engine tests
go test -run TestPastTense ./...     # Run a single test by name
go test -bench=. ./...               # Benchmarks
go vet ./...                         # Static analysis
golangci-lint run ./...              # Lint (govet, errcheck, staticcheck, gocritic, gofmt, etc.)
```

Integration tests live in a separate module:
```bash
cd integration && go test ./...      # Integration tests (requires go-inference)
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

If you flatten this to `"gram.verb.delete.past": "deleted"`, the grammar engine breaks silently. **This is the #1 cause of agent-introduced bugs.** The loader's `flattenWithGrammar()` relies on path prefixes (`gram.verb.*`, `gram.noun.*`, etc.) to route objects into typed Go structs. Flattening bypasses this routing and causes silent data loss.

### This library does not manage consumer translations

go-i18n provides grammar primitives. Apps using it (core/cli, etc.) manage their own translation files. Do not add app-specific translation keys to `locales/en.json` — only `gram.*` grammar data belongs there.

## Architecture

| Package | Purpose |
|---------|---------|
| Root (`i18n`) | Forward composition: T(), grammar primitives, handlers, service, loader |
| `reversal/` | Reverse grammar: tokeniser, imprint, reference distributions, anomaly detection, multiplier |
| `locales/` | Grammar tables (JSON, embedded via `embed.FS`) — only `gram.*` data |
| `integration/` | Separate Go module for integration tests (depends on `go-inference`) |
| `docs/` | Architecture, development, history, grammar-table-spec |

See `docs/architecture.md` for full technical detail.

### Service Pattern

The root package uses a **singleton + instance** pattern. Package-level functions (`T()`, `PastTense()`, `SetLanguage()`, etc.) delegate to a default `Service` stored in an `atomic.Pointer`. `NewService()` creates instances; `SetDefault()` installs one as the package-level singleton. The `Service` holds the message store, grammar cache, key handlers, and language state.

### Three-Tier Grammar Lookup

Every grammar primitive follows the same resolution order:
1. JSON grammar tables (`gram.verb`, `gram.noun`) loaded for the current language
2. Go built-in irregular maps (`irregularVerbs`, `irregularNouns`) — English only
3. Regular morphological rules (algorithmic) — English only

Non-English languages must provide comprehensive JSON tables since tiers 2 and 3 are English-only fallbacks.

### Key Handler Chain

`T(key, args...)` passes keys through registered `KeyHandler` implementations before falling back to the message store. The built-in `i18n.*` namespace auto-composes output (e.g., `i18n.progress.build` → `"Building..."`, `i18n.done.delete` → `"File deleted"`).

## Key API

- `T(key, args...)` — Translate with namespace handlers
- `PastTense(verb)`, `Gerund(verb)`, `Pluralize(noun, n)`, `Article(word)` — Grammar primitives
- `Progress(verb)`, `ActionResult(verb, subject)`, `ActionFailed(verb, subject)` — Composite functions
- `reversal.NewTokeniser().Tokenise(text)` — Reverse grammar lookup
- `reversal.NewImprint(tokens)` — Feature vector projection
- `reversal.NewMultiplier().Expand(text)` — Training data augmentation
- `ClassifyCorpus(ctx, model, input, output, opts...)` — 1B model pre-sort pipeline (uses `go-inference`)

## Coding Standards

- UK English (colour, organisation, centre)
- `go test ./...` must pass before commit
- Errors use `log.E(op, msg, err)` from `forge.lthn.ai/core/go-log`, not `fmt.Errorf`
- Conventional commits: `type(scope): description`
- Co-Author: `Co-Authored-By: Virgil <virgil@lethean.io>`
