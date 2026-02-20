package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Package-level T() ---

func TestT_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	got := T("prompt.yes")
	assert.Equal(t, "y", got)
}

func TestT_Good_WithHandlers(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	got := T("i18n.label.status")
	assert.Equal(t, "Status:", got)
}

func TestT_Good_MissingKey(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	got := T("nonexistent.key.test")
	assert.Equal(t, "nonexistent.key.test", got)
}

// --- Package-level Raw() ---

func TestRaw_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	got := Raw("prompt.yes")
	assert.Equal(t, "y", got)
}

func TestRaw_Good_BypassesHandlers(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	// i18n.label.status is not in messages, handlers don't apply in Raw
	got := Raw("i18n.label.status")
	assert.Equal(t, "i18n.label.status", got)
}

// --- SetLanguage / CurrentLanguage ---

func TestSetLanguage_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	err = SetLanguage("en")
	assert.NoError(t, err)
	assert.Contains(t, CurrentLanguage(), "en")
}

func TestSetLanguage_Bad_Unsupported(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	_ = SetLanguage("xx")
}

func TestCurrentLanguage_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	lang := CurrentLanguage()
	assert.NotEmpty(t, lang)
}

// --- SetMode / CurrentMode ---

func TestSetMode_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	SetMode(ModeCollect)
	assert.Equal(t, ModeCollect, CurrentMode())

	SetMode(ModeNormal)
	assert.Equal(t, ModeNormal, CurrentMode())
}

func TestCurrentMode_Good_Default(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	assert.Equal(t, ModeNormal, CurrentMode())
}

// --- N() numeric shorthand ---

func TestN_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	tests := []struct {
		name   string
		format string
		value  any
		want   string
	}{
		{"number", "number", int64(1234567), "1,234,567"},
		{"percent", "percent", 0.85, "85%"},
		{"bytes", "bytes", int64(1536000), "1.5 MB"},
		{"ordinal", "ordinal", 1, "1st"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := N(tt.format, tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- AddHandler / PrependHandler ---

func TestAddHandler_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	initialCount := len(svc.Handlers())

	AddHandler(LabelHandler{})
	assert.Equal(t, initialCount+1, len(svc.Handlers()))
}

func TestPrependHandler_Good(t *testing.T) {
	svc, err := New(WithHandlers()) // start with no handlers
	require.NoError(t, err)
	_ = Init()
	SetDefault(svc)

	PrependHandler(LabelHandler{})
	assert.Equal(t, 1, len(svc.Handlers()))

	// Prepend another
	PrependHandler(ProgressHandler{})
	handlers := svc.Handlers()
	assert.Equal(t, 2, len(handlers))
}

// --- executeIntentTemplate ---

func TestExecuteIntentTemplate_Good(t *testing.T) {
	data := templateData{
		Subject: "config.yaml",
		Noun:    "file",
		Count:   1,
	}

	got := executeIntentTemplate("Delete {{.Subject}}?", data)
	assert.Equal(t, "Delete config.yaml?", got)
}

func TestExecuteIntentTemplate_Good_Empty(t *testing.T) {
	got := executeIntentTemplate("", templateData{})
	assert.Equal(t, "", got)
}

func TestExecuteIntentTemplate_Bad_InvalidTemplate(t *testing.T) {
	got := executeIntentTemplate("{{.Invalid", templateData{})
	assert.Equal(t, "{{.Invalid", got, "should return raw template on parse error")
}

func TestExecuteIntentTemplate_Good_Cached(t *testing.T) {
	data := templateData{Subject: "test"}
	tmplStr := "Hello {{.Subject}} intent"

	// First call caches
	got1 := executeIntentTemplate(tmplStr, data)
	assert.Equal(t, "Hello test intent", got1)

	// Second call hits cache
	got2 := executeIntentTemplate(tmplStr, data)
	assert.Equal(t, "Hello test intent", got2)
}

func TestExecuteIntentTemplate_Good_WithFuncs(t *testing.T) {
	data := templateData{Subject: "build"}

	got := executeIntentTemplate("{{past .Subject}}!", data)
	assert.Equal(t, "built!", got)
}

// --- applyTemplate ---

func TestApplyTemplate_Good(t *testing.T) {
	got := applyTemplate("Hello {{.Name}}", map[string]any{"Name": "World"})
	assert.Equal(t, "Hello World", got)
}

func TestApplyTemplate_Good_NoTemplate(t *testing.T) {
	got := applyTemplate("plain text", nil)
	assert.Equal(t, "plain text", got, "should return text as-is without template markers")
}

func TestApplyTemplate_Bad_InvalidTemplate(t *testing.T) {
	got := applyTemplate("{{.Invalid", nil)
	assert.Equal(t, "{{.Invalid", got, "should return raw text on parse error")
}

func TestApplyTemplate_Good_Cached(t *testing.T) {
	tmpl := "Cached {{.Val}} template test"
	data := map[string]any{"Val": "test"}

	got1 := applyTemplate(tmpl, data)
	assert.Equal(t, "Cached test template test", got1)

	// Hit cache
	got2 := applyTemplate(tmpl, data)
	assert.Equal(t, "Cached test template test", got2)
}

func TestApplyTemplate_Bad_ExecuteError(t *testing.T) {
	// A template that references a method on a non-struct type
	got := applyTemplate("{{.Missing}}", "not a map")
	assert.Equal(t, "{{.Missing}}", got)
}

// --- ErrServiceNotInitialized ---

func TestErrServiceNotInitialized_Good(t *testing.T) {
	assert.Equal(t, "i18n: service not initialized", ErrServiceNotInitialized.Error())
}
