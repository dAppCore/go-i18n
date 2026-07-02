package i18n

import "testing"

// TestGrammarRegular_applyRegularPastTense_Good drives every morphological
// branch of the regular past-tense rule directly, bypassing the JSON/irregular
// tiers that the public PastTense consults first.
//
//	applyRegularPastTense("panic") // "panicked"
//	applyRegularPastTense("stop")  // "stopped"
func TestGrammarRegular_applyRegularPastTense_Good(t *testing.T) {
	tests := []struct {
		verb string
		want string
	}{
		{"walk", "walked"},     // default suffix
		{"delete", "deleted"},  // ends in e → +d
		{"carry", "carried"},   // consonant + y → ied
		{"play", "played"},     // vowel + y → +ed (no ied)
		{"panic", "panicked"},  // ends in c → +ked
		{"stop", "stopped"},    // CVC short word → double consonant
		{"deleted", "deleted"}, // already past, consonant before -ed → unchanged
	}
	for _, tt := range tests {
		if got := applyRegularPastTense(tt.verb); got != tt.want {
			t.Errorf("applyRegularPastTense(%q) = %q, want %q", tt.verb, got, tt.want)
		}
	}
}

// TestGrammarRegular_applyRegularPastTense_Ugly drives the already-`ed` guard
// where the third-from-end character is a vowel (so the word is NOT treated as
// already past) and the very short word path.
func TestGrammarRegular_applyRegularPastTense_Ugly(t *testing.T) {
	// "freed" ends in "ed" but the third-from-end char is the vowel 'e', so the
	// already-past guard does NOT short-circuit and the default "+ed" rule fires.
	if got := applyRegularPastTense("freed"); got != "freeded" {
		t.Errorf("applyRegularPastTense(freed) = %q, want %q", got, "freeded")
	}
	if got := applyRegularPastTense("a"); got != "aed" {
		t.Errorf("applyRegularPastTense(a) = %q, want %q", got, "aed")
	}
}

// TestGrammarRegular_applyRegularGerund_Good covers the -ie, -e, -c and
// consonant-doubling gerund branches.
//
//	applyRegularGerund("die")   // "dying"
//	applyRegularGerund("panic") // "panicking"
func TestGrammarRegular_applyRegularGerund_Good(t *testing.T) {
	tests := []struct {
		verb string
		want string
	}{
		{"die", "dying"},       // -ie → ying
		{"delete", "deleting"}, // -e (consonant before) → drop e + ing
		{"see", "seeing"},      // -ee → keep e
		{"panic", "panicking"}, // -c → +king
		{"stop", "stopping"},   // CVC → double
		{"walk", "walking"},    // default
		{"echo", "echoing"},    // ends in vowel+o, no special rule
	}
	for _, tt := range tests {
		if got := applyRegularGerund(tt.verb); got != tt.want {
			t.Errorf("applyRegularGerund(%q) = %q, want %q", tt.verb, got, tt.want)
		}
	}
}

// TestGrammarRegular_applyRegularPlural_Good covers the sibilant, consonant-y
// and -o special-case plural branches. The f/fe → ves plural is deliberately
// ABSENT here: it is a closed Old English class owned by irregularNouns
// (PluralForm resolves leaf → leaves at that tier), and the productive rule
// takes -s — which is why the hockey team is the Maple Leafs.
//
//	applyRegularPlural("leaf")  // "leafs" (PluralForm("leaf") → "leaves")
//	applyRegularPlural("hero")  // "heroes"
func TestGrammarRegular_applyRegularPlural_Good(t *testing.T) {
	tests := []struct {
		noun string
		want string
	}{
		{"box", "boxes"},       // sibilant x → es
		{"bus", "buses"},       // s → es
		{"dish", "dishes"},     // sh → es
		{"church", "churches"}, // ch → es
		{"buzz", "buzzes"},     // zz → es (already doubled)
		{"quiz", "quizzes"},    // single z after vowel → double + es
		{"whiz", "whizzes"},    // single z after vowel → double + es
		{"waltz", "waltzes"},   // z after consonant → plain es
		{"city", "cities"},     // consonant + y → ies
		{"key", "keys"},        // vowel + y → +s
		{"leaf", "leafs"},      // f is REGULAR +s; leaves lives in irregularNouns
		{"knife", "knifes"},    // fe is REGULAR +s; knives lives in irregularNouns
		{"roof", "roofs"},      // modern f-final takes -s, never "rooves"
		{"chief", "chiefs"},    // not "chieves"
		{"safe", "safes"},      // not "saves"
		{"cliff", "cliffs"},    // ff-final takes -s
		{"hero", "heroes"},     // -o special case
		{"potato", "potatoes"}, // -o special case
		{"piano", "pianos"},    // -o NOT special → +s
		{"server", "servers"},  // default → +s
	}
	for _, tt := range tests {
		if got := applyRegularPlural(tt.noun); got != tt.want {
			t.Errorf("applyRegularPlural(%q) = %q, want %q", tt.noun, got, tt.want)
		}
	}
}

// TestGrammarRegular_shouldDoubleConsonant covers the CVC / non-CVC decision
// branches, including the explicit no-double set and the long-word path.
func TestGrammarRegular_shouldDoubleConsonant(t *testing.T) {
	tests := []struct {
		verb string
		want bool
	}{
		{"stop", true},      // CVC short word
		{"run", true},       // CVC short word
		{"quiz", true},      // qu is a /kw/ onset, so quiz IS CVC → quizzed
		{"equal", true},     // same qu logic feeds the -l rule → equalled
		{"go", false},       // too short (< 3)
		{"play", false},     // ends in y
		{"fix", false},      // ends in x
		{"flow", false},     // ends in w
		{"hello", false},    // ends in vowel
		{"commit", false},   // long-word path (the table carries committed)
		{"visit", false},    // long word, no double
		{"marshal", true},   // en-GB -l doubling → marshalled
		{"signal", true},    // en-GB -l doubling → signalled
		{"total", true},     // en-GB -l doubling → totalled
		{"parallel", false}, // in noDoubleConsonant: both dialects prefer paralleled
		{"reveal", false},   // vowel digraph before -l → revealed
		{"email", false},    // vowel digraph before -l → emailed
		{"curl", false},     // consonant before -l → curled
	}
	for _, tt := range tests {
		if got := shouldDoubleConsonant(tt.verb); got != tt.want {
			t.Errorf("shouldDoubleConsonant(%q) = %v, want %v", tt.verb, got, tt.want)
		}
	}
}
