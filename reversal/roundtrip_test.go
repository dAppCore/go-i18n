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
