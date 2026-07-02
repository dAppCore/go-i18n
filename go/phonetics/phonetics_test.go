package phonetics

import "testing"

// TestPhonetics_Lookup_Good pins headword resolution: case folding,
// variant collection, and the GB override riding first for herb.
//
//	Lookup("Herb") // [[HH ER1 B] [ER1 B] ...], true
func TestPhonetics_Lookup_Good(t *testing.T) {
	prons, ok := Lookup("Quiz")
	if !ok || len(prons) == 0 {
		t.Fatalf("Lookup(Quiz) missing")
	}
	if got := prons[0].OnsetKey(); got != "K W" {
		t.Errorf("quiz onset = %q, want %q", got, "K W")
	}

	herb, ok := Lookup("herb")
	if !ok || len(herb) < 2 {
		t.Fatalf("Lookup(herb) = %v, want GB override + American variants", herb)
	}
	if got := herb[0][0]; got != "HH" {
		t.Errorf("herb primary opens %q, want %q (the English of England sounds its h)", got, "HH")
	}
	if got := herb[1][0]; got != "ER1" {
		t.Errorf("herb second pronunciation opens %q, want the American %q", got, "ER1")
	}

	tomato, ok := Lookup("tomato")
	if !ok || len(tomato) < 2 {
		t.Fatalf("Lookup(tomato) = %v, want both pronunciations (you say...)", tomato)
	}
}

// TestPhonetics_Lookup_Bad pins the miss cases: unknown words, blanks.
func TestPhonetics_Lookup_Bad(t *testing.T) {
	if _, ok := Lookup("zzzxqjw"); ok {
		t.Error("Lookup(zzzxqjw) = ok, want miss")
	}
	if _, ok := Lookup("   "); ok {
		t.Error("Lookup(blank) = ok, want miss")
	}
	if _, ok := Primary(""); ok {
		t.Error("Primary(empty) = ok, want miss")
	}
}

// TestPhonetics_Size_Good pins that the full dictionary actually loaded.
func TestPhonetics_Size_Good(t *testing.T) {
	if n := Size(); n < 120000 {
		t.Errorf("Size() = %d, want the full dictionary (>120k headwords)", n)
	}
}

// TestPhonetics_parseDictLine_Ugly drives comment lines, variant markers
// and malformed input straight through the parser.
func TestPhonetics_parseDictLine_Ugly(t *testing.T) {
	if _, _, ok := parseDictLine(";;; comment"); ok {
		t.Error("comment line parsed as entry")
	}
	if _, _, ok := parseDictLine("# comment"); ok {
		t.Error("hash comment parsed as entry")
	}
	if _, _, ok := parseDictLine("lonely"); ok {
		t.Error("phoneme-less line parsed as entry")
	}
	word, pron, ok := parseDictLine("require(2) R IY0 K W AY1 R")
	if !ok || word != "require" || len(pron) != 6 {
		t.Errorf("variant line = %q %v %v, want require with 6 phones", word, pron, ok)
	}
	word, pron, ok = parseDictLine("word W ER1 D # some note")
	if !ok || word != "word" || len(pron) != 3 {
		t.Errorf("commented line = %q %v %v, want word with 3 phones", word, pron, ok)
	}
}
