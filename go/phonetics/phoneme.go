package phonetics

import "dappco.re/go"

// Pronunciation is one way of saying a word: an ordered list of ARPABET
// phonemes, vowels carrying a trailing stress digit (0 none, 1 primary,
// 2 secondary).
//
//	Pronunciation{"K", "AH0", "M", "IH1", "T"} // commit
type Pronunciation []string

// arpabetVowels is the full ARPABET vowel inventory (stress digit stripped).
var arpabetVowels = map[string]bool{
	"AA": true, "AE": true, "AH": true, "AO": true, "AW": true,
	"AY": true, "EH": true, "ER": true, "EY": true, "IH": true,
	"IY": true, "OW": true, "OY": true, "UH": true, "UW": true,
}

// basePhone strips the stress digit from a phoneme: "IH1" → "IH".
//
//	basePhone("AY2") // "AY"
func basePhone(phone string) string {
	if phone == "" {
		return phone
	}
	last := phone[len(phone)-1]
	if last >= '0' && last <= '9' {
		return phone[:len(phone)-1]
	}
	return phone
}

// phoneStress reads a vowel phoneme's stress digit, -1 for consonants.
//
//	phoneStress("IH1") // 1
//	phoneStress("K")   // -1
func phoneStress(phone string) int {
	if phone == "" {
		return -1
	}
	last := phone[len(phone)-1]
	if last < '0' || last > '9' {
		return -1
	}
	return int(last - '0')
}

// IsVowelPhone reports whether a phoneme is a vowel.
//
//	phonetics.IsVowelPhone("UW1") // true
//	phonetics.IsVowelPhone("Y")   // false — the glide in "ewe"
func IsVowelPhone(phone string) bool {
	return arpabetVowels[basePhone(phone)]
}

// StartsWithVowelSound reports whether this pronunciation opens on a vowel
// phoneme. Glides (Y, W) are consonants, which is the whole point: "ewe"
// and "one" open on consonants however they are spelled.
//
//	p, _ := phonetics.Primary("hour"); p.StartsWithVowelSound() // true
func (p Pronunciation) StartsWithVowelSound() bool {
	if len(p) == 0 {
		return false
	}
	return IsVowelPhone(p[0])
}

// OnsetKey returns the initial consonant cluster (stress digits stripped),
// empty for vowel-initial words — the unit alliteration compares.
//
//	p, _ := phonetics.Primary("quiz"); p.OnsetKey() // "K W"
func (p Pronunciation) OnsetKey() string {
	onset := make([]string, 0, 3)
	for _, phone := range p {
		if IsVowelPhone(phone) {
			break
		}
		onset = append(onset, basePhone(phone))
	}
	return core.Join(" ", onset...)
}
