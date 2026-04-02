package i18n

import (
	"bytes"
	"strings"
	"text/template"

	"dappco.re/go/core"
)

// T translates a message using the default service.
func T(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.T(messageID, args...)
	}
	return messageID
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

// SetMode sets the translation mode for the default service.
func SetMode(m Mode) {
	if svc := Default(); svc != nil {
		svc.SetMode(m)
	}
}

// CurrentMode returns the current translation mode.
func CurrentMode() Mode {
	if svc := Default(); svc != nil {
		return svc.Mode()
	}
	return ModeNormal
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

// N formats a number using the i18n.numeric.* namespace.
//
//	N("number", 1234567)   // "1,234,567"
//	N("percent", 0.85)     // "85%"
//	N("bytes", 1536000)    // "1.46 MB"
//	N("ordinal", 1)        // "1st"
func N(format string, value any) string {
	return T("i18n.numeric."+format, value)
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
	if idx := strings.IndexAny(key, "-_"); idx > 0 {
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
		rendered := map[string]any{
			"Context":   v.Context,
			"Gender":    v.Gender,
			"Location":  v.Location,
			"Formality": v.Formality,
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
