package reversal

import (
	"testing"

	i18n "forge.lthn.ai/core/go-i18n"
)

func setup(t *testing.T) {
	t.Helper()
	svc, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)
}

func TestTokeniser_MatchVerb_Irregular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word    string
		wantOK  bool
		wantBase string
		wantTense string
	}{
		// Irregular past tense
		{"deleted", true, "delete", "past"},
		{"deleting", true, "delete", "gerund"},
		{"went", true, "go", "past"},
		{"going", true, "go", "gerund"},
		{"was", true, "be", "past"},
		{"being", true, "be", "gerund"},
		{"ran", true, "run", "past"},
		{"running", true, "run", "gerund"},
		{"wrote", true, "write", "past"},
		{"writing", true, "write", "gerund"},
		{"built", true, "build", "past"},
		{"building", true, "build", "gerund"},
		{"committed", true, "commit", "past"},
		{"committing", true, "commit", "gerund"},

		// Base forms
		{"delete", true, "delete", "base"},
		{"go", true, "go", "base"},

		// Unknown words return false
		{"xyzzy", false, "", ""},
		{"flurble", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchVerb(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchVerb(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchVerb(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Tense != tt.wantTense {
				t.Errorf("MatchVerb(%q).Tense = %q, want %q", tt.word, match.Tense, tt.wantTense)
			}
		})
	}
}

func TestTokeniser_MatchNoun_Irregular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word       string
		wantOK     bool
		wantBase   string
		wantPlural bool
	}{
		{"files", true, "file", true},
		{"file", true, "file", false},
		{"people", true, "person", true},
		{"person", true, "person", false},
		{"children", true, "child", true},
		{"child", true, "child", false},
		{"repositories", true, "repository", true},
		{"repository", true, "repository", false},
		{"branches", true, "branch", true},
		{"branch", true, "branch", false},
		{"xyzzy", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchNoun(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchNoun(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchNoun(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Plural != tt.wantPlural {
				t.Errorf("MatchNoun(%q).Plural = %v, want %v", tt.word, match.Plural, tt.wantPlural)
			}
		})
	}
}

func TestTokeniser_MatchNoun_Regular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word       string
		wantOK     bool
		wantBase   string
		wantPlural bool
	}{
		// Regular nouns NOT in grammar tables — detected by reverse morphology + round-trip
		{"servers", true, "server", true},
		{"processes", true, "process", true},
		{"entries", true, "entry", true},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchNoun(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchNoun(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchNoun(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Plural != tt.wantPlural {
				t.Errorf("MatchNoun(%q).Plural = %v, want %v", tt.word, match.Plural, tt.wantPlural)
			}
		})
	}
}

func TestTokeniser_MatchWord(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word    string
		wantCat string
		wantOK  bool
	}{
		{"URL", "url", true},
		{"url", "url", true},
		{"ID", "id", true},
		{"SSH", "ssh", true},
		{"PHP", "php", true},
		{"xyzzy", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			cat, ok := tok.MatchWord(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchWord(%q) ok=%v, want %v", tt.word, ok, tt.wantOK)
			}
			if ok && cat != tt.wantCat {
				t.Errorf("MatchWord(%q) = %q, want %q", tt.word, cat, tt.wantCat)
			}
		})
	}
}

func TestTokeniser_MatchArticle(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word     string
		wantType string
		wantOK   bool
	}{
		{"a", "indefinite", true},
		{"an", "indefinite", true},
		{"the", "definite", true},
		{"A", "indefinite", true},
		{"The", "definite", true},
		{"foo", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			artType, ok := tok.MatchArticle(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchArticle(%q) ok=%v, want %v", tt.word, ok, tt.wantOK)
			}
			if ok && artType != tt.wantType {
				t.Errorf("MatchArticle(%q) = %q, want %q", tt.word, artType, tt.wantType)
			}
		})
	}
}

func TestTokeniser_Tokenise(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("Deleted the configuration files")

	if len(tokens) != 4 {
		t.Fatalf("Tokenise() returned %d tokens, want 4", len(tokens))
	}

	// "Deleted" → verb, past tense
	if tokens[0].Type != TokenVerb {
		t.Errorf("tokens[0].Type = %v, want TokenVerb", tokens[0].Type)
	}
	if tokens[0].VerbInfo.Tense != "past" {
		t.Errorf("tokens[0].VerbInfo.Tense = %q, want %q", tokens[0].VerbInfo.Tense, "past")
	}

	// "the" → article
	if tokens[1].Type != TokenArticle {
		t.Errorf("tokens[1].Type = %v, want TokenArticle", tokens[1].Type)
	}

	// "configuration" → unknown
	if tokens[2].Type != TokenUnknown {
		t.Errorf("tokens[2].Type = %v, want TokenUnknown", tokens[2].Type)
	}

	// "files" → noun, plural
	if tokens[3].Type != TokenNoun {
		t.Errorf("tokens[3].Type = %v, want TokenNoun", tokens[3].Type)
	}
	if !tokens[3].NounInfo.Plural {
		t.Errorf("tokens[3].NounInfo.Plural = false, want true")
	}
}

func TestTokeniser_Tokenise_Punctuation(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("Building project...")
	hasPunct := false
	for _, tok := range tokens {
		if tok.Type == TokenPunctuation {
			hasPunct = true
		}
	}
	if !hasPunct {
		t.Error("did not detect punctuation in \"Building project...\"")
	}
}

func TestTokeniser_Tokenise_Empty(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("")
	if len(tokens) != 0 {
		t.Errorf("Tokenise(\"\") returned %d tokens, want 0", len(tokens))
	}
}

func TestTokeniser_MatchVerb_Regular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word      string
		wantOK    bool
		wantBase  string
		wantTense string
	}{
		// Regular verbs NOT in grammar tables — detected by reverse morphology + round-trip
		{"walked", true, "walk", "past"},
		{"walking", true, "walk", "gerund"},
		{"processed", true, "process", "past"},
		{"processing", true, "process", "gerund"},
		{"copied", true, "copy", "past"},
		{"copying", true, "copy", "gerund"},
		{"stopped", true, "stop", "past"},
		{"stopping", true, "stop", "gerund"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchVerb(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchVerb(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchVerb(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Tense != tt.wantTense {
				t.Errorf("MatchVerb(%q).Tense = %q, want %q", tt.word, match.Tense, tt.wantTense)
			}
		})
	}
}

func TestTokeniser_WithSignals(t *testing.T) {
	setup(t)
	tok := NewTokeniser(WithSignals())
	_ = tok // verify it compiles and accepts the option
}

func TestTokeniser_DualClassDetection(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	dualClass := []string{"commit", "run", "test", "check", "file", "build"}
	for _, word := range dualClass {
		if !tok.IsDualClass(word) {
			t.Errorf("%q should be dual-class", word)
		}
	}

	notDual := []string{"delete", "go", "push", "branch", "repo"}
	for _, word := range notDual {
		if tok.IsDualClass(word) {
			t.Errorf("%q should not be dual-class", word)
		}
	}
}

func TestToken_ConfidenceField(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Deleted the branch")

	for _, token := range tokens {
		if token.Type != TokenUnknown && token.Confidence == 0 {
			t.Errorf("token %q (type %d) has zero Confidence", token.Raw, token.Type)
		}
	}
}
