// SPDX-Licence-Identifier: EUPL-1.2

package i18n

import (
	"context"
	"fmt"
	"strings"

	"forge.lthn.ai/core/go-inference"
	corelog "forge.lthn.ai/core/go-log"
)

// ArticlePair holds a noun and its proposed article for validation.
type ArticlePair struct {
	Noun    string
	Article string
}

// ArticleResult reports whether a given article usage is grammatically correct.
type ArticleResult struct {
	Noun      string // the noun being checked
	Given     string // the article provided by the caller
	Predicted string // what the model predicted
	Valid     bool   // Given == Predicted
	Prompt    string // the prompt used (for debugging)
}

// IrregularForm holds a verb, tense, and proposed inflected form for validation.
type IrregularForm struct {
	Verb  string
	Tense string
	Form  string
}

// IrregularResult reports whether a given irregular verb form is correct.
type IrregularResult struct {
	Verb      string // base verb
	Tense     string // tense being checked (e.g. "past", "past participle")
	Given     string // the form provided by the caller
	Predicted string // what the model predicted
	Valid     bool   // Given == Predicted
	Prompt    string // the prompt used (for debugging)
}

// articlePrompt builds a fill-in-the-blank prompt for article prediction.
func articlePrompt(noun string) string {
	return fmt.Sprintf(
		"Complete with the correct article (a/an/the): ___ %s. Answer with just the article:",
		noun,
	)
}

// irregularPrompt builds a fill-in-the-blank prompt for irregular verb prediction.
func irregularPrompt(verb, tense string) string {
	return fmt.Sprintf(
		"What is the %s form of the verb '%s'? Answer with just the word:",
		tense, verb,
	)
}

// collectGenerated runs a single-token generation and returns the trimmed, lowercased output.
func collectGenerated(ctx context.Context, m inference.TextModel, prompt string) (string, error) {
	var sb strings.Builder
	for tok := range m.Generate(ctx, prompt, inference.WithMaxTokens(1), inference.WithTemperature(0.05)) {
		sb.WriteString(tok.Text)
	}
	if err := m.Err(); err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.ToLower(sb.String())), nil
}

// ValidateArticle checks whether a given article usage is grammatically correct
// by asking the model to predict the correct article in context.
// Uses single-token generation with near-zero temperature for deterministic output.
func ValidateArticle(ctx context.Context, m inference.TextModel, noun string, article string) (ArticleResult, error) {
	prompt := articlePrompt(noun)
	predicted, err := collectGenerated(ctx, m, prompt)
	if err != nil {
		return ArticleResult{}, corelog.E("i18n.ValidateArticle", fmt.Sprintf("validate article %q", noun), err)
	}
	given := strings.TrimSpace(strings.ToLower(article))
	return ArticleResult{
		Noun:      noun,
		Given:     given,
		Predicted: predicted,
		Valid:     given == predicted,
		Prompt:    prompt,
	}, nil
}

// ValidateIrregular checks whether a given irregular verb form is correct
// by asking the model to predict the correct form in context.
// Uses single-token generation with near-zero temperature for deterministic output.
func ValidateIrregular(ctx context.Context, m inference.TextModel, verb string, tense string, form string) (IrregularResult, error) {
	prompt := irregularPrompt(verb, tense)
	predicted, err := collectGenerated(ctx, m, prompt)
	if err != nil {
		return IrregularResult{}, corelog.E("i18n.ValidateIrregular", fmt.Sprintf("validate irregular %q (%s)", verb, tense), err)
	}
	given := strings.TrimSpace(strings.ToLower(form))
	return IrregularResult{
		Verb:      verb,
		Tense:     tense,
		Given:     given,
		Predicted: predicted,
		Valid:     given == predicted,
		Prompt:    prompt,
	}, nil
}

// BatchValidateArticles validates multiple article-noun pairs efficiently.
// Each pair is validated independently via single-token generation.
func BatchValidateArticles(ctx context.Context, m inference.TextModel, pairs []ArticlePair) ([]ArticleResult, error) {
	results := make([]ArticleResult, 0, len(pairs))
	for _, p := range pairs {
		r, err := ValidateArticle(ctx, m, p.Noun, p.Article)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}

// BatchValidateIrregulars validates multiple irregular verb forms efficiently.
// Each form is validated independently via single-token generation.
func BatchValidateIrregulars(ctx context.Context, m inference.TextModel, forms []IrregularForm) ([]IrregularResult, error) {
	results := make([]IrregularResult, 0, len(forms))
	for _, f := range forms {
		r, err := ValidateIrregular(ctx, m, f.Verb, f.Tense, f.Form)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}
