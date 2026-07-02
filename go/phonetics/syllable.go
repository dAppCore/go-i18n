package phonetics

// SyllableCount counts the vowel phonemes — in ARPABET every syllable has
// exactly one.
//
//	p, _ := phonetics.Primary("unicorn"); p.SyllableCount() // 3
func (p Pronunciation) SyllableCount() int {
	count := 0
	for _, phone := range p {
		if IsVowelPhone(phone) {
			count++
		}
	}
	return count
}

// StressPattern renders the stress digits of the vowels in order: commit
// is "01", visit is "10", university is "20100". This is the raw material
// for metre scanning.
//
//	p, _ := phonetics.Primary("delete"); p.StressPattern() // "01"
func (p Pronunciation) StressPattern() string {
	pattern := make([]byte, 0, 6)
	for _, phone := range p {
		stress := phoneStress(phone)
		if stress < 0 {
			continue
		}
		pattern = append(pattern, byte('0'+stress))
	}
	return string(pattern)
}

// FinalSyllableStressed reports whether the last syllable carries primary
// or secondary stress — the true condition behind English consonant
// doubling: commit → committed (IH1), visit → visited (IH0).
//
//	p, _ := phonetics.Primary("refer"); p.FinalSyllableStressed() // true
func (p Pronunciation) FinalSyllableStressed() bool {
	for i := len(p) - 1; i >= 0; i-- {
		stress := phoneStress(p[i])
		if stress < 0 {
			continue
		}
		return stress > 0
	}
	return false
}

// SyllableCount answers for a word via its primary pronunciation.
//
//	n, ok := phonetics.SyllableCount("vulnerability") // 6, true
func SyllableCount(word string) (int, bool) {
	pron, ok := Primary(word)
	if !ok {
		return 0, false
	}
	return pron.SyllableCount(), true
}

// FinalSyllableStressed answers for a word via its primary pronunciation.
// The second return reports whether the dictionary knows the word at all —
// callers fall back to their own heuristics when it does not.
//
//	stressed, known := phonetics.FinalSyllableStressed("visit") // false, true
func FinalSyllableStressed(word string) (bool, bool) {
	pron, ok := Primary(word)
	if !ok {
		return false, false
	}
	return pron.FinalSyllableStressed(), true
}

// StartsWithVowelSound answers for a word via its primary pronunciation.
//
//	vowel, known := phonetics.StartsWithVowelSound("x-ray") // true, true
func StartsWithVowelSound(word string) (bool, bool) {
	pron, ok := Primary(word)
	if !ok {
		return false, false
	}
	return pron.StartsWithVowelSound(), true
}
