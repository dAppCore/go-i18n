# 1B Pre-Sort Pipeline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `ClassifyCorpus()` function that batch-classifies JSONL text through a Gemma3-1B model and writes back JSONL with `domain_1b` labels.

**Architecture:** Streaming JSONL reader batches prompts into groups of 8, sends each batch to `inference.TextModel.Classify()`, maps the single-token response to a 4-way domain label, and writes the augmented line to output. Caller manages model lifecycle.

**Tech Stack:** go-inference (TextModel interface), go-mlx (Metal backend, integration tests only)

---

### Task 1: Add go-inference dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add go-inference require and replace directives**

```bash
cd /Users/snider/Code/go-i18n
go mod edit -require forge.lthn.ai/core/go-inference@v0.0.0
go mod edit -replace forge.lthn.ai/core/go-inference=../go-inference
```

**Step 2: Verify module resolves**

Run: `go mod tidy`
Expected: clean exit, go.sum updated

**Step 3: Verify existing tests still pass**

Run: `go test ./...`
Expected: all PASS

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add go-inference dependency

Co-Authored-By: Virgil <virgil@lethean.io>
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 2: Write classify types and token-to-domain mapper

**Files:**
- Create: `classify.go`
- Create: `classify_test.go`

**Step 1: Write the failing test for mapTokenToDomain**

In `classify_test.go`:

```go
package i18n

import "testing"

func TestMapTokenToDomain(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"technical", "technical"},
		{"Technical", "technical"},
		{"tech", "technical"},
		{"creative", "creative"},
		{"Creative", "creative"},
		{"cre", "creative"},
		{"ethical", "ethical"},
		{"Ethical", "ethical"},
		{"eth", "ethical"},
		{"casual", "casual"},
		{"Casual", "casual"},
		{"cas", "casual"},
		{"unknown", "unknown"},
		{"", "unknown"},
		{"foo", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := mapTokenToDomain(tt.token)
			if got != tt.want {
				t.Errorf("mapTokenToDomain(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestMapTokenToDomain ./...`
Expected: FAIL — `mapTokenToDomain` undefined

**Step 3: Write classify.go with types and mapper**

In `classify.go`:

```go
package i18n

import (
	"context"
	"io"
	"time"

	"forge.lthn.ai/core/go-inference"
)

// ClassifyStats reports metrics from a ClassifyCorpus run.
type ClassifyStats struct {
	Total         int
	Skipped       int            // malformed or missing prompt field
	ByDomain      map[string]int // domain_1b label -> count
	Duration      time.Duration
	PromptsPerSec float64
}

// ClassifyOption configures ClassifyCorpus behaviour.
type ClassifyOption func(*classifyConfig)

type classifyConfig struct {
	batchSize      int
	promptField    string
	promptTemplate string
}

func defaultClassifyConfig() classifyConfig {
	return classifyConfig{
		batchSize:   8,
		promptField: "prompt",
		promptTemplate: "Classify this text into exactly one category: technical, creative, ethical, casual.\n\nText: %s\n\nCategory:",
	}
}

// WithBatchSize sets the number of prompts per Classify call. Default 8.
func WithBatchSize(n int) ClassifyOption {
	return func(c *classifyConfig) { c.batchSize = n }
}

// WithPromptField sets which JSONL field contains the text to classify. Default "prompt".
func WithPromptField(field string) ClassifyOption {
	return func(c *classifyConfig) { c.promptField = field }
}

// WithPromptTemplate sets the classification prompt. Use %s for the text placeholder. Default is a 4-way domain classifier.
func WithPromptTemplate(tmpl string) ClassifyOption {
	return func(c *classifyConfig) { c.promptTemplate = tmpl }
}

// mapTokenToDomain maps a model output token to a 4-way domain label.
func mapTokenToDomain(token string) string {
	if len(token) == 0 {
		return "unknown"
	}
	lower := strings.ToLower(token)
	switch {
	case strings.HasPrefix(lower, "tech"):
		return "technical"
	case strings.HasPrefix(lower, "cre"):
		return "creative"
	case strings.HasPrefix(lower, "eth"):
		return "ethical"
	case strings.HasPrefix(lower, "cas"):
		return "casual"
	default:
		return "unknown"
	}
}

// Ensure inference import is used (ClassifyCorpus added in Task 3).
var _ inference.TextModel
```

Note: Add `"strings"` to the import block.

**Step 4: Run test to verify it passes**

Run: `go test -run TestMapTokenToDomain ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add classify.go classify_test.go
git commit -m "feat: add classify types and token-to-domain mapper

Co-Authored-By: Virgil <virgil@lethean.io>
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 3: Write ClassifyCorpus with mock tests

**Files:**
- Modify: `classify.go` — add `ClassifyCorpus` function
- Modify: `classify_test.go` — add mock model and pipeline tests

**Step 1: Write failing tests for ClassifyCorpus**

Add to `classify_test.go`:

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"forge.lthn.ai/core/go-inference"
)

// mockModel satisfies inference.TextModel for testing.
type mockModel struct {
	classifyFunc func(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error)
}

func (m *mockModel) Classify(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error) {
	return m.classifyFunc(ctx, prompts, opts...)
}

// Stub methods to satisfy interface.
func (m *mockModel) Generate(ctx context.Context, prompt string, opts ...inference.GenerateOption) iter.Seq[inference.Token] { return nil }
func (m *mockModel) Chat(ctx context.Context, msgs []inference.Message, opts ...inference.GenerateOption) iter.Seq[inference.Token] { return nil }
func (m *mockModel) BatchGenerate(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.BatchResult, error) { return nil, nil }
func (m *mockModel) ModelType() string { return "mock" }
func (m *mockModel) Info() inference.ModelInfo { return inference.ModelInfo{} }
func (m *mockModel) Metrics() inference.GenerateMetrics { return inference.GenerateMetrics{} }
func (m *mockModel) Err() error { return nil }
func (m *mockModel) Close() error { return nil }

func TestClassifyCorpus_Basic(t *testing.T) {
	mock := &mockModel{
		classifyFunc: func(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error) {
			results := make([]inference.ClassifyResult, len(prompts))
			for i := range prompts {
				results[i] = inference.ClassifyResult{Token: inference.Token{Text: "technical"}}
			}
			return results, nil
		},
	}

	input := strings.NewReader(`{"seed_id":"test1","domain":"Technical","prompt":"Delete the file"}
{"seed_id":"test2","domain":"Ethics","prompt":"We should be fair"}
`)

	var output bytes.Buffer
	stats, err := ClassifyCorpus(context.Background(), mock, input, &output, WithBatchSize(2))
	if err != nil {
		t.Fatalf("ClassifyCorpus error: %v", err)
	}

	if stats.Total != 2 {
		t.Errorf("Total = %d, want 2", stats.Total)
	}
	if stats.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", stats.Skipped)
	}

	// Parse output lines
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("output lines = %d, want 2", len(lines))
	}
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("unmarshal output: %v", err)
		}
		if m["domain_1b"] != "technical" {
			t.Errorf("domain_1b = %v, want technical", m["domain_1b"])
		}
		// Original domain field preserved
		if _, ok := m["domain"]; !ok {
			t.Error("original domain field missing")
		}
	}
}

func TestClassifyCorpus_SkipsMalformed(t *testing.T) {
	mock := &mockModel{
		classifyFunc: func(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error) {
			results := make([]inference.ClassifyResult, len(prompts))
			for i := range prompts {
				results[i] = inference.ClassifyResult{Token: inference.Token{Text: "technical"}}
			}
			return results, nil
		},
	}

	input := strings.NewReader(`not json at all
{"seed_id":"test1","prompt":"Delete the file"}
{"seed_id":"test2","no_prompt_field":"oops"}
`)

	var output bytes.Buffer
	stats, err := ClassifyCorpus(context.Background(), mock, input, &output)
	if err != nil {
		t.Fatalf("ClassifyCorpus error: %v", err)
	}

	if stats.Total != 1 {
		t.Errorf("Total = %d, want 1", stats.Total)
	}
	if stats.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", stats.Skipped)
	}
}

func TestClassifyCorpus_DomainMapping(t *testing.T) {
	// Mock returns different domains based on prompt content
	mock := &mockModel{
		classifyFunc: func(ctx context.Context, prompts []string, opts ...inference.GenerateOption) ([]inference.ClassifyResult, error) {
			results := make([]inference.ClassifyResult, len(prompts))
			for i, p := range prompts {
				if strings.Contains(p, "Delete") {
					results[i] = inference.ClassifyResult{Token: inference.Token{Text: "technical"}}
				} else {
					results[i] = inference.ClassifyResult{Token: inference.Token{Text: "ethical"}}
				}
			}
			return results, nil
		},
	}

	input := strings.NewReader(`{"prompt":"Delete the file"}
{"prompt":"We should be fair"}
`)

	var output bytes.Buffer
	stats, err := ClassifyCorpus(context.Background(), mock, input, &output, WithBatchSize(4))
	if err != nil {
		t.Fatalf("ClassifyCorpus error: %v", err)
	}

	if stats.ByDomain["technical"] != 1 {
		t.Errorf("technical count = %d, want 1", stats.ByDomain["technical"])
	}
	if stats.ByDomain["ethical"] != 1 {
		t.Errorf("ethical count = %d, want 1", stats.ByDomain["ethical"])
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run "TestClassifyCorpus" ./...`
Expected: FAIL — `ClassifyCorpus` undefined

**Step 3: Implement ClassifyCorpus**

Add to `classify.go` (replace the `var _ inference.TextModel` line):

```go
// ClassifyCorpus reads JSONL from input, batch-classifies each entry through
// model, and writes JSONL with domain_1b field added to output.
func ClassifyCorpus(ctx context.Context, model inference.TextModel,
	input io.Reader, output io.Writer, opts ...ClassifyOption) (*ClassifyStats, error) {

	cfg := defaultClassifyConfig()
	for _, o := range opts {
		o(&cfg)
	}

	stats := &ClassifyStats{ByDomain: make(map[string]int)}
	start := time.Now()

	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB line limit

	type pending struct {
		record map[string]any
		prompt string
	}

	var batch []pending

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		prompts := make([]string, len(batch))
		for i, p := range batch {
			prompts[i] = fmt.Sprintf(cfg.promptTemplate, p.prompt)
		}
		results, err := model.Classify(ctx, prompts, inference.WithMaxTokens(1))
		if err != nil {
			return fmt.Errorf("classify batch: %w", err)
		}
		for i, r := range results {
			domain := mapTokenToDomain(r.Token.Text)
			batch[i].record["domain_1b"] = domain
			stats.ByDomain[domain]++
			stats.Total++

			line, err := json.Marshal(batch[i].record)
			if err != nil {
				return fmt.Errorf("marshal output: %w", err)
			}
			if _, err := fmt.Fprintf(output, "%s\n", line); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		}
		batch = batch[:0]
		return nil
	}

	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			stats.Skipped++
			continue
		}

		promptVal, ok := record[cfg.promptField]
		if !ok {
			stats.Skipped++
			continue
		}
		prompt, ok := promptVal.(string)
		if !ok || prompt == "" {
			stats.Skipped++
			continue
		}

		batch = append(batch, pending{record: record, prompt: prompt})

		if len(batch) >= cfg.batchSize {
			if err := flush(); err != nil {
				return stats, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return stats, fmt.Errorf("read input: %w", err)
	}

	// Flush remaining
	if err := flush(); err != nil {
		return stats, err
	}

	stats.Duration = time.Since(start)
	if stats.Duration > 0 {
		stats.PromptsPerSec = float64(stats.Total) / stats.Duration.Seconds()
	}

	return stats, nil
}
```

Note: Add `"bufio"`, `"encoding/json"`, `"fmt"` to imports.

**Step 4: Run tests to verify they pass**

Run: `go test -run "TestClassifyCorpus|TestMapTokenToDomain" ./...`
Expected: all PASS

**Step 5: Run full test suite**

Run: `go test ./...`
Expected: all PASS

**Step 6: Commit**

```bash
git add classify.go classify_test.go
git commit -m "feat: implement ClassifyCorpus with streaming batch classification

Co-Authored-By: Virgil <virgil@lethean.io>
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 4: Add integration test (build-tagged)

**Files:**
- Create: `classify_integration_test.go`

**Step 1: Write integration test**

In `classify_integration_test.go`:

```go
//go:build integration

package i18n

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"forge.lthn.ai/core/go-inference"
	_ "forge.lthn.ai/core/go-mlx" // registers Metal backend
)

func TestClassifyCorpus_Integration(t *testing.T) {
	model, err := inference.LoadModel("/Volumes/Data/lem/LEM-Gemma3-1B-layered-v2")
	if err != nil {
		t.Skipf("model not available: %v", err)
	}
	defer model.Close()

	// Build test corpus from classification benchmark sentences
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, fmt.Sprintf(`{"id":%d,"prompt":"Delete the configuration file and rebuild the project"}`, i))
	}
	input := strings.NewReader(strings.Join(lines, "\n") + "\n")

	var output bytes.Buffer
	start := time.Now()
	stats, err := ClassifyCorpus(context.Background(), model, input, &output, WithBatchSize(8))
	if err != nil {
		t.Fatalf("ClassifyCorpus: %v", err)
	}
	elapsed := time.Since(start)

	t.Logf("Classified %d prompts in %v (%.1f prompts/sec)", stats.Total, elapsed, stats.PromptsPerSec)
	t.Logf("By domain: %v", stats.ByDomain)

	if stats.Total != 50 {
		t.Errorf("Total = %d, want 50", stats.Total)
	}
	if stats.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", stats.Skipped)
	}
}
```

**Step 2: Verify it doesn't run without build tag**

Run: `go test ./...`
Expected: all PASS (integration test excluded)

**Step 3: Add go-mlx dependency**

```bash
go mod edit -require forge.lthn.ai/core/go-mlx@v0.0.0
go mod edit -replace forge.lthn.ai/core/go-mlx=../go-mlx
go mod tidy
```

**Step 4: Run integration test (if GPU available)**

Run: `go test -tags integration -run TestClassifyCorpus_Integration -v -timeout 5m ./...`
Expected: PASS with throughput logged

**Step 5: Commit**

```bash
git add classify_integration_test.go go.mod go.sum
git commit -m "test: add integration test for ClassifyCorpus with real model

Co-Authored-By: Virgil <virgil@lethean.io>
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 5: Update TODO.md and FINDINGS.md

**Files:**
- Modify: `TODO.md` — mark 1B pre-sort pipeline as done
- Modify: `FINDINGS.md` — add integration results

**Step 1: Update TODO.md**

Mark the task complete with commit hash.

**Step 2: Update FINDINGS.md**

Add section with integration test results (throughput, domain distribution, any surprises).

**Step 3: Commit**

```bash
git add TODO.md FINDINGS.md
git commit -m "docs: mark 1B pre-sort pipeline complete

Co-Authored-By: Virgil <virgil@lethean.io>
Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```
