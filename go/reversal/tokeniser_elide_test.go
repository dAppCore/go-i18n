package reversal

import (
	"testing"

	i18n "dappco.re/go/i18n"
)

// TestTokeniserElide_classifyElidedFrenchWord drives the reachable
// classification branches of the elided-word classifier directly: article,
// dual-class base form (ambiguous), verb-only, noun-only, word-category and
// unknown. The matchers operate against the current language's tables, so
// English inputs exercise the same control flow used for French elided words.
//
//	classifyElidedFrenchWord("the")     // TokenArticle
//	classifyElidedFrenchWord("deleted") // TokenVerb
//	classifyElidedFrenchWord("xyzzy")   // TokenUnknown
func TestTokeniserElide_classifyElidedFrenchWord(t *testing.T) {
	svc, err := valueFromResult[*i18n.Service](i18n.New())
	if err != nil {
		t.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)
	tk := NewTokeniser()

	tests := []struct {
		word      string
		wantType  TokenType
		wantConf  float64
		dualClass bool
	}{
		{"the", TokenArticle, 1.0, false},
		{"deleted", TokenVerb, 1.0, false},
		{"files", TokenNoun, 1.0, false},
		{"file", tokenAmbiguous, 0.0, true}, // dual-class base form → ambiguous
		{"run", tokenAmbiguous, 0.0, true},
		{"xyzzy", TokenUnknown, 0.0, false},
	}
	for _, tt := range tests {
		tok := tk.classifyElidedFrenchWord(tt.word)
		if tok.Type != tt.wantType {
			t.Errorf("classifyElidedFrenchWord(%q).Type = %d, want %d", tt.word, tok.Type, tt.wantType)
		}
		if tok.Confidence != tt.wantConf {
			t.Errorf("classifyElidedFrenchWord(%q).Confidence = %.1f, want %.1f", tt.word, tok.Confidence, tt.wantConf)
		}
		if tt.dualClass {
			// Ambiguous dual-class tokens carry both verb and noun info.
			if tok.VerbInfo.Base == "" || tok.NounInfo.Base == "" {
				t.Errorf("classifyElidedFrenchWord(%q) dual-class should populate both VerbInfo and NounInfo", tt.word)
			}
		}
	}
}

// TestTokeniserElide_classifyElidedFrenchWord_WordCategory covers the
// MatchWord branch using a word that is neither verb, noun nor article but is
// present in the grammar word map.
func TestTokeniserElide_classifyElidedFrenchWord_WordCategory(t *testing.T) {
	svc, err := valueFromResult[*i18n.Service](i18n.New())
	if err != nil {
		t.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)
	tk := NewTokeniser()

	// "status" is a grammar word-map entry, not a verb/noun match path.
	tok := tk.classifyElidedFrenchWord("status")
	if tok.Type == TokenUnknown {
		t.Errorf("classifyElidedFrenchWord(status) = TokenUnknown, want a recognised type")
	}
}
