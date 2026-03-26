package i18n

import (
	"bytes"
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

// N formats a number using the i18n.numeric.* namespace.
//
//	N("number", 1234567)   // "1,234,567"
//	N("percent", 0.85)     // "85%"
//	N("bytes", 1536000)    // "1.5 MB"
//	N("ordinal", 1)        // "1st"
func N(format string, value any) string {
	return T("i18n.numeric."+format, value)
}

// AddHandler appends a handler to the default service's handler chain.
func AddHandler(h KeyHandler) {
	if svc := Default(); svc != nil {
		svc.AddHandler(h)
	}
}

// PrependHandler inserts a handler at the start of the default service's handler chain.
func PrependHandler(h KeyHandler) {
	if svc := Default(); svc != nil {
		svc.PrependHandler(h)
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
	if cached, ok := templateCache.Load(text); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return text
		}
		return buf.String()
	}
	tmpl, err := template.New("").Parse(text)
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
