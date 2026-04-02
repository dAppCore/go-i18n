package i18n

import (
	"bytes"
	"io/fs"
	"text/template"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
)

// T translates a message using the default service.
func T(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.T(messageID, args...)
	}
	return messageID
}

// Translate translates a message using the default service and returns a Core result.
func Translate(messageID string, args ...any) core.Result {
	if svc := Default(); svc != nil {
		return svc.Translate(messageID, args...)
	}
	return core.Result{Value: messageID, OK: false}
}

// Raw translates without i18n.* namespace magic.
func Raw(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.Raw(messageID, args...)
	}
	return messageID
}

// ErrServiceNotInitialised is returned when the service is not initialised.
var ErrServiceNotInitialised = core.NewError("i18n: service not initialised")

// ErrServiceNotInitialized is deprecated: use ErrServiceNotInitialised.
var ErrServiceNotInitialized = ErrServiceNotInitialised

// SetLanguage sets the language for the default service.
func SetLanguage(lang string) error {
	svc := Default()
	if svc == nil {
		return ErrServiceNotInitialised
	}
	return svc.SetLanguage(lang)
}

// CurrentLanguage returns the current language code.
func CurrentLanguage() string {
	if svc := Default(); svc != nil {
		return svc.Language()
	}
	return "en"
}

// AvailableLanguages returns the loaded language tags on the default service.
func AvailableLanguages() []string {
	if svc := Default(); svc != nil {
		langs := svc.AvailableLanguages()
		if len(langs) == 0 {
			return nil
		}
		return append([]string(nil), langs...)
	}
	return nil
}

// SetMode sets the translation mode for the default service.
func SetMode(m Mode) {
	if svc := Default(); svc != nil {
		svc.SetMode(m)
	}
}

// SetFallback sets the fallback language for the default service.
func SetFallback(lang string) {
	if svc := Default(); svc != nil {
		svc.SetFallback(lang)
	}
}

// CurrentMode returns the current translation mode.
func CurrentMode() Mode {
	if svc := Default(); svc != nil {
		return svc.Mode()
	}
	return ModeNormal
}

// CurrentFallback returns the current fallback language.
func CurrentFallback() string {
	if svc := Default(); svc != nil {
		return svc.Fallback()
	}
	return "en"
}

// CurrentFormality returns the current default formality.
func CurrentFormality() Formality {
	if svc := Default(); svc != nil {
		return svc.Formality()
	}
	return FormalityNeutral
}

// CurrentDebug reports whether debug mode is enabled on the default service.
func CurrentDebug() bool {
	if svc := Default(); svc != nil {
		return svc.Debug()
	}
	return false
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
	return T("i18n.numeric."+format, append([]any{value}, args...)...)
}

// Prompt translates a prompt key from the prompt namespace.
//
//	Prompt("yes")      // "y"
//	Prompt("confirm")  // "Are you sure?"
func Prompt(key string) string {
	key = normalizeLookupKey(key)
	if key == "" {
		return ""
	}
	return T("prompt." + key)
}

// Lang translates a language label from the lang namespace.
//
//	Lang("de")  // "German"
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
	return T("lang." + key)
}

func normalizeLookupKey(key string) string {
	return core.Lower(core.Trim(key))
}

// AddHandler appends one or more handlers to the default service's handler chain.
func AddHandler(handlers ...KeyHandler) {
	if svc := Default(); svc != nil {
		svc.AddHandler(handlers...)
	}
}

// LoadFS loads additional translations from an fs.FS into the default service.
//
// Call this from init() in packages that ship their own locale files:
//
//	//go:embed locales/*.json
//	var localeFS embed.FS
//
//	func init() { i18n.LoadFS(localeFS, "locales") }
func LoadFS(fsys fs.FS, dir string) {
	if svc := Default(); svc != nil {
		if err := svc.LoadFS(fsys, dir); err != nil {
			log.Error("i18n: LoadFS failed", "dir", dir, "err", err)
		}
	}
}

// PrependHandler inserts one or more handlers at the start of the default service's handler chain.
func PrependHandler(handlers ...KeyHandler) {
	if svc := Default(); svc != nil {
		svc.PrependHandler(handlers...)
	}
}

// ClearHandlers removes all handlers from the default service.
func ClearHandlers() {
	if svc := Default(); svc != nil {
		svc.ClearHandlers()
	}
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
