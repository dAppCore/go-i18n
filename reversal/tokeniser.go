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
	"unicode/utf8"

	"dappco.re/go/core"
	i18n "dappco.re/go/core/i18n"
)

var frenchElisionPrefixes = []string{"l", "d", "j", "m", "t", "s", "n", "c", "qu"}

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
	Raw        string           // Original text as it appeared in input
	Lower      string           // Lowercased form
	Type       TokenType        // Classification
	Confidence float64          // 0.0-1.0 classification confidence
	AltType    TokenType        // Runner-up classification (dual-class only)
	AltConf    float64          // Runner-up confidence
	VerbInfo   VerbMatch        // Set when Type OR AltType == TokenVerb
	NounInfo   NounMatch        // Set when Type OR AltType == TokenNoun
	WordCat    string           // Set when Type == TokenWord
	ArtType    string           // Set when Type == TokenArticle
	PunctType  string           // Set when Type == TokenPunctuation
	Signals    *SignalBreakdown // Non-nil only when WithSignals() option is set
}

// SignalBreakdown provides detailed scoring for dual-class disambiguation.
type SignalBreakdown struct {
	VerbScore  float64
	NounScore  float64
	Components []SignalComponent
}

// SignalComponent describes a single signal's contribution to disambiguation.
type SignalComponent struct {
	Name    string  // "noun_determiner", "verb_auxiliary", etc.
	Weight  float64 // Signal weight (0.0-1.0)
	Value   float64 // Signal firing strength (0.0-1.0)
	Contrib float64 // weight x value
	Reason  string  // Human-readable: "preceded by 'the'"
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
	phraseLen    int               // longest multi-word gram.word entry
	lang         string

	dualClass   map[string]bool    // words in both verb AND noun tables
	nounDet     map[string]bool    // signal: noun determiners
	verbAux     map[string]bool    // signal: verb auxiliaries
	verbInf     map[string]bool    // signal: infinitive markers
	verbNeg     map[string]bool    // signal: negation cues
	withSignals bool               // allocate SignalBreakdown on ambiguous tokens
	weights     map[string]float64 // signal weights (F3: configurable)
}

// TokeniserOption configures a Tokeniser.
type TokeniserOption func(*Tokeniser)

// WithSignals enables detailed SignalBreakdown on ambiguous tokens.
func WithSignals() TokeniserOption {
	return func(t *Tokeniser) { t.withSignals = true }
}

// WithWeights overrides the default signal weights for disambiguation.
// Omitted keys keep their default weights so partial overrides stay safe.
func WithWeights(w map[string]float64) TokeniserOption {
	return func(t *Tokeniser) {
		if len(w) == 0 {
			t.weights = nil
			return
		}
		// Start from the defaults so callers can override only the weights they
		// care about without accidentally disabling the rest of the signal set.
		copied := defaultWeights()
		for key, value := range w {
			copied[key] = value
		}
		t.weights = copied
	}
}

// NewTokeniser creates a Tokeniser for English ("en").
func NewTokeniser(opts ...TokeniserOption) *Tokeniser {
	return NewTokeniserForLang("en", opts...)
}

// NewTokeniserForLang creates a Tokeniser for the specified language,
// building inverse indexes from the grammar data.
func NewTokeniserForLang(lang string, opts ...TokeniserOption) *Tokeniser {
	t := &Tokeniser{
		pastToBase:   make(map[string]string),
		gerundToBase: make(map[string]string),
		baseVerbs:    make(map[string]bool),
		pluralToBase: make(map[string]string),
		baseNouns:    make(map[string]bool),
		words:        make(map[string]string),
		lang:         lang,
	}
	for _, opt := range opts {
		opt(t)
	}
	t.buildVerbIndex()
	t.buildNounIndex()
	t.buildWordIndex()
	t.buildDualClassIndex()
	t.buildSignalIndex()
	if t.weights == nil {
		t.weights = defaultWeights()
	}
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

	// Tier 2b: Seed additional regular dual-class bases that are common in
	// dev/ops text. These are regular forms, but they need to behave like
	// known bases so the dual-class resolver can disambiguate them.
	for base, forms := range i18n.DualClassVerbs() {
		t.baseVerbs[base] = true
		if forms.Past != "" && t.pastToBase[forms.Past] == "" {
			t.pastToBase[forms.Past] = base
		}
		if forms.Gerund != "" && t.gerundToBase[forms.Gerund] == "" {
			t.gerundToBase[forms.Gerund] = base
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
			if skipDeprecatedEnglishGrammarEntry(base) {
				continue
			}
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

	// Tier 2b: Seed additional regular dual-class bases that are common in
	// dev/ops text. The plural forms are regular, but the entries need to
	// appear in the base noun set so the ambiguous-token pass can see them.
	for base, plural := range i18n.DualClassNouns() {
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
	word = core.Lower(core.Trim(word))
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
	if core.HasSuffix(word, "ies") && len(word) > 3 {
		base := word[:len(word)-3] + "y"
		candidates = append(candidates, base)
	}

	// Rule: "ves" → "f" or "fe" (e.g., "wolves" → "wolf", "knives" → "knife")
	if core.HasSuffix(word, "ves") && len(word) > 3 {
		candidates = append(candidates, word[:len(word)-3]+"f")
		candidates = append(candidates, word[:len(word)-3]+"fe")
	}

	// Rule: sibilant + "es" (e.g., "processes" → "process", "branches" → "branch")
	if core.HasSuffix(word, "ses") || core.HasSuffix(word, "xes") ||
		core.HasSuffix(word, "zes") || core.HasSuffix(word, "ches") ||
		core.HasSuffix(word, "shes") {
		base := word[:len(word)-2] // strip "es"
		candidates = append(candidates, base)
	}

	// Rule: drop "s" (e.g., "servers" → "server")
	if core.HasSuffix(word, "s") && len(word) > 1 {
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
	word = core.Lower(core.Trim(word))
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
		if !core.HasSuffix(m, "e") {
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

	if !core.HasSuffix(word, "ed") {
		return candidates
	}

	// Rule: consonant + "ied" → consonant + "y" (e.g., "copied" → "copy")
	if core.HasSuffix(word, "ied") && len(word) > 3 {
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

	if !core.HasSuffix(word, "ing") || len(word) < 4 {
		return candidates
	}

	stem := word[:len(word)-3] // strip "ing"

	// Rule: "ying" → "ie" (e.g., "dying" → "die")
	if core.HasSuffix(word, "ying") && len(word) > 4 {
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
		if skipDeprecatedEnglishGrammarEntry(key) {
			continue
		}
		// Map the key itself (already lowercase)
		t.words[core.Lower(key)] = key
		// Map the display form (e.g., "URL" → "url", "SSH" → "ssh")
		lowerDisplay := core.Lower(display)
		t.words[lowerDisplay] = key
		if words := strings.Fields(lowerDisplay); len(words) > 1 && len(words) > t.phraseLen {
			t.phraseLen = len(words)
		}
	}
}

// IsDualClass returns true if the word exists in both verb and noun tables.
func (t *Tokeniser) IsDualClass(word string) bool {
	return t.dualClass[core.Lower(word)]
}

func (t *Tokeniser) buildDualClassIndex() {
	t.dualClass = make(map[string]bool)
	for base := range t.baseVerbs {
		if t.baseNouns[base] {
			t.dualClass[base] = true
		}
	}
}

func (t *Tokeniser) buildSignalIndex() {
	t.nounDet = make(map[string]bool)
	t.verbAux = make(map[string]bool)
	t.verbInf = make(map[string]bool)
	t.verbNeg = make(map[string]bool)

	data := i18n.GetGrammarData(t.lang)

	// Guard each signal list independently so partial locale data
	// falls back per-field rather than silently disabling signals.
	if data != nil && len(data.Signals.NounDeterminers) > 0 {
		for _, w := range data.Signals.NounDeterminers {
			t.nounDet[core.Lower(w)] = true
		}
	} else {
		for _, w := range []string{
			"the", "a", "an", "this", "that", "these", "those",
			"my", "your", "his", "her", "its", "our", "their",
			"every", "each", "some", "any", "no",
			"many", "few", "several", "all", "both",
		} {
			t.nounDet[w] = true
		}
	}

	if data != nil && len(data.Signals.VerbAuxiliaries) > 0 {
		for _, w := range data.Signals.VerbAuxiliaries {
			t.verbAux[core.Lower(w)] = true
		}
	} else {
		for _, w := range defaultVerbAuxiliaries() {
			t.verbAux[w] = true
		}
	}

	if data != nil && len(data.Signals.VerbInfinitive) > 0 {
		for _, w := range data.Signals.VerbInfinitive {
			t.verbInf[core.Lower(w)] = true
		}
	} else {
		t.verbInf["to"] = true
	}

	if data != nil && len(data.Signals.VerbNegation) > 0 {
		for _, w := range data.Signals.VerbNegation {
			t.verbNeg[core.Lower(w)] = true
		}
	} else {
		// Keep the fallback conservative: these are weak cues, not hard
		// negation parsing.
		for _, w := range []string{"not", "never"} {
			t.verbNeg[w] = true
		}
	}
}

func defaultVerbAuxiliaries() []string {
	return []string{
		"am", "is", "are", "was", "were",
		"has", "had", "have",
		"do", "does", "did",
		"will", "would", "could", "should",
		"can", "may", "might", "shall", "must",
		"don't", "can't", "won't", "shouldn't", "couldn't", "wouldn't",
		"doesn't", "didn't", "isn't", "aren't", "wasn't", "weren't",
		"hasn't", "hadn't", "haven't",
	}
}

func defaultWeights() map[string]float64 {
	return map[string]float64{
		"noun_determiner":   0.35,
		"verb_auxiliary":    0.25,
		"verb_negation":     0.05,
		"following_class":   0.15,
		"sentence_position": 0.10,
		"verb_saturation":   0.10,
		"inflection_echo":   0.03,
		"default_prior":     0.02,
	}
}

func skipDeprecatedEnglishGrammarEntry(key string) bool {
	switch core.Lower(key) {
	case "passed", "failed", "skipped":
		return true
	default:
		return false
	}
}

// MatchWord performs a case-insensitive lookup in the words map.
// Returns the category key and true if found, or ("", false) otherwise.
func (t *Tokeniser) MatchWord(word string) (string, bool) {
	cat, ok := t.words[core.Lower(word)]
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

	if base, _ := splitTrailingPunct(word); base != "" {
		word = base
	}
	lower := core.Lower(word)

	if artType, ok := matchConfiguredArticleText(lower, data); ok {
		return artType, true
	}
	if t.isFrenchLanguage() {
		if artType, ok := matchFrenchLeadingArticlePhrase(lower); ok {
			return artType, true
		}
		if artType, ok := matchFrenchArticleText(lower); ok {
			return artType, true
		}
		if artType, ok := matchFrenchAttachedArticle(lower); ok {
			return artType, true
		}
		switch lower {
		case "l'", "l’", "les", "au", "aux":
			return "definite", true
		case "d'", "d’", "de l'", "de l’", "de la", "du", "des":
			return "indefinite", true
		case "j'", "j’", "m'", "m’", "t'", "t’", "s'", "s’", "n'", "n’", "c'", "c’", "qu'", "qu’":
			return "definite", true
		case "un", "une":
			return "indefinite", true
		}
	}

	return "", false
}

func matchConfiguredArticleText(lower string, data *i18n.GrammarData) (string, bool) {
	if data == nil {
		return "", false
	}

	if lower == core.Lower(data.Articles.IndefiniteDefault) ||
		lower == core.Lower(data.Articles.IndefiniteVowel) {
		return "indefinite", true
	}
	if lower == core.Lower(data.Articles.Definite) {
		return "definite", true
	}
	for _, article := range data.Articles.ByGender {
		if lower == core.Lower(article) {
			return "definite", true
		}
	}

	if idx := strings.IndexAny(lower, " \t"); idx > 0 {
		prefix := core.Trim(lower[:idx])
		if prefix == "" {
			return "", false
		}
		if prefix == core.Lower(data.Articles.IndefiniteDefault) ||
			prefix == core.Lower(data.Articles.IndefiniteVowel) {
			return "indefinite", true
		}
		if prefix == core.Lower(data.Articles.Definite) {
			return "definite", true
		}
		for _, article := range data.Articles.ByGender {
			if prefix == core.Lower(article) {
				return "definite", true
			}
		}
	}

	return "", false
}

func matchFrenchLeadingArticlePhrase(lower string) (string, bool) {
	switch {
	case lower == "le", lower == "la", lower == "les",
		lower == "l'", lower == "l’", lower == "au", lower == "aux":
		return "definite", true
	case lower == "un", lower == "une", lower == "du", lower == "des":
		return "indefinite", true
	}

	for _, prefix := range []struct {
		text string
		kind string
	}{
		{text: "le ", kind: "definite"},
		{text: "la ", kind: "definite"},
		{text: "les ", kind: "definite"},
		{text: "un ", kind: "indefinite"},
		{text: "une ", kind: "indefinite"},
		{text: "du ", kind: "indefinite"},
		{text: "des ", kind: "indefinite"},
		{text: "au ", kind: "definite"},
		{text: "aux ", kind: "definite"},
		{text: "l'", kind: "definite"},
		{text: "l’", kind: "definite"},
		{text: "d'", kind: "indefinite"},
		{text: "d’", kind: "indefinite"},
	} {
		if strings.HasPrefix(lower, prefix.text) {
			return prefix.kind, true
		}
	}

	return "", false
}

func matchFrenchArticleText(lower string) (string, bool) {
	switch {
	case strings.HasPrefix(lower, "de l'"), strings.HasPrefix(lower, "de l’"):
		return "indefinite", true
	case strings.HasPrefix(lower, "de la "), strings.HasPrefix(lower, "de le "), strings.HasPrefix(lower, "de les "), strings.HasPrefix(lower, "du "), strings.HasPrefix(lower, "des "):
		return "indefinite", true
	case strings.HasPrefix(lower, "au "), strings.HasPrefix(lower, "aux "):
		return "definite", true
	}

	fields := strings.Fields(lower)
	if len(fields) == 0 {
		return "", false
	}

	switch fields[0] {
	case "l'", "l’", "les", "au", "aux":
		return "definite", true
	case "un", "une":
		return "indefinite", true
	case "du", "des":
		return "indefinite", true
	case "de":
		if len(fields) >= 2 {
			switch fields[1] {
			case "la", "le", "les", "l'", "l’":
				return "indefinite", true
			case "du", "des":
				return "definite", true
			}
		}
	case "d'", "d’":
		return "indefinite", true
	case "j'", "j’", "m'", "m’", "t'", "t’", "s'", "s’", "n'", "n’", "c'", "c’", "qu'", "qu’":
		return "definite", true
	}

	if artType, ok := matchFrenchAttachedArticle(lower); ok {
		return artType, true
	}

	return "", false
}

func matchFrenchAttachedArticle(lower string) (string, bool) {
	for _, prefix := range frenchElisionPrefixes {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		rest := strings.TrimPrefix(lower, prefix)
		if rest == "" {
			continue
		}
		if !strings.HasPrefix(rest, "'") && !strings.HasPrefix(rest, "’") {
			continue
		}
		switch prefix {
		case "d":
			return "indefinite", true
		case "l":
			return "definite", true
		default:
			return "definite", true
		}
	}
	return "", false
}

// tokenAmbiguous is an internal sentinel used during Pass 1 to mark
// dual-class base forms that need disambiguation in Pass 2.
const tokenAmbiguous TokenType = -1

// clauseBoundaries lists words that delimit clause boundaries for
// the verb_saturation signal (D2 review fix).
var clauseBoundaries = map[string]bool{
	"and": true, "or": true, "but": true, "because": true,
	"when": true, "while": true, "if": true, "then": true, "so": true,
}

// Tokenise splits text on whitespace and classifies each word using a
// two-pass algorithm:
//
// Pass 1 classifies unambiguous tokens and marks dual-class base forms.
// Pass 2 resolves ambiguous tokens using weighted disambiguation signals.
func (t *Tokeniser) Tokenise(text string) []Token {
	text = core.Trim(text)
	if text == "" {
		return nil
	}

	parts := strings.Fields(text)
	var tokens []Token

	// --- Pass 1: Classify & Mark ---
	for i := 0; i < len(parts); i++ {
		if consumed, tok, punctTok := t.matchWordPhrase(parts, i); consumed > 0 {
			tokens = append(tokens, tok)
			if punctTok != nil {
				tokens = append(tokens, *punctTok)
			}
			i += consumed - 1
			continue
		}
		if consumed, tok, extraTok, punctTok := t.matchFrenchArticlePhrase(parts, i); consumed > 0 {
			tokens = append(tokens, tok)
			if extraTok != nil {
				tokens = append(tokens, *extraTok)
			}
			if punctTok != nil {
				tokens = append(tokens, *punctTok)
			}
			i += consumed - 1
			continue
		}

		raw := parts[i]
		if prefix, rest, ok := t.splitFrenchElision(raw); ok {
			if artType, ok := t.MatchArticle(prefix); ok {
				tokens = append(tokens, Token{
					Raw:        prefix,
					Lower:      core.Lower(prefix),
					Type:       TokenArticle,
					ArtType:    artType,
					Confidence: 1.0,
				})
			}
			raw = rest
			if raw == "" {
				continue
			}
		}

		// Strip trailing punctuation to get the clean word.
		word, punct := splitTrailingPunct(raw)

		// Classify the word portion (if any).
		if word != "" {
			tok := Token{Raw: raw, Lower: core.Lower(word)}

			if artType, ok := t.MatchArticle(word); ok {
				// Articles are unambiguous.
				tok.Type = TokenArticle
				tok.ArtType = artType
				tok.Confidence = 1.0
			} else {
				// For non-articles, check BOTH verb and noun.
				vm, verbOK := t.MatchVerb(word)
				nm, nounOK := t.MatchNoun(word)

				if verbOK && nounOK && t.dualClass[tok.Lower] {
					// Dual-class word: check for self-resolving inflections.
					if vm.Tense != "base" {
						// Inflected verb form self-resolves.
						tok.Type = TokenVerb
						tok.VerbInfo = vm
						tok.NounInfo = nm
						tok.Confidence = 1.0
					} else if nm.Plural {
						// Inflected noun form self-resolves.
						tok.Type = TokenNoun
						tok.VerbInfo = vm
						tok.NounInfo = nm
						tok.Confidence = 1.0
					} else {
						// Base form: ambiguous, stash both and defer to Pass 2.
						tok.Type = tokenAmbiguous
						tok.VerbInfo = vm
						tok.NounInfo = nm
					}
				} else if verbOK {
					tok.Type = TokenVerb
					tok.VerbInfo = vm
					tok.Confidence = 1.0
				} else if nounOK {
					tok.Type = TokenNoun
					tok.NounInfo = nm
					tok.Confidence = 1.0
				} else if cat, ok := t.MatchWord(word); ok {
					tok.Type = TokenWord
					tok.WordCat = cat
					tok.Confidence = 1.0
				} else {
					tok.Type = TokenUnknown
				}
			}
			tokens = append(tokens, tok)
		}

		// Emit a punctuation token if trailing punctuation was found.
		if punct != "" {
			if punctType, ok := matchPunctuation(punct); ok {
				tokens = append(tokens, Token{
					Raw:        punct,
					Lower:      punct,
					Type:       TokenPunctuation,
					PunctType:  punctType,
					Confidence: 1.0,
				})
			}
		}
	}

	// --- Pass 2: Resolve Ambiguous ---
	t.resolveAmbiguous(tokens)

	return tokens
}

func (t *Tokeniser) matchWordPhrase(parts []string, start int) (int, Token, *Token) {
	if t.phraseLen < 2 || start >= len(parts) {
		return 0, Token{}, nil
	}

	maxLen := t.phraseLen
	if remaining := len(parts) - start; remaining < maxLen {
		maxLen = remaining
	}

	for n := maxLen; n >= 2; n-- {
		phraseWords := make([]string, 0, n)
		rawParts := make([]string, 0, n)
		var punct string
		valid := true

		for j := 0; j < n; j++ {
			part := parts[start+j]
			if prefix, _, ok := t.splitFrenchElision(part); ok && prefix != part {
				valid = false
				break
			}

			word, partPunct := splitTrailingPunct(part)
			if word == "" {
				valid = false
				break
			}
			if partPunct != "" && j != n-1 {
				valid = false
				break
			}

			rawParts = append(rawParts, word)
			phraseWords = append(phraseWords, core.Lower(word))
			if j == n-1 {
				punct = partPunct
			}
		}

		if !valid {
			continue
		}

		phrase := strings.Join(phraseWords, " ")
		cat, ok := t.words[phrase]
		if !ok {
			continue
		}

		tok := Token{
			Raw:        strings.Join(rawParts, " "),
			Lower:      phrase,
			Type:       TokenWord,
			WordCat:    cat,
			Confidence: 1.0,
		}

		if punct != "" {
			if punctType, ok := matchPunctuation(punct); ok {
				punctTok := Token{
					Raw:        punct,
					Lower:      punct,
					Type:       TokenPunctuation,
					PunctType:  punctType,
					Confidence: 1.0,
				}
				return n, tok, &punctTok
			}
		}

		return n, tok, nil
	}

	return 0, Token{}, nil
}

func (t *Tokeniser) matchFrenchArticlePhrase(parts []string, start int) (int, Token, *Token, *Token) {
	if !t.isFrenchLanguage() || start+1 >= len(parts) {
		return 0, Token{}, nil, nil
	}

	first, firstPunct := splitTrailingPunct(parts[start])
	if first == "" || firstPunct != "" {
		return 0, Token{}, nil, nil
	}
	second, secondPunct := splitTrailingPunct(parts[start+1])
	if second == "" {
		return 0, Token{}, nil, nil
	}

	switch core.Lower(first) {
	case "de":
		switch core.Lower(second) {
		case "la", "le", "les", "du", "des":
			tok := Token{
				Raw:        first + " " + second,
				Lower:      core.Lower(first + " " + second),
				Type:       TokenArticle,
				ArtType:    "indefinite",
				Confidence: 1.0,
			}
			if secondPunct != "" {
				if punctType, ok := matchPunctuation(secondPunct); ok {
					punctTok := Token{
						Raw:        secondPunct,
						Lower:      secondPunct,
						Type:       TokenPunctuation,
						PunctType:  punctType,
						Confidence: 1.0,
					}
					return 2, tok, nil, &punctTok
				}
			}
			return 2, tok, nil, nil
		default:
			if prefix, rest, ok := t.splitFrenchElision(second); ok && (prefix == "l'" || prefix == "l’") && rest != "" {
				tok := Token{
					Raw:        first + " " + prefix,
					Lower:      core.Lower(first + " " + prefix),
					Type:       TokenArticle,
					ArtType:    "indefinite",
					Confidence: 1.0,
				}
				extra := t.classifyElidedFrenchWord(rest)
				var punctTok *Token
				if secondPunct != "" {
					if punctType, ok := matchPunctuation(secondPunct); ok {
						punctTok = &Token{
							Raw:        secondPunct,
							Lower:      secondPunct,
							Type:       TokenPunctuation,
							PunctType:  punctType,
							Confidence: 1.0,
						}
					}
				}
				return 2, tok, &extra, punctTok
			}
			// Handle spaced elision forms such as "de l' enfant" or "de l’ enfant".
			if (second == "l'" || second == "l’") && start+2 < len(parts) {
				third, thirdPunct := splitTrailingPunct(parts[start+2])
				if third != "" {
					tok := Token{
						Raw:        first + " " + second,
						Lower:      core.Lower(first + " " + second),
						Type:       TokenArticle,
						ArtType:    "indefinite",
						Confidence: 1.0,
					}
					extra := t.classifyElidedFrenchWord(third)
					var punctTok *Token
					if thirdPunct != "" {
						if punctType, ok := matchPunctuation(thirdPunct); ok {
							punctTok = &Token{
								Raw:        thirdPunct,
								Lower:      thirdPunct,
								Type:       TokenPunctuation,
								PunctType:  punctType,
								Confidence: 1.0,
							}
						}
					}
					return 3, tok, &extra, punctTok
				}
			}
			return 0, Token{}, nil, nil
		}
	}

	return 0, Token{}, nil, nil
}

func (t *Tokeniser) classifyElidedFrenchWord(word string) Token {
	tok := Token{Raw: word, Lower: core.Lower(word)}

	if artType, ok := t.MatchArticle(word); ok {
		tok.Type = TokenArticle
		tok.ArtType = artType
		tok.Confidence = 1.0
		return tok
	}

	vm, verbOK := t.MatchVerb(word)
	nm, nounOK := t.MatchNoun(word)
	if verbOK && nounOK && t.dualClass[tok.Lower] {
		if vm.Tense != "base" {
			tok.Type = TokenVerb
			tok.VerbInfo = vm
			tok.NounInfo = nm
			tok.Confidence = 1.0
		} else if nm.Plural {
			tok.Type = TokenNoun
			tok.VerbInfo = vm
			tok.NounInfo = nm
			tok.Confidence = 1.0
		} else {
			tok.Type = tokenAmbiguous
			tok.VerbInfo = vm
			tok.NounInfo = nm
		}
		return tok
	}
	if verbOK {
		tok.Type = TokenVerb
		tok.VerbInfo = vm
		tok.Confidence = 1.0
		return tok
	}
	if nounOK {
		tok.Type = TokenNoun
		tok.NounInfo = nm
		tok.Confidence = 1.0
		return tok
	}
	if cat, ok := t.MatchWord(word); ok {
		tok.Type = TokenWord
		tok.WordCat = cat
		tok.Confidence = 1.0
		return tok
	}

	tok.Type = TokenUnknown
	return tok
}

// resolveAmbiguous iterates all tokens and resolves any marked as
// tokenAmbiguous using the weighted scoring function.
func (t *Tokeniser) resolveAmbiguous(tokens []Token) {
	for i := range tokens {
		if tokens[i].Type != tokenAmbiguous {
			continue
		}
		verbScore, nounScore, components := t.scoreAmbiguous(tokens, i)
		t.resolveToken(&tokens[i], verbScore, nounScore, components)
	}
}

// scoreAmbiguous evaluates 8 weighted signals to determine whether an
// ambiguous token should be classified as verb or noun.
func (t *Tokeniser) scoreAmbiguous(tokens []Token, idx int) (float64, float64, []SignalComponent) {
	var verbScore, nounScore float64
	var components []SignalComponent

	// 1. noun_determiner: preceding token is a noun determiner
	if w, ok := t.weights["noun_determiner"]; ok && idx > 0 {
		prev := tokens[idx-1]
		if t.nounDet[prev.Lower] {
			nounScore += w * 1.0
			if t.withSignals {
				components = append(components, SignalComponent{
					Name: "noun_determiner", Weight: w, Value: 1.0, Contrib: w,
					Reason: "preceded by '" + prev.Lower + "'",
				})
			}
		}
	}

	// 2. verb_auxiliary: preceding token is an auxiliary or infinitive marker
	if w, ok := t.weights["verb_auxiliary"]; ok && idx > 0 {
		prev := tokens[idx-1]
		if t.verbAux[prev.Lower] || t.verbInf[prev.Lower] {
			verbScore += w * 1.0
			if t.withSignals {
				components = append(components, SignalComponent{
					Name: "verb_auxiliary", Weight: w, Value: 1.0, Contrib: w,
					Reason: "preceded by '" + prev.Lower + "'",
				})
			}
		}
	}

	// 3. verb_negation: preceding negation weakly signals a verb
	if w, ok := t.weights["verb_negation"]; ok && idx > 0 {
		prev := tokens[idx-1]
		if t.verbNeg[prev.Lower] || t.hasNoLongerBefore(tokens, idx) {
			verbScore += w * 1.0
			if t.withSignals {
				reason := "preceded by '" + prev.Lower + "'"
				if t.hasNoLongerBefore(tokens, idx) {
					reason = "preceded by 'no longer'"
				}
				components = append(components, SignalComponent{
					Name: "verb_negation", Weight: w, Value: 1.0, Contrib: w,
					Reason: reason,
				})
			}
		}
	}

	// 4. following_class: next token's class informs this token's role
	if w, ok := t.weights["following_class"]; ok && idx+1 < len(tokens) {
		next := tokens[idx+1]
		if next.Type != tokenAmbiguous {
			if next.Type == TokenArticle || t.nounDet[next.Lower] || next.Type == TokenNoun {
				// Followed by article/determiner/noun → verb signal
				verbScore += w * 1.0
				if t.withSignals {
					components = append(components, SignalComponent{
						Name: "following_class", Weight: w, Value: 1.0, Contrib: w,
						Reason: "followed by " + next.Lower + " (article/noun)",
					})
				}
			} else if next.Type == TokenVerb {
				// Followed by verb → noun signal
				nounScore += w * 1.0
				if t.withSignals {
					components = append(components, SignalComponent{
						Name: "following_class", Weight: w, Value: 1.0, Contrib: w,
						Reason: "followed by verb '" + next.Lower + "'",
					})
				}
			}
		}
	}

	// 5. sentence_position: first token in sentence → verb signal (imperative)
	if w, ok := t.weights["sentence_position"]; ok && idx == 0 {
		verbScore += w * 1.0
		if t.withSignals {
			components = append(components, SignalComponent{
				Name: "sentence_position", Weight: w, Value: 1.0, Contrib: w,
				Reason: "sentence-initial position (imperative)",
			})
		}
	}

	// 6. verb_saturation: if a confident verb already exists in the same clause
	if w, ok := t.weights["verb_saturation"]; ok {
		if t.hasConfidentVerbInClause(tokens, idx) {
			nounScore += w * 1.0
			if t.withSignals {
				components = append(components, SignalComponent{
					Name: "verb_saturation", Weight: w, Value: 1.0, Contrib: w,
					Reason: "confident verb already in clause",
				})
			}
		}
	}

	// 7. inflection_echo: another token shares the same base in inflected form
	if w, ok := t.weights["inflection_echo"]; ok {
		echoVerb, echoNoun := t.checkInflectionEcho(tokens, idx)
		if echoNoun {
			// Another token uses same base as inflected noun → signal verb
			verbScore += w * 1.0
			if t.withSignals {
				components = append(components, SignalComponent{
					Name: "inflection_echo", Weight: w, Value: 1.0, Contrib: w,
					Reason: "inflected noun echo found",
				})
			}
		}
		if echoVerb {
			// Another token uses same base as inflected verb → signal noun
			nounScore += w * 1.0
			if t.withSignals {
				components = append(components, SignalComponent{
					Name: "inflection_echo", Weight: w, Value: 1.0, Contrib: w,
					Reason: "inflected verb echo found",
				})
			}
		}
	}

	// 8. default_prior: corpus-derived priors take precedence; otherwise fall back to the static verb prior.
	if priorVerb, priorNoun, ok := t.corpusPrior(tokens[idx].Lower); ok {
		verbScore += priorVerb
		nounScore += priorNoun
		if t.withSignals {
			components = append(components, SignalComponent{
				Name: "default_prior", Weight: 1.0, Value: priorVerb, Contrib: priorVerb,
				Reason: "corpus-derived prior",
			})
			if priorNoun > 0 {
				components = append(components, SignalComponent{
					Name: "default_prior", Weight: 1.0, Value: priorNoun, Contrib: priorNoun,
					Reason: "corpus-derived prior",
				})
			}
		}
	} else if w, ok := t.weights["default_prior"]; ok {
		verbScore += w * 1.0
		if t.withSignals {
			components = append(components, SignalComponent{
				Name: "default_prior", Weight: w, Value: 1.0, Contrib: w,
				Reason: "default verb prior",
			})
		}
	}

	return verbScore, nounScore, components
}

func (t *Tokeniser) hasNoLongerBefore(tokens []Token, idx int) bool {
	if idx < 2 {
		return false
	}
	return tokens[idx-2].Lower == "no" && tokens[idx-1].Lower == "longer"
}

func (t *Tokeniser) corpusPrior(word string) (float64, float64, bool) {
	data := i18n.GetGrammarData(t.lang)
	if data == nil || len(data.Signals.Priors) == 0 {
		return 0, 0, false
	}
	bucket, ok := data.Signals.Priors[core.Lower(word)]
	if !ok || len(bucket) == 0 {
		return 0, 0, false
	}
	verb := bucket["verb"]
	noun := bucket["noun"]
	total := verb + noun
	if total <= 0 {
		return 0, 0, false
	}
	return verb / total, noun / total, true
}

// hasConfidentVerbInClause scans for a confident verb (Confidence >= 1.0)
// within the same clause as the token at idx. Clause boundaries are
// punctuation tokens and clause-boundary conjunctions/subordinators (D2).
func (t *Tokeniser) hasConfidentVerbInClause(tokens []Token, idx int) bool {
	// Scan backwards from idx to find clause start.
	start := 0
	for i := idx - 1; i >= 0; i-- {
		if tokens[i].Type == TokenPunctuation || clauseBoundaries[tokens[i].Lower] {
			start = i + 1
			break
		}
	}
	// Scan forwards from idx to find clause end.
	end := len(tokens)
	for i := idx + 1; i < len(tokens); i++ {
		if tokens[i].Type == TokenPunctuation || clauseBoundaries[tokens[i].Lower] {
			end = i
			break
		}
	}
	// Look for a confident verb in [start, end), excluding idx itself.
	for i := start; i < end; i++ {
		if i == idx {
			continue
		}
		if tokens[i].Type == TokenVerb && tokens[i].Confidence >= 1.0 {
			return true
		}
	}
	return false
}

// checkInflectionEcho checks whether another token shares the same base
// as this ambiguous token but in an inflected form. Returns (echoVerb, echoNoun)
// where echoVerb means another token has the same base as an inflected verb,
// and echoNoun means another token has the same base as an inflected noun.
func (t *Tokeniser) checkInflectionEcho(tokens []Token, idx int) (bool, bool) {
	target := tokens[idx]
	var echoVerb, echoNoun bool

	for i, tok := range tokens {
		if i == idx {
			continue
		}
		// Check if another token is a verb with the same base
		if tok.Type == TokenVerb && tok.VerbInfo.Base == target.VerbInfo.Base && tok.VerbInfo.Tense != "base" {
			echoVerb = true
		}
		// Check if another token is a noun with the same base
		if tok.Type == TokenNoun && tok.NounInfo.Base == target.NounInfo.Base && tok.NounInfo.Plural {
			echoNoun = true
		}
	}

	return echoVerb, echoNoun
}

// resolveToken assigns the final classification to an ambiguous token
// based on verb and noun scores from disambiguation signals.
func (t *Tokeniser) resolveToken(tok *Token, verbScore, nounScore float64, components []SignalComponent) {
	total := verbScore + nounScore

	// B3 review fix: if total < 0.10 (only default prior fired),
	// use low-information confidence floor.
	if total < 0.10 {
		if verbScore >= nounScore {
			tok.Type = TokenVerb
			tok.Confidence = 0.55
			tok.AltType = TokenNoun
			tok.AltConf = 0.45
		} else {
			tok.Type = TokenNoun
			tok.Confidence = 0.55
			tok.AltType = TokenVerb
			tok.AltConf = 0.45
		}
	} else {
		if verbScore >= nounScore {
			tok.Type = TokenVerb
			tok.Confidence = verbScore / total
			tok.AltType = TokenNoun
			tok.AltConf = nounScore / total
		} else {
			tok.Type = TokenNoun
			tok.Confidence = nounScore / total
			tok.AltType = TokenVerb
			tok.AltConf = verbScore / total
		}
	}

	if t.withSignals {
		tok.Signals = &SignalBreakdown{
			VerbScore:  verbScore,
			NounScore:  nounScore,
			Components: components,
		}
	}
}

// splitTrailingPunct separates a word from its trailing punctuation.
// Returns the word and the punctuation suffix. It also recognises
// standalone punctuation tokens such as "." and ")".
func splitTrailingPunct(s string) (string, string) {
	// Standalone punctuation token.
	if _, ok := matchPunctuation(s); ok {
		return "", s
	}

	// Check for "..." suffix first (3-char pattern).
	if core.HasSuffix(s, "...") {
		return s[:len(s)-3], "..."
	}
	// Check single-char trailing punctuation.
	if len(s) > 1 {
		last := s[len(s)-1]
		if last == '?' || last == ':' || last == '!' || last == ';' || last == ',' || last == '.' || last == ')' || last == ']' || last == '}' {
			return s[:len(s)-1], string(last)
		}
	}
	return s, ""
}

func (t *Tokeniser) splitFrenchElision(raw string) (string, string, bool) {
	if !t.isFrenchLanguage() || len(raw) == 0 {
		return "", raw, false
	}

	lower := core.Lower(raw)
	if len(lower) < 2 {
		return "", raw, false
	}

	for _, prefix := range frenchElisionPrefixes {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		idx := len(prefix)
		if idx >= len(raw) {
			continue
		}
		if idx < len(raw) {
			r, size := utf8.DecodeRuneInString(raw[idx:])
			if r != '\'' && r != '’' {
				continue
			}
			if size > 0 {
				return raw[:idx+size], raw[idx+size:], true
			}
		}
	}

	return "", raw, false
}

func (t *Tokeniser) isFrenchLanguage() bool {
	lang := core.Lower(t.lang)
	return lang == "fr" || core.HasPrefix(lang, "fr-")
}

// matchPunctuation detects known punctuation patterns.
// Returns the punctuation type and true if recognised.
func matchPunctuation(punct string) (string, bool) {
	switch punct {
	case "...":
		return "progress", true
	case "?":
		return "question", true
	case "!":
		return "exclamation", true
	case ":":
		return "label", true
	case ";":
		return "separator", true
	case ",":
		return "comma", true
	case ".":
		return "sentence_end", true
	case ")":
		return "close_paren", true
	case "]":
		return "close_bracket", true
	case "}":
		return "close_brace", true
	}
	return "", false
}

// DisambiguationStats provides aggregate statistics about token disambiguation.
type DisambiguationStats struct {
	TotalTokens     int
	AmbiguousTokens int
	ResolvedAsVerb  int
	ResolvedAsNoun  int
	AvgConfidence   float64
	LowConfidence   int // count where confidence < 0.7
}

// DisambiguationStatsFromTokens computes aggregate disambiguation stats from a token slice.
func DisambiguationStatsFromTokens(tokens []Token) DisambiguationStats {
	var s DisambiguationStats
	s.TotalTokens = len(tokens)
	var confSum float64
	var confCount int

	for _, tok := range tokens {
		if tok.AltType != 0 && tok.AltConf > 0 {
			s.AmbiguousTokens++
			if tok.Type == TokenVerb {
				s.ResolvedAsVerb++
			} else if tok.Type == TokenNoun {
				s.ResolvedAsNoun++
			}
		}
		if tok.Type != TokenUnknown && tok.Confidence > 0 {
			confSum += tok.Confidence
			confCount++
			if tok.Confidence < 0.7 {
				s.LowConfidence++
			}
		}
	}

	if confCount > 0 {
		s.AvgConfidence = confSum / float64(confCount)
	}
	return s
}

// DisambiguationStats returns aggregate disambiguation stats for a token slice.
func (t *Tokeniser) DisambiguationStats(tokens []Token) DisambiguationStats {
	return DisambiguationStatsFromTokens(tokens)
}
