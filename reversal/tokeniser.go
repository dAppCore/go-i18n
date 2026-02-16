// Package reversal provides reverse grammar lookups.
//
// The forward engine (go-i18n) maps base forms to inflected forms:
//
//	PastTense("delete") → "deleted"
//	Gerund("run")       → "running"
//
// The reversal engine reads those same tables backwards, turning
// inflected forms back into base forms with tense metadata:
//
//	MatchVerb("deleted")  → {Base: "delete", Tense: "past"}
//	MatchVerb("running")  → {Base: "run",    Tense: "gerund"}
//
// 3-tier lookup: JSON grammar data → irregular verb maps → regular
// morphology rules (verified by round-tripping through forward functions).
package reversal

import (
	"strings"

	i18n "forge.lthn.ai/core/go-i18n"
)

// VerbMatch holds the result of a reverse verb lookup.
type VerbMatch struct {
	Base  string // Base form of the verb ("delete", "run")
	Tense string // "past", "gerund", or "base"
	Form  string // The original inflected form
}

// NounMatch holds the result of a reverse noun lookup.
type NounMatch struct {
	Base   string // Base/singular form of the noun
	Plural bool   // Whether the matched form was plural
	Form   string // The original form
}

// TokenType classifies a token identified during tokenisation.
type TokenType int

const (
	TokenUnknown     TokenType = iota // Unrecognised word
	TokenVerb                         // Matched verb (see VerbInfo)
	TokenNoun                         // Matched noun (see NounInfo)
	TokenArticle                      // Matched article ("a", "an", "the")
	TokenWord                         // Matched word from grammar word map
	TokenPunctuation                  // Punctuation ("...", "?")
)

// Token represents a single classified token from a text string.
type Token struct {
	Raw       string    // Original text as it appeared in input
	Lower     string    // Lowercased form
	Type      TokenType // Classification
	VerbInfo  VerbMatch // Set when Type == TokenVerb
	NounInfo  NounMatch // Set when Type == TokenNoun
	WordCat   string    // Set when Type == TokenWord
	ArtType   string    // Set when Type == TokenArticle
	PunctType string    // Set when Type == TokenPunctuation
}

// Tokeniser provides reverse grammar lookups by maintaining inverse
// indexes built from the forward grammar tables.
type Tokeniser struct {
	pastToBase   map[string]string // "deleted" → "delete"
	gerundToBase map[string]string // "deleting" → "delete"
	baseVerbs    map[string]bool   // "delete" → true
	pluralToBase map[string]string // "files" → "file"
	baseNouns    map[string]bool   // "file" → true
	words        map[string]string // word translations
	lang         string
}

// NewTokeniser creates a Tokeniser for English ("en").
func NewTokeniser() *Tokeniser {
	return NewTokeniserForLang("en")
}

// NewTokeniserForLang creates a Tokeniser for the specified language,
// building inverse indexes from the grammar data.
func NewTokeniserForLang(lang string) *Tokeniser {
	t := &Tokeniser{
		pastToBase:   make(map[string]string),
		gerundToBase: make(map[string]string),
		baseVerbs:    make(map[string]bool),
		pluralToBase: make(map[string]string),
		baseNouns:    make(map[string]bool),
		words:        make(map[string]string),
		lang:         lang,
	}
	t.buildVerbIndex()
	t.buildNounIndex()
	t.buildWordIndex()
	return t
}

// buildVerbIndex reads grammar tables and irregular verb maps to build
// inverse lookup maps: inflected form → base form.
func (t *Tokeniser) buildVerbIndex() {
	// Tier 1: Read from JSON grammar data (via GetGrammarData).
	data := i18n.GetGrammarData(t.lang)
	if data != nil && data.Verbs != nil {
		for base, forms := range data.Verbs {
			t.baseVerbs[base] = true
			if forms.Past != "" {
				t.pastToBase[forms.Past] = base
			}
			if forms.Gerund != "" {
				t.gerundToBase[forms.Gerund] = base
			}
		}
	}

	// Tier 2: Read from the exported irregularVerbs map.
	// Build inverse maps directly from the authoritative source.
	for base, forms := range i18n.IrregularVerbs() {
		t.baseVerbs[base] = true
		if forms.Past != "" {
			if _, exists := t.pastToBase[forms.Past]; !exists {
				t.pastToBase[forms.Past] = base
			}
		}
		if forms.Gerund != "" {
			if _, exists := t.gerundToBase[forms.Gerund]; !exists {
				t.gerundToBase[forms.Gerund] = base
			}
		}
	}
}

// buildNounIndex reads grammar tables and irregular noun maps to build
// inverse lookup maps: plural form → base form.
func (t *Tokeniser) buildNounIndex() {
	// Tier 1: Read from JSON grammar data (via GetGrammarData).
	data := i18n.GetGrammarData(t.lang)
	if data != nil && data.Nouns != nil {
		for base, forms := range data.Nouns {
			t.baseNouns[base] = true
			if forms.Other != "" && forms.Other != base {
				t.pluralToBase[forms.Other] = base
			}
		}
	}

	// Tier 2: Read from the exported irregularNouns map.
	for base, plural := range i18n.IrregularNouns() {
		t.baseNouns[base] = true
		if plural != base {
			if _, exists := t.pluralToBase[plural]; !exists {
				t.pluralToBase[plural] = base
			}
		}
	}
}

// MatchNoun performs a 3-tier reverse lookup for a noun form.
//
// Tier 1: Check if the word is a known base noun.
// Tier 2: Check the pluralToBase inverse map.
// Tier 3: Try reverse morphology rules and round-trip verify via
// the forward function PluralForm().
func (t *Tokeniser) MatchNoun(word string) (NounMatch, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return NounMatch{}, false
	}

	// Tier 1: Is it a base noun?
	if t.baseNouns[word] {
		return NounMatch{Base: word, Plural: false, Form: word}, true
	}

	// Tier 2: Check inverse map from grammar tables + irregular nouns.
	if base, ok := t.pluralToBase[word]; ok {
		return NounMatch{Base: base, Plural: true, Form: word}, true
	}

	// Tier 3: Reverse morphology with round-trip verification.
	candidates := t.reverseRegularPlural(word)
	for _, c := range candidates {
		if i18n.PluralForm(c) == word {
			return NounMatch{Base: c, Plural: true, Form: word}, true
		}
	}

	return NounMatch{}, false
}

// reverseRegularPlural generates candidate base forms by reversing regular
// plural suffixes. Returns multiple candidates ordered by likelihood.
//
// The forward engine applies rules in this order:
//  1. ends in s/ss/sh/ch/x/z → +es
//  2. ends in consonant+y → ies
//  3. ends in f → ves, fe → ves
//  4. default → +s
//
// We generate candidates for each possible reverse rule. Round-trip
// verification ensures only correct candidates pass.
func (t *Tokeniser) reverseRegularPlural(word string) []string {
	var candidates []string

	// Rule: consonant + "ies" → consonant + "y" (e.g., "entries" → "entry")
	if strings.HasSuffix(word, "ies") && len(word) > 3 {
		base := word[:len(word)-3] + "y"
		candidates = append(candidates, base)
	}

	// Rule: "ves" → "f" or "fe" (e.g., "wolves" → "wolf", "knives" → "knife")
	if strings.HasSuffix(word, "ves") && len(word) > 3 {
		candidates = append(candidates, word[:len(word)-3]+"f")
		candidates = append(candidates, word[:len(word)-3]+"fe")
	}

	// Rule: sibilant + "es" (e.g., "processes" → "process", "branches" → "branch")
	if strings.HasSuffix(word, "ses") || strings.HasSuffix(word, "xes") ||
		strings.HasSuffix(word, "zes") || strings.HasSuffix(word, "ches") ||
		strings.HasSuffix(word, "shes") {
		base := word[:len(word)-2] // strip "es"
		candidates = append(candidates, base)
	}

	// Rule: drop "s" (e.g., "servers" → "server")
	if strings.HasSuffix(word, "s") && len(word) > 1 {
		base := word[:len(word)-1]
		candidates = append(candidates, base)
	}

	return candidates
}

// MatchVerb performs a 3-tier reverse lookup for a verb form.
//
// Tier 1: Check if the word is a known base verb.
// Tier 2: Check the pastToBase and gerundToBase inverse maps.
// Tier 3: Try reverse morphology rules and round-trip verify via
// the forward functions PastTense() and Gerund().
func (t *Tokeniser) MatchVerb(word string) (VerbMatch, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return VerbMatch{}, false
	}

	// Tier 1: Is it a base verb?
	if t.baseVerbs[word] {
		return VerbMatch{Base: word, Tense: "base", Form: word}, true
	}

	// Tier 2: Check inverse maps from grammar tables + irregular verbs.
	if base, ok := t.pastToBase[word]; ok {
		return VerbMatch{Base: base, Tense: "past", Form: word}, true
	}
	if base, ok := t.gerundToBase[word]; ok {
		return VerbMatch{Base: base, Tense: "gerund", Form: word}, true
	}

	// Tier 3: Reverse morphology with round-trip verification.
	// Try past tense candidates.
	if base := t.bestRoundTrip(word, t.reverseRegularPast(word), i18n.PastTense); base != "" {
		return VerbMatch{Base: base, Tense: "past", Form: word}, true
	}

	// Try gerund candidates.
	if base := t.bestRoundTrip(word, t.reverseRegularGerund(word), i18n.Gerund); base != "" {
		return VerbMatch{Base: base, Tense: "gerund", Form: word}, true
	}

	return VerbMatch{}, false
}

// bestRoundTrip selects the best candidate from a list by round-tripping
// each through a forward function. When multiple candidates round-trip
// successfully (ambiguity), it uses the following priority:
//  1. Candidates that are known base verbs (in grammar tables / irregular maps)
//  2. Candidates ending in a VCe pattern (vowel-consonant-e, the "magic e"
//     pattern common in real English verbs like "delete", "create", "use").
//     This avoids phantom verbs like "walke" or "processe" which have a
//     CCe pattern (consonant-consonant-e) that doesn't occur naturally.
//  3. Candidates NOT ending in "e" (the default morphology path)
//  4. First match in candidate order as final tiebreaker
func (t *Tokeniser) bestRoundTrip(target string, candidates []string, forward func(string) string) string {
	var matches []string
	for _, c := range candidates {
		if forward(c) == target {
			matches = append(matches, c)
		}
	}
	if len(matches) == 0 {
		return ""
	}
	if len(matches) == 1 {
		return matches[0]
	}

	// Priority 1: known base verb
	for _, m := range matches {
		if t.baseVerbs[m] {
			return m
		}
	}

	// Priority 2: prefer VCe-ending candidate (real English verb pattern)
	for _, m := range matches {
		if hasVCeEnding(m) {
			return m
		}
	}

	// Priority 3: prefer candidate not ending in "e" (avoids phantom verbs
	// with CCe endings like "walke", "processe")
	for _, m := range matches {
		if !strings.HasSuffix(m, "e") {
			return m
		}
	}

	return matches[0]
}

// hasVCeEnding returns true if the word ends in a vowel-consonant-e pattern
// (the "magic e" pattern). This is characteristic of real English verbs like
// "delete" (-ete), "create" (-ate), "use" (-use), "close" (-ose).
// Phantom verbs produced by naive suffix stripping like "walke" (-lke) or
// "processe" (-sse) end in consonant-consonant-e and return false.
func hasVCeEnding(word string) bool {
	if len(word) < 3 || word[len(word)-1] != 'e' {
		return false
	}
	lastConsonant := word[len(word)-2]
	vowelBefore := word[len(word)-3]
	return !isVowelByte(lastConsonant) && isVowelByte(vowelBefore)
}

func isVowelByte(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// reverseRegularPast generates candidate base forms by reversing regular
// past tense suffixes. Returns multiple candidates ordered by likelihood.
//
// The forward engine applies rules in this order:
//  1. ends in "e" → +d  (create → created)
//  2. ends in "y" + consonant → ied  (copy → copied)
//  3. shouldDoubleConsonant → double+ed  (stop → stopped)
//  4. default → +ed  (walk → walked)
//
// We generate candidates for each possible reverse rule. Round-trip
// verification (in bestRoundTrip) ensures only correct candidates pass.
func (t *Tokeniser) reverseRegularPast(word string) []string {
	var candidates []string

	if !strings.HasSuffix(word, "ed") {
		return candidates
	}

	// Rule: consonant + "ied" → consonant + "y" (e.g., "copied" → "copy")
	if strings.HasSuffix(word, "ied") && len(word) > 3 {
		base := word[:len(word)-3] + "y"
		candidates = append(candidates, base)
	}

	// Rule: doubled consonant + "ed" → single consonant (e.g., "stopped" → "stop")
	if len(word) > 4 {
		beforeEd := word[:len(word)-2]
		lastChar := beforeEd[len(beforeEd)-1]
		if len(beforeEd) >= 2 && beforeEd[len(beforeEd)-2] == lastChar {
			base := beforeEd[:len(beforeEd)-1]
			candidates = append(candidates, base)
		}
	}

	// Rule: stem + "d" where stem ends in "e" (e.g., "created" → "create")
	if len(word) > 2 {
		stemPlusE := word[:len(word)-1] // strip "d", leaving stem + "e"
		candidates = append(candidates, stemPlusE)
	}

	// Rule: stem + "ed" (e.g., "walked" → "walk")
	if len(word) > 2 {
		stem := word[:len(word)-2]
		candidates = append(candidates, stem)
	}

	return candidates
}

// reverseRegularGerund generates candidate base forms by reversing regular
// gerund suffixes. Returns multiple candidates ordered by likelihood.
//
// Rules reversed:
//   - verb + "ing"          (e.g., "walking" → "walk")
//   - verb[:-1] + "ing"     (e.g., "creating" → "create", drop e)
//   - doubled consonant     (e.g., "stopping" → "stop")
//   - verb[:-2] + "ying"    (e.g., "dying" → "die")
func (t *Tokeniser) reverseRegularGerund(word string) []string {
	var candidates []string

	if !strings.HasSuffix(word, "ing") || len(word) < 4 {
		return candidates
	}

	stem := word[:len(word)-3] // strip "ing"

	// Rule: "ying" → "ie" (e.g., "dying" → "die")
	if strings.HasSuffix(word, "ying") && len(word) > 4 {
		base := word[:len(word)-4] + "ie"
		candidates = append(candidates, base)
	}

	// Rule: doubled consonant + "ing" → single consonant (e.g., "stopping" → "stop")
	if len(stem) >= 2 && stem[len(stem)-1] == stem[len(stem)-2] {
		base := stem[:len(stem)-1]
		candidates = append(candidates, base)
	}

	// Rule: direct strip "ing" (e.g., "walking" → "walk")
	// This must come before the stem+"e" rule to avoid false positives
	// like "walke" round-tripping through Gerund("walke") = "walking".
	candidates = append(candidates, stem)

	// Rule: stem + "e" was dropped before "ing" (e.g., "creating" → "create")
	// Try adding "e" back.
	candidates = append(candidates, stem+"e")

	return candidates
}

// buildWordIndex reads GrammarData.Words and builds a reverse lookup map.
// Both the key (e.g., "url") and the display form (e.g., "URL") map back
// to the key, enabling case-insensitive lookups.
func (t *Tokeniser) buildWordIndex() {
	data := i18n.GetGrammarData(t.lang)
	if data == nil || data.Words == nil {
		return
	}
	for key, display := range data.Words {
		// Map the key itself (already lowercase)
		t.words[strings.ToLower(key)] = key
		// Map the display form (e.g., "URL" → "url", "SSH" → "ssh")
		t.words[strings.ToLower(display)] = key
	}
}

// MatchWord performs a case-insensitive lookup in the words map.
// Returns the category key and true if found, or ("", false) otherwise.
func (t *Tokeniser) MatchWord(word string) (string, bool) {
	cat, ok := t.words[strings.ToLower(word)]
	return cat, ok
}

// MatchArticle checks whether a word is an article (definite or indefinite).
// Returns the article type ("indefinite" or "definite") and true if matched,
// or ("", false) otherwise.
func (t *Tokeniser) MatchArticle(word string) (string, bool) {
	data := i18n.GetGrammarData(t.lang)
	if data == nil {
		return "", false
	}

	lower := strings.ToLower(word)

	if lower == strings.ToLower(data.Articles.IndefiniteDefault) ||
		lower == strings.ToLower(data.Articles.IndefiniteVowel) {
		return "indefinite", true
	}
	if lower == strings.ToLower(data.Articles.Definite) {
		return "definite", true
	}

	return "", false
}

// Tokenise splits text on whitespace and classifies each word.
// Priority: punctuation → article → verb → noun → word → unknown.
// Trailing punctuation is stripped from words before matching.
func (t *Tokeniser) Tokenise(text string) []Token {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	parts := strings.Fields(text)
	var tokens []Token

	for _, raw := range parts {
		// Strip trailing punctuation to get the clean word.
		word, punct := splitTrailingPunct(raw)

		// Classify the word portion (if any).
		if word != "" {
			tok := Token{Raw: raw, Lower: strings.ToLower(word)}

			if artType, ok := t.MatchArticle(word); ok {
				tok.Type = TokenArticle
				tok.ArtType = artType
			} else if vm, ok := t.MatchVerb(word); ok {
				tok.Type = TokenVerb
				tok.VerbInfo = vm
			} else if nm, ok := t.MatchNoun(word); ok {
				tok.Type = TokenNoun
				tok.NounInfo = nm
			} else if cat, ok := t.MatchWord(word); ok {
				tok.Type = TokenWord
				tok.WordCat = cat
			} else {
				tok.Type = TokenUnknown
			}
			tokens = append(tokens, tok)
		}

		// Emit a punctuation token if trailing punctuation was found.
		if punct != "" {
			if punctType, ok := matchPunctuation(punct); ok {
				tokens = append(tokens, Token{
					Raw:       punct,
					Lower:     punct,
					Type:      TokenPunctuation,
					PunctType: punctType,
				})
			}
		}
	}

	return tokens
}

// splitTrailingPunct separates a word from its trailing punctuation.
// Returns the word and the punctuation suffix. Punctuation patterns
// recognised: "..." (progress), "?" (question), ":" (label).
func splitTrailingPunct(s string) (string, string) {
	// Check for "..." suffix first (3-char pattern).
	if strings.HasSuffix(s, "...") {
		return s[:len(s)-3], "..."
	}
	// Check single-char trailing punctuation.
	if len(s) > 1 {
		last := s[len(s)-1]
		if last == '?' || last == ':' {
			return s[:len(s)-1], string(last)
		}
	}
	return s, ""
}

// matchPunctuation detects known punctuation patterns.
// Returns the punctuation type and true if recognised.
func matchPunctuation(punct string) (string, bool) {
	switch punct {
	case "...":
		return "progress", true
	case "?":
		return "question", true
	case ":":
		return "label", true
	}
	return "", false
}
