# Development Guide

## Prerequisites

- Go 1.25 or later (the module uses `go 1.25.5`)
- `golang.org/x/text` (only external dependency for the core engine)
- `forge.lthn.ai/core/go-inference` (replaced via local path `../go-inference` in `go.mod` — required for the `classify.go` and `calibrate.go` files and integration tests)

For integration tests only:
- Models on `/Volumes/Data/lem/` — specifically `LEM-Gemma3-1B-layered-v2` and `LEM-Gemma3-27B` (or compatible models served via the `go-inference` interface)

The `go-inference` package provides the `TextModel` interface used by `ClassifyCorpus()` and `CalibrateDomains()`. Unit tests use a mock implementation and do not require real models.

---

## Build and Test

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./reversal/

# Run a single test by name
go test -run TestName ./...

# Run benchmarks
go test -bench=. ./...

# Run benchmarks for a specific package
go test -bench=. -benchmem ./reversal/

# Run with race detector
go test -race ./...
```

All tests must pass before committing. The race detector must report clean.

---

## Integration Tests

Integration tests require real model instances on `/Volumes/Data/lem/` and are kept in the `integration/` directory, separate from unit tests. They are not run by `go test ./...` from the module root (the integration package is excluded from the default build tag set).

```bash
# 1B classification pipeline (50 prompts, approximately 1 second on M3 Ultra)
cd integration && go test -v -run TestClassifyCorpus_Integration

# 1B vs 27B calibration (500 sentences, approximately 2-5 minutes with 27B)
cd integration && go test -v -run TestCalibrateDomains_1Bvs27B
```

If models are unavailable, the integration tests skip automatically via `testing.Short()` or an explicit model-presence check. Do not convert integration tests to unit tests — they have real runtime cost and external dependencies.

---

## Test Patterns

### Unit tests

Unit tests for the reversal package follow the `_Good`, `_Bad`, `_Ugly` naming pattern inherited from the broader Core Go ecosystem:

- `_Good`: happy path
- `_Bad`: expected error conditions
- `_Ugly`: panic or edge cases

Tests for the root package use standard Go test function naming.

### Mock models

`ClassifyCorpus()` and `CalibrateDomains()` accept the `inference.TextModel` interface. Unit tests construct a mock that returns controlled token sequences without loading any model. The mock implements `Classify(ctx, prompts, opts...) ([]Result, error)`.

### Round-trip tests

`reversal/roundtrip_test.go` validates the round-trip property: every verb in `irregularVerbs` and every noun in `irregularNouns` must survive a reverse lookup and recover the original base form. Add any new irregular entries to the maps in `types.go` and the round-trip tests will automatically cover them.

### Disambiguation tests

Nine named scenario tests cover the key disambiguation signal interactions:
- Noun after determiner (noun_determiner fires)
- Imperative verb at sentence start (sentence_position fires)
- Verb saturation within clause
- Clause boundary isolation
- Contraction auxiliary (`don't`, `can't`, etc.)

Twelve dual-class round-trip tests cover all six dual-class words (`commit`, `run`, `test`, `check`, `file`, `build`) in both verb and noun roles.

### Benchmark baselines

Benchmark baselines were measured on M3 Ultra, arm64. See `FINDINGS.md` (archived in `docs/history.md`) for the full table. When adding new benchmarks, include `b.ReportAllocs()` and compare against the baseline table.

---

## Coding Standards

### Language

UK English throughout. Correct spellings: `colour`, `organisation`, `centre`, `analyse`, `recognise`, `optimise`, `initialise`, `synchronise`, `cancelling`, `modelled`, `labelled`, `travelling`. These spellings appear in the `irregularVerbs` map and must remain consistent.

### Go style

- `declare(strict_types=1)` equivalent: use explicit types on all declarations where the type is not obvious from context
- All parameters and return types must be named and typed
- Prefer `fmt.Errorf("context: %w", err)` for error wrapping
- Use `errors.Is()` for error comparison, not string matching
- No global mutable state beyond the `grammarCache` and `templateCache` (which are already protected by synchronisation primitives)

### Grammar table rules

**Never flatten `gram.*` keys in locale JSON.** The loader (`flattenWithGrammar()`) depends on the nested `gram.verb.*`, `gram.noun.*` etc. path structure to route objects into typed Go structs. Flattening to `"gram.verb.delete.past": "deleted"` causes silent data loss — the key is treated as a plain translation message, not a verb form.

**Dual-class words** must appear in both `gram.verb` and `gram.noun` in the JSON. The tokeniser builds the `dualClass` index by intersecting `baseVerbs` and `baseNouns` at construction time.

**Only `gram.*` grammar data belongs in `locales/en.json` and `locales/fr.json`.** Consumer app translation keys (`prompt.*`, `time.*`, etc.) are managed by consumers, not this library.

### File organisation

| File | Contents |
|------|----------|
| `types.go` | All types, interfaces, package variables, irregular maps |
| `grammar.go` | Forward composition functions |
| `loader.go` | FSLoader, JSON parsing, flattenWithGrammar |
| `classify.go` | ClassifyCorpus, ClassifyStats, ClassifyOption |
| `calibrate.go` | CalibrateDomains, CalibrationStats, CalibrationResult |
| `reversal/tokeniser.go` | Tokeniser, Tokenise, two-pass disambiguation |
| `reversal/imprint.go` | GrammarImprint, NewImprint, Similar |
| `reversal/reference.go` | ReferenceSet, BuildReferences, Compare, Classify, distance metrics |
| `reversal/anomaly.go` | DetectAnomalies, AnomalyResult, AnomalyStats |
| `reversal/multiplier.go` | Multiplier, Expand |

Do not put grammar functions in `types.go` or type definitions in `grammar.go`. Keep the split clean.

---

## Conventional Commits

Format: `type(scope): description`

Common types: `feat`, `fix`, `test`, `bench`, `refactor`, `docs`, `chore`

Common scopes: `tokeniser`, `imprint`, `reference`, `anomaly`, `multiplier`, `grammar`, `loader`, `classify`, `calibrate`, `fr` (for French grammar table changes)

Examples:
```
feat(tokeniser): add two-pass disambiguation for dual-class words
fix(imprint): floor confidence at 0.55/0.45 when only prior fires
test(reference): add Mahalanobis fallback to Euclidean test
bench(grammar): add PastTense and Gerund baselines
```

---

## Co-Author

All commits must include the co-author trailer:

```
Co-Authored-By: Virgil <virgil@lethean.io>
```

---

## Licence

EUPL-1.2. Do not add dependencies with incompatible licences. The only external runtime dependency is `golang.org/x/text` (BSD-3-Clause, compatible). `go-inference` is an internal Core module.

---

## Adding a New Language

1. Create `locales/<lang>.json` with a complete `gram` block following `docs/grammar-table-spec.md`.
2. Populate `gram.verb` comprehensively — tiers 2 and 3 of the fallback chain are English-only.
3. Populate `gram.noun` with gender fields if the language has grammatical gender.
4. Set `gram.article.by_gender` for gendered article systems.
5. Set `gram.punct.label` correctly — French uses `" :"` (space before colon), English uses `":"`.
6. Populate `gram.signal` lists so the disambiguation tokeniser has language-appropriate determiners and auxiliaries. Without these, the tokeniser uses hardcoded English defaults.
7. Add a plural rule function to the `pluralRules` map in `types.go` if the language has non-standard plural categories (beyond one/other).
8. Run `go test ./...` and confirm all existing tests still pass. Add grammar data tests that verify the loaded counts and known values.
9. If the language needs reversal support, verify that `NewTokeniserForLang("<lang>")` builds indexes correctly and that `MatchVerb` / `MatchNoun` return correct results for a sample of forms.

---

## Performance Notes

- `Imprint.Similar` is zero-alloc. Keep it that way — it is called in tight loops during reference comparison.
- `WithSignals()` allocates `SignalBreakdown` on every ambiguous token. It is for diagnostics only; never enable it in the hot path.
- `Multiplier.Expand` allocates heavily (63 allocs for a four-word sentence). If it becomes a bottleneck, pool the token slices.
- The `grammarCache` uses `sync.RWMutex` with read-biased locking. Languages are loaded once at startup and then read-only; this is the intended pattern.
