package i18n

import (
	"maps"
	"strings"
	"text/template"
	"unicode"

	"dappco.re/go/core"
)

// GetGrammarData returns the grammar data for the specified language.
func GetGrammarData(lang string) *GrammarData {
	grammarCacheMu.RLock()
	defer grammarCacheMu.RUnlock()
	return grammarCache[lang]
}

// SetGrammarData sets the grammar data for a language, replacing any existing data.
func SetGrammarData(lang string, data *GrammarData) {
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	grammarCache[lang] = data
}

// MergeGrammarData merges grammar data into the existing data for a language.
// New entries are added; existing entries are overwritten per-key.
func MergeGrammarData(lang string, data *GrammarData) {
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	existing := grammarCache[lang]
	if existing == nil {
		grammarCache[lang] = data
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
	maps.Copy(existing.Verbs, data.Verbs)
	maps.Copy(existing.Nouns, data.Nouns)
	maps.Copy(existing.Words, data.Words)
	mergeArticleForms(&existing.Articles, data.Articles)
	mergePunctuationRules(&existing.Punct, data.Punct)
	mergeSignalData(&existing.Signals, data.Signals)
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
	if len(src.NounDeterminers) > 0 {
		dst.NounDeterminers = append(dst.NounDeterminers, src.NounDeterminers...)
	}
	if len(src.VerbAuxiliaries) > 0 {
		dst.VerbAuxiliaries = append(dst.VerbAuxiliaries, src.VerbAuxiliaries...)
	}
	if len(src.VerbInfinitive) > 0 {
		dst.VerbInfinitive = append(dst.VerbInfinitive, src.VerbInfinitive...)
	}
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
		len(data.Articles.ByGender) > 0 {
		return true
	}
	if data.Punct.LabelSuffix != "" || data.Punct.ProgressSuffix != "" {
		return true
	}
	if len(data.Signals.NounDeterminers) > 0 ||
		len(data.Signals.VerbAuxiliaries) > 0 ||
		len(data.Signals.VerbInfinitive) > 0 ||
		len(data.Signals.Priors) > 0 {
		return true
	}
	return data.Number != (NumberFormat{})
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

func getVerbForm(lang, verb, form string) string {
	data := GetGrammarData(lang)
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
	data := GetGrammarData(lang)
	if data == nil || data.Words == nil {
		return ""
	}
	return data.Words[core.Lower(word)]
}

func getPunct(lang, rule, defaultVal string) string {
	data := GetGrammarData(lang)
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
	data := GetGrammarData(lang)
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
	verb = core.Lower(core.Trim(verb))
	if verb == "" {
		return ""
	}
	if form := getVerbForm(currentLangForGrammar(), verb, "past"); form != "" {
		return form
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
	secondLast := rune(verb[len(verb)-2])
	if isVowel(lastChar) || lastChar == 'w' || lastChar == 'x' || lastChar == 'y' {
		return false
	}
	if !isVowel(secondLast) {
		return false
	}
	if len(verb) <= 4 {
		thirdLast := rune(verb[len(verb)-3])
		return !isVowel(thirdLast)
	}
	return false
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
			if unicode.IsUpper(rune(noun[0])) && len(form) > 0 {
				return core.Upper(string(form[0])) + form[1:]
			}
			return form
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
		if unicode.IsUpper(rune(noun[0])) && len(form) > 0 {
			return core.Upper(string(form[0])) + form[1:]
		}
		return form
	}
	if plural, ok := irregularNouns[lower]; ok {
		if unicode.IsUpper(rune(noun[0])) {
			return core.Upper(string(plural[0])) + plural[1:]
		}
		return plural
	}
	return applyRegularPlural(noun)
}

func applyRegularPlural(noun string) string {
	lower := core.Lower(noun)
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
	if core.HasSuffix(lower, "f") {
		return noun[:len(noun)-1] + "ves"
	}
	if core.HasSuffix(lower, "fe") {
		return noun[:len(noun)-2] + "ves"
	}
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

// Article returns the appropriate article for the current language.
// English falls back to phonetic "a"/"an" heuristics. Locale grammar data
// can override this with language-specific article forms.
//
//	Article("file")     // "a"
//	Article("error")    // "an"
//	Article("user")     // "a" (sounds like "yoo-zer")
//	Article("hour")     // "an" (silent h)
func Article(word string) string {
	if word == "" {
		return ""
	}
	trimmed := core.Trim(word)
	lower := core.Lower(trimmed)
	if article, ok := articleForCurrentLanguage(lower, word); ok {
		return article
	}
	if isInitialism(trimmed) {
		if initialismUsesVowelSound(trimmed) {
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

func articleForCurrentLanguage(lowerWord, originalWord string) (string, bool) {
	lang := currentLangForGrammar()
	data := GetGrammarData(lang)
	if data == nil {
		return "", false
	}

	if article, ok := articleForPluralForm(data, lowerWord, lang); ok {
		return article, true
	}
	if article, ok := articleForFrenchPluralGuess(data, lowerWord, originalWord, lang); ok {
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
	if !isFrenchLanguage(lang) {
		return "", false
	}
	if !isKnownPluralNoun(data, lowerWord) {
		return "", false
	}
	return "les", true
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
	if data == nil || len(data.Nouns) == 0 {
		return false
	}
	for _, forms := range data.Nouns {
		if forms.Other == "" || core.Lower(forms.Other) != lowerWord {
			continue
		}
		if forms.One != "" && core.Lower(forms.One) == lowerWord {
			continue
		}
		return true
	}
	return false
}

func articleFromGrammarForms(data *GrammarData, word string) (string, bool) {
	if data.Articles.IndefiniteDefault == "" && data.Articles.IndefiniteVowel == "" {
		return "", false
	}
	if usesVowelSoundArticle(word) && data.Articles.IndefiniteVowel != "" {
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

func usesVowelSoundArticle(word string) bool {
	trimmed := core.Trim(word)
	if trimmed == "" {
		return false
	}
	if isInitialism(trimmed) {
		return initialismUsesVowelSound(trimmed)
	}
	lower := core.Lower(trimmed)
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

func looksLikeFrenchPlural(word string) bool {
	trimmed := core.Trim(word)
	if trimmed == "" || strings.ContainsAny(trimmed, " \t") || isInitialism(trimmed) {
		return false
	}
	lower := core.Lower(trimmed)
	if core.HasSuffix(lower, "aux") || core.HasSuffix(lower, "eaux") {
		return true
	}
	return core.HasSuffix(lower, "s") || core.HasSuffix(lower, "x")
}

func startsWithVowelSound(word string) bool {
	lower := core.Lower(core.Trim(word))
	if lower == "" {
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

func renderWordOrTitle(lang, word string) string {
	if translated := getWord(lang, word); translated != "" {
		return translated
	}
	return Title(word)
}

// Quote wraps a string in double quotes.
func Quote(s string) string {
	return `"` + s + `"`
}

// ArticlePhrase prefixes a noun phrase with the correct article.
func ArticlePhrase(word string) string {
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	word = renderWord(lang, word)
	article := Article(word)
	if article == "" {
		return ""
	}
	if strings.HasSuffix(article, "'") {
		return article + word
	}
	return article + " " + word
}

// DefiniteArticle returns the language-specific definite article for a word.
// For languages such as French, this respects gendered articles, plural forms,
// and elision rules when grammar data is available.
func DefiniteArticle(word string) string {
	if word == "" {
		return ""
	}
	trimmed := core.Trim(word)
	lower := core.Lower(trimmed)
	if article, ok := definiteArticleForCurrentLanguage(lower, trimmed); ok {
		return article
	}
	lang := currentLangForGrammar()
	data := GetGrammarData(lang)
	if data != nil && data.Articles.Definite != "" {
		return data.Articles.Definite
	}
	return "the"
}

// DefinitePhrase prefixes a noun phrase with the correct definite article.
func DefinitePhrase(word string) string {
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	word = renderWord(lang, word)
	article := DefiniteArticle(word)
	if article == "" {
		return ""
	}
	if strings.HasSuffix(article, "'") {
		return article + word
	}
	return article + " " + word
}

func definiteArticleForCurrentLanguage(lowerWord, originalWord string) (string, bool) {
	lang := currentLangForGrammar()
	data := GetGrammarData(lang)
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

func definiteArticleFromGrammarForms(data *GrammarData, lowerWord, originalWord, lang string) (string, bool) {
	if data == nil || data.Articles.Definite == "" {
		return "", false
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
		"lower":           core.Lower,
		"upper":           core.Upper,
		"past":            PastTense,
		"gerund":          Gerund,
		"plural":          Pluralize,
		"pluralForm":      PluralForm,
		"article":         ArticlePhrase,
		"definite":        DefinitePhrase,
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
	if subject == "" {
		return Title(g) + suffix
	}
	return Title(g) + " " + renderWord(lang, subject) + suffix
}

// ActionResult returns a completion message: "File deleted"
func ActionResult(verb, subject string) string {
	p := PastTense(verb)
	if p == "" {
		return ""
	}
	if subject == "" {
		return Title(p)
	}
	return renderWordOrTitle(currentLangForGrammar(), subject) + " " + p
}

// ActionFailed returns a failure message: "Failed to delete file"
func ActionFailed(verb, subject string) string {
	if verb == "" {
		return ""
	}
	if subject == "" {
		return "Failed to " + verb
	}
	return "Failed to " + verb + " " + renderWord(currentLangForGrammar(), subject)
}

// Label returns a label with suffix: "Status:" (EN) or "Statut :" (FR)
func Label(word string) string {
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	translated := renderWordOrTitle(lang, word)
	suffix := getPunct(lang, "label", ":")
	return translated + suffix
}
