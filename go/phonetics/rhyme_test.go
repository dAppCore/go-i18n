package phonetics

import "testing"

// TestPhonetics_Rhymes_Good pins perfect rhymes — including the pairs where
// spelling and sound disagree in both directions.
//
//	Rhymes("commit", "submit") // true
//	Rhymes("cough", "bough")   // false — eye-rhyme only
func TestPhonetics_Rhymes_Good(t *testing.T) {
	rhymingPairs := [][2]string{
		{"commit", "submit"},
		{"deploy", "enjoy"},
		{"code", "load"},   // different spellings, same sound
		{"queue", "few"},   // spelling could not be less helpful
		{"cache", "stash"}, // the pun every engineer has made
	}
	for _, pair := range rhymingPairs {
		if !Rhymes(pair[0], pair[1]) {
			t.Errorf("Rhymes(%q, %q) = false, want true", pair[0], pair[1])
		}
	}
	nonRhymes := [][2]string{
		{"cough", "bough"},  // eye-rhyme: spelling matches, sound does not
		{"commit", "visit"}, // stress lands differently
		{"file", "fill"},
	}
	for _, pair := range nonRhymes {
		if Rhymes(pair[0], pair[1]) {
			t.Errorf("Rhymes(%q, %q) = true, want false", pair[0], pair[1])
		}
	}
}

// TestPhonetics_Rhymes_Bad pins the refusals: identity, unknowns, blanks.
func TestPhonetics_Rhymes_Bad(t *testing.T) {
	if Rhymes("code", "code") {
		t.Error("a word does not rhyme with itself")
	}
	if Rhymes("code", "zzzxqjw") {
		t.Error("unknown words cannot rhyme")
	}
	if Rhymes("", "code") {
		t.Error("blank cannot rhyme")
	}
}

// TestPhonetics_Alliterate_Good pins onset-cluster matching: whole cluster,
// not first letter.
//
//	Alliterate("quiz", "quick") // true — both /kw/
func TestPhonetics_Alliterate_Good(t *testing.T) {
	if !Alliterate("quiz", "quick") {
		t.Error("Alliterate(quiz, quick) = false, want true (both open /kw/)")
	}
	if !Alliterate("ship", "shape") {
		t.Error("Alliterate(ship, shape) = false, want true (both open /ʃ/)")
	}
	if Alliterate("cough", "quiz") {
		t.Error("Alliterate(cough, quiz) = true, want false (K vs K W)")
	}
	if Alliterate("apple", "error") {
		t.Error("Alliterate(apple, error) = true, want false (vowel onsets have no cluster)")
	}
}

// TestPhonetics_RhymeKey_Ugly drives the no-primary-stress fallback.
func TestPhonetics_RhymeKey_Ugly(t *testing.T) {
	// Synthetic pronunciation with no stress digits at all: fall back to
	// the last vowel.
	p := Pronunciation{"K", "AH", "T"}
	if got := p.RhymeKey(); got != "AH T" {
		t.Errorf("RhymeKey(no stress) = %q, want %q", got, "AH T")
	}
	empty := Pronunciation{"K", "T"}
	if got := empty.RhymeKey(); got != "" {
		t.Errorf("RhymeKey(no vowel) = %q, want empty", got)
	}
}
