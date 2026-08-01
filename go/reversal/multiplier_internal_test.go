package reversal

import (
	"testing"

	i18n "dappco.re/go/i18n"
)

// TestMultiplierInternal_reconstruct_Good rejoins classified tokens with single
// spaces and confirms punctuation tokens stay attached to the preceding word.
//
//	reconstruct([]Token{{Raw: "Delete"}, {Raw: "file"}, {Raw: "?", Type: TokenPunctuation}})
//	// "Delete file?"
func TestMultiplierInternal_reconstruct_Good(t *testing.T) {
	tokens := []Token{
		{Raw: "Delete", Lower: "delete", Type: TokenVerb},
		{Raw: "the", Lower: "the", Type: TokenArticle},
		{Raw: "file", Lower: "file", Type: TokenNoun},
		{Raw: "?", Type: TokenPunctuation, PunctType: "?"},
	}
	got := reconstruct(tokens)
	want := "Delete the file?"
	if got != want {
		t.Errorf("reconstruct() = %q, want %q", got, want)
	}
}

// TestMultiplierInternal_reconstruct_Bad confirms a single-token slice rejoins
// to exactly that token with no leading or trailing space.
func TestMultiplierInternal_reconstruct_Bad(t *testing.T) {
	got := reconstruct([]Token{{Raw: "configuration", Lower: "configuration", Type: TokenNoun}})
	if got != "configuration" {
		t.Errorf("reconstruct(single) = %q, want %q", got, "configuration")
	}
}

// TestMultiplierInternal_reconstruct_Ugly confirms an empty token slice rejoins
// to the empty string without panicking.
func TestMultiplierInternal_reconstruct_Ugly(t *testing.T) {
	if got := reconstruct(nil); got != "" {
		t.Errorf("reconstruct(nil) = %q, want empty", got)
	}
	if got := reconstruct([]Token{}); got != "" {
		t.Errorf("reconstruct(empty) = %q, want empty", got)
	}
}

// TestMultiplierInternal_applyNounTransform_Good confirms the public-shape
// wrapper toggles a singular noun to plural via the default English rules.
func TestMultiplierInternal_applyNounTransform_Good(t *testing.T) {
	svc, _ := valueFromResult[*i18n.Service](i18n.New())
	i18n.SetDefault(svc)
	m := NewMultiplier()

	tokens := []Token{
		{Raw: "file", Lower: "file", Type: TokenNoun, NounInfo: NounMatch{Base: "file", Plural: false}},
	}
	out := m.applyNounTransform(tokens, 0)
	if len(out) != 1 {
		t.Fatalf("applyNounTransform returned %d tokens, want 1", len(out))
	}
	if !out[0].NounInfo.Plural {
		t.Errorf("expected plural toggle, got %+v", out[0].NounInfo)
	}
	if out[0].Raw != "files" {
		t.Errorf("applyNounTransform Raw = %q, want %q", out[0].Raw, "files")
	}
}

// TestMultiplierInternal_applyNounTransform_Bad confirms an already-plural noun
// reverts to its singular base form.
func TestMultiplierInternal_applyNounTransform_Bad(t *testing.T) {
	svc, _ := valueFromResult[*i18n.Service](i18n.New())
	i18n.SetDefault(svc)
	m := NewMultiplier()

	tokens := []Token{
		{Raw: "files", Lower: "files", Type: TokenNoun, NounInfo: NounMatch{Base: "file", Plural: true}},
	}
	out := m.applyNounTransform(tokens, 0)
	if out[0].NounInfo.Plural {
		t.Errorf("expected singular revert, got plural %+v", out[0].NounInfo)
	}
	if out[0].Raw != "file" {
		t.Errorf("applyNounTransform Raw = %q, want %q", out[0].Raw, "file")
	}
}

// TestMultiplierInternal_applyNounTransform_Ugly confirms an empty noun base
// (nothing to inflect) leaves the original token unchanged.
func TestMultiplierInternal_applyNounTransform_Ugly(t *testing.T) {
	svc, _ := valueFromResult[*i18n.Service](i18n.New())
	i18n.SetDefault(svc)
	m := NewMultiplier()

	tokens := []Token{
		{Raw: "xyzzy", Lower: "xyzzy", Type: TokenNoun, NounInfo: NounMatch{Base: "", Plural: false}},
	}
	out := m.applyNounTransform(tokens, 0)
	if out[0].Raw != "xyzzy" {
		t.Errorf("applyNounTransform with empty base mutated Raw to %q", out[0].Raw)
	}
}

// TestMultiplierInternal_isAllUpper_Good exercises the rune-based all-upper test
// directly (the non-ASCII slow path of preserveCase).
func TestMultiplierInternal_isAllUpper_Good(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"DELETE", true},
		{"ÉLÉMENT", true}, // accented uppercase, drives the rune path
		{"Delete", false},
		{"delete", false},
		{"D3LETE", true}, // digits are not letters
		{"", true},       // no letters at all
		{"123!?", true},  // no letters
		{"Élément", false},
	}
	for _, tt := range tests {
		if got := isAllUpper(tt.in); got != tt.want {
			t.Errorf("isAllUpper(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// TestMultiplierInternal_upperASCII_Good confirms ASCII lowercasing is flipped
// while non-letters and already-upper bytes are preserved.
func TestMultiplierInternal_upperASCII_Good(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"delete", "DELETE"},
		{"DELETE", "DELETE"}, // already upper — returned unchanged
		{"Delete", "DELETE"},
		{"file-2", "FILE-2"}, // non-letters preserved
		{"", ""},
		{"123", "123"}, // no letters — returned unchanged
	}
	for _, tt := range tests {
		if got := upperASCII(tt.in); got != tt.want {
			t.Errorf("upperASCII(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestMultiplierInternal_preserveCase_NonASCII drives preserveCase through its
// non-ASCII slow path, which is what reaches isAllUpper. An all-uppercase
// accented original uppercases the replacement; a capitalised accented original
// capitalises only the first rune.
//
//	preserveCase("ÉLÉMENT", "delete") // "DELETE"
//	preserveCase("Élément", "delete") // "Delete"
func TestMultiplierInternal_preserveCase_NonASCII(t *testing.T) {
	if got := preserveCase("ÉLÉMENT", "delete"); got != "DELETE" {
		t.Errorf("preserveCase(all-upper non-ASCII) = %q, want %q", got, "DELETE")
	}
	if got := preserveCase("Élément", "delete"); got != "Delete" {
		t.Errorf("preserveCase(capitalised non-ASCII) = %q, want %q", got, "Delete")
	}
	if got := preserveCase("élément", "Delete"); got != "delete" {
		t.Errorf("preserveCase(lower non-ASCII) = %q, want %q", got, "delete")
	}
}

// TestMultiplierInternal_preserveCase_Ugly confirms the empty-string guards in
// preserveCase return the replacement untouched.
func TestMultiplierInternal_preserveCase_Ugly(t *testing.T) {
	if got := preserveCase("", "delete"); got != "delete" {
		t.Errorf("preserveCase(empty original) = %q, want %q", got, "delete")
	}
	if got := preserveCase("Delete", ""); got != "" {
		t.Errorf("preserveCase(empty replacement) = %q, want empty", got)
	}
}
