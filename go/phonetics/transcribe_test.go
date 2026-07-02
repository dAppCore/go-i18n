package phonetics

import "testing"

// catSentence is the house demonstration sentence (Snider, 2026-07-02).
const catSentence = "The cat was in the garden, when a cloud came by and rained, " +
	"the cat rushed inside but bonked his head on the cat flap trying to avoid water"

// TestPhonetics_TranscribeWord_Good pins single-word rendering in all
// three styles.
//
//	TranscribeWord("water", StyleRespell) // "wortuh"
func TestPhonetics_TranscribeWord_Good(t *testing.T) {
	tests := []struct {
		word  string
		style Style
		want  string
	}{
		{"cat", StyleARPABET, "K AE1 T"},
		{"cat", StyleIPA, "kˈæt"},
		{"cat", StyleRespell, "kat"},
		{"water", StyleRespell, "wortuh"},
		{"garden", StyleRespell, "gahrduhn"}, // CMU is rhotic; the r is real
		{"cloud", StyleRespell, "klowd"},
		{"queue", StyleRespell, "kyoo"},
	}
	for _, tt := range tests {
		got, ok := TranscribeWord(tt.word, tt.style)
		if !ok {
			t.Errorf("TranscribeWord(%q) unknown", tt.word)
			continue
		}
		if got != tt.want {
			t.Errorf("TranscribeWord(%q, %v) = %q, want %q", tt.word, tt.style, got, tt.want)
		}
	}
}

// TestPhonetics_TranscribeWord_Bad pins the pass-through contract for
// words the dictionary does not know.
func TestPhonetics_TranscribeWord_Bad(t *testing.T) {
	got, ok := TranscribeWord("zzzxqjw", StyleRespell)
	if ok || got != "zzzxqjw" {
		t.Errorf("unknown word = %q/%v, want pass-through/false", got, ok)
	}
}

// TestPhonetics_Transcribe_Good drives the house sentence through the
// transcriber and logs all three renderings for the reader.
func TestPhonetics_Transcribe_Good(t *testing.T) {
	respelled := Transcribe(catSentence, StyleRespell)
	ipa := Transcribe(catSentence, StyleIPA)
	arpabet := Transcribe(catSentence, StyleARPABET)
	t.Logf("respell: %s", respelled)
	t.Logf("ipa:     %s", ipa)
	t.Logf("arpabet: %s", arpabet)
	if respelled == "" || ipa == "" || arpabet == "" {
		t.Fatal("transcriptions came back empty")
	}
	// The sentence must not leak raw ARPABET into the respelled register.
	if core_containsUpper(respelled) {
		t.Errorf("respell contains uppercase ARPABET leakage: %q", respelled)
	}
}

func core_containsUpper(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

// TestPhonetics_ScanLine_Good pins the metrical reading of plain prose and
// a deliberately iambic line.
//
//	ScanLine("the cat avoids the dog").Pattern // "010101"
func TestPhonetics_ScanLine_Good(t *testing.T) {
	s := ScanLine("the cat avoids the dog")
	if s.Pattern != "010101" {
		t.Errorf("Pattern = %q, want %q", s.Pattern, "010101")
	}
	if s.Feet != "iamb" || s.Regularity != 1.0 {
		t.Errorf("Feet/Regularity = %q/%v, want perfect iamb", s.Feet, s.Regularity)
	}

	whole := ScanLine(catSentence)
	t.Logf("cat sentence: pattern=%s feet=%s regularity=%.2f unknown=%v",
		whole.Pattern, whole.Feet, whole.Regularity, whole.Unknown)
	if len(whole.Pattern) < 20 {
		t.Errorf("cat sentence pattern too short: %q", whole.Pattern)
	}
}

// TestPhonetics_ScanLine_Bad pins empty and unknown-only input.
func TestPhonetics_ScanLine_Bad(t *testing.T) {
	empty := ScanLine("")
	if empty.Pattern != "" || empty.Feet != "" {
		t.Errorf("empty line scanned to %+v", empty)
	}
	unknown := ScanLine("zzzxqjw qqqjxz")
	if unknown.Pattern != "" || len(unknown.Unknown) != 2 {
		t.Errorf("unknown-only line = %+v, want empty pattern + 2 unknowns", unknown)
	}
}

// TestPhonetics_dominantFoot_Ugly drives the tie-break and short patterns.
func TestPhonetics_dominantFoot_Ugly(t *testing.T) {
	foot, score := dominantFoot("10101")
	if foot != "trochee" {
		t.Errorf("10101 = %q (%.2f), want trochee", foot, score)
	}
	foot, score = dominantFoot("11111")
	if foot != "spondee" || score != 1.0 {
		t.Errorf("11111 = %q (%.2f), want perfect spondee", foot, score)
	}
	if foot, _ := dominantFoot("001001"); foot != "anapaest" {
		t.Errorf("001001 = %q, want anapaest", foot)
	}
}
