package i18n

import (
	// Note: AX-6 — grammar data merge/copy needs map cloning; pinned core has no map-copy primitive.
	"maps"
	// Note: AX-6 — grammar quote helper needs Go string escaping; pinned core has no quote primitive.
	"strconv"
	// Note: AX-6 — grammar interpolation exposes template.FuncMap; pinned core has no template primitive.
	"text/template"
	// Note: AX-6 — grammar casing and initialism rules need Unicode rune classes; pinned core has no rune helpers.
	"unicode"

	"dappco.re/go"
)

// GetGrammarData returns the grammar data for the specified language.
func GetGrammarData(lang string) *GrammarData {
	lang = normalizeLanguageTag(lang)
	if lang == "" {
		return nil
	}
	grammarCacheMu.RLock()
	defer grammarCacheMu.RUnlock()
	if data, ok := grammarCache[lang]; ok && data != nil {
		return data
	}
	if base := baseLanguageTag(lang); base != "" {
		if data, ok := grammarCache[base]; ok && data != nil {
			return data
		}
	}
	return nil
}

// SetGrammarData sets the grammar data for a language, replacing any existing data.
func SetGrammarData(lang string, data *GrammarData) {
	lang = normalizeLanguageTag(lang)
	if lang == "" {
		return
	}
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	if data == nil {
		delete(grammarCache, lang)
		return
	}
	grammarCache[lang] = cloneGrammarData(data)
}

// MergeGrammarData merges grammar data into the existing data for a language.
// New entries are added; existing entries are overwritten per-key.
func MergeGrammarData(lang string, data *GrammarData) {
	lang = normalizeLanguageTag(lang)
	if lang == "" || data == nil {
		return
	}
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	existing := grammarCache[lang]
	if existing == nil {
		grammarCache[lang] = cloneGrammarData(data)
		return
	}
	if existing.Verbs == nil {
		existing.Verbs = make(map[string]VerbForms, len(data.Verbs))
	}
	if existing.Nouns == nil {
		existing.Nouns = make(map[string]NounForms, len(data.Nouns))
	}
	if existing.Words == nil {
		existing.Words = make(map[string]string, len(data.Words))
	}
	if existing.Intents == nil {
		existing.Intents = make(map[string]Intent, len(data.Intents))
	}
	maps.Copy(existing.Verbs, data.Verbs)
	maps.Copy(existing.Nouns, data.Nouns)
	maps.Copy(existing.Words, data.Words)
	mergeArticleForms(&existing.Articles, data.Articles)
	mergePunctuationRules(&existing.Punct, data.Punct)
	mergeSignalData(&existing.Signals, data.Signals)
	existing.Intents = mergeIntentData(existing.Intents, data.Intents)
	if data.Number.ThousandsSep != "" {
		existing.Number.ThousandsSep = data.Number.ThousandsSep
	}
	if data.Number.DecimalSep != "" {
		existing.Number.DecimalSep = data.Number.DecimalSep
	}
	if data.Number.PercentFmt != "" {
		existing.Number.PercentFmt = data.Number.PercentFmt
	}
}

func mergeArticleForms(dst *ArticleForms, src ArticleForms) {
	if dst == nil {
		return
	}
	if src.IndefiniteDefault != "" {
		dst.IndefiniteDefault = src.IndefiniteDefault
	}
	if src.IndefiniteVowel != "" {
		dst.IndefiniteVowel = src.IndefiniteVowel
	}
	if src.Definite != "" {
		dst.Definite = src.Definite
	}
	if src.DefinitePlural != "" {
		dst.DefinitePlural = src.DefinitePlural
	}
	dst.VowelSoundWords = appendUniqueStrings(dst.VowelSoundWords, src.VowelSoundWords...)
	dst.ConsonantSoundWords = appendUniqueStrings(dst.ConsonantSoundWords, src.ConsonantSoundWords...)
	if len(src.IndefiniteByGender) > 0 {
		if dst.IndefiniteByGender == nil {
			dst.IndefiniteByGender = make(map[string]string, len(src.IndefiniteByGender))
		}
		maps.Copy(dst.IndefiniteByGender, src.IndefiniteByGender)
	}
	if len(src.DefinitePluralByGender) > 0 {
		if dst.DefinitePluralByGender == nil {
			dst.DefinitePluralByGender = make(map[string]string, len(src.DefinitePluralByGender))
		}
		maps.Copy(dst.DefinitePluralByGender, src.DefinitePluralByGender)
	}
	if len(src.ByGender) == 0 {
		return
	}
	if dst.ByGender == nil {
		dst.ByGender = make(map[string]string, len(src.ByGender))
	}
	maps.Copy(dst.ByGender, src.ByGender)
}

func mergePunctuationRules(dst *PunctuationRules, src PunctuationRules) {
	if dst == nil {
		return
	}
	if src.LabelSuffix != "" {
		dst.LabelSuffix = src.LabelSuffix
	}
	if src.ProgressSuffix != "" {
		dst.ProgressSuffix = src.ProgressSuffix
	}
}

func mergeSignalData(dst *SignalData, src SignalData) {
	if dst == nil {
		return
	}
	dst.NounDeterminers = appendUniqueStrings(dst.NounDeterminers, src.NounDeterminers...)
	dst.VerbAuxiliaries = appendUniqueStrings(dst.VerbAuxiliaries, src.VerbAuxiliaries...)
	dst.VerbInfinitive = appendUniqueStrings(dst.VerbInfinitive, src.VerbInfinitive...)
	dst.VerbNegation = appendUniqueStrings(dst.VerbNegation, src.VerbNegation...)
	if len(src.Priors) == 0 {
		return
	}
	if dst.Priors == nil {
		dst.Priors = make(map[string]map[string]float64, len(src.Priors))
	}
	for word, priors := range src.Priors {
		if dst.Priors[word] == nil {
			dst.Priors[word] = make(map[string]float64, len(priors))
		}
		maps.Copy(dst.Priors[word], priors)
	}
}

func appendUniqueStrings(dst []string, values ...string) []string {
	if len(values) == 0 {
		return dst
	}
	seen := make(map[string]struct{}, len(dst))
	for _, value := range dst {
		seen[value] = struct{}{}
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		dst = append(dst, value)
	}
	return dst
}

func grammarDataHasContent(data *GrammarData) bool {
	if data == nil {
		return false
	}
	if len(data.Verbs) > 0 || len(data.Nouns) > 0 || len(data.Words) > 0 {
		return true
	}
	if data.Articles.IndefiniteDefault != "" ||
		data.Articles.IndefiniteVowel != "" ||
		data.Articles.Definite != "" ||
		data.Articles.DefinitePlural != "" ||
		len(data.Articles.DefinitePluralByGender) > 0 ||
		len(data.Articles.ByGender) > 0 ||
		len(data.Articles.IndefiniteByGender) > 0 ||
		len(data.Articles.VowelSoundWords) > 0 ||
		len(data.Articles.ConsonantSoundWords) > 0 {
		return true
	}
	if data.Punct.LabelSuffix != "" || data.Punct.ProgressSuffix != "" {
		return true
	}
	if len(data.Signals.NounDeterminers) > 0 ||
		len(data.Signals.VerbAuxiliaries) > 0 ||
		len(data.Signals.VerbInfinitive) > 0 ||
		len(data.Signals.VerbNegation) > 0 ||
		len(data.Signals.Priors) > 0 {
		return true
	}
	if len(data.Intents) > 0 {
		return true
	}
	return data.Number != (NumberFormat{})
}

func cloneGrammarData(data *GrammarData) *GrammarData {
	if data == nil {
		return nil
	}
	clone := &GrammarData{
		Articles: ArticleForms{
			IndefiniteDefault:   data.Articles.IndefiniteDefault,
			IndefiniteVowel:     data.Articles.IndefiniteVowel,
			Definite:            data.Articles.Definite,
			DefinitePlural:      data.Articles.DefinitePlural,
			VowelSoundWords:     append([]string(nil), data.Articles.VowelSoundWords...),
			ConsonantSoundWords: append([]string(nil), data.Articles.ConsonantSoundWords...),
		},
		Punct: data.Punct,
		Signals: SignalData{
			NounDeterminers: append([]string(nil), data.Signals.NounDeterminers...),
			VerbAuxiliaries: append([]string(nil), data.Signals.VerbAuxiliaries...),
			VerbInfinitive:  append([]string(nil), data.Signals.VerbInfinitive...),
			VerbNegation:    append([]string(nil), data.Signals.VerbNegation...),
			Priors:          make(map[string]map[string]float64, len(data.Signals.Priors)),
		},
		Number: data.Number,
	}
	if len(data.Verbs) > 0 {
		clone.Verbs = make(map[string]VerbForms, len(data.Verbs))
		maps.Copy(clone.Verbs, data.Verbs)
	}
	if len(data.Nouns) > 0 {
		clone.Nouns = make(map[string]NounForms, len(data.Nouns))
		maps.Copy(clone.Nouns, data.Nouns)
	}
	if len(data.Words) > 0 {
		clone.Words = make(map[string]string, len(data.Words))
		maps.Copy(clone.Words, data.Words)
	}
	if len(data.Articles.ByGender) > 0 {
		clone.Articles.ByGender = make(map[string]string, len(data.Articles.ByGender))
		maps.Copy(clone.Articles.ByGender, data.Articles.ByGender)
	}
	if len(data.Articles.IndefiniteByGender) > 0 {
		clone.Articles.IndefiniteByGender = make(map[string]string, len(data.Articles.IndefiniteByGender))
		maps.Copy(clone.Articles.IndefiniteByGender, data.Articles.IndefiniteByGender)
	}
	if len(data.Articles.DefinitePluralByGender) > 0 {
		clone.Articles.DefinitePluralByGender = make(map[string]string, len(data.Articles.DefinitePluralByGender))
		maps.Copy(clone.Articles.DefinitePluralByGender, data.Articles.DefinitePluralByGender)
	}
	if len(data.Signals.Priors) > 0 {
		for word, priors := range data.Signals.Priors {
			if len(priors) == 0 {
				continue
			}
			clone.Signals.Priors[word] = make(map[string]float64, len(priors))
			maps.Copy(clone.Signals.Priors[word], priors)
		}
	}
	if len(data.Intents) > 0 {
		clone.Intents = cloneIntentMap(data.Intents)
	}
	return clone
}

func mergeIntentData(dst map[string]Intent, src map[string]Intent) map[string]Intent {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = make(map[string]Intent, len(src))
	}
	for key, intent := range src {
		dst[key] = cloneIntent(intent)
	}
	return dst
}

func cloneIntentMap(src map[string]Intent) map[string]Intent {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]Intent, len(src))
	for key, intent := range src {
		dst[key] = cloneIntent(intent)
	}
	return dst
}

func cloneIntent(intent Intent) Intent {
	if len(intent.Meta.Supports) > 0 {
		intent.Meta.Supports = append([]string(nil), intent.Meta.Supports...)
	}
	return intent
}

// IrregularVerbs returns a copy of the irregular verb forms map.
func IrregularVerbs() map[string]VerbForms {
	result := make(map[string]VerbForms, len(irregularVerbs))
	maps.Copy(result, irregularVerbs)
	return result
}

// IrregularNouns returns a copy of the irregular nouns map.
func IrregularNouns() map[string]string {
	result := make(map[string]string, len(irregularNouns))
	maps.Copy(result, irregularNouns)
	return result
}

// DualClassVerbs returns a copy of the additional regular verbs that also act
// as common nouns in dev/ops text.
func DualClassVerbs() map[string]VerbForms {
	result := make(map[string]VerbForms, len(dualClassVerbs))
	maps.Copy(result, dualClassVerbs)
	return result
}

// DualClassNouns returns a copy of the additional regular nouns that also act
// as common verbs in dev/ops text.
func DualClassNouns() map[string]string {
	result := make(map[string]string, len(dualClassNouns))
	maps.Copy(result, dualClassNouns)
	return result
}

// Lower returns the lowercase form of s.
func Lower(s string) string {
	return core.Lower(s)
}

// Upper returns the uppercase form of s.
func Upper(s string) string {
	return core.Upper(s)
}

func getVerbForm(lang, verb, form string) string {
	data := grammarDataForLang(lang)
	if data == nil || data.Verbs == nil {
		return ""
	}
	verb = core.Lower(verb)
	if forms, ok := data.Verbs[verb]; ok {
		switch form {
		case "past":
			return forms.Past
		case "gerund":
			return forms.Gerund
		}
	}
	return ""
}

func getWord(lang, word string) string {
	data := grammarDataForLang(lang)
	if data == nil || data.Words == nil {
		return ""
	}
	return data.Words[core.Lower(word)]
}

func getPunct(lang, rule, defaultVal string) string {
	data := grammarDataForLang(lang)
	if data == nil {
		return defaultVal
	}
	switch rule {
	case "label":
		if data.Punct.LabelSuffix != "" {
			return data.Punct.LabelSuffix
		}
	case "progress":
		if data.Punct.ProgressSuffix != "" {
			return data.Punct.ProgressSuffix
		}
	}
	return defaultVal
}

func getNounForm(lang, noun, form string) string {
	data := grammarDataForLang(lang)
	if data == nil || data.Nouns == nil {
		return ""
	}
	noun = core.Lower(noun)
	if forms, ok := data.Nouns[noun]; ok {
		switch form {
		case "one":
			return forms.One
		case "other":
			return forms.Other
		case "gender":
			return forms.Gender
		}
	}
	return ""
}

func currentLangForGrammar() string {
	if svc := Default(); svc != nil {
		return svc.Language()
	}
	return "en"
}

// PastTense returns the past tense of a verb.
// 3-tier fallback: JSON locale -> irregular verbs -> regular rules.
//
//	PastTense("delete") // "deleted"
//	PastTense("run")    // "ran"
//	PastTense("copy")   // "copied"
func PastTense(verb string) string {
	return pastTenseForLanguages([]string{currentLangForGrammar()}, verb)
}

func pastTenseForLanguages(langs []string, verb string) string {
	verb = core.Lower(core.Trim(verb))
	if verb == "" {
		return ""
	}
	for _, lang := range languageFallbackOrder(langs) {
		if form := getVerbForm(lang, verb, "past"); form != "" {
			return form
		}
	}
	if forms, ok := irregularVerbs[verb]; ok {
		return forms.Past
	}
	return applyRegularPastTense(verb)
}

func applyRegularPastTense(verb string) string {
	if core.HasSuffix(verb, "ed") && len(verb) > 2 {
		thirdFromEnd := verb[len(verb)-3]
		if !isVowel(rune(thirdFromEnd)) && thirdFromEnd != 'e' {
			return verb
		}
	}
	if core.HasSuffix(verb, "e") {
		return verb + "d"
	}
	if core.HasSuffix(verb, "y") && len(verb) > 1 {
		prev := rune(verb[len(verb)-2])
		if !isVowel(prev) {
			return verb[:len(verb)-1] + "ied"
		}
	}
	// The -c → -cked insertion protects a hard /k/ between vowels (panic →
	// panicked, mimic → mimicked). After a consonant the c needs no
	// protection: sync → synced, arc → arced.
	if core.HasSuffix(verb, "c") && len(verb) > 1 && isVowel(rune(verb[len(verb)-2])) {
		return verb + "ked"
	}
	if len(verb) >= 2 && shouldDoubleConsonant(verb) {
		return verb + string(verb[len(verb)-1]) + "ed"
	}
	return verb + "ed"
}

func shouldDoubleConsonant(verb string) bool {
	if len(verb) < 3 {
		return false
	}
	if noDoubleConsonant[verb] {
		return false
	}
	lastChar := rune(verb[len(verb)-1])
	if isVowel(lastChar) || lastChar == 'w' || lastChar == 'x' || lastChar == 'y' {
		return false
	}
	if !isVowelLetterAt(verb, len(verb)-2) {
		return false
	}
	// en-GB doubles a final single -l after a single vowel regardless of
	// stress: marshalled, signalled, totalled, equalled. en-US single-l
	// forms override per-verb via locale data (locales/en-US.json).
	if lastChar == 'l' && !isVowelLetterAt(verb, len(verb)-3) {
		return true
	}
	if len(verb) <= 4 {
		return !isVowelLetterAt(verb, len(verb)-3)
	}
	return false
}

// isVowelLetterAt reports whether the byte at index i sounds as a vowel in
// the CVC sense. A "u" straight after "q" is part of the /kw/ onset cluster,
// not a vowel — that is why quiz doubles (quizzed) and equal doubles
// (equalled) even though their spellings read vowel-vowel.
func isVowelLetterAt(word string, i int) bool {
	if i < 0 || i >= len(word) {
		return false
	}
	r := rune(word[i])
	if !isVowel(r) {
		return false
	}
	if r == 'u' && i > 0 && word[i-1] == 'q' {
		return false
	}
	return true
}

// Gerund returns the present participle (-ing form) of a verb.
//
//	Gerund("delete")  // "deleting"
//	Gerund("run")     // "running"
//	Gerund("die")     // "dying"
func Gerund(verb string) string {
	verb = core.Lower(core.Trim(verb))
	if verb == "" {
		return ""
	}
	if form := getVerbForm(currentLangForGrammar(), verb, "gerund"); form != "" {
		return form
	}
	if forms, ok := irregularVerbs[verb]; ok {
		return forms.Gerund
	}
	return applyRegularGerund(verb)
}

func applyRegularGerund(verb string) string {
	if core.HasSuffix(verb, "ie") {
		return verb[:len(verb)-2] + "ying"
	}
	if core.HasSuffix(verb, "e") && len(verb) > 1 {
		secondLast := rune(verb[len(verb)-2])
		if secondLast != 'e' && secondLast != 'y' && secondLast != 'o' {
			return verb[:len(verb)-1] + "ing"
		}
	}
	// Same vowel gate as the past tense: panicking but syncing.
	if core.HasSuffix(verb, "c") && len(verb) > 1 && isVowel(rune(verb[len(verb)-2])) {
		return verb + "king"
	}
	if shouldDoubleConsonant(verb) {
		return verb + string(verb[len(verb)-1]) + "ing"
	}
	return verb + "ing"
}

// Pluralize returns the plural form of a noun based on count.
//
//	Pluralize("file", 1)    // "file"
//	Pluralize("file", 5)    // "files"
//	Pluralize("child", 3)   // "children"
func Pluralize(noun string, count int) string {
	if count == 1 {
		// Honour locale-provided singular forms before falling back to the input.
		noun = core.Trim(noun)
		if noun == "" {
			return ""
		}
		lower := core.Lower(noun)
		if form := getNounForm(currentLangForGrammar(), lower, "one"); form != "" {
			return preserveInitialCapitalization(noun, form)
		}
		return noun
	}
	return PluralForm(noun)
}

// PluralForm returns the plural form of a noun.
func PluralForm(noun string) string {
	noun = core.Trim(noun)
	if noun == "" {
		return ""
	}
	lower := core.Lower(noun)
	if form := getNounForm(currentLangForGrammar(), lower, "other"); form != "" {
		return preserveInitialCapitalization(noun, form)
	}
	if plural, ok := irregularNouns[lower]; ok {
		return preserveInitialCapitalization(noun, plural)
	}
	return applyRegularPlural(noun)
}

func applyRegularPlural(noun string) string {
	lower := core.Lower(noun)
	// Greek -sis pluralises -ses across the whole class: basis → bases,
	// synopsis → synopses, emphasis → emphases. Checked before the sibilant
	// branch, which would otherwise produce "basises". (iris/bias end -is
	// but not -sis, so they keep the regular -es path.)
	if core.HasSuffix(lower, "sis") && len(noun) > 3 {
		return noun[:len(noun)-2] + "es"
	}
	// A single final z after a vowel doubles before -es: quizzes, whizzes.
	// After a consonant or an existing zz it takes plain -es: waltzes, buzzes.
	if core.HasSuffix(lower, "z") && !core.HasSuffix(lower, "zz") &&
		len(lower) > 1 && isVowel(rune(lower[len(lower)-2])) {
		return noun + string(noun[len(noun)-1]) + "es"
	}
	if core.HasSuffix(lower, "s") ||
		core.HasSuffix(lower, "ss") ||
		core.HasSuffix(lower, "sh") ||
		core.HasSuffix(lower, "ch") ||
		core.HasSuffix(lower, "x") ||
		core.HasSuffix(lower, "z") {
		return noun + "es"
	}
	if core.HasSuffix(lower, "y") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			return noun[:len(noun)-1] + "ies"
		}
	}
	// No f/fe → ves branch: that plural is a closed Old English class carried
	// by irregularNouns (wolf, knife, elf...). Productive English takes -s
	// (roofs, chiefs, safes, cliffs, chefs, beliefs).
	if core.HasSuffix(lower, "o") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			if lower == "hero" || lower == "potato" || lower == "tomato" || lower == "echo" || lower == "veto" {
				return noun + "es"
			}
		}
	}
	return noun + "s"
}

// Article returns the appropriate article token for the current language.
// English falls back to phonetic "a"/"an" heuristics. Locale grammar data
// can override this with language-specific article forms.
//
// Use ArticlePhrase when you want the noun phrase prefixed with the article.
//
//	Article("file")     // "a"
//	Article("error")    // "an"
//	Article("user")     // "a" (sounds like "yoo-zer")
//	Article("hour")     // "an" (silent h)
func Article(word string) string {
	word = core.Trim(word)
	if word == "" {
		return ""
	}
	lower := core.Lower(word)
	if article, ok := articleForCurrentLanguage(lower, word); ok {
		return article
	}
	if isInitialism(word) {
		if initialismUsesVowelSound(word) {
			return "an"
		}
		return "a"
	}
	// Locale sound-word lists still apply when the locale declares no
	// indefinite forms of its own (a pure phonetic-delta file).
	if vowel, ok := articleSoundOverride(grammarDataForLang(currentLangForGrammar()), lower); ok {
		if vowel {
			return "an"
		}
		return "a"
	}
	for key := range consonantSounds {
		if core.HasPrefix(lower, key) {
			return "a"
		}
	}
	for key := range vowelSounds {
		if core.HasPrefix(lower, key) {
			return "an"
		}
	}
	if len(lower) > 0 && isVowel(rune(lower[0])) {
		return "an"
	}
	return "a"
}

// ArticleToken is an explicit alias for Article.
func ArticleToken(word string) string {
	return Article(word)
}

func articleForCurrentLanguage(lowerWord, originalWord string) (string, bool) {
	lang := currentLangForGrammar()
	data := grammarDataForLang(lang)
	if data == nil {
		return "", false
	}

	if article, ok := articleForPluralForm(data, lowerWord, lang); ok {
		return article, true
	}
	if article, ok := articleForFrenchPluralGuess(data, lowerWord, originalWord, lang); ok {
		return article, true
	}
	// True gendered indefinites (un/une, ein/eine) outrank the gendered
	// definite fallback so Article() stays an indefinite everywhere once a
	// locale declares them.
	if article, ok := indefiniteArticleByGender(data, lowerWord); ok {
		return article, true
	}
	if article, ok := articleByGender(data, lowerWord, originalWord, lang); ok {
		return article, true
	}
	if article, ok := articleFromGrammarForms(data, originalWord); ok {
		return article, true
	}
	return "", false
}

// indefiniteArticleByGender resolves the gendered indefinite article for a
// known noun: un/une in French, ein/eine/ein in German. Indefinite articles
// do not elide, so the word passes through untouched.
//
//	indefiniteArticleByGender(frData, "branche") // ("une", true)
func indefiniteArticleByGender(data *GrammarData, lowerWord string) (string, bool) {
	if len(data.Articles.IndefiniteByGender) == 0 {
		return "", false
	}
	forms, ok := data.Nouns[lowerWord]
	if !ok || forms.Gender == "" {
		return "", false
	}
	article, ok := data.Articles.IndefiniteByGender[forms.Gender]
	if !ok || article == "" {
		return "", false
	}
	return article, true
}

func articleByGender(data *GrammarData, lowerWord, originalWord, lang string) (string, bool) {
	if len(data.Articles.ByGender) == 0 {
		return "", false
	}
	forms, ok := data.Nouns[lowerWord]
	if !ok || forms.Gender == "" {
		return "", false
	}
	article, ok := data.Articles.ByGender[forms.Gender]
	if !ok || article == "" {
		return "", false
	}
	return maybeElideArticle(article, originalWord, lang), true
}

func articleForPluralForm(data *GrammarData, lowerWord, lang string) (string, bool) {
	gender, ok := pluralNounGender(data, lowerWord)
	if !ok {
		return "", false
	}
	// Known plurals take the locale's definite plural article — gendered
	// where the language demands it (los/las), a single form otherwise
	// (les, die). The French literal remains as a fallback for data
	// without either field.
	if article, ok := data.Articles.DefinitePluralByGender[gender]; ok && article != "" {
		return article, true
	}
	if data.Articles.DefinitePlural != "" {
		return data.Articles.DefinitePlural, true
	}
	if isFrenchLanguage(lang) {
		return "les", true
	}
	return "", false
}

func articleForFrenchPluralGuess(data *GrammarData, lowerWord, originalWord, lang string) (string, bool) {
	if !isFrenchLanguage(lang) {
		return "", false
	}
	if isKnownPluralNoun(data, lowerWord) {
		return "", false
	}
	if !looksLikeFrenchPlural(originalWord) {
		return "", false
	}
	return "des", true
}

func isKnownPluralNoun(data *GrammarData, lowerWord string) bool {
	_, ok := pluralNounGender(data, lowerWord)
	return ok
}

// pluralNounGender reports whether the word is a known plural form and, when
// it is, the gender of the noun it pluralises — Spanish needs this because
// the plural definite article is itself gendered (los archivos, las tareas).
func pluralNounGender(data *GrammarData, lowerWord string) (string, bool) {
	if data == nil || len(data.Nouns) == 0 {
		return "", false
	}
	for _, forms := range data.Nouns {
		if forms.Other == "" || core.Lower(forms.Other) != lowerWord {
			continue
		}
		if forms.One != "" && core.Lower(forms.One) == lowerWord {
			continue
		}
		return forms.Gender, true
	}
	return "", false
}

func articleFromGrammarForms(data *GrammarData, word string) (string, bool) {
	if data.Articles.IndefiniteDefault == "" && data.Articles.IndefiniteVowel == "" {
		return "", false
	}
	if usesVowelSoundArticle(data, word) && data.Articles.IndefiniteVowel != "" {
		return data.Articles.IndefiniteVowel, true
	}
	if data.Articles.IndefiniteDefault != "" {
		return data.Articles.IndefiniteDefault, true
	}
	if data.Articles.IndefiniteVowel != "" {
		return data.Articles.IndefiniteVowel, true
	}
	return "", false
}

func maybeElideArticle(article, word, lang string) string {
	if !isFrenchLanguage(lang) {
		return article
	}
	if !startsWithVowelSound(word) {
		return article
	}
	switch core.Lower(article) {
	case "le", "la", "de", "je", "me", "te", "se", "ne", "ce":
		// French elision keeps the leading consonant and replaces the final
		// vowel with an apostrophe: le/la -> l', de -> d', je -> j', etc.
		return core.Lower(article[:1]) + "'"
	case "que":
		return "qu'"
	}
	return article
}

func usesVowelSoundArticle(data *GrammarData, word string) bool {
	trimmed := core.Trim(word)
	if trimmed == "" {
		return false
	}
	if isInitialism(trimmed) {
		return initialismUsesVowelSound(trimmed)
	}
	lower := core.Lower(trimmed)
	// Locale-declared phonetic exceptions outrank the built-in tables so a
	// dialect file can re-hear a word: en-US declares "herb" vowel-sound
	// (silent h) while base en keeps the British /h/.
	if vowel, ok := articleSoundOverride(data, lower); ok {
		return vowel
	}
	for key := range consonantSounds {
		if core.HasPrefix(lower, key) {
			return false
		}
	}
	for key := range vowelSounds {
		if core.HasPrefix(lower, key) {
			return true
		}
	}
	for _, r := range lower {
		return isVowel(r)
	}
	return false
}

// articleSoundOverride matches a lowercased word against the locale's
// declared sound-word prefixes. The second return reports whether any
// locale entry decided the word.
//
//	articleSoundOverride(usData, "herb") // (true, true) under en-US
func articleSoundOverride(data *GrammarData, lower string) (bool, bool) {
	if data == nil {
		return false, false
	}
	for _, key := range data.Articles.ConsonantSoundWords {
		if core.HasPrefix(lower, key) {
			return false, true
		}
	}
	for _, key := range data.Articles.VowelSoundWords {
		if core.HasPrefix(lower, key) {
			return true, true
		}
	}
	return false, false
}

func looksLikeFrenchPlural(word string) bool {
	trimmed := core.Trim(word)
	if trimmed == "" || core.Contains(trimmed, " ") || core.Contains(trimmed, "\t") || isInitialism(trimmed) {
		return false
	}
	lower := core.Lower(trimmed)
	if isFrenchAspiratedHWord(lower) {
		return false
	}
	if core.HasSuffix(lower, "aux") || core.HasSuffix(lower, "eaux") {
		return true
	}
	return core.HasSuffix(lower, "s") || core.HasSuffix(lower, "x")
}

func startsWithVowelSound(word string) bool {
	trimmed := core.Trim(word)
	lower := core.Lower(trimmed)
	if lower == "" {
		return false
	}
	if isFrenchAspiratedHWord(lower) {
		return false
	}
	r := []rune(lower)
	switch r[0] {
	case 'a', 'e', 'i', 'o', 'u', 'y',
		'à', 'â', 'ä', 'æ', 'é', 'è', 'ê', 'ë',
		'î', 'ï', 'ô', 'ö', 'ù', 'û', 'ü', 'œ', 'h':
		return true
	}
	return false
}

func isFrenchAspiratedHWord(word string) bool {
	switch word {
	case "haricot", "héron", "héros", "honte", "hache", "hasard", "hibou", "houx", "hurluberlu":
		return true
	default:
		return false
	}
}

func isFrenchLanguage(lang string) bool {
	lang = core.Lower(lang)
	return lang == "fr" || core.HasPrefix(lang, "fr-")
}

func isInitialism(word string) bool {
	if len(word) < 2 {
		return false
	}
	hasLetter := false
	for _, r := range word {
		if !unicode.IsLetter(r) {
			return false
		}
		hasLetter = true
		if unicode.IsLower(r) {
			return false
		}
	}
	return hasLetter
}

func preserveInitialCapitalization(original, form string) string {
	if original == "" || form == "" {
		return form
	}
	originalRunes := []rune(original)
	formRunes := []rune(form)
	if len(originalRunes) == 0 || len(formRunes) == 0 {
		return form
	}
	if !unicode.IsUpper(originalRunes[0]) {
		return form
	}
	formRunes[0] = unicode.ToUpper(formRunes[0])
	return string(formRunes)
}

func initialismUsesVowelSound(word string) bool {
	if word == "" {
		return false
	}
	switch unicode.ToUpper([]rune(word)[0]) {
	case 'A', 'E', 'F', 'H', 'I', 'L', 'M', 'N', 'O', 'R', 'S', 'X':
		return true
	default:
		return false
	}
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// Title capitalises the first letter of each word-like segment.
//
// Hyphens and whitespace start a new segment; punctuation inside identifiers
// such as dots and underscores is preserved so filenames stay readable.
func Title(s string) string {
	b := core.NewBuilder()
	b.Grow(len(s))
	capNext := true
	for _, r := range s {
		if unicode.IsLetter(r) && capNext {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(r)
		}
		switch r {
		case ' ', '\t', '\n', '\r', '-':
			capNext = true
		default:
			capNext = false
		}
	}
	return b.String()
}

func renderWord(lang, word string) string {
	if translated := getWord(lang, word); translated != "" {
		return translated
	}
	return word
}

// renderWordOrTitle keeps a mapped word's AUTHORED case verbatim — the word
// table carries display forms (go.mod, URL, pnpm) whose casing is the point.
// Only unmapped words are title-cased.
func renderWordOrTitle(lang, word string) string {
	if translated := getWord(lang, word); translated != "" {
		return translated
	}
	return Title(word)
}

// Quote wraps a string in double quotes.
func Quote(s string) string {
	return strconv.Quote(s)
}

// ArticlePhrase prefixes a noun phrase with the correct article.
func ArticlePhrase(word string) string {
	word = core.Trim(word)
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	word = renderWord(lang, word)
	article := Article(word)
	return prefixWithArticle(article, word)
}

// DefiniteArticle returns the language-specific definite article token for a word.
// For languages such as French, this respects gendered articles, plural forms,
// and elision rules when grammar data is available.
func DefiniteArticle(word string) string {
	word = core.Trim(word)
	if word == "" {
		return ""
	}
	lower := core.Lower(word)
	if article, ok := definiteArticleForCurrentLanguage(lower, word); ok {
		return article
	}
	lang := currentLangForGrammar()
	data := grammarDataForLang(lang)
	if data != nil && data.Articles.Definite != "" {
		return data.Articles.Definite
	}
	return "the"
}

// DefiniteToken is an explicit alias for DefiniteArticle.
func DefiniteToken(word string) string {
	return DefiniteArticle(word)
}

// DefinitePhrase prefixes a noun phrase with the correct definite article.
func DefinitePhrase(word string) string {
	word = core.Trim(word)
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	word = renderWord(lang, word)
	article := DefiniteArticle(word)
	return prefixWithArticle(article, word)
}

func definiteArticleForCurrentLanguage(lowerWord, originalWord string) (string, bool) {
	lang := currentLangForGrammar()
	data := grammarDataForLang(lang)
	if data == nil {
		return "", false
	}
	if article, ok := articleByGender(data, lowerWord, originalWord, lang); ok {
		return article, true
	}
	if article, ok := definiteArticleFromGrammarForms(data, lowerWord, originalWord, lang); ok {
		return article, true
	}
	return "", false
}

func grammarDataForLang(lang string) *GrammarData {
	if data := GetGrammarData(lang); data != nil {
		return data
	}
	if base := baseLanguageTag(lang); base != "" {
		return GetGrammarData(base)
	}
	return nil
}

func languageFallbackOrder(langs []string) []string {
	// Build cache key from input. len==1 is the dominant case — skip
	// the Join allocation entirely there. len==0 maps to "" which
	// caches the universal "en" fallback.
	var key string
	switch len(langs) {
	case 0:
		key = ""
	case 1:
		key = langs[0]
	default:
		key = core.Join("|", langs...)
	}
	if cached, ok := fallbackOrderCache.Load(key); ok {
		return cached.([]string)
	}

	ordered := make([]string, 0, len(langs)*2)
	seen := make(map[string]struct{}, len(langs)*2)
	var add func(string)
	add = func(lang string) {
		lang = normalizeLanguageTag(lang)
		if lang == "" {
			return
		}
		if _, ok := seen[lang]; ok {
			return
		}
		seen[lang] = struct{}{}
		ordered = append(ordered, lang)
		if base := baseLanguageTag(lang); base != "" && base != lang {
			add(base)
		}
	}
	for _, lang := range langs {
		add(lang)
	}
	if len(ordered) == 0 {
		ordered = append(ordered, "en")
	}
	fallbackOrderCache.Store(key, ordered)
	return ordered
}

func baseLanguageTag(lang string) string {
	if idx := indexAny(lang, "-_"); idx > 0 {
		return lang[:idx]
	}
	return ""
}

func definiteArticleFromGrammarForms(data *GrammarData, lowerWord, originalWord, lang string) (string, bool) {
	if data == nil || data.Articles.Definite == "" {
		return "", false
	}
	if gender, ok := pluralNounGender(data, lowerWord); ok {
		if article, ok := data.Articles.DefinitePluralByGender[gender]; ok && article != "" {
			return article, true
		}
		if data.Articles.DefinitePlural != "" {
			return data.Articles.DefinitePlural, true
		}
	}
	if isFrenchLanguage(lang) {
		if isKnownPluralNoun(data, lowerWord) || looksLikeFrenchPlural(originalWord) {
			return "les", true
		}
		return maybeElideArticle(data.Articles.Definite, originalWord, lang), true
	}
	return data.Articles.Definite, true
}

// TemplateFuncs returns the template.FuncMap with all grammar functions.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"title":           Title,
		"lower":           Lower,
		"upper":           Upper,
		"n":               N,
		"number":          Number,
		"int":             Number,
		"decimal":         Decimal,
		"float":           Decimal,
		"percent":         Percent,
		"pct":             Percent,
		byteUnitName():    Bytes,
		"size":            Bytes,
		"ordinal":         Ordinal,
		"ord":             Ordinal,
		"ago":             Ago,
		"past":            PastTense,
		"gerund":          Gerund,
		"plural":          Pluralize,
		"pluralForm":      PluralForm,
		"article":         ArticlePhrase,
		"articleToken":    ArticleToken,
		"articlePhrase":   ArticlePhrase,
		"definiteArticle": DefiniteArticle,
		"definiteToken":   DefiniteToken,
		"definite":        DefinitePhrase,
		"definitePhrase":  DefinitePhrase,
		"quote":           Quote,
		"label":           Label,
		"progress":        Progress,
		"progressSubject": ProgressSubject,
		"actionResult":    ActionResult,
		"actionFailed":    ActionFailed,
		"prompt":          Prompt,
		"lang":            Lang,
		"timeAgo":         TimeAgo,
		"formatAgo":       FormatAgo,
	}
}

// Number formats an integer using the current locale's number rules.
func Number(value any) string {
	return FormatNumber(toInt64(value))
}

// Decimal formats a decimal using the current locale's number rules.
func Decimal(value any) string {
	return FormatDecimal(toFloat64(value))
}

// Percent formats a percentage using the current locale's number rules.
func Percent(value any) string {
	return FormatPercent(toFloat64(value))
}

// Bytes formats a byte count using the current locale's number rules.
func Bytes(value any) string {
	return FormatBytes(toInt64(value))
}

// Ordinal formats a number as an ordinal using the current locale.
func Ordinal(value any) string {
	return FormatOrdinal(toInt(value))
}

// Ago formats a relative time using the current locale's ago rules.
func Ago(count int, unit string) string {
	return FormatAgo(count, unit)
}

func prefixWithArticle(article, word string) string {
	if article == "" || word == "" {
		return ""
	}
	if core.HasSuffix(article, "'") {
		return article + word
	}
	return article + " " + word
}

// Progress returns a progress message: "Building..."
func Progress(verb string) string {
	lang := currentLangForGrammar()
	word := renderWord(lang, verb)
	g := Gerund(word)
	if g == "" {
		return ""
	}
	suffix := getPunct(lang, "progress", "...")
	return Title(g) + suffix
}

// ProgressSubject returns a progress message with subject: "Building project..."
func ProgressSubject(verb, subject string) string {
	lang := currentLangForGrammar()
	word := renderWord(lang, verb)
	g := Gerund(word)
	if g == "" {
		return ""
	}
	suffix := getPunct(lang, "progress", "...")
	subject = core.Trim(subject)
	if subject == "" {
		return Title(g) + suffix
	}
	return Title(g) + " " + renderWord(lang, subject) + suffix
}

// ActionResult returns a completion message: "File deleted"
func ActionResult(verb, subject string) string {
	return actionResultForLanguages([]string{currentLangForGrammar()}, verb, subject)
}

func actionResultForLanguages(langs []string, verb, subject string) string {
	// Map the verb through the gram.word bridge first, exactly as Progress
	// and ActionFailed do — "delete" resolves to the locale's own verb
	// (supprimer, löschen) and is then conjugated in that locale's table.
	verb = renderWordForLanguages(langs, core.Lower(core.Trim(verb)))
	p := pastTenseForLanguages(langs, verb)
	if p == "" {
		return ""
	}
	subject = core.Trim(subject)
	if subject == "" {
		return Title(p)
	}
	return renderSubjectLeadForLanguages(langs, subject) + " " + p
}

// renderSubjectLeadForLanguages resolves a sentence-leading subject. The
// noun table outranks the word bridge so a verb-biased dual key ("run" →
// ausführen) still renders as its noun (Lauf); noun forms are title-cased
// for the lead position (fichier → "Fichier supprimé"). Word-bridge values
// keep their authored case (go.mod, URL).
func renderSubjectLeadForLanguages(langs []string, word string) string {
	if one, ok := nounOneForLanguages(langs, word); ok {
		return Title(one)
	}
	return renderWordOrTitleForLanguages(langs, word)
}

// renderSubjectTailForLanguages resolves a mid-sentence subject: the same
// noun-first lookup, without the lead capitalisation ("Impossible de
// supprimer fichier"; German nouns are authored capitalised regardless).
func renderSubjectTailForLanguages(langs []string, word string) string {
	if one, ok := nounOneForLanguages(langs, word); ok {
		return one
	}
	return renderWordForLanguages(langs, word)
}

func nounOneForLanguages(langs []string, word string) (string, bool) {
	lower := core.Lower(core.Trim(word))
	if lower == "" {
		return "", false
	}
	for _, lang := range languageFallbackOrder(langs) {
		if one := getNounForm(lang, lower, "one"); one != "" {
			return one, true
		}
	}
	return "", false
}

// ActionFailed returns a failure message: "Failed to delete file"
func ActionFailed(verb, subject string) string {
	return actionFailedForLanguages([]string{currentLangForGrammar()}, verb, subject)
}

func actionFailedForLanguages(langs []string, verb, subject string) string {
	verb = core.Trim(verb)
	if verb == "" {
		return ""
	}
	// Keep the failure verb in sentence case when no locale override exists.
	verb = renderWordForLanguages(langs, core.Lower(verb))
	prefix := failedPrefixForLanguages(langs)
	subject = core.Trim(subject)
	if subject == "" {
		return prefix + " " + verb
	}
	return prefix + " " + verb + " " + renderSubjectTailForLanguages(langs, subject)
}

func failedPrefix(lang string) string {
	return failedPrefixForLanguages([]string{lang})
}

func failedPrefixForLanguages(langs []string) string {
	for _, lang := range languageFallbackOrder(langs) {
		prefix := renderWord(lang, "failed_to")
		if prefix != "" && prefix != "failed_to" {
			return prefix
		}
	}
	return "Failed to"
}

func renderWordForLanguages(langs []string, word string) string {
	for _, lang := range languageFallbackOrder(langs) {
		if translated := getWord(lang, word); translated != "" {
			return translated
		}
	}
	return word
}

func renderWordOrTitleForLanguages(langs []string, word string) string {
	for _, lang := range languageFallbackOrder(langs) {
		if translated := getWord(lang, word); translated != "" {
			return translated
		}
	}
	return Title(word)
}

// Label returns a label with suffix: "Status:" (EN) or "Statut :" (FR)
func Label(word string) string {
	word = core.Trim(word)
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	translated := renderWordOrTitle(lang, word)
	suffix := getPunct(lang, "label", ":")
	return translated + suffix
}
