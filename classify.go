package i18n

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
		batchSize:      8,
		promptField:    "prompt",
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

// WithPromptTemplate sets the classification prompt. Use %s for the text placeholder.
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
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

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
	if err := flush(); err != nil {
		return stats, err
	}

	stats.Duration = time.Since(start)
	if stats.Duration > 0 {
		stats.PromptsPerSec = float64(stats.Total) / stats.Duration.Seconds()
	}

	return stats, nil
}
