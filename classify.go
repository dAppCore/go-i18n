package i18n

import (
	"context"
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

// Ensure imports are used (ClassifyCorpus will be added in Task 3).
var (
	_ = (*inference.Token)(nil)
	_ context.Context
	_ io.Reader
	_ time.Duration
)
