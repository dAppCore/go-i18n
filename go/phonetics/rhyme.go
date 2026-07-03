package phonetics

import "dappco.re/go"

// RhymeKey returns the phoneme tail from the last PRIMARY-stressed vowel to
// the end, stress digits stripped — the classical definition of a perfect
// rhyme. Words rhyme when their keys match and their onsets differ (cat/cat
// is identity, not rhyme).
//
//	p, _ := phonetics.Primary("deploy"); p.RhymeKey() // "OY"
//	q, _ := phonetics.Primary("enjoy");  q.RhymeKey() // "OY"
func (p Pronunciation) RhymeKey() string {
	start := -1
	for i, phone := range p {
		if phoneStress(phone) == 1 {
			start = i
		}
	}
	if start < 0 {
		// No primary stress recorded — fall back to the last vowel.
		for i, phone := range p {
			if IsVowelPhone(phone) {
				start = i
			}
		}
	}
	if start < 0 {
		return ""
	}
	tail := make([]string, 0, len(p)-start)
	for _, phone := range p[start:] {
		tail = append(tail, basePhone(phone))
	}
	return core.Join(" ", tail...)
}

// Rhymes reports whether two words form a perfect rhyme on ANY of their
// pronunciation pairs: matching rhyme keys, and not the identical word.
//
//	phonetics.Rhymes("commit", "submit") // true
//	phonetics.Rhymes("cough", "bough")   // false — spelling lies, sound decides
func Rhymes(a, b string) bool {
	aLower := core.Lower(core.Trim(a))
	bLower := core.Lower(core.Trim(b))
	if aLower == "" || bLower == "" || aLower == bLower {
		return false
	}
	aProns, ok := Lookup(aLower)
	if !ok {
		return false
	}
	bProns, ok := Lookup(bLower)
	if !ok {
		return false
	}
	for _, ap := range aProns {
		aKey := ap.RhymeKey()
		if aKey == "" {
			continue
		}
		for _, bp := range bProns {
			if aKey == bp.RhymeKey() {
				return true
			}
		}
	}
	return false
}

// Alliterate reports whether two words share a non-empty initial consonant
// cluster — quiz and quick both open /kw/.
//
//	phonetics.Alliterate("quiz", "quick") // true
//	phonetics.Alliterate("cough", "quiz") // false — K vs K W
func Alliterate(a, b string) bool {
	ap, ok := Primary(a)
	if !ok {
		return false
	}
	bp, ok := Primary(b)
	if !ok {
		return false
	}
	aKey := ap.OnsetKey()
	return aKey != "" && aKey == bp.OnsetKey()
}
