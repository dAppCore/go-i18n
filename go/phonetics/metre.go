package phonetics

import "dappco.re/go"

// Scansion is the metrical reading of a line: the stress pattern across all
// syllables, the best-fitting classical foot, and how regular the fit is.
//
//	s := phonetics.ScanLine("The cat was in the garden")
//	s.Feet       // "iamb"
//	s.Regularity // 0.0–1.0
type Scansion struct {
	Pattern    string   // one digit per syllable: 0 weak, 1 strong
	Words      []string // dictionary-known words, in order
	Unknown    []string // words the dictionary missed (skipped from Pattern)
	Feet       string   // dominant foot: iamb, trochee, anapaest, dactyl, spondee
	Regularity float64  // fraction of syllables matching the dominant foot
}

// metricalFeet maps the classical feet to weak/strong templates.
var metricalFeet = map[string]string{
	"iamb":     "01",
	"trochee":  "10",
	"anapaest": "001",
	"dactyl":   "100",
	"spondee":  "11",
}

// functionWords are the closed-class monosyllables prosody demotes to weak
// positions regardless of their dictionary stress mark — "was" carries AA1
// in the dictionary but no weight in a spoken line.
var functionWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "but": true, "or": true,
	"nor": true, "of": true, "in": true, "on": true, "at": true, "to": true,
	"by": true, "up": true, "was": true, "is": true, "are": true, "were": true,
	"be": true, "been": true, "am": true, "his": true, "her": true, "its": true,
	"my": true, "your": true, "their": true, "our": true, "when": true,
	"then": true, "than": true, "that": true, "this": true, "with": true,
	"for": true, "from": true, "as": true, "he": true, "she": true, "it": true,
	"we": true, "they": true, "you": true, "i": true, "so": true, "if": true,
	"had": true, "has": true, "have": true, "did": true, "do": true, "not": true,
}

// ScanLine reads a line of text into a Scansion. Words are looked up in the
// pronouncing dictionary; monosyllabic function words are demoted to weak,
// and stress digits collapse to weak (0) or strong (1/2). Unknown words are
// reported, not guessed.
//
//	phonetics.ScanLine("the cat rushed inside").Pattern // "0101"
func ScanLine(line string) Scansion {
	scansion := Scansion{}
	pattern := make([]byte, 0, 32)
	for _, word := range splitWords(line) {
		pron, ok := Primary(word)
		if !ok {
			scansion.Unknown = append(scansion.Unknown, word)
			continue
		}
		scansion.Words = append(scansion.Words, word)
		stresses := pron.StressPattern()
		if len(stresses) == 1 && functionWords[word] {
			pattern = append(pattern, '0')
			continue
		}
		for i := 0; i < len(stresses); i++ {
			// Binary scansion: only PRIMARY stress fills a strong slot;
			// secondary stress reads weak (inside 21 → 01, an iamb).
			if stresses[i] == '1' {
				pattern = append(pattern, '1')
			} else {
				pattern = append(pattern, '0')
			}
		}
	}
	scansion.Pattern = string(pattern)
	scansion.Feet, scansion.Regularity = dominantFoot(scansion.Pattern)
	return scansion
}

// dominantFoot tiles each classical foot template across the pattern and
// returns the best fit with its match ratio.
//
//	dominantFoot("010101") // ("iamb", 1.0)
func dominantFoot(pattern string) (string, float64) {
	if pattern == "" {
		return "", 0
	}
	bestFoot := ""
	bestScore := -1.0
	// Fixed evaluation order keeps ties deterministic.
	for _, foot := range []string{"iamb", "trochee", "anapaest", "dactyl", "spondee"} {
		template := metricalFeet[foot]
		matches := 0
		for i := 0; i < len(pattern); i++ {
			if pattern[i] == template[i%len(template)] {
				matches++
			}
		}
		score := float64(matches) / float64(len(pattern))
		if score > bestScore {
			bestFoot = foot
			bestScore = score
		}
	}
	return bestFoot, bestScore
}

// splitWords lowers a line and splits it into dictionary-shaped words,
// keeping in-word apostrophes and hyphens ('em, x-ray) and dropping other
// punctuation.
func splitWords(line string) []string {
	lower := core.Lower(line)
	words := make([]string, 0, 16)
	current := make([]rune, 0, 24)
	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0]
		}
	}
	for _, r := range lower {
		switch {
		case (r >= 'a' && r <= 'z') || r == '\'' || r == '-':
			current = append(current, r)
		default:
			flush()
		}
	}
	flush()
	return words
}
