package phonetics

import "testing"

// TestPhonetics_SyllableCount_Good pins syllable counting across sizes.
//
//	SyllableCount("vulnerability") // 6
func TestPhonetics_SyllableCount_Good(t *testing.T) {
	tests := []struct {
		word string
		want int
	}{
		{"quiz", 1},
		{"commit", 2},
		{"unicorn", 3},
		{"repository", 5},
		{"vulnerability", 6},
	}
	for _, tt := range tests {
		got, ok := SyllableCount(tt.word)
		if !ok {
			t.Errorf("SyllableCount(%q) unknown", tt.word)
			continue
		}
		if got != tt.want {
			t.Errorf("SyllableCount(%q) = %d, want %d", tt.word, got, tt.want)
		}
	}
}

// TestPhonetics_StressPattern_Good pins the raw stress strings that metre
// scanning will consume.
func TestPhonetics_StressPattern_Good(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"commit", "01"},
		{"visit", "10"},
		{"delete", "01"},
		{"unicorn", "102"},
	}
	for _, tt := range tests {
		pron, ok := Primary(tt.word)
		if !ok {
			t.Fatalf("Primary(%q) unknown", tt.word)
		}
		if got := pron.StressPattern(); got != tt.want {
			t.Errorf("StressPattern(%q) = %q, want %q", tt.word, got, tt.want)
		}
	}
}

// TestPhonetics_FinalSyllableStressed_Good pins the doubling condition on
// the classic minimal pairs the length heuristic cannot tell apart.
//
//	FinalSyllableStressed("commit") // true  → committed
//	FinalSyllableStressed("visit")  // false → visited
func TestPhonetics_FinalSyllableStressed_Good(t *testing.T) {
	stressedWords := []string{"commit", "refer", "occur", "equip", "regret", "begin", "admit"}
	for _, word := range stressedWords {
		stressed, known := FinalSyllableStressed(word)
		if !known || !stressed {
			t.Errorf("FinalSyllableStressed(%q) = %v/%v, want stressed+known", word, stressed, known)
		}
	}
	unstressedWords := []string{"visit", "edit", "orbit", "target", "budget", "open", "enter", "suffer"}
	for _, word := range unstressedWords {
		stressed, known := FinalSyllableStressed(word)
		if !known || stressed {
			t.Errorf("FinalSyllableStressed(%q) = %v/%v, want unstressed+known", word, stressed, known)
		}
	}
}

// TestPhonetics_FinalSyllableStressed_Bad pins the unknown-word contract:
// callers must be told the dictionary has no opinion.
func TestPhonetics_FinalSyllableStressed_Bad(t *testing.T) {
	if _, known := FinalSyllableStressed("zzzxqjw"); known {
		t.Error("FinalSyllableStressed(nonsense) claims knowledge")
	}
}

// TestPhonetics_StartsWithVowelSound_Good pins the article decision on the
// exact words the spelling heuristics used to get wrong.
//
//	StartsWithVowelSound("ewe")   // false — /juː/, "a ewe"
//	StartsWithVowelSound("hour")  // true  — silent h, "an hour"
func TestPhonetics_StartsWithVowelSound_Good(t *testing.T) {
	consonantOnsets := []string{"ewe", "unicorn", "user", "europe", "one", "once", "utopia", "ukulele", "herb"}
	for _, word := range consonantOnsets {
		vowel, known := StartsWithVowelSound(word)
		if !known {
			t.Errorf("StartsWithVowelSound(%q) unknown", word)
			continue
		}
		if vowel {
			t.Errorf("StartsWithVowelSound(%q) = vowel, want consonant onset", word)
		}
	}
	vowelOnsets := []string{"hour", "honest", "heir", "x-ray", "umbrella", "error", "apple"}
	for _, word := range vowelOnsets {
		vowel, known := StartsWithVowelSound(word)
		if !known {
			t.Errorf("StartsWithVowelSound(%q) unknown", word)
			continue
		}
		if !vowel {
			t.Errorf("StartsWithVowelSound(%q) = consonant, want vowel onset", word)
		}
	}
}
