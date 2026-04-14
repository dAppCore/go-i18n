package i18n

import (
	"unicode"

	"dappco.re/go/core"
)

// LabelHandler handles i18n.label.{word} -> "Status:" patterns.
type LabelHandler struct{}

func (h LabelHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.label.")
}

func (h LabelHandler) Handle(key string, args []any, next func() string) string {
	word := core.TrimPrefix(key, "i18n.label.")
	if got := Label(word); got != "" {
		return got
	}
	if next != nil {
		return next()
	}
	return ""
}

// ProgressHandler handles i18n.progress.{verb} -> "Building..." patterns.
type ProgressHandler struct{}

func (h ProgressHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.progress.")
}

func (h ProgressHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.progress.")
	if len(args) > 0 {
		if subj := subjectArgText(args[0]); subj != "" {
			if got := ProgressSubject(verb, subj); got != "" {
				return got
			}
		}
	}
	if got := Progress(verb); got != "" {
		return got
	}
	if next != nil {
		return next()
	}
	return ""
}

// CountHandler handles i18n.count.{noun} -> "5 files" patterns.
type CountHandler struct{}

func (h CountHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.count.")
}

func (h CountHandler) Handle(key string, args []any, next func() string) string {
	noun := core.TrimPrefix(key, "i18n.count.")
	lang := currentLangForGrammar()
	if core.Trim(noun) == "" {
		if next != nil {
			return next()
		}
		return ""
	}
	if len(args) > 0 {
		count := getCount(args[0])
		return core.Sprintf("%s %s", FormatNumber(int64(count)), countWordForm(lang, noun, count))
	}
	return countWordForm(lang, noun, 1)
}

// DoneHandler handles i18n.done.{verb} -> "File deleted" patterns.
type DoneHandler struct{}

func (h DoneHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.done.")
}

func (h DoneHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.done.")
	if len(args) > 0 {
		if subj := subjectArgText(args[0]); subj != "" {
			if got := ActionResult(verb, subj); got != "" {
				return got
			}
		}
	}
	if got := Title(PastTense(verb)); got != "" {
		return got
	}
	if next != nil {
		return next()
	}
	return ""
}

// FailHandler handles i18n.fail.{verb} -> "Failed to delete file" patterns.
type FailHandler struct{}

func (h FailHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.fail.")
}

func (h FailHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.fail.")
	if len(args) > 0 {
		if subj := subjectArgText(args[0]); subj != "" {
			if got := ActionFailed(verb, subj); got != "" {
				return got
			}
		}
	}
	if got := ActionFailed(verb, ""); got != "" {
		return got
	}
	if next != nil {
		return next()
	}
	return ""
}

// NumericHandler handles i18n.numeric.{format} -> formatted numbers.
type NumericHandler struct{}

func (h NumericHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.numeric.")
}

func (h NumericHandler) Handle(key string, args []any, next func() string) string {
	if len(args) == 0 {
		if next != nil {
			return next()
		}
		return ""
	}
	format := core.TrimPrefix(key, "i18n.numeric.")
	switch format {
	case "number", "int":
		return FormatNumber(toInt64(args[0]))
	case "decimal", "float":
		return FormatDecimal(toFloat64(args[0]))
	case "percent", "pct":
		return FormatPercent(toFloat64(args[0]))
	case "bytes", "size":
		return FormatBytes(toInt64(args[0]))
	case "ordinal", "ord":
		return FormatOrdinal(toInt(args[0]))
	case "ago":
		if len(args) >= 2 {
			if unit, ok := args[1].(string); ok {
				return FormatAgo(toInt(args[0]), unit)
			}
		}
	}
	if next != nil {
		return next()
	}
	return ""
}

// DefaultHandlers returns the built-in i18n.* namespace handlers.
//
//	handlers := i18n.DefaultHandlers()
func DefaultHandlers() []KeyHandler {
	return []KeyHandler{
		LabelHandler{},
		ProgressHandler{},
		CountHandler{},
		DoneHandler{},
		FailHandler{},
		NumericHandler{},
	}
}

func countWordForm(lang, noun string, count int) string {
	if hasGrammarCountForms(lang, noun) {
		return Pluralize(noun, count)
	}

	display := renderWord(lang, noun)
	if display == "" {
		return Pluralize(noun, count)
	}
	if count == 1 {
		return display
	}
	if !isPluralisableWordDisplay(display) {
		return display
	}
	if isUpperAcronymPlural(display) {
		return display
	}
	return Pluralize(display, count)
}

func hasGrammarCountForms(lang, noun string) bool {
	data := GetGrammarData(lang)
	if data == nil || len(data.Nouns) == 0 {
		return false
	}
	forms, ok := data.Nouns[core.Lower(noun)]
	if !ok {
		return false
	}
	return forms.One != "" || forms.Other != ""
}

func isPluralisableWordDisplay(s string) bool {
	hasLetter := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsSpace(r):
			// Multi-word vocabulary entries should stay exact. The count handler
			// prefixes the quantity, but does not invent a plural form for phrases.
			return false
		default:
			return false
		}
	}
	return hasLetter
}

func isUpperAcronymPlural(s string) bool {
	if len(s) < 2 || !core.HasSuffix(s, "s") {
		return false
	}
	hasLetter := false
	for _, r := range s[:len(s)-1] {
		if !unicode.IsLetter(r) {
			continue
		}
		hasLetter = true
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return hasLetter
}

func isAllUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		hasLetter = true
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return hasLetter
}

func subjectArgText(arg any) string {
	switch v := arg.(type) {
	case string:
		return v
	case *Subject:
		if v == nil {
			return ""
		}
		return v.String()
	case *TranslationContext:
		if v == nil {
			return ""
		}
		if text := core.Trim(v.String()); text != "" {
			return text
		}
		if v.Extra != nil {
			if text := contextArgText(v.Extra); text != "" {
				return text
			}
		}
		return ""
	case map[string]any:
		return contextArgText(v)
	case map[string]string:
		return contextArgText(v)
	case stringer:
		return v.String()
	default:
		return ""
	}
}

func contextArgText(values any) string {
	for _, key := range []string{"Subject", "subject", "Value", "value", "Text", "text", "Context", "context", "Noun", "noun"} {
		if text, ok := mapValueString(values, key); ok {
			return text
		}
	}
	return ""
}

// RunHandlerChain executes a chain of handlers for a key.
func RunHandlerChain(handlers []KeyHandler, key string, args []any, fallback func() string) string {
	for i, h := range handlers {
		if h == nil {
			continue
		}
		if h.Match(key) {
			next := func() string {
				remaining := handlers[i+1:]
				if len(remaining) > 0 {
					return RunHandlerChain(remaining, key, args, fallback)
				}
				return fallback()
			}
			return h.Handle(key, args, next)
		}
	}
	return fallback()
}

func filterNilHandlers(handlers []KeyHandler) []KeyHandler {
	if len(handlers) == 0 {
		return nil
	}
	filtered := make([]KeyHandler, 0, len(handlers))
	for _, h := range handlers {
		if h != nil {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

var (
	_ KeyHandler = LabelHandler{}
	_ KeyHandler = ProgressHandler{}
	_ KeyHandler = CountHandler{}
	_ KeyHandler = DoneHandler{}
	_ KeyHandler = FailHandler{}
	_ KeyHandler = NumericHandler{}
)
