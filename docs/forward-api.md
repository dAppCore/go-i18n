---
title: Forward Composition API
description: The consumer-facing API for composing grammatically correct text from base forms.
---

# Forward Composition API

The forward API composes grammatically correct text from base forms. This is the consumer-facing side of go-i18n -- applications call these functions to build messages.

## Service Setup

```go
// Option 1: Default service (lazy-initialised, English)
svc := i18n.Default()

// Option 2: Explicit creation with options
svc, err := i18n.New(
    i18n.WithFallback("en"),
    i18n.WithDefaultHandlers(),
)
i18n.SetDefault(svc)

// Option 3: Custom loader (your own filesystem)
loader := i18n.NewFSLoader(myFS, "locales")
svc, err := i18n.NewWithLoader(loader)

// Option 4: Load from an arbitrary fs.FS
svc, err := i18n.NewWithFS(myFS, "locales")
```

The service automatically detects the system language from `LANG`, `LC_ALL`, or `LC_MESSAGES` environment variables using BCP 47 tag matching.

### Options

| Option | Effect |
|--------|--------|
| `WithFallback("en")` | Set fallback language for missing translations |
| `WithDefaultHandlers()` | Register the six built-in `i18n.*` namespace handlers |
| `WithHandlers(h...)` | Replace handlers entirely |
| `WithMode(ModeStrict)` | Panic on missing keys (useful in CI) |
| `WithFormality(FormalityFormal)` | Default formality level |
| `WithDebug(true)` | Prefix output with key path for debugging |

### Translation Modes

| Mode | Behaviour |
|------|-----------|
| `ModeNormal` | Returns key as-is when missing (production) |
| `ModeStrict` | Panics on missing key (dev/CI) |
| `ModeCollect` | Dispatches `MissingKey` events, returns `[key]` (QA) |

## Grammar Primitives

### PastTense(verb) -> string

Returns the past tense of a verb using a 3-tier lookup: JSON grammar data, then irregular Go map, then regular morphology rules.

```go
i18n.PastTense("delete")  // "deleted"
i18n.PastTense("commit")  // "committed"
i18n.PastTense("go")      // "went"       (irregular)
i18n.PastTense("run")     // "ran"        (irregular)
i18n.PastTense("copy")    // "copied"     (regular rule: consonant+y -> ied)
```

Regular rules applied in order:
1. Already ends in `-ed` with non-vowel, non-e third-from-end -- return as-is
2. Ends in `-e` -- append `d`
3. Ends in consonant + `y` -- replace `y` with `ied`
4. CVC doubling applies -- double final consonant + `ed`
5. Default -- append `ed`

### Gerund(verb) -> string

Returns the present participle (-ing form) of a verb.

```go
i18n.Gerund("delete")  // "deleting"
i18n.Gerund("commit")  // "committing"
i18n.Gerund("run")     // "running"
i18n.Gerund("die")     // "dying"       (ie -> ying)
```

### Pluralize(noun, count) -> string

Returns singular for count=1, plural otherwise.

```go
i18n.Pluralize("file", 1)     // "file"
i18n.Pluralize("file", 5)     // "files"
i18n.Pluralize("person", 3)   // "people"    (irregular)
i18n.Pluralize("child", 2)    // "children"  (irregular)
```

### PluralForm(noun) -> string

Always returns the plural form (no count check).

```go
i18n.PluralForm("repository")  // "repositories"
i18n.PluralForm("child")       // "children"
i18n.PluralForm("wolf")        // "wolves"
```

### Article(word) -> string

Returns `"a"` or `"an"` based on phonetic rules, not spelling.

```go
i18n.Article("file")     // "a"
i18n.Article("error")    // "an"
i18n.Article("user")     // "a"   (sounds like "yoo-zer")
i18n.Article("hour")     // "an"  (silent h)
i18n.Article("SSH")      // "an"  (sounds like "ess-ess-aitch")
```

Uses consonant/vowel sound exception maps, falling back to first-letter vowel check.

### Utility Functions

```go
i18n.Title("hello world")  // "Hello World"
i18n.Quote("config.yaml")  // "\"config.yaml\""
i18n.Progress("build")     // "Building..."
i18n.ProgressSubject("build", "project")  // "Building project..."
i18n.ActionResult("delete", "file")       // "File deleted"
i18n.ActionFailed("push", "commits")      // "Failed to push commits"
i18n.Label("status")                      // "Status:"
```

## T() -- Core Translation Function

`T()` resolves message IDs through a handler chain, then falls back to direct key lookup with language fallback.

```go
i18n.T("greeting")                    // Direct key lookup
i18n.T("i18n.label.status")           // Via LabelHandler -> "Status:"
i18n.T("i18n.progress.build")         // Via ProgressHandler -> "Building..."
i18n.T("i18n.count.file", 5)          // Via CountHandler -> "5 files"
```

Resolution order:
1. Run through handler chain (stops at first match)
2. Look up key in current language messages
3. Look up key in fallback language messages
4. Try `common.action.{verb}` and `common.{verb}` variants
5. Handle missing key according to current mode

### Raw()

`Raw()` translates without running the `i18n.*` namespace handler chain. Useful when you want direct key lookup only.

```go
i18n.Raw("my.custom.key")  // Direct lookup, no handler magic
```

## Magic Namespace Handlers

Six built-in handlers are registered by `WithDefaultHandlers()`. They intercept keys matching their `i18n.*` prefix and compose output from grammar primitives.

### i18n.label.* -- LabelHandler

Produces labelled output with locale-specific suffix.

```go
T("i18n.label.status")    // "Status:"
T("i18n.label.progress")  // "Progress:"
```

The suffix is language-specific: English uses `:`, French uses ` :` (space before colon).

### i18n.progress.* -- ProgressHandler

Produces gerund-form progress messages.

```go
T("i18n.progress.build")              // "Building..."
T("i18n.progress.build", "project")   // "Building project..."
T("i18n.progress.delete", "cache")    // "Deleting cache..."
```

### i18n.count.* -- CountHandler

Produces pluralised count messages.

```go
T("i18n.count.file", 1)     // "1 file"
T("i18n.count.file", 5)     // "5 files"
T("i18n.count.person", 3)   // "3 people"
```

### i18n.done.* -- DoneHandler

Produces past-tense completion messages.

```go
T("i18n.done.delete", "config.yaml")  // "Config.yaml deleted"
T("i18n.done.push", "commits")        // "Commits pushed"
T("i18n.done.delete")                 // "Deleted"
```

### i18n.fail.* -- FailHandler

Produces failure messages.

```go
T("i18n.fail.push", "commits")   // "Failed to push commits"
T("i18n.fail.delete")            // "Failed to delete"
```

### i18n.numeric.* -- NumericHandler

Locale-aware number formatting.

```go
T("i18n.numeric.number", 1234567)   // "1,234,567"
T("i18n.numeric.decimal", 3.14)     // "3.14"
T("i18n.numeric.percent", 0.85)     // "85%"
T("i18n.numeric.bytes", 1536000)    // "1.46 MB"
T("i18n.numeric.ordinal", 3)        // "3rd"
T("i18n.numeric.ago", 5, "minutes") // "5 minutes ago"
```

The shorthand `N()` function wraps this namespace:

```go
i18n.N("number", 1234567)  // "1,234,567"
i18n.N("percent", 0.85)    // "85%"
i18n.N("bytes", 1536000)   // "1.46 MB"
i18n.N("ordinal", 1)       // "1st"
```

## Subject Builder -- S()

Builds semantic context for complex translations:

```go
s := i18n.S("file", "config.yaml")
s.Count(3)          // Set plural count
s.Gender("neuter")  // Grammatical gender (for gendered languages)
s.In("workspace")   // Location context
s.Formal()          // Formal register

// All methods chain:
i18n.S("file", "config.yaml").Count(3).In("workspace").Formal()
```

The Subject carries metadata that gendered/formal language systems (French, German, Japanese) use to select correct grammatical forms. English mostly ignores gender/formality, but the API is language-agnostic.

## Translation Context -- C()

Provides disambiguation for translations where the same key has different meanings in different contexts:

```go
i18n.T("direction.right", i18n.C("navigation"))  // "rechts" (German)
i18n.T("status.right", i18n.C("correctness"))    // "richtig" (German)
```

Context can carry gender and formality:

```go
ctx := i18n.C("greeting").WithGender("feminine").Formal()
i18n.T("welcome", ctx)
```

## Custom Handlers

Implement `KeyHandler` to add your own namespace handlers:

```go
type KeyHandler interface {
    Match(key string) bool
    Handle(key string, args []any, next func() string) string
}
```

Register them on the service:

```go
i18n.AddHandler(myHandler)       // Append to chain
i18n.PrependHandler(myHandler)   // Insert at start
```

Each handler receives a `next` function to delegate to the rest of the chain -- this is a middleware pattern.

## Registering External Locales

Packages can register their own locale files to be loaded when the default service initialises:

```go
//go:embed locales/*.json
var localeFS embed.FS

func init() {
    i18n.RegisterLocales(localeFS, "locales")
}
```

If the service is already initialised, the locales are loaded immediately. Otherwise they are queued and loaded during `Init()`.

## CLDR Plural Rules

The service supports CLDR plural categories with rules for English, German, French, Spanish, Russian, Polish, Arabic, Chinese, Japanese, and Korean:

| Category | Example Languages |
|----------|-------------------|
| `PluralZero` | Arabic (n=0) |
| `PluralOne` | Most languages (n=1) |
| `PluralTwo` | Arabic, Welsh (n=2) |
| `PluralFew` | Slavic (2-4), Arabic (3-10) |
| `PluralMany` | Slavic (5+), Arabic (11-99) |
| `PluralOther` | Default/fallback |

Messages can define plural forms in locale JSON:

```json
{
  "item_count": {
    "one": "{{.Count}} item",
    "other": "{{.Count}} items",
    "zero": "No items"
  }
}
```

## Time and Relative Dates

```go
i18n.TimeAgo(time.Now().Add(-5 * time.Minute))  // "5 minutes ago"
i18n.FormatAgo(3, "hour")                        // "3 hours ago"
```

## Template Functions

All grammar functions are available as Go template functions via `TemplateFuncs()`:

```go
template.New("").Funcs(i18n.TemplateFuncs())
```

Available functions: `title`, `lower`, `upper`, `past`, `gerund`, `plural`, `pluralForm`, `article`, `quote`, `label`, `progress`, `progressSubject`, `actionResult`, `actionFailed`, `timeAgo`, `formatAgo`.
