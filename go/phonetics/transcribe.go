package phonetics

import "dappco.re/go"

// Style selects the transcription alphabet.
type Style int

const (
	// StyleARPABET renders raw dictionary phonemes: "DH AH0 K AE1 T".
	StyleARPABET Style = iota
	// StyleIPA renders the International Phonetic Alphabet: "ðə kæt".
	StyleIPA
	// StyleRespell renders readable phonetic English — the register a
	// narrator like Swindells' Daz writes in: "dhuh kat".
	StyleRespell
)

// ipaPhones maps ARPABET (stress-stripped) to IPA. Vowel values follow the
// house dialect — the English of England (OW → əʊ, not the American oʊ).
var ipaPhones = map[string]string{
	"AA": "ɑː", "AE": "æ", "AO": "ɔː", "AW": "aʊ", "AY": "aɪ",
	"EH": "ɛ", "ER": "ɜː", "EY": "eɪ", "IH": "ɪ", "IY": "iː",
	"OW": "əʊ", "OY": "ɔɪ", "UH": "ʊ", "UW": "uː",
	"B": "b", "CH": "tʃ", "D": "d", "DH": "ð", "F": "f", "G": "ɡ",
	"HH": "h", "JH": "dʒ", "K": "k", "L": "l", "M": "m", "N": "n",
	"NG": "ŋ", "P": "p", "R": "r", "S": "s", "SH": "ʃ", "T": "t",
	"TH": "θ", "V": "v", "W": "w", "Y": "j", "Z": "z", "ZH": "ʒ",
}

// respellPhones maps ARPABET to readable English graphemes. Chosen so a
// whole sentence stays legible: water → "wortuh", garden → "gahduhn".
var respellPhones = map[string]string{
	"AA": "ah", "AE": "a", "AO": "or", "AW": "ow", "AY": "y",
	"EH": "e", "ER": "uh", "EY": "ay", "IH": "i", "IY": "ee",
	"OW": "oh", "OY": "oy", "UH": "oo", "UW": "oo",
	"B": "b", "CH": "ch", "D": "d", "DH": "dh", "F": "f", "G": "g",
	"HH": "h", "JH": "j", "K": "k", "L": "l", "M": "m", "N": "n",
	"NG": "ng", "P": "p", "R": "r", "S": "s", "SH": "sh", "T": "t",
	"TH": "th", "V": "v", "W": "w", "Y": "y", "Z": "z", "ZH": "zh",
}

// TranscribeWord renders one word in the requested style. Unknown words
// come back unchanged with ok=false so callers can pass them through.
//
//	phonetics.TranscribeWord("water", phonetics.StyleRespell) // "wortuh", true
//	phonetics.TranscribeWord("cat", phonetics.StyleIPA)       // "kæt", true
func TranscribeWord(word string, style Style) (string, bool) {
	pron, ok := Primary(word)
	if !ok {
		return word, false
	}
	switch style {
	case StyleIPA:
		return pron.IPA(), true
	case StyleRespell:
		return pron.Respell(), true
	default:
		return core.Join(" ", []string(pron)...), true
	}
}

// Transcribe renders a whole sentence, preserving word order and passing
// unknown words through untouched. Punctuation between words is dropped;
// words are joined with single spaces.
//
//	phonetics.Transcribe("The cat was in the garden", phonetics.StyleRespell)
//	// "dhuh kat wahz in dhuh gahduhn"
func Transcribe(text string, style Style) string {
	words := splitWords(text)
	out := make([]string, 0, len(words))
	for _, word := range words {
		rendered, _ := TranscribeWord(word, style)
		out = append(out, rendered)
	}
	return core.Join(" ", out...)
}

// IPA renders the pronunciation in the International Phonetic Alphabet,
// with the primary-stress mark placed before the stressed vowel (an
// approximation of syllable onset placement).
//
//	p, _ := phonetics.Primary("garden"); p.IPA() // "ˈɡɑːrdən"
func (p Pronunciation) IPA() string {
	b := core.NewBuilder()
	for _, phone := range p {
		stress := phoneStress(phone)
		base := basePhone(phone)
		if stress == 1 {
			b.WriteString("ˈ")
		}
		if stress == 2 {
			b.WriteString("ˌ")
		}
		// Unstressed AH is the schwa — the most common sound in English
		// and the reason phonetic spelling looks nothing like writing.
		if base == "AH" {
			if stress == 0 {
				b.WriteString("ə")
			} else {
				b.WriteString("ʌ")
			}
			continue
		}
		if ipa, ok := ipaPhones[base]; ok {
			b.WriteString(ipa)
			continue
		}
		b.WriteString(core.Lower(base))
	}
	return b.String()
}

// Respell renders the pronunciation as readable phonetic English.
//
//	p, _ := phonetics.Primary("cloud"); p.Respell() // "klowd"
func (p Pronunciation) Respell() string {
	b := core.NewBuilder()
	for _, phone := range p {
		stress := phoneStress(phone)
		base := basePhone(phone)
		if base == "AH" {
			if stress == 0 {
				b.WriteString("uh")
			} else {
				b.WriteString("u")
			}
			continue
		}
		if spelled, ok := respellPhones[base]; ok {
			b.WriteString(spelled)
			continue
		}
		b.WriteString(core.Lower(base))
	}
	return b.String()
}
