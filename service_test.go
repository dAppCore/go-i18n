package i18n

import (
	"testing"
	"testing/fstest"
)

func TestNewService(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	lang := svc.Language()
	if lang == "" {
		t.Error("Language() is empty")
	}
	// Language matcher may return canonical form with region (e.g. "en-u-rg-uszzzz")
	// depending on LANG environment variable. Just check it starts with "en".
	if lang[:2] != "en" {
		t.Errorf("Language() = %q, expected to start with 'en'", lang)
	}

	langs := svc.AvailableLanguages()
	if len(langs) == 0 {
		t.Error("AvailableLanguages() is empty")
	}
}

func TestServiceT(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	// Label handler
	got := svc.T("i18n.label.status")
	if got != "Status:" {
		t.Errorf("T(i18n.label.status) = %q, want 'Status:'", got)
	}

	// Progress handler
	got = svc.T("i18n.progress.build")
	if got != "Building..." {
		t.Errorf("T(i18n.progress.build) = %q, want 'Building...'", got)
	}

	// Count handler
	got = svc.T("i18n.count.file", 5)
	if got != "5 files" {
		t.Errorf("T(i18n.count.file, 5) = %q, want '5 files'", got)
	}

	// Done handler
	got = svc.T("i18n.done.delete", "config.yaml")
	if got != "Config.Yaml deleted" {
		t.Errorf("T(i18n.done.delete, config.yaml) = %q, want 'Config.Yaml deleted'", got)
	}

	// Fail handler
	got = svc.T("i18n.fail.push", "commits")
	if got != "Failed to push commits" {
		t.Errorf("T(i18n.fail.push, commits) = %q, want 'Failed to push commits'", got)
	}
}

func TestServiceTDirectKeys(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Direct JSON keys
	got := svc.T("prompt.yes")
	if got != "y" {
		t.Errorf("T(prompt.yes) = %q, want 'y'", got)
	}

	got = svc.T("prompt.confirm")
	if got != "Are you sure?" {
		t.Errorf("T(prompt.confirm) = %q, want 'Are you sure?'", got)
	}

	got = svc.T("lang.de")
	if got != "German" {
		t.Errorf("T(lang.de) = %q, want 'German'", got)
	}
}

func TestServiceTPluralMessage(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// time.ago.second has one/other forms
	got := svc.T("time.ago.second", map[string]any{"Count": 1})
	if got != "1 second ago" {
		t.Errorf("T(time.ago.second, 1) = %q, want '1 second ago'", got)
	}

	got = svc.T("time.ago.second", map[string]any{"Count": 5})
	if got != "5 seconds ago" {
		t.Errorf("T(time.ago.second, 5) = %q, want '5 seconds ago'", got)
	}
}

func TestServiceRaw(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Raw bypasses handlers
	got := svc.Raw("prompt.yes")
	if got != "y" {
		t.Errorf("Raw(prompt.yes) = %q, want 'y'", got)
	}

	// Raw doesn't process i18n.* keys as handlers would
	got = svc.Raw("i18n.label.status")
	// Should return the key since it's not in the messages map
	if got != "i18n.label.status" {
		t.Errorf("Raw(i18n.label.status) = %q, want key returned", got)
	}
}

func TestServiceModes(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Default mode
	if svc.Mode() != ModeNormal {
		t.Errorf("default Mode() = %v, want ModeNormal", svc.Mode())
	}

	// Normal mode returns key for missing
	got := svc.T("nonexistent.key")
	if got != "nonexistent.key" {
		t.Errorf("ModeNormal missing key = %q, want key", got)
	}

	// Collect mode returns [key] and dispatches event
	svc.SetMode(ModeCollect)
	var missing MissingKey
	OnMissingKey(func(m MissingKey) { missing = m })
	got = svc.T("nonexistent.key")
	if got != "[nonexistent.key]" {
		t.Errorf("ModeCollect missing key = %q, want '[nonexistent.key]'", got)
	}
	if missing.Key != "nonexistent.key" {
		t.Errorf("MissingKey.Key = %q, want 'nonexistent.key'", missing.Key)
	}

	// Strict mode panics
	svc.SetMode(ModeStrict)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("ModeStrict should panic on missing key")
		}
	}()
	svc.T("nonexistent.key")
}

func TestServiceDebug(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	svc.SetDebug(true)
	got := svc.T("prompt.yes")
	if got != "[prompt.yes] y" {
		t.Errorf("debug T() = %q, want '[prompt.yes] y'", got)
	}

	svc.SetDebug(false)
	got = svc.T("prompt.yes")
	if got != "y" {
		t.Errorf("non-debug T() = %q, want 'y'", got)
	}
}

func TestServiceFormality(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.Formality() != FormalityNeutral {
		t.Errorf("default Formality() = %v, want FormalityNeutral", svc.Formality())
	}

	svc.SetFormality(FormalityFormal)
	if svc.Formality() != FormalityFormal {
		t.Errorf("Formality() = %v, want FormalityFormal", svc.Formality())
	}
}

func TestServiceDirection(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.Direction() != DirLTR {
		t.Errorf("Direction() = %v, want DirLTR", svc.Direction())
	}

	if svc.IsRTL() {
		t.Error("IsRTL() should be false for English")
	}
}

func TestServiceAddMessages(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"custom.greeting": "Hello!",
	})

	got := svc.T("custom.greeting")
	if got != "Hello!" {
		t.Errorf("T(custom.greeting) = %q, want 'Hello!'", got)
	}
}

func TestServiceHandlers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	initial := len(svc.Handlers())
	if initial != 6 {
		t.Errorf("default handlers = %d, want 6", initial)
	}

	svc.ClearHandlers()
	if len(svc.Handlers()) != 0 {
		t.Error("ClearHandlers() should remove all handlers")
	}
}

func TestServiceWithOptions(t *testing.T) {
	svc, err := New(
		WithFallback("en"),
		WithFormality(FormalityFormal),
		WithMode(ModeCollect),
		WithDebug(true),
	)
	if err != nil {
		t.Fatalf("New() with options failed: %v", err)
	}

	if svc.Formality() != FormalityFormal {
		t.Errorf("Formality = %v, want FormalityFormal", svc.Formality())
	}
	if svc.Mode() != ModeCollect {
		t.Errorf("Mode = %v, want ModeCollect", svc.Mode())
	}
	if !svc.Debug() {
		t.Error("Debug should be true")
	}
}

func TestNewWithFS(t *testing.T) {
	fs := fstest.MapFS{
		"i18n/custom.json": &fstest.MapFile{
			Data: []byte(`{"hello": "Hola!"}`),
		},
	}

	svc, err := NewWithFS(fs, "i18n", WithFallback("custom"))
	if err != nil {
		t.Fatalf("NewWithFS failed: %v", err)
	}

	got := svc.T("hello")
	if got != "Hola!" {
		t.Errorf("T(hello) = %q, want 'Hola!'", got)
	}
}

func TestServicePluralCategory(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.PluralCategory(1) != PluralOne {
		t.Errorf("PluralCategory(1) = %v, want PluralOne", svc.PluralCategory(1))
	}
	if svc.PluralCategory(5) != PluralOther {
		t.Errorf("PluralCategory(5) = %v, want PluralOther", svc.PluralCategory(5))
	}
}
