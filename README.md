[![Go Reference](https://pkg.go.dev/badge/dappco.re/go/core/i18n.svg)](https://pkg.go.dev/dappco.re/go/core/i18n)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](go.mod)

# go-i18n

Grammar engine for Go. Provides forward composition primitives (PastTense, Gerund, Pluralize, Article, composite progress and label functions), a `T()` translation entry point with namespace key handlers, and a reversal engine that recovers base forms and grammatical roles from inflected text. The reversal package produces `GrammarImprint` feature vectors for semantic similarity scoring, builds reference domain distributions, performs anomaly detection, and includes a 1B model pre-sort pipeline for training data classification. Consumers bring their own translation keys; this library provides the grammatical machinery.

Bundled locales: `en` (the English of England — bare `en` IS en-GB), `en-US` (dialect override), `fr`, `de` and `es` — all composing full output from English keys through the `gram.word` bridge: `T("i18n.done.delete", "file")` → "File deleted" / "Fichier supprimé" / "Datei gelöscht" / "Archivo eliminado". The Spanish locale is pure data — no Spanish-specific code exists in the engine.

The `phonetics` package gives the engine an ear: the vendored CMU Pronouncing Dictionary (134k words with stress markers) backs `Article()` and consonant doubling with actual phonemes ("an yttrium sample", commit → committed but visit → visited by stress, not by table), plus rhyme, alliteration, syllable and stress primitives, metre scanning (`ScanLine` → dominant foot + regularity), and whole-sentence transcription in ARPABET, IPA or readable phonetic respelling.

Novelty locales double as architecture proofs: `en-x-pirate` (vocabulary-only dialect riding the `en` fallback chain — "Blimey! Couldn't heave plank") and `tlh` (Klingon: canonical Okrand vocabulary, aspect suffixes filling the past/gerund slots mechanically, and the `article.none` mechanism article-less languages like Japanese and Russian will reuse — "De' teqpu'").

**Module**: `dappco.re/go/core/i18n`
**Licence**: EUPL-1.2
**Language**: Go 1.25

## Quick Start

```go
import "dappco.re/go/core/i18n"

// Grammar primitives
fmt.Println(i18n.PastTense("delete"))   // "deleted"
fmt.Println(i18n.Gerund("build"))       // "building"
fmt.Println(i18n.Pluralize("file", 3))  // "files"

// Translation with auto-composed output
fmt.Println(i18n.T("i18n.progress.build"))  // "Building..."
fmt.Println(i18n.T("i18n.done.delete", "file"))  // "File deleted"

// Reversal: recover grammar from text
tokeniser := reversal.NewTokeniser()
tokens := tokeniser.Tokenise("deleted the files")
imprint := reversal.NewImprint(tokens)
```

## Documentation

- [Architecture](docs/architecture.md) — grammar primitives, T() handlers, reversal engine, GrammarImprint, reference distributions, 1B pipeline
- [Development Guide](docs/development.md) — building, testing, grammar table structure (critical: do not flatten JSON)
- [Project History](docs/history.md) — completed phases and known limitations

## Build & Test

```bash
go test ./...
go test -v ./reversal/
go test -bench=. ./...
go build ./...
```

For repeatable local runs in a clean workspace, the repo also ships a
`Makefile` with the standard workflow targets:

```bash
make build
make vet
make test
make cover
make tidy
```

## Licence

European Union Public Licence 1.2 — see [LICENCE](LICENCE) for details.
