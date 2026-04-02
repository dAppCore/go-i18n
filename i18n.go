package i18n

import (
	"bytes"
	"io/fs"
	"text/template"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
)

// T translates a message using the default service.
//
// Example:
//
//	i18n.T("greeting")
func T(messageID string, args ...any) string {
	return defaultServiceValue(messageID, func(svc *Service) string {
		return svc.T(messageID, args...)
	})
}

// Translate translates a message using the default service and returns a Core result.
//
// Example:
//
//	result := i18n.Translate("greeting")
func Translate(messageID string, args ...any) core.Result {
	return defaultServiceValue(core.Result{Value: messageID, OK: false}, func(svc *Service) core.Result {
		return svc.Translate(messageID, args...)
	})
}

// Raw translates without i18n.* namespace magic.
//
// Example:
//
//	i18n.Raw("prompt.yes")
func Raw(messageID string, args ...any) string {
	return defaultServiceValue(messageID, func(svc *Service) string {
		return svc.Raw(messageID, args...)
	})
}

// ErrServiceNotInitialised is returned when the service is not initialised.
var ErrServiceNotInitialised = core.NewError("i18n: service not initialised")

// ErrServiceNotInitialized is deprecated: use ErrServiceNotInitialised.
var ErrServiceNotInitialized = ErrServiceNotInitialised

// SetLanguage sets the language for the default service.
//
// Example:
//
//	_ = i18n.SetLanguage("fr")
func SetLanguage(lang string) error {
	return defaultServiceValue(ErrServiceNotInitialised, func(svc *Service) error {
		return svc.SetLanguage(lang)
	})
}

// CurrentLanguage returns the current language code.
//
// Example:
//
//	lang := i18n.CurrentLanguage()
func CurrentLanguage() string {
	return Language()
}

// Language returns the current language code.
//
// Example:
//
//	lang := i18n.Language()
func Language() string {
	return defaultServiceValue("en", func(svc *Service) string {
		return svc.Language()
	})
}

// AvailableLanguages returns the loaded language tags on the default service.
//
// Example:
//
//	langs := i18n.AvailableLanguages()
func AvailableLanguages() []string {
	return defaultServiceValue([]string{}, func(svc *Service) []string {
		langs := svc.AvailableLanguages()
		if len(langs) == 0 {
			return []string{}
		}
		return append([]string(nil), langs...)
	})
}

// CurrentAvailableLanguages returns the loaded language tags on the default
// service.
//
// Example:
//
//	langs := i18n.CurrentAvailableLanguages()
func CurrentAvailableLanguages() []string {
	return AvailableLanguages()
}

// SetMode sets the translation mode for the default service.
//
// Example:
//
//	i18n.SetMode(i18n.ModeCollect)
func SetMode(m Mode) {
	withDefaultService(func(svc *Service) { svc.SetMode(m) })
}

// SetFallback sets the fallback language for the default service.
//
// Example:
//
//	i18n.SetFallback("en")
func SetFallback(lang string) {
	withDefaultService(func(svc *Service) { svc.SetFallback(lang) })
}

// Fallback returns the current fallback language.
//
// Example:
//
//	fallback := i18n.Fallback()
func Fallback() string {
	return defaultServiceValue("en", func(svc *Service) string {
		return svc.Fallback()
	})
}

// CurrentMode returns the current translation mode.
//
// Example:
//
//	mode := i18n.CurrentMode()
func CurrentMode() Mode {
	return defaultServiceValue(ModeNormal, func(svc *Service) Mode { return svc.Mode() })
}

// CurrentFallback returns the current fallback language.
//
// Example:
//
//	fallback := i18n.CurrentFallback()
func CurrentFallback() string {
	return Fallback()
}

// CurrentFormality returns the current default formality.
//
// Example:
//
//	formality := i18n.CurrentFormality()
func CurrentFormality() Formality {
	return defaultServiceValue(FormalityNeutral, func(svc *Service) Formality { return svc.Formality() })
}

// CurrentDebug reports whether debug mode is enabled on the default service.
//
// Example:
//
//	debug := i18n.CurrentDebug()
func CurrentDebug() bool {
	return Debug()
}

// Debug reports whether debug mode is enabled on the default service.
//
// Example:
//
//	debug := i18n.Debug()
func Debug() bool {
	return defaultServiceValue(false, func(svc *Service) bool {
		return svc.Debug()
	})
}

// N formats a value using the i18n.numeric.* namespace.
//
//	N("number", 1234567)   // "1,234,567"
//	N("percent", 0.85)     // "85%"
//	N("bytes", 1536000)    // "1.46 MB"
//	N("ordinal", 1)        // "1st"
//
// Multi-argument formats such as "ago" also pass through unchanged:
//
//	N("ago", 5, "minutes")  // "5 minutes ago"
func N(format string, value any, args ...any) string {
	format = normalizeLookupKey(format)
	switch format {
	case "number", "int":
		return FormatNumber(toInt64(value))
	case "decimal", "float":
		return FormatDecimal(toFloat64(value))
	case "percent", "pct":
		return FormatPercent(toFloat64(value))
	case "bytes", "size":
		return FormatBytes(toInt64(value))
	case "ordinal", "ord":
		return FormatOrdinal(toInt(value))
	case "ago":
		if len(args) > 0 {
			if unit, ok := args[0].(string); ok {
				return FormatAgo(toInt(value), unit)
			}
		}
	}
	return T("i18n.numeric."+format, append([]any{value}, args...)...)
}

// Prompt translates a prompt key from the prompt namespace.
//
// Example:
//
//	  i18n.Prompt("confirm")
//
//		Prompt("yes")      // "y"
//		Prompt("confirm")  // "Are you sure?"
func Prompt(key string) string {
	key = normalizeLookupKey(key)
	if key == "" {
		return ""
	}
	return T("prompt." + key)
}

// Lang translates a language label from the lang namespace.
//
// Example:
//
//	  i18n.Lang("de")
//
//		Lang("de")  // "German"
func Lang(key string) string {
	key = normalizeLookupKey(key)
	if key == "" {
		return ""
	}
	if got := T("lang." + key); got != "lang."+key {
		return got
	}
	if idx := indexAny(key, "-_"); idx > 0 {
		if base := key[:idx]; base != "" {
			if got := T("lang." + base); got != "lang."+base {
				return got
			}
		}
	}
	return "lang." + key
}

func normalizeLookupKey(key string) string {
	return core.Lower(core.Trim(key))
}

// AddHandler appends one or more handlers to the default service's handler chain.
//
// Example:
//
//	i18n.AddHandler(MyHandler{})
func AddHandler(handlers ...KeyHandler) {
	withDefaultService(func(svc *Service) { svc.AddHandler(handlers...) })
}

// SetHandlers replaces the default service's handler chain.
//
// Example:
//
//	i18n.SetHandlers(i18n.LabelHandler{}, i18n.ProgressHandler{})
func SetHandlers(handlers ...KeyHandler) {
	withDefaultService(func(svc *Service) { svc.SetHandlers(handlers...) })
}

// LoadFS loads additional translations from an fs.FS into the default service.
//
// Example:
//
//	i18n.LoadFS(os.DirFS("."), "locales")
//
// Call this from init() in packages that ship their own locale files:
//
//	//go:embed locales/*.json
//	var localeFS embed.FS
//
//	func init() { i18n.LoadFS(localeFS, "locales") }
func LoadFS(fsys fs.FS, dir string) {
	withDefaultService(func(svc *Service) {
		if err := svc.AddLoader(NewFSLoader(fsys, dir)); err != nil {
			log.Error("i18n: LoadFS failed", "dir", dir, "err", err)
		}
	})
}

// AddMessages adds message strings to the default service for a language.
//
// Example:
//
//	i18n.AddMessages("en", map[string]string{"custom.greeting": "Hello!"})
func AddMessages(lang string, messages map[string]string) {
	withDefaultService(func(svc *Service) { svc.AddMessages(lang, messages) })
}

// PrependHandler inserts one or more handlers at the start of the default service's handler chain.
//
// Example:
//
//	i18n.PrependHandler(MyHandler{})
func PrependHandler(handlers ...KeyHandler) {
	withDefaultService(func(svc *Service) { svc.PrependHandler(handlers...) })
}

// CurrentHandlers returns a copy of the default service's handler chain.
//
// Example:
//
//	handlers := i18n.CurrentHandlers()
func CurrentHandlers() []KeyHandler {
	return Handlers()
}

// Handlers returns a copy of the default service's handler chain.
//
// Example:
//
//	handlers := i18n.Handlers()
func Handlers() []KeyHandler {
	return defaultServiceValue([]KeyHandler{}, func(svc *Service) []KeyHandler {
		return svc.Handlers()
	})
}

// ClearHandlers removes all handlers from the default service.
//
// Example:
//
//	i18n.ClearHandlers()
func ClearHandlers() {
	withDefaultService(func(svc *Service) { svc.ClearHandlers() })
}

// ResetHandlers restores the built-in default handler chain on the default
// service.
//
// Example:
//
//	i18n.ResetHandlers()
func ResetHandlers() {
	withDefaultService(func(svc *Service) { svc.ResetHandlers() })
}

func executeIntentTemplate(tmplStr string, data templateData) string {
	if tmplStr == "" {
		return ""
	}
	if cached, ok := templateCache.Load(tmplStr); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return tmplStr
		}
		return buf.String()
	}
	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}
	templateCache.Store(tmplStr, tmpl)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return tmplStr
	}
	return buf.String()
}

func applyTemplate(text string, data any) string {
	if !core.Contains(text, "{{") {
		return text
	}
	data = templateDataForRendering(data)
	if cached, ok := templateCache.Load(text); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return text
		}
		return buf.String()
	}
	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(text)
	if err != nil {
		return text
	}
	templateCache.Store(text, tmpl)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return text
	}
	return buf.String()
}

func templateDataForRendering(data any) any {
	switch v := data.(type) {
	case *TranslationContext:
		if v == nil {
			return nil
		}
		count, explicit := v.countValue()
		if !explicit && v.Extra != nil {
			if c, ok := v.Extra["Count"]; ok {
				count = toInt(c)
			} else if c, ok := v.Extra["count"]; ok {
				count = toInt(c)
			}
		}
		rendered := map[string]any{
			"Context":   v.Context,
			"Gender":    v.Gender,
			"Location":  v.Location,
			"Formality": v.Formality,
			"Count":     count,
			"IsPlural":  count != 1,
			"Extra":     v.Extra,
		}
		for key, value := range v.Extra {
			if _, exists := rendered[key]; !exists {
				rendered[key] = value
			}
		}
		return rendered
	case *Subject:
		if v == nil {
			return nil
		}
		return map[string]any{
			"Subject":   v.String(),
			"Noun":      v.Noun,
			"Count":     v.count,
			"Gender":    v.gender,
			"Location":  v.location,
			"Formality": v.formality,
			"IsFormal":  v.formality == FormalityFormal,
			"IsPlural":  v.count != 1,
			"Value":     v.Value,
		}
	default:
		return data
	}
}
