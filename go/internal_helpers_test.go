package i18n

import "testing"

// TestInternalHelpers_failedPrefix_Good confirms the single-language failedPrefix
// convenience returns the default English prefix when no grammar override is
// loaded for the language.
//
//	failedPrefix("en") // "Failed to"
func TestInternalHelpers_failedPrefix_Good(t *testing.T) {
	if got := failedPrefix("en"); got != "Failed to" {
		t.Errorf("failedPrefix(en) = %q, want %q", got, "Failed to")
	}
}

// TestInternalHelpers_failedPrefix_Bad confirms an unknown language still falls
// back to the English default rather than returning empty.
func TestInternalHelpers_failedPrefix_Bad(t *testing.T) {
	if got := failedPrefix("zz"); got != "Failed to" {
		t.Errorf("failedPrefix(zz) = %q, want %q", got, "Failed to")
	}
}

// TestInternalHelpers_failedPrefix_Ugly confirms the empty-language case is
// handled without panic and still yields the default prefix.
func TestInternalHelpers_failedPrefix_Ugly(t *testing.T) {
	if got := failedPrefix(""); got != "Failed to" {
		t.Errorf("failedPrefix(empty) = %q, want %q", got, "Failed to")
	}
}

// TestInternalHelpers_firstNonEmptyString_Good returns the first trimmed
// non-empty value, skipping leading blanks and whitespace-only entries.
//
//	firstNonEmptyString("", "  ", "a", "b") // "a"
func TestInternalHelpers_firstNonEmptyString_Good(t *testing.T) {
	if got := firstNonEmptyString("", "  ", "a", "b"); got != "a" {
		t.Errorf("firstNonEmptyString = %q, want %q", got, "a")
	}
	if got := firstNonEmptyString("  first  ", "second"); got != "first" {
		t.Errorf("firstNonEmptyString trims = %q, want %q", got, "first")
	}
}

// TestInternalHelpers_firstNonEmptyString_Bad confirms all-empty and
// whitespace-only inputs yield the empty string.
func TestInternalHelpers_firstNonEmptyString_Bad(t *testing.T) {
	if got := firstNonEmptyString("", "   ", "\t"); got != "" {
		t.Errorf("firstNonEmptyString(all blank) = %q, want empty", got)
	}
}

// TestInternalHelpers_firstNonEmptyString_Ugly confirms a zero-argument call
// returns empty without panic.
func TestInternalHelpers_firstNonEmptyString_Ugly(t *testing.T) {
	if got := firstNonEmptyString(); got != "" {
		t.Errorf("firstNonEmptyString() = %q, want empty", got)
	}
}

// TestInternalHelpers_isAllUpper_Good confirms the handler-side all-upper check
// recognises acronyms and rejects mixed or lowercase words.
func TestInternalHelpers_isAllUpper_Good(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"API", true},
		{"URL", true},
		{"Api", false},
		{"api", false},
		{"A1B2", true},  // digits ignored, all letters upper
		{"", false},     // no letters → false (differs from multiplier's isAllUpper)
		{"123", false},  // no letters → false
		{"FILE-PATH", true},
	}
	for _, tt := range tests {
		if got := isAllUpper(tt.in); got != tt.want {
			t.Errorf("isAllUpper(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
