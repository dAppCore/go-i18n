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

func TestImprint_Similar_SameText(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()
	// Use "branch" (noun-only) to avoid dual-class ambiguity with "file" (now both verb and noun).
	tokens := tok.Tokenise("Delete the configuration branch")
	imp1 := NewImprint(tokens)
	imp2 := NewImprint(tokens)

	sim := imp1.Similar(imp2)
	if sim != 1.0 {
		t.Errorf("Same text similarity = %f, want 1.0", sim)
	}
}

func TestImprint_Similar_SimilarText(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	// Use "branch" (noun-only) to avoid dual-class ambiguity with "file" (now both verb and noun).
	imp1 := NewImprint(tok.Tokenise("Delete the configuration branch"))
	imp2 := NewImprint(tok.Tokenise("Deleted the configuration branches"))

	sim := imp1.Similar(imp2)
	if sim < 0.3 {
		t.Errorf("Similar text similarity = %f, want >= 0.3", sim)
	}
	if sim >= 1.0 {
		t.Errorf("Different text similarity = %f, want < 1.0", sim)
	}
}

func TestImprint_Similar_DifferentText(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	imp1 := NewImprint(tok.Tokenise("Delete the configuration branch"))
	imp2 := NewImprint(tok.Tokenise("Building the project successfully"))

	sim := imp1.Similar(imp2)
	if sim > 0.7 {
		t.Errorf("Different text similarity = %f, want <= 0.7", sim)
	}
}

func TestImprint_Similar_Empty(t *testing.T) {
	imp1 := NewImprint(nil)
	imp2 := NewImprint(nil)
	sim := imp1.Similar(imp2)
	if sim != 1.0 {
		t.Errorf("Empty imprint similarity = %f, want 1.0", sim)
	}
}
