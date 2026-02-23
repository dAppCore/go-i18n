[![Go Reference](https://pkg.go.dev/badge/forge.lthn.ai/core/go-i18n.svg)](https://pkg.go.dev/forge.lthn.ai/core/go-i18n)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](go.mod)

# go-i18n

Grammar engine for Go. Provides forward composition primitives (PastTense, Gerund, Pluralize, Article, composite progress and label functions), a `T()` translation entry point with namespace key handlers, and a reversal engine that recovers base forms and grammatical roles from inflected text. The reversal package produces `GrammarImprint` feature vectors for semantic similarity scoring, builds reference domain distributions, performs anomaly detection, and includes a 1B model pre-sort pipeline for training data classification. Consumers bring their own translation keys; this library provides the grammatical machinery.

**Module**: `forge.lthn.ai/core/go-i18n`
**Licence**: EUPL-1.2
**Language**: Go 1.25

## Quick Start

```go
import "forge.lthn.ai/core/go-i18n"

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

## Licence

European Union Public Licence 1.2 — see [LICENCE](LICENCE) for details.
