package i18n

import (
	"strings"
	"text/template"
	"unicode"
)

// GetGrammarData returns the grammar data for the specified language.
func GetGrammarData(lang string) *GrammarData {
	grammarCacheMu.RLock()
	defer grammarCacheMu.RUnlock()
	return grammarCache[lang]
}

// SetGrammarData sets the grammar data for a language.
func SetGrammarData(lang string, data *GrammarData) {
	grammarCacheMu.Lock()
	defer grammarCacheMu.Unlock()
	grammarCache[lang] = data
}

// IrregularVerbs returns a copy of the irregular verb forms map.
func IrregularVerbs() map[string]VerbForms {
	result := make(map[string]VerbForms, len(irregularVerbs))
	for k, v := range irregularVerbs {
		result[k] = v
	}
	return result
}

// IrregularNouns returns a copy of the irregular nouns map.
func IrregularNouns() map[string]string {
	result := make(map[string]string, len(irregularNouns))
	for k, v := range irregularNouns {
		result[k] = v
	}
	return result
}

func getVerbForm(lang, verb, form string) string {
	data := GetGrammarData(lang)
	if data == nil || data.Verbs == nil {
		return ""
	}
	verb = strings.ToLower(verb)
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
	return data.Words[strings.ToLower(word)]
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
	noun = strings.ToLower(noun)
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
	verb = strings.ToLower(strings.TrimSpace(verb))
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
	if strings.HasSuffix(verb, "ed") && len(verb) > 2 {
		thirdFromEnd := verb[len(verb)-3]
		if !isVowel(rune(thirdFromEnd)) && thirdFromEnd != 'e' {
			return verb
		}
	}
	if strings.HasSuffix(verb, "e") {
		return verb + "d"
	}
	if strings.HasSuffix(verb, "y") && len(verb) > 1 {
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
	verb = strings.ToLower(strings.TrimSpace(verb))
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
	if strings.HasSuffix(verb, "ie") {
		return verb[:len(verb)-2] + "ying"
	}
	if strings.HasSuffix(verb, "e") && len(verb) > 1 {
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
		return noun
	}
	return PluralForm(noun)
}

// PluralForm returns the plural form of a noun.
func PluralForm(noun string) string {
	noun = strings.TrimSpace(noun)
	if noun == "" {
		return ""
	}
	lower := strings.ToLower(noun)
	if form := getNounForm(currentLangForGrammar(), lower, "other"); form != "" {
		if unicode.IsUpper(rune(noun[0])) && len(form) > 0 {
			return strings.ToUpper(string(form[0])) + form[1:]
		}
		return form
	}
	if plural, ok := irregularNouns[lower]; ok {
		if unicode.IsUpper(rune(noun[0])) {
			return strings.ToUpper(string(plural[0])) + plural[1:]
		}
		return plural
	}
	return applyRegularPlural(noun)
}

func applyRegularPlural(noun string) string {
	lower := strings.ToLower(noun)
	if strings.HasSuffix(lower, "s") ||
		strings.HasSuffix(lower, "ss") ||
		strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") {
		return noun + "es"
	}
	if strings.HasSuffix(lower, "y") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			return noun[:len(noun)-1] + "ies"
		}
	}
	if strings.HasSuffix(lower, "f") {
		return noun[:len(noun)-1] + "ves"
	}
	if strings.HasSuffix(lower, "fe") {
		return noun[:len(noun)-2] + "ves"
	}
	if strings.HasSuffix(lower, "o") && len(noun) > 1 {
		prev := rune(lower[len(lower)-2])
		if !isVowel(prev) {
			if lower == "hero" || lower == "potato" || lower == "tomato" || lower == "echo" || lower == "veto" {
				return noun + "es"
			}
		}
	}
	return noun + "s"
}

// Article returns the appropriate indefinite article ("a" or "an").
//
//	Article("file")     // "a"
//	Article("error")    // "an"
//	Article("user")     // "a" (sounds like "yoo-zer")
//	Article("hour")     // "an" (silent h)
func Article(word string) string {
	if word == "" {
		return ""
	}
	lower := strings.ToLower(strings.TrimSpace(word))
	for key := range consonantSounds {
		if strings.HasPrefix(lower, key) {
			return "a"
		}
	}
	for key := range vowelSounds {
		if strings.HasPrefix(lower, key) {
			return "an"
		}
	}
	if len(lower) > 0 && isVowel(rune(lower[0])) {
		return "an"
	}
	return "a"
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// Title capitalises the first letter of each word.
func Title(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prev := ' '
	for _, r := range s {
		if !unicode.IsLetter(prev) && unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(r)
		}
		prev = r
	}
	return b.String()
}

// Quote wraps a string in double quotes.
func Quote(s string) string {
	return `"` + s + `"`
}

// TemplateFuncs returns the template.FuncMap with all grammar functions.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"title":      Title,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"past":       PastTense,
		"gerund":     Gerund,
		"plural":     Pluralize,
		"pluralForm": PluralForm,
		"article":    Article,
		"quote":      Quote,
	}
}

// Progress returns a progress message: "Building..."
func Progress(verb string) string {
	lang := currentLangForGrammar()
	word := getWord(lang, verb)
	if word == "" {
		word = verb
	}
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
	word := getWord(lang, verb)
	if word == "" {
		word = verb
	}
	g := Gerund(word)
	if g == "" {
		return ""
	}
	suffix := getPunct(lang, "progress", "...")
	return Title(g) + " " + subject + suffix
}

// ActionResult returns a completion message: "File deleted"
func ActionResult(verb, subject string) string {
	p := PastTense(verb)
	if p == "" || subject == "" {
		return ""
	}
	return Title(subject) + " " + p
}

// ActionFailed returns a failure message: "Failed to delete file"
func ActionFailed(verb, subject string) string {
	if verb == "" {
		return ""
	}
	if subject == "" {
		return "Failed to " + verb
	}
	return "Failed to " + verb + " " + subject
}

// Label returns a label with suffix: "Status:" (EN) or "Statut :" (FR)
func Label(word string) string {
	if word == "" {
		return ""
	}
	lang := currentLangForGrammar()
	translated := getWord(lang, word)
	if translated == "" {
		translated = word
	}
	suffix := getPunct(lang, "label", ":")
	return Title(translated) + suffix
}
