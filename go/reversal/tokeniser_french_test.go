package reversal

import "testing"

// TestTokeniserFrench_matchPunctuation covers every recognised punctuation
// pattern plus the unrecognised default.
//
//	matchPunctuation("...") // "progress", true
//	matchPunctuation("@")   // "", false
func TestTokeniserFrench_matchPunctuation(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"...", "progress", true},
		{"?", "question", true},
		{"!", "exclamation", true},
		{":", "label", true},
		{";", "separator", true},
		{",", "comma", true},
		{".", "sentence_end", true},
		{")", "close_paren", true},
		{"]", "close_bracket", true},
		{"}", "close_brace", true},
		{"@", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := matchPunctuation(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("matchPunctuation(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

// TestTokeniserFrench_matchFrenchArticleText covers the prefix forms (de l',
// partitive du/des, au/aux), the leading-field articles, the de + second-field
// branches and the elided-pronoun set.
func TestTokeniserFrench_matchFrenchArticleText(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"de l'ami", "indefinite", true},
		{"de la branche", "indefinite", true},
		{"du fichier", "indefinite", true},
		{"des amis", "indefinite", true},
		{"au bureau", "definite", true},
		{"aux bureaux", "definite", true},
		{"les fichiers", "definite", true},
		{"un fichier", "indefinite", true},
		{"une branche", "indefinite", true},
		{"de la", "indefinite", true}, // de + second field "la"
		{"de du", "definite", true},   // de + second field "du"
		{"d'accord", "indefinite", true},
		{"j'aime", "definite", true},
		{"qu'il", "definite", true},
		{"bonjour", "", false}, // no article
	}
	for _, tt := range tests {
		got, ok := matchFrenchArticleText(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("matchFrenchArticleText(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

// TestTokeniserFrench_matchFrenchAttachedArticle covers the d/l/other elision
// prefixes, the missing-apostrophe rejection and the empty-rest rejection.
func TestTokeniserFrench_matchFrenchAttachedArticle(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"d'accord", "indefinite", true}, // d prefix → indefinite
		{"l'ami", "definite", true},      // l prefix → definite
		{"j'aime", "definite", true},     // other prefix → definite
		{"qu'il", "definite", true},      // multi-char prefix
		{"daccord", "", false},           // no apostrophe after prefix
		{"bonjour", "", false},           // no matching prefix
		{"l", "", false},                 // prefix with empty rest
	}
	for _, tt := range tests {
		got, ok := matchFrenchAttachedArticle(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("matchFrenchAttachedArticle(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

// TestTokeniserFrench_normaliseFrenchApostrophes confirms the typographic and
// modifier-letter apostrophes are folded to the straight ASCII form, and the
// no-apostrophe fast path returns the input unchanged.
func TestTokeniserFrench_normaliseFrenchApostrophes(t *testing.T) {
	if got := normaliseFrenchApostrophes("l’ami"); got != "l'ami" {
		t.Errorf("normaliseFrenchApostrophes(typographic) = %q, want %q", got, "l'ami")
	}
	if got := normaliseFrenchApostrophes("dʼaccord"); got != "d'accord" {
		t.Errorf("normaliseFrenchApostrophes(modifier) = %q, want %q", got, "d'accord")
	}
	if got := normaliseFrenchApostrophes("bonjour"); got != "bonjour" {
		t.Errorf("normaliseFrenchApostrophes(none) = %q, want unchanged", got)
	}
}
