package reversal

import (
	"testing"

	i18n "forge.lthn.ai/core/go-i18n"
)

func TestNewImprint(t *testing.T) {
	svc, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)

	tok := NewTokeniser()
	tokens := tok.Tokenise("Deleted the configuration files successfully")
	imp := NewImprint(tokens)

	if imp.TokenCount != 5 {
		t.Errorf("TokenCount = %d, want 5", imp.TokenCount)
	}
	if imp.UniqueVerbs == 0 {
		t.Error("UniqueVerbs = 0, want > 0")
	}
	if imp.UniqueNouns == 0 {
		t.Error("UniqueNouns = 0, want > 0")
	}
	if imp.TenseDistribution["past"] == 0 {
		t.Error("TenseDistribution[\"past\"] = 0, want > 0")
	}
	if imp.ArticleUsage["definite"] == 0 {
		t.Error("ArticleUsage[\"definite\"] = 0, want > 0")
	}
}

func TestNewImprint_Empty(t *testing.T) {
	imp := NewImprint(nil)
	if imp.TokenCount != 0 {
		t.Errorf("TokenCount = %d, want 0", imp.TokenCount)
	}
}

func TestNewImprint_PluralRatio(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	// All plural nouns
	tokens := tok.Tokenise("files branches repositories")
	imp := NewImprint(tokens)
	if imp.PluralRatio < 0.5 {
		t.Errorf("PluralRatio = %f for all-plural input, want >= 0.5", imp.PluralRatio)
	}

	// All singular nouns
	tokens = tok.Tokenise("file branch repository")
	imp = NewImprint(tokens)
	if imp.PluralRatio > 0.5 {
		t.Errorf("PluralRatio = %f for all-singular input, want <= 0.5", imp.PluralRatio)
	}
}
