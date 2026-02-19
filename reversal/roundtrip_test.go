package reversal

import (
	"testing"

	i18n "forge.lthn.ai/core/go-i18n"
)

// TestRoundTrip_ForwardThenReverse — go-i18n composed output → reversal → verify correct tokens
func TestRoundTrip_ForwardThenReverse(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	tests := []struct {
		name      string
		text      string
		wantVerb  string
		wantTense string
	}{
		{
			name:      "Progress pattern",
			text:      i18n.Progress("build"),
			wantVerb:  "build",
			wantTense: "gerund",
		},
		{
			name:      "ActionResult pattern",
			text:      i18n.ActionResult("delete", "file"),
			wantVerb:  "delete",
			wantTense: "past",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tok.Tokenise(tt.text)
			foundVerb := false
			for _, tok := range tokens {
				if tok.Type == TokenVerb && tok.VerbInfo.Base == tt.wantVerb {
					foundVerb = true
					if tok.VerbInfo.Tense != tt.wantTense {
						t.Errorf("verb %q tense = %q, want %q", tt.wantVerb, tok.VerbInfo.Tense, tt.wantTense)
					}
				}
			}
			if !foundVerb {
				t.Errorf("did not find verb %q in tokens from %q", tt.wantVerb, tt.text)
			}
		})
	}
}

// TestRoundTrip_MultiplierImprints — variants should be similar to original
func TestRoundTrip_MultiplierImprints(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()
	m := NewMultiplier()

	original := "Delete the configuration file"
	variants := m.Expand(original)
	origImprint := NewImprint(tok.Tokenise(original))

	for _, v := range variants {
		if v == original {
			continue
		}
		varImprint := NewImprint(tok.Tokenise(v))
		sim := origImprint.Similar(varImprint)
		if sim < 0.2 {
			t.Errorf("Variant %q similarity to original = %f, want >= 0.2", v, sim)
		}
	}
}

// TestRoundTrip_SimilarDocuments — similar docs → higher similarity than different docs
func TestRoundTrip_SimilarDocuments(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	imp1 := NewImprint(tok.Tokenise("Delete the configuration file"))
	imp2 := NewImprint(tok.Tokenise("Delete the old file"))
	imp3 := NewImprint(tok.Tokenise("Building the project successfully"))

	simSame := imp1.Similar(imp2)
	simDiff := imp1.Similar(imp3)

	if simSame <= simDiff {
		t.Errorf("Similar documents (%f) should score higher than different (%f)", simSame, simDiff)
	}
}

func TestRoundTrip_DualClassDisambiguation(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	tests := []struct {
		name     string
		text     string
		word     string
		wantType TokenType
	}{
		{"commit as noun", "Delete the commit", "commit", TokenNoun},
		{"commit as verb", "Commit the changes", "commit", TokenVerb},
		{"run as verb", "Run the tests", "run", TokenVerb},
		{"test as noun", "The test passed", "test", TokenNoun},
		{"build as verb", "Build the project", "build", TokenVerb},
		{"build as noun", "The build failed", "build", TokenNoun},
		{"check as noun", "The check passed", "check", TokenNoun},
		{"check as verb", "Check the logs", "check", TokenVerb},
		{"file as noun", "Delete the file", "file", TokenNoun},
		{"file as verb", "File the report", "file", TokenVerb},
		{"test as verb after aux", "will test the system", "test", TokenVerb},
		{"run as noun with possessive", "his run was fast", "run", TokenNoun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tok.Tokenise(tt.text)
			found := false
			for _, token := range tokens {
				if token.Lower == tt.word {
					found = true
					if token.Type != tt.wantType {
						t.Errorf("%q in %q: got Type %v, want %v (Confidence: %.2f)",
							tt.word, tt.text, token.Type, tt.wantType, token.Confidence)
					}
				}
			}
			if !found {
				t.Errorf("did not find %q in tokens from %q", tt.word, tt.text)
			}
		})
	}
}

func TestRoundTrip_DualClassImprintConvergence(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	// Two texts using "commit" as noun should produce similar imprints
	imp1 := NewImprint(tok.Tokenise("the commit was approved"))
	imp2 := NewImprint(tok.Tokenise("the commit was merged"))

	sim := imp1.Similar(imp2)
	if sim < 0.7 {
		t.Errorf("Same-role imprint similarity = %f, want >= 0.7", sim)
	}

	// Text using "commit" as verb should diverge more
	imp3 := NewImprint(tok.Tokenise("Commit the changes now"))
	simDiff := imp1.Similar(imp3)

	if simDiff >= sim {
		t.Errorf("Different-role similarity (%f) should be less than same-role (%f)",
			simDiff, sim)
	}
}

func TestDisambiguationStats(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	tokens := tok.Tokenise("The commit was approved")
	stats := DisambiguationStatsFromTokens(tokens)

	if stats.TotalTokens != 4 {
		t.Errorf("TotalTokens = %d, want 4", stats.TotalTokens)
	}
	if stats.AmbiguousTokens < 1 {
		t.Errorf("AmbiguousTokens = %d, want >= 1 ('commit' is dual-class)", stats.AmbiguousTokens)
	}
	if stats.ResolvedAsNoun < 1 {
		t.Errorf("ResolvedAsNoun = %d, want >= 1", stats.ResolvedAsNoun)
	}
	if stats.AvgConfidence == 0 {
		t.Error("AvgConfidence = 0, want > 0")
	}
}

func TestDisambiguationStats_NoAmbiguity(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	tok := NewTokeniser()

	tokens := tok.Tokenise("Deleted the files")
	stats := DisambiguationStatsFromTokens(tokens)

	if stats.AmbiguousTokens != 0 {
		t.Errorf("AmbiguousTokens = %d, want 0 (no dual-class words)", stats.AmbiguousTokens)
	}
	if stats.AvgConfidence != 1.0 {
		t.Errorf("AvgConfidence = %f, want 1.0 (all unambiguous)", stats.AvgConfidence)
	}
}
