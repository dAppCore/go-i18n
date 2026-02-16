# go-i18n Reversal + go-html — Combined Design

**Date:** 2026-02-16
**Status:** Approved
**Author:** Snider + Claude
**License:** EUPL-1.2
**Depends On:** go-i18n (shipped), RFC-022 (go-html spec), RFC-001 (HLCRF)
**Heritage:** WorkStation Commerce (2008) → CorePHP → Go

---

## Context

go-i18n is shipped (`forge.lthn.ai/core/go-i18n`) with grammar tables, handlers, and en.json. go-html has a spec (RFC-022). CorePHP has a working HLCRF implementation (`Core\Front\Components\Layout`) and the "Modern Flexy" view protocol. Both modules share grammar tables and compose into the same binary (CoreGo + FrankenPHP + Octane + CorePHP + Wails).

Go eventually replaces PHP for rendering. PHP stays for business logic (CoreCommerce = WorkStation v2).

## Decisions

- **Priority:** Bottom-up (reversal first, then go-html on top)
- **Module boundaries:** Grammar reversal lives inside go-i18n as `reversal/` sub-package
- **Scope:** Reversal Layers 1-2 only (tokeniser + imprint). TIM/calibration/Poindexter are future work.
- **Interleaved:** Both modules planned together since they share grammar tables

---

## Phase 1: go-i18n Reversal (Layers 1-2)

### What it does

Takes text in, returns a GrammarImprint out. Uses the same grammar tables (verb forms, noun forms, articles, punctuation, word maps) as the forward engine, read backwards as pattern matchers.

### Structure

```
go-i18n/
├── reversal/
│   ├── tokeniser.go      # Reverse grammar tables → pattern matchers
│   ├── imprint.go         # GrammarImprint struct + comparison
│   ├── multiplier.go      # Training data augmentation
│   ├── tokeniser_test.go
│   ├── imprint_test.go
│   └── multiplier_test.go
├── grammar.go             # existing (forward)
├── types.go               # existing (shared types)
└── ...
```

### Core types

```go
// GrammarImprint — the "mould of the key"
type GrammarImprint struct {
    VerbDistribution   map[string]float64  // verb → frequency
    TenseDistribution  map[string]float64  // past/gerund/base → ratio
    NounDistribution   map[string]float64  // noun → frequency
    PluralRatio        float64             // plural vs singular
    FormalityScore     float64             // from existing Formality system
    DomainVocabulary   map[string]int      // gram.word hits by category
    ArticleUsage       map[string]float64  // a/an/the distribution
    PunctuationPattern map[string]float64  // label/progress/question ratios
    TokenCount         int
    UniqueVerbs        int
    UniqueNouns        int
}

// Tokeniser — reverses grammar tables into matchers
type Tokeniser struct {
    verbs    map[string]VerbMatch    // "committed" → {base:"commit", tense:"past"}
    nouns    map[string]NounMatch    // "files" → {base:"file", plural:true}
    words    map[string]string       // gram.word reverse lookup
    service  *i18n.Service           // parent service for grammar data
}

// VerbMatch — result of matching a word against verb tables
type VerbMatch struct {
    Base   string // "commit"
    Tense  string // "past", "gerund", "base"
    Form   string // the matched form
}

// NounMatch — result of matching a word against noun tables
type NounMatch struct {
    Base   string // "file"
    Plural bool   // true if plural form
    Form   string // the matched form
}

// Multiplier — grammatical augmentation for training data
type Multiplier struct {
    tokeniser *Tokeniser
}
```

### Flow

```
Text → Tokeniser.Tokenise(text)
         ├── Split to words
         ├── Match against verb tables (irregular → regular rules)
         ├── Match against noun tables (irregular → regular rules)
         ├── Match against gram.word maps
         ├── Detect articles, punctuation patterns
         └── → []Token

[]Token → NewImprint(tokens)
         ├── Calculate distributions
         ├── Score formality
         ├── Count domain vocabulary hits
         └── → GrammarImprint

GrammarImprint.Similar(other) → float64  // 0.0-1.0 similarity
```

### Training data multiplier

```go
m := reversal.NewMultiplier(service)

// "Delete the configuration file"
variants := m.Expand("Delete the configuration file")
// → "Deleted the configuration file"      (past)
// → "Deleting the configuration file"     (gerund)
// → "Delete the configuration files"      (plural)
// → "Deleted the configuration files"     (past+plural)
// → "Deleting the configuration files"    (gerund+plural)
```

Uses go-i18n's existing `PastTense()`, `Gerund()`, `Pluralize()` functions. Zero API calls. Deterministic. 88K seeds × variants = 500K+ training examples.

### Properties as a linguistic hash function

| Property | Cryptographic Hash | Grammar Imprint |
|----------|-------------------|-----------------|
| Deterministic | Same input → same hash | Same document → same imprint |
| One-way | Can't reconstruct input | Can't reconstruct document |
| Fixed output | 256/512 bits | Grammar feature vector |
| Collision-resistant | Different inputs → different hashes | Different documents → different imprints |
| Semantic-preserving | No | **Yes** — similar documents → similar imprints |

The surjection property gives privacy. The similarity property gives utility.

### Validation

- Feed known sentence through forward (go-i18n) + reverse (tokeniser) → confirm correct token extraction
- Compare imprints of semantically similar documents → confirm similarity score > threshold
- Compare imprints of unrelated documents → confirm low similarity
- Feed multiplier output through reversal → confirm imprints differ only in transformed dimension

---

## Phase 2: go-html Core (`forge.lthn.ai/core/go-html`)

### Heritage

Direct port of CorePHP's `Core\Front\Components\Layout` + "Modern Flexy" view protocol. Same architecture, compiled Go, WASM-capable. Go replaces PHP for rendering; PHP stays for business logic.

### Structure

```
go-html/
├── go.mod                 # forge.lthn.ai/core/go-html
├── layout.go              # HLCRF parser + Layout type
├── node.go                # Node interface + core types
├── path.go                # Path-based ID generation
├── render.go              # Tree → valid HTML string
├── context.go             # Render context (identity, locale, entitlements)
├── responsive.go          # Multi-variant responsive layouts
├── layout_test.go
├── node_test.go
├── path_test.go
├── render_test.go
└── responsive_test.go
```

### Core types

```go
// Node — everything renderable
type Node interface {
    Render(ctx *Context) string
}

// Layout — HLCRF compositor
type Layout struct {
    variant  string            // "HLCRF", "HCF", "C", etc.
    path     string            // "" for root, "C-0-" for nested
    slots    map[byte][]Node   // H, L, C, R, F → children
    attrs    map[string]string
}

// Construction — mirrors CorePHP's fluent API
func NewLayout(variant string) *Layout
func (l *Layout) H(nodes ...Node) *Layout
func (l *Layout) L(nodes ...Node) *Layout
func (l *Layout) C(nodes ...Node) *Layout
func (l *Layout) R(nodes ...Node) *Layout
func (l *Layout) F(nodes ...Node) *Layout
func (l *Layout) Render(ctx *Context) string

// Context — render-time state
type Context struct {
    Identity     string                    // @name.lthn
    Locale       string                    // language code
    Formality    i18n.Formality            // from go-i18n
    Entitlements func(feature string) bool // RFC-004 checker
    Data         map[string]any            // controller data
    service      *i18n.Service             // go-i18n for text
}
```

### Node types

```go
func El(tag string, children ...Node) Node          // HTML element
func Text(key string, args ...any) Node              // go-i18n composed text
func Raw(content string) Node                         // escape hatch
func If(fn func(*Context) bool, node Node) Node      // conditional
func Unless(fn func(*Context) bool, node Node) Node  // inverse conditional
func Switch(fn func(*Context) string, cases map[string]Node) Node
func Each[T any](items []T, fn func(T) Node) Node   // iteration
func Entitled(feature string, node Node) Node         // RFC-004 gating
func Slot(name string) Node                           // named placeholder
```

### Path-based IDs

Every element gets a deterministic address:

```
Layout("HLCRF")
├── H-0           data-block="H-0"
├── L-0           data-block="L-0"
│   └── Layout("HCF")  ← nested
│       ├── L-0-H-0    data-block="L-0-H-0"
│       ├── L-0-C-0    data-block="L-0-C-0"
│       └── L-0-F-0    data-block="L-0-F-0"
├── C-0           data-block="C-0"
├── R-0           data-block="R-0"
└── F-0           data-block="F-0"
```

### go-i18n integration

Every `Text()` node flows through the grammar engine:

```go
El("h1", Text("i18n.progress.build", project))
```

Every rendered page is automatically:
- Grammatically correct (go-i18n forward)
- Reversible to GrammarImprint (go-i18n reversal)
- Localisable (swap locale JSON)

Safety by structure: text bypasses grammar pipeline only via explicit `Raw()`.

### Output guarantees

| Guarantee | Mechanism |
|-----------|-----------|
| Valid HTML | Tree rendering, not string concat |
| No XSS | Text nodes escaped by default |
| No orphaned tags | Tree structure enforces nesting |
| Semantic elements | HLCRF → header/aside/main/footer |
| Accessible | HLCRF regions → ARIA landmarks |

### Validation

- `NewLayout("HLCRF").H(...).C(...).F(...)` produces valid HTML with correct data-block IDs
- Nested layouts generate correct path chains
- `Text()` nodes produce grammar-composed output matching go-i18n forward engine
- `Entitled()` nodes are absent (not hidden) when entitlement missing
- HTML output validates against spec (no orphaned tags, no unescaped user content)

---

## Phase 3: Integration + WASM

| Deliverable | Description |
|-------------|-------------|
| Responsive variants | Multi-variant layouts (desktop/tablet/mobile) |
| Entitlement gating | RFC-004 integration |
| WASM build | `GOOS=js GOARCH=wasm`, <2MB target |
| Full pipeline | Render → reverse → GrammarImprint |

### Validation

- Same content renders differently per variant string
- WASM module loads in browser, renders to DOM
- Rendered page produces valid GrammarImprint via reversal

---

## Phase 4: CoreDeno + Web Components (future)

| Deliverable | Description |
|-------------|-------------|
| CoreDeno | Deno runtime for client-side (heritage: dAppServer POC) |
| Slot composition | `<slot name="L-C">` Web Components |
| Shadow DOM | Encapsulation per HLCRF region |
| Custom elements | Registration from WASM or Deno |

Deno's permission model (`--allow-net`, `--allow-read`) = I/O fortress principle at the client runtime level.

**Not designed in detail** — plan when Phases 1-3 are proven.

---

## The Full Binary (Context)

Both modules compose into the single binary vision:

```
Core Binary (Go)
├── CoreGUI (Wails — native WebView)
├── FrankenPHP (embedded PHP runtime)
│   ├── Octane (hot in memory)
│   └── CorePHP (framework) → CoreCommerce (WorkStation v2)
├── go-i18n (grammar forward + reversal)
├── go-html (HLCRF rendering)
├── SMSG (encryption)
├── Bouncer (I/O fortress)
└── Entitlements (permission matrix)
```

Phase 4 adds CoreDeno for client-side Web Components.

Go replaces PHP for rendering. PHP stays for business logic. The binary builds binaries — on the machine, for the machine, safe.

---

## Lineage

```
WorkStation Commerce (2008) — I/O fortress, chainable ORM, Bouncer, GearMan, Flexy
    → CorePHP (WorkStation v2) — Laravel as library, HLCRF, Modern Flexy, entitlements
        → go-i18n (grammar forward) — shipped
        → go-i18n/reversal (grammar reverse) — Phase 1
        → go-html (HLCRF in Go) — Phase 2
        → Integration + WASM — Phase 3
        → CoreDeno + Web Components — Phase 4
```

"Dream lofty dreams, and as you dream, so shall you become." — James Allen
