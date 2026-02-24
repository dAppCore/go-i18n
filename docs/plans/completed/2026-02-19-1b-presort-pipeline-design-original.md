# 1B Pre-Sort Pipeline Design

## Purpose

Batch-classify a JSONL corpus through Gemma3-1B to add `domain_1b` labels ({technical, creative, ethical, casual}). This pre-sorts the 88K Phase 0 seeds for reference distribution building (Phase 2b).

## API

```go
// ClassifyCorpus reads JSONL from input, batch-classifies each entry's text
// through model, and writes JSONL with domain_1b field added to output.
// The original domain field is preserved for comparison.
func ClassifyCorpus(ctx context.Context, model inference.TextModel,
    input io.Reader, output io.Writer, opts ...ClassifyOption) (*ClassifyStats, error)
```

Caller manages model lifecycle — load once, classify multiple files.

## Configuration

```go
type ClassifyOption func(*classifyConfig)

type classifyConfig struct {
    BatchSize      int    // default 8
    PromptField    string // JSONL field containing text to classify (default "prompt")
    PromptTemplate string // template with %s placeholder for text
}

func WithBatchSize(n int) ClassifyOption
func WithPromptField(field string) ClassifyOption
func WithPromptTemplate(tmpl string) ClassifyOption
```

## Data Flow

```
input JSONL → read line → unmarshal to map[string]any → extract prompt field
    → accumulate batch (8 prompts) → model.Classify(ctx, batch, WithMaxTokens(1))
    → parse token → map to 4-way domain → add "domain_1b" to map → marshal → write line
```

## Prompt Template

Default:

```
Classify this text into exactly one category: technical, creative, ethical, casual.

Text: %s

Category:
```

Model generates one token. Map first character: t→technical, c→creative/casual (disambiguate by checking full token), e→ethical. Fallback: "unknown".

## Token-to-Domain Mapping

```go
func mapTokenToDomain(token string) string
```

Strategy: lowercase the token text, match prefix:
- "tech" → technical
- "cre" → creative
- "eth" → ethical
- "cas" → casual
- anything else → "unknown"

## Return Type

```go
type ClassifyStats struct {
    Total         int
    Skipped       int            // malformed lines
    ByDomain      map[string]int // domain_1b label → count
    Duration      time.Duration
    PromptsPerSec float64
}
```

## Error Handling

- Malformed JSON lines: skip, increment Skipped counter
- Missing prompt field: skip, increment Skipped
- Model Classify error: return immediately (GPU errors are fatal)
- Partial batch at EOF: flush normally
- Context cancellation: return partial stats + context error

## File Location

`classify.go` in root package. Build tag: none (but `inference` import means consumers need go-inference).

**Important**: go-inference is a soft dependency. The classify.go file imports it, but the core grammar engine (PastTense, Gerund, etc.) does not. Consumers who only need grammar primitives never touch classify.go.

## Testing

- `classify_test.go`: unit tests with mock TextModel (satisfies interface, returns canned tokens)
- `classify_integration_test.go` with `//go:build integration`: real model, real JSONL, validates throughput target (152+ prompts/sec)

## Performance Target

88K corpus at 152 prompts/sec = ~10 minutes. Batch size 8 balances throughput and memory.

## Dependencies Added

```
require forge.lthn.ai/core/go-inference v0.0.0
require forge.lthn.ai/core/go-mlx v0.0.0

replace forge.lthn.ai/core/go-inference => ../go-inference
replace forge.lthn.ai/core/go-mlx => ../go-mlx
```

go-mlx registers the "metal" backend via `init()` — import with blank identifier in integration tests and any CLI caller.
