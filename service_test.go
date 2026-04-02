package i18n

import (
	"testing"
	"testing/fstest"
	"time"

	"dappco.re/go/core"
)

type messageBaseFallbackLoader struct{}

func (messageBaseFallbackLoader) Languages() []string {
	return []string{"en-GB", "en", "fr"}
}

func (messageBaseFallbackLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	return map[string]Message{}, nil, nil
}

type serviceMutatingHandler struct {
	svc *Service
}

func (h serviceMutatingHandler) Match(key string) bool {
	return key == "custom.mutate.language"
}

func (h serviceMutatingHandler) Handle(key string, args []any, next func() string) string {
	if h.svc != nil {
		_ = h.svc.SetLanguage("fr")
	}
	return "mutated"
}

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
	if got != "Config.yaml deleted" {
		t.Errorf("T(i18n.done.delete, config.yaml) = %q, want 'Config.yaml deleted'", got)
	}

	// Fail handler
	got = svc.T("i18n.fail.push", "commits")
	if got != "Failed to push commits" {
		t.Errorf("T(i18n.fail.push, commits) = %q, want 'Failed to push commits'", got)
	}
}

func TestServiceTranslate(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	result := svc.Translate("prompt.yes")
	if !result.OK {
		t.Fatalf("Translate(prompt.yes) returned not OK: %#v", result)
	}
	if got := result.Value; got != "y" {
		t.Fatalf("Translate(prompt.yes) = %#v, want %q", got, "y")
	}

	var _ core.Translator = (*Service)(nil)
}

func TestServiceTranslateMissingKey(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	result := svc.Translate("missing.translation.key")
	if result.OK {
		t.Fatalf("Translate(missing.translation.key) returned OK, want false: %#v", result)
	}
	if got, want := result.Value, "missing.translation.key"; got != want {
		t.Fatalf("Translate(missing.translation.key) = %#v, want %q", got, want)
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

func TestServiceTMapContextNestedExtra(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	svc.AddMessages("en", map[string]string{
		"welcome._nav._scope._admin": "Admin navigation",
	})

	got := svc.T("welcome", map[string]any{
		"Context": "nav",
		"Extra": map[string]any{
			"Scope": "admin",
		},
	})
	if got != "Admin navigation" {
		t.Fatalf("T(welcome, nested extra) = %q, want %q", got, "Admin navigation")
	}
}

func TestServiceTMapContextStringExtras(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	svc.AddMessages("en", map[string]string{
		"welcome._greeting":                   "hello",
		"welcome._greeting._region._europe":   "bonjour",
		"welcome._greeting._region._americas": "howdy",
	})

	if got := svc.T("welcome", map[string]string{"Context": "greeting"}); got != "hello" {
		t.Fatalf("T(welcome, map[string]string{Context:greeting}) = %q, want %q", got, "hello")
	}

	if got := svc.T("welcome", map[string]string{"Context": "greeting", "region": "europe"}); got != "bonjour" {
		t.Fatalf("T(welcome, map[string]string{Context:greeting region:europe}) = %q, want %q", got, "bonjour")
	}

	if got := svc.T("welcome", map[string]string{"Context": "greeting", "region": "americas"}); got != "howdy" {
		t.Fatalf("T(welcome, map[string]string{Context:greeting region:americas}) = %q, want %q", got, "howdy")
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

func TestServiceT_CustomHandlerCanMutateService(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.PrependHandler(serviceMutatingHandler{svc: svc})

	done := make(chan string, 1)
	go func() {
		done <- svc.T("custom.mutate.language")
	}()

	select {
	case got := <-done:
		if got != "mutated" {
			t.Fatalf("T(custom.mutate.language) = %q, want %q", got, "mutated")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("T(custom.mutate.language) timed out while handler mutated service state")
	}

	if got := svc.Language(); got != "fr" {
		t.Fatalf("Language() = %q, want %q", got, "fr")
	}
}

func TestServiceRaw_DoesNotUseCommonFallbacks(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.messages["en"]["common.action.status"] = Message{Text: "Common status"}

	got := svc.Raw("missing.status")
	if got != "missing.status" {
		t.Errorf("Raw(missing.status) = %q, want key returned", got)
	}
}

func TestServiceRaw_MissingKeyHandlersCanMutateService(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	prev := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prev)
	})

	OnMissingKey(func(m MissingKey) {
		_ = svc.SetLanguage("fr")
	})
	svc.SetMode(ModeCollect)
	svc.SetDebug(true)
	t.Cleanup(func() {
		svc.SetDebug(false)
	})

	done := make(chan string, 1)
	go func() {
		done <- svc.Raw("missing.raw.key")
	}()

	select {
	case got := <-done:
		if got != "[missing.raw.key] [missing.raw.key]" {
			t.Fatalf("Raw(missing.raw.key) = %q, want %q", got, "[missing.raw.key] [missing.raw.key]")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Raw(missing.raw.key) timed out while missing-key handler mutated service state")
	}

	if got := svc.Language(); got != "fr" {
		t.Fatalf("Language() = %q, want %q", got, "fr")
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

func TestServiceFallback(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.Fallback() != "en" {
		t.Errorf("default Fallback() = %q, want en", svc.Fallback())
	}

	svc.SetFallback("fr")
	if svc.Fallback() != "fr" {
		t.Errorf("Fallback() = %q, want fr", svc.Fallback())
	}
}

func TestServiceFallbackNormalisesLanguageTag(t *testing.T) {
	svc, err := New(WithFallback("fr_CA"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if got, want := svc.Fallback(), "fr-CA"; got != want {
		t.Fatalf("WithFallback(fr_CA) = %q, want %q", got, want)
	}

	svc.SetFallback("en_US")
	if got, want := svc.Fallback(), "en-US"; got != want {
		t.Fatalf("SetFallback(en_US) = %q, want %q", got, want)
	}
}

func TestServiceMessageFallbackUsesBaseLanguageTagBeforeConfiguredFallback(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"greeting": "hello",
	})
	svc.AddMessages("fr", map[string]string{
		"greeting": "bonjour",
	})

	if err := svc.SetLanguage("en-GB"); err != nil {
		t.Fatalf("SetLanguage(en-GB) failed: %v", err)
	}
	svc.SetFallback("fr")

	if got := svc.T("greeting"); got != "hello" {
		t.Fatalf("T(greeting) = %q, want %q", got, "hello")
	}
}

func TestServiceMessageFallbackUsesConfiguredFallbackBaseLanguageTag(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	svc.AddMessages("fr", map[string]string{
		"greeting": "bonjour",
	})

	if err := svc.SetLanguage("en-GB"); err != nil {
		t.Fatalf("SetLanguage(en-GB) failed: %v", err)
	}
	svc.SetFallback("fr-CA")

	if got := svc.T("greeting"); got != "bonjour" {
		t.Fatalf("T(greeting) = %q, want %q", got, "bonjour")
	}
}

func TestServiceCommonFallbackUsesBaseLanguageTagBeforeConfiguredFallback(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"common.action.status": "Common status",
	})
	svc.AddMessages("fr", map[string]string{
		"common.action.status": "Statut commun",
	})

	if err := svc.SetLanguage("en-GB"); err != nil {
		t.Fatalf("SetLanguage(en-GB) failed: %v", err)
	}
	svc.SetFallback("fr")

	if got := svc.T("missing.status"); got != "Common status" {
		t.Fatalf("T(missing.status) = %q, want %q", got, "Common status")
	}
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

func TestServiceLocation(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.Location() != "" {
		t.Errorf("default Location() = %q, want empty", svc.Location())
	}

	svc.SetLocation("workspace")
	if svc.Location() != "workspace" {
		t.Errorf("Location() = %q, want workspace", svc.Location())
	}
}

func TestServiceTranslationContext(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"direction.right._navigation":                   "right",
		"direction.right._navigation._formal":           "right, sir",
		"direction.right._navigation._feminine":         "right, ma'am",
		"direction.right._navigation._feminine._formal": "right, madam",
		"direction.right._correctness":                  "correct",
		"direction.right._correctness._formal":          "correct, sir",
		"welcome._workspace":                            "welcome aboard",
		"welcome._feminine":                             "welcome, ma'am",
	})

	if got := svc.T("direction.right", C("navigation")); got != "right" {
		t.Errorf("T(direction.right, C(navigation)) = %q, want %q", got, "right")
	}

	if got := svc.T("direction.right", C("navigation").Formal()); got != "right, sir" {
		t.Errorf("T(direction.right, C(navigation).Formal()) = %q, want %q", got, "right, sir")
	}

	if got := svc.T("direction.right", C("correctness")); got != "correct" {
		t.Errorf("T(direction.right, C(correctness)) = %q, want %q", got, "correct")
	}

	if got := svc.T("direction.right", C("correctness").Formal()); got != "correct, sir" {
		t.Errorf("T(direction.right, C(correctness).Formal()) = %q, want %q", got, "correct, sir")
	}

	if got := svc.T("direction.right", C("navigation").WithGender("feminine")); got != "right, ma'am" {
		t.Errorf("T(direction.right, C(navigation).WithGender(feminine)) = %q, want %q", got, "right, ma'am")
	}

	if got := svc.T("direction.right", C("navigation").WithGender("feminine").Formal()); got != "right, madam" {
		t.Errorf("T(direction.right, C(navigation).WithGender(feminine).Formal()) = %q, want %q", got, "right, madam")
	}

	if got := svc.T("welcome", C("greeting").In("workspace")); got != "welcome aboard" {
		t.Errorf("T(welcome, C(greeting).In(workspace)) = %q, want %q", got, "welcome aboard")
	}

	if got := svc.T("welcome", S("user", "Alice").Gender("feminine")); got != "welcome, ma'am" {
		t.Errorf("T(welcome, S(user, Alice).Gender(feminine)) = %q, want %q", got, "welcome, ma'am")
	}

	if got := svc.T("welcome", S("user", "Alice").In("workspace")); got != "welcome aboard" {
		t.Errorf("T(welcome, S(user, Alice).In(workspace)) = %q, want %q", got, "welcome aboard")
	}
}

func TestServiceDefaultLocationContext(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	svc.AddMessages("en", map[string]string{
		"welcome._workspace": "welcome aboard",
	})

	svc.SetLocation("workspace")

	if got := svc.T("welcome"); got != "welcome aboard" {
		t.Errorf("T(welcome) with default location = %q, want %q", got, "welcome aboard")
	}
}

func TestServiceTranslationContextExtrasInTemplates(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome": "Hello {{.name}} from {{.Context}} in {{.city}}",
	})

	ctx := C("greeting").
		Set("name", "World").
		Set("city", "Paris")

	got := svc.T("welcome", ctx)
	if got != "Hello World from greeting in Paris" {
		t.Errorf("T(welcome, ctx) = %q, want %q", got, "Hello World from greeting in Paris")
	}
}

func TestServiceTranslationContextExtrasInLookup(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome._greeting":                   "hello",
		"welcome._greeting._region._europe":   "bonjour",
		"welcome._greeting._region._americas": "howdy",
	})

	if got := svc.T("welcome", C("greeting")); got != "hello" {
		t.Errorf("T(welcome, C(greeting)) = %q, want %q", got, "hello")
	}

	if got := svc.T("welcome", C("greeting").Set("region", "europe")); got != "bonjour" {
		t.Errorf("T(welcome, C(greeting).Set(region, europe)) = %q, want %q", got, "bonjour")
	}

	if got := svc.T("welcome", C("greeting").Set("region", "americas")); got != "howdy" {
		t.Errorf("T(welcome, C(greeting).Set(region, americas)) = %q, want %q", got, "howdy")
	}
}

func TestServiceMapContextExtrasInLookup(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome._greeting":                   "hello",
		"welcome._greeting._region._europe":   "bonjour",
		"welcome._greeting._region._americas": "howdy",
	})

	if got := svc.T("welcome", map[string]any{"Context": "greeting"}); got != "hello" {
		t.Errorf("T(welcome, map[Context:greeting]) = %q, want %q", got, "hello")
	}

	if got := svc.T("welcome", map[string]any{"Context": "greeting", "region": "europe"}); got != "bonjour" {
		t.Errorf("T(welcome, map[Context:greeting region:europe]) = %q, want %q", got, "bonjour")
	}

	if got := svc.T("welcome", map[string]any{"Context": "greeting", "region": "americas"}); got != "howdy" {
		t.Errorf("T(welcome, map[Context:greeting region:americas]) = %q, want %q", got, "howdy")
	}
}

func TestServiceDefaultLocationAppliesToMapData(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome._workspace": "welcome aboard",
	})
	svc.SetLocation("workspace")

	if got := svc.T("welcome", map[string]any{}); got != "welcome aboard" {
		t.Errorf("T(welcome, map[]) with default location = %q, want %q", got, "welcome aboard")
	}
}

func TestServiceSubjectCountPlurals(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if err := svc.loadJSON("en", []byte(`{
		"item_count": {
			"one": "{{.Count}} item",
			"other": "{{.Count}} items"
		}
	}`)); err != nil {
		t.Fatalf("loadJSON() failed: %v", err)
	}

	if got := svc.T("item_count", S("item", "config.yaml").Count(1)); got != "1 item" {
		t.Errorf("T(item_count, Count(1)) = %q, want %q", got, "1 item")
	}

	if got := svc.T("item_count", S("item", "config.yaml").Count(3)); got != "3 items" {
		t.Errorf("T(item_count, Count(3)) = %q, want %q", got, "3 items")
	}
}

func TestServiceLoadJSONPartialVerbForms(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	const lang = "zz"
	prevDefault := Default()
	prevGrammar := GetGrammarData(lang)
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prevDefault)
		SetGrammarData(lang, prevGrammar)
	})

	svc.currentLang = lang

	if err := svc.loadJSON(lang, []byte(`{
		"gram": {
			"verb": {
				"render": { "past": "rendered" },
				"stream": { "gerund": "streaming" }
			}
		}
	}`)); err != nil {
		t.Fatalf("loadJSON() failed: %v", err)
	}

	if v, ok := GetGrammarData(lang).Verbs["render"]; !ok || v.Past != "rendered" || v.Gerund != "" {
		t.Fatalf("partial past verb not loaded correctly: %+v", v)
	}
	if v, ok := GetGrammarData(lang).Verbs["stream"]; !ok || v.Past != "" || v.Gerund != "streaming" {
		t.Fatalf("partial gerund verb not loaded correctly: %+v", v)
	}
	if got := PastTense("render"); got != "rendered" {
		t.Fatalf("PastTense(render) = %q, want %q", got, "rendered")
	}
	if got := Gerund("stream"); got != "streaming" {
		t.Fatalf("Gerund(stream) = %q, want %q", got, "streaming")
	}
}

func TestServiceTemplatesSupportGrammarFuncs(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"build.status": "{{past .Verb}} complete",
	})

	got := svc.T("build.status", map[string]any{"Verb": "build"})
	if got != "built complete" {
		t.Errorf("T(build.status) = %q, want %q", got, "built complete")
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

func TestServiceAddMessages_RegistersLanguage(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("fr_CA", map[string]string{
		"custom.greeting": "Salut!",
	})

	if err := svc.SetLanguage("fr_CA"); err != nil {
		t.Fatalf("SetLanguage(fr_CA) failed: %v", err)
	}

	if got := svc.T("custom.greeting"); got != "Salut!" {
		t.Fatalf("T(custom.greeting) = %q, want %q", got, "Salut!")
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

func TestWithDefaultHandlers_Idempotent(t *testing.T) {
	svc, err := New(WithDefaultHandlers())
	if err != nil {
		t.Fatalf("New() with WithDefaultHandlers() failed: %v", err)
	}

	if got := len(svc.Handlers()); got != 6 {
		t.Fatalf("len(Handlers()) = %d, want 6", got)
	}
}

func TestServiceWithOptions(t *testing.T) {
	svc, err := New(
		WithFallback("en"),
		WithFormality(FormalityFormal),
		WithLocation("workspace"),
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
	if svc.Location() != "workspace" {
		t.Errorf("Location = %q, want workspace", svc.Location())
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

func TestServiceAddLoader_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	extra := fstest.MapFS{
		"en.json": &fstest.MapFile{
			Data: []byte(`{"extra.key": "extra value"}`),
		},
	}
	if err := svc.AddLoader(NewFSLoader(extra, ".")); err != nil {
		t.Fatalf("AddLoader() failed: %v", err)
	}

	got := svc.T("extra.key")
	if got != "extra value" {
		t.Errorf("T(extra.key) = %q, want 'extra value'", got)
	}
}

func TestServiceAddLoader_Bad(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Loader with a valid dir listing but broken JSON
	broken := fstest.MapFS{
		"broken.json": &fstest.MapFile{
			Data: []byte(`{invalid json}`),
		},
	}
	if err := svc.AddLoader(NewFSLoader(broken, ".")); err == nil {
		t.Error("AddLoader() should fail with invalid JSON")
	}
}

func TestPackageLevelAddLoader(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	extra := fstest.MapFS{
		"en.json": &fstest.MapFile{
			Data: []byte(`{"pkg.hello": "from package"}`),
		},
	}
	AddLoader(NewFSLoader(extra, "."))

	got := T("pkg.hello")
	if got != "from package" {
		t.Errorf("T(pkg.hello) = %q, want 'from package'", got)
	}
}

func TestServiceLoadFS_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	extra := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"loaded": "yes"}`),
		},
	}
	if err := svc.LoadFS(extra, "locales"); err != nil {
		t.Fatalf("LoadFS() failed: %v", err)
	}

	got := svc.T("loaded")
	if got != "yes" {
		t.Errorf("T(loaded) = %q, want 'yes'", got)
	}
}

func TestServiceLoadFS_Bad(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	empty := fstest.MapFS{}
	if err := svc.LoadFS(empty, "nonexistent"); err == nil {
		t.Error("LoadFS() should fail with bad directory")
	}
}

func TestNewWithLoaderNoLanguages(t *testing.T) {
	empty := fstest.MapFS{}
	_, err := NewWithFS(empty, "empty")
	if err == nil {
		t.Error("NewWithFS with empty dir should fail")
	}
}

func TestServiceIsRTL(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if svc.IsRTL() {
		t.Error("IsRTL() should be false for English")
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
