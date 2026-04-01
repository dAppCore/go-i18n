package i18n

import (
	"strings"
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
	return Label(word)
}

// ProgressHandler handles i18n.progress.{verb} -> "Building..." patterns.
type ProgressHandler struct{}

func (h ProgressHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.progress.")
}

func (h ProgressHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.progress.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ProgressSubject(verb, subj)
		}
	}
	return Progress(verb)
}

// CountHandler handles i18n.count.{noun} -> "5 files" patterns.
type CountHandler struct{}

func (h CountHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.count.")
}

func (h CountHandler) Handle(key string, args []any, next func() string) string {
	noun := core.TrimPrefix(key, "i18n.count.")
	lang := currentLangForGrammar()
	if len(args) > 0 {
		count := toInt(args[0])
		return core.Sprintf("%s %s", FormatNumber(int64(count)), countWordForm(lang, noun, count))
	}
	return renderWord(lang, noun)
}

// DoneHandler handles i18n.done.{verb} -> "File deleted" patterns.
type DoneHandler struct{}

func (h DoneHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.done.")
}

func (h DoneHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.done.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ActionResult(verb, subj)
		}
	}
	return Title(PastTense(verb))
}

// FailHandler handles i18n.fail.{verb} -> "Failed to delete file" patterns.
type FailHandler struct{}

func (h FailHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.fail.")
}

func (h FailHandler) Handle(key string, args []any, next func() string) string {
	verb := core.TrimPrefix(key, "i18n.fail.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ActionFailed(verb, subj)
		}
	}
	return ActionFailed(verb, "")
}

// NumericHandler handles i18n.numeric.{format} -> formatted numbers.
type NumericHandler struct{}

func (h NumericHandler) Match(key string) bool {
	return core.HasPrefix(key, "i18n.numeric.")
}

func (h NumericHandler) Handle(key string, args []any, next func() string) string {
	if len(args) == 0 {
		return next()
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
	return next()
}

// DefaultHandlers returns the built-in i18n.* namespace handlers.
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

func isPluralisableWordDisplay(s string) bool {
	hasLetter := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsSpace(r):
			continue
		default:
			return false
		}
	}
	return hasLetter
}

func isUpperAcronymPlural(s string) bool {
	if len(s) < 2 || !strings.HasSuffix(s, "s") {
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

// RunHandlerChain executes a chain of handlers for a key.
func RunHandlerChain(handlers []KeyHandler, key string, args []any, fallback func() string) string {
	for i, h := range handlers {
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

var (
	_ KeyHandler = LabelHandler{}
	_ KeyHandler = ProgressHandler{}
	_ KeyHandler = CountHandler{}
	_ KeyHandler = DoneHandler{}
	_ KeyHandler = FailHandler{}
	_ KeyHandler = NumericHandler{}
)
