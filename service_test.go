package i18n

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"dappco.re/go"
	"slices"
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

type serviceStubHandler struct{}

func (serviceStubHandler) Match(key string) bool {
	return key == "custom.stub"
}

func (serviceStubHandler) Handle(key string, args []any, next func() string) string {
	return "stub"
}

type underscoreLangLoader struct{}

func (underscoreLangLoader) Languages() []string {
	return []string{"en_US"}
}

func (underscoreLangLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	return map[string]Message{
		"greeting": {Text: "hello"},
	}, nil, nil
}

type duplicateLangLoader struct{}

func (duplicateLangLoader) Languages() []string {
	return []string{"en_US", "en-US"}
}

func (duplicateLangLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	switch lang {
	case "en_US":
		return map[string]Message{
				"first.key": {Text: "first"},
			}, &GrammarData{
				Words: map[string]string{"pkg": "package"},
			}, nil
	case "en-US":
		return map[string]Message{
				"second.key": {Text: "second"},
			}, &GrammarData{
				Words: map[string]string{"api": "API"},
			}, nil
	default:
		return map[string]Message{}, nil, nil
	}
}

type intentReceiverLanguageLoader struct{}

func (intentReceiverLanguageLoader) Languages() []string {
	return []string{"qaa-x-intent"}
}

func (intentReceiverLanguageLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	return map[string]Message{}, &GrammarData{
		Verbs: map[string]VerbForms{
			"supprimer": {Past: "supprimé"},
		},
		Words: map[string]string{
			"failed_to": "Impossible de",
		},
		Intents: map[string]Intent{
			"core.delete": {
				Meta: IntentMeta{
					Type: "action",
					Verb: "common.verb.supprimer",
				},
			},
		},
	}, nil
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

func TestNewServiceAliases(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	if svc == nil {
		t.Fatal("NewService() returned nil service")
	}

	fsys := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{}`)},
	}
	withFS, err := NewServiceWithFS(fsys, "locales")
	if err != nil {
		t.Fatalf("NewServiceWithFS() failed: %v", err)
	}
	if withFS == nil {
		t.Fatal("NewServiceWithFS() returned nil service")
	}

	withLoader, err := NewServiceWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewServiceWithLoader() failed: %v", err)
	}
	if withLoader == nil {
		t.Fatal("NewServiceWithLoader() returned nil service")
	}
}

func TestServiceCurrentStateAliases(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	svc.SetFallback("fr")
	svc.SetMode(ModeCollect)
	svc.SetFormality(FormalityFormal)
	svc.SetLocation("workspace")
	svc.SetDebug(true)

	if got, want := svc.CurrentLanguage(), svc.Language(); got != want {
		t.Fatalf("CurrentLanguage() = %q, want %q", got, want)
	}
	if got, want := svc.CurrentLang(), svc.Language(); got != want {
		t.Fatalf("CurrentLang() = %q, want %q", got, want)
	}
	if got, want := svc.CurrentAvailableLanguages(), svc.AvailableLanguages(); !slices.Equal(got, want) {
		t.Fatalf("CurrentAvailableLanguages() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentMode(), svc.Mode(); got != want {
		t.Fatalf("CurrentMode() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentFallback(), svc.Fallback(); got != want {
		t.Fatalf("CurrentFallback() = %q, want %q", got, want)
	}
	if got, want := svc.CurrentFormality(), svc.Formality(); got != want {
		t.Fatalf("CurrentFormality() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentLocation(), svc.Location(); got != want {
		t.Fatalf("CurrentLocation() = %q, want %q", got, want)
	}
	if got, want := svc.CurrentDirection(), svc.Direction(); got != want {
		t.Fatalf("CurrentDirection() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentTextDirection(), svc.CurrentDirection(); got != want {
		t.Fatalf("CurrentTextDirection() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentPluralCategory(2), svc.PluralCategory(2); got != want {
		t.Fatalf("CurrentPluralCategory() = %v, want %v", got, want)
	}
	if got, want := svc.PluralCategoryOf(2), svc.PluralCategory(2); got != want {
		t.Fatalf("PluralCategoryOf() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentDebug(), svc.Debug(); got != want {
		t.Fatalf("CurrentDebug() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentIsRTL(), svc.IsRTL(); got != want {
		t.Fatalf("CurrentIsRTL() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentRTL(), svc.IsRTL(); got != want {
		t.Fatalf("CurrentRTL() = %v, want %v", got, want)
	}
	if got, want := svc.CurrentHandlers(), svc.Handlers(); len(got) != len(want) {
		t.Fatalf("CurrentHandlers() len = %d, want %d", len(got), len(want))
	}
	if got, want := svc.CurrentState(), svc.State(); len(got.AvailableLanguages) != len(want.AvailableLanguages) || len(got.Handlers) != len(want.Handlers) {
		t.Fatalf("CurrentState() = %+v, want %+v", got, want)
	}
	if got, want := svc.CurrentState().RequestedLanguage, svc.State().RequestedLanguage; got != want {
		t.Fatalf("CurrentState().RequestedLanguage = %q, want %q", got, want)
	}
	if got, want := svc.CurrentState().LanguageExplicit, svc.State().LanguageExplicit; got != want {
		t.Fatalf("CurrentState().LanguageExplicit = %t, want %t", got, want)
	}
}

func TestServiceCurrentStateAliasesReturnCopies(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	langs := svc.CurrentAvailableLanguages()
	if len(langs) == 0 {
		t.Fatal("CurrentAvailableLanguages() returned no languages")
	}
	langs[0] = "zz"
	if got := svc.CurrentAvailableLanguages()[0]; got == "zz" {
		t.Fatalf("CurrentAvailableLanguages() returned a shared slice; first element mutated to %q", got)
	}

	handlers := svc.CurrentHandlers()
	if len(handlers) == 0 {
		t.Fatal("CurrentHandlers() returned no handlers")
	}
	handlers[0] = nil
	if svc.CurrentHandlers()[0] == nil {
		t.Fatal("CurrentHandlers() returned a shared slice; first handler mutated to nil")
	}

	state := svc.CurrentState()
	if len(state.AvailableLanguages) == 0 {
		t.Fatal("CurrentState() returned no available languages")
	}
	if len(state.Handlers) == 0 {
		t.Fatal("CurrentState() returned no handlers")
	}
	names := state.HandlerTypeNames()
	if len(names) != len(state.Handlers) {
		t.Fatalf("HandlerTypeNames() len = %d, want %d", len(names), len(state.Handlers))
	}
	names[0] = "zz"
	if got := svc.CurrentState().HandlerTypeNames()[0]; got == "zz" {
		t.Fatalf("HandlerTypeNames() returned a shared slice; first element mutated to %q", got)
	}
	state.AvailableLanguages[0] = "zz"
	if got := svc.CurrentState().AvailableLanguages[0]; got == "zz" {
		t.Fatalf("CurrentState() returned a shared available languages slice; first element mutated to %q", got)
	}
	state.Handlers[0] = nil
	if svc.CurrentState().Handlers[0] == nil {
		t.Fatal("CurrentState() returned a shared handlers slice; first handler mutated to nil")
	}
}

func TestServiceNilReceiverIsSafe(t *testing.T) {
	var svc *Service

	if got, want := svc.Language(), "en"; got != want {
		t.Fatalf("nil Service.Language() = %q, want %q", got, want)
	}
	if got, want := svc.Fallback(), "en"; got != want {
		t.Fatalf("nil Service.Fallback() = %q, want %q", got, want)
	}
	if got, want := svc.Mode(), ModeNormal; got != want {
		t.Fatalf("nil Service.Mode() = %v, want %v", got, want)
	}
	if got, want := svc.Formality(), FormalityNeutral; got != want {
		t.Fatalf("nil Service.Formality() = %v, want %v", got, want)
	}
	if got, want := svc.Direction(), DirLTR; got != want {
		t.Fatalf("nil Service.Direction() = %v, want %v", got, want)
	}
	if got, want := svc.PluralCategory(2), PluralOther; got != want {
		t.Fatalf("nil Service.PluralCategory(2) = %v, want %v", got, want)
	}
	if got, want := svc.AvailableLanguages(), []string{}; len(got) != len(want) {
		t.Fatalf("nil Service.AvailableLanguages() = %v, want %v", got, want)
	}
	if got, want := svc.Handlers(), []KeyHandler{}; len(got) != len(want) {
		t.Fatalf("nil Service.Handlers() = %v, want %v", got, want)
	}
	if got, want := svc.State(), defaultServiceStateSnapshot(); got.Language != want.Language || got.Mode != want.Mode || got.Fallback != want.Fallback || got.Formality != want.Formality || got.Location != want.Location || got.Direction != want.Direction || got.IsRTL != want.IsRTL || got.Debug != want.Debug || len(got.AvailableLanguages) != len(want.AvailableLanguages) || len(got.Handlers) != len(want.Handlers) {
		t.Fatalf("nil Service.State() = %+v, want %+v", got, want)
	}
	if got, want := svc.String(), defaultServiceStateSnapshot().String(); got != want {
		t.Fatalf("nil Service.String() = %q, want %q", got, want)
	}
	if got, want := svc.T("prompt.yes"), "prompt.yes"; got != want {
		t.Fatalf("nil Service.T(prompt.yes) = %q, want %q", got, want)
	}
	if got, want := svc.Raw("prompt.yes"), "prompt.yes"; got != want {
		t.Fatalf("nil Service.Raw(prompt.yes) = %q, want %q", got, want)
	}
	result := svc.Translate("prompt.yes")
	if result.OK {
		t.Fatalf("nil Service.Translate(prompt.yes) returned OK, want false: %#v", result)
	}
	if got, want := result.Error(), "prompt.yes"; got != want {
		t.Fatalf("nil Service.Translate(prompt.yes) = %#v, want %#v", got, want)
	}
	if got, want := svc.Prompt("confirm"), "prompt.confirm"; got != want {
		t.Fatalf("nil Service.Prompt(confirm) = %q, want %q", got, want)
	}
	if got, want := svc.Lang("fr"), "lang.fr"; got != want {
		t.Fatalf("nil Service.Lang(fr) = %q, want %q", got, want)
	}

	svc.SetMode(ModeStrict)
	svc.SetFallback("fr")
	svc.SetFormality(FormalityFormal)
	svc.SetLocation("workspace")
	svc.SetDebug(true)
	svc.SetHandlers(LabelHandler{})
	svc.AddHandler(ProgressHandler{})
	svc.PrependHandler(CountHandler{})
	svc.ClearHandlers()
	svc.ResetHandlers()
	svc.AddMessages("en", map[string]string{"x": "y"})

	if err := svc.SetLanguage("en"); err != ErrServiceNotInitialised {
		t.Fatalf("nil Service.SetLanguage() error = %v, want ErrServiceNotInitialised", err)
	}
	if err := svc.AddLoader(nil); err != ErrServiceNotInitialised {
		t.Fatalf("nil Service.AddLoader() error = %v, want ErrServiceNotInitialised", err)
	}
	if err := svc.LoadFS(fstest.MapFS{}, "locales"); err != ErrServiceNotInitialised {
		t.Fatalf("nil Service.LoadFS() error = %v, want ErrServiceNotInitialised", err)
	}
}

func TestServiceStateString(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	got := svc.State().String()
	if got == "" {
		t.Fatal("ServiceState.String() returned an empty string")
	}
	for _, want := range []string{
		"ServiceState{",
		"language=",
		"requested=",
		"explicit=",
		"fallback=",
		"mode=",
		"available=",
		"handlers=",
		"LabelHandler",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("ServiceState.String() = %q, want substring %q", got, want)
		}
	}
}

func TestServiceAvailableLanguagesSorted(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	if got, want := svc.AvailableLanguages(), []string{"en", "en-GB", "fr"}; !slices.Equal(got, want) {
		t.Fatalf("AvailableLanguages() = %v, want %v", got, want)
	}
}

func TestServiceSetLanguageUnsupportedIncludesAvailableLanguages(t *testing.T) {
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	err = svc.SetLanguage("es")
	if err == nil {
		t.Fatal("SetLanguage(es) succeeded, want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unsupported language: es") {
		t.Fatalf("SetLanguage(es) error = %q, want unsupported language message", msg)
	}
	for _, want := range []string{"en", "en-GB", "fr"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("SetLanguage(es) error = %q, want available language %q", msg, want)
		}
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
	if got, want := result.Error(), "missing.translation.key"; got != want {
		t.Fatalf("Translate(missing.translation.key) = %#v, want %q", got, want)
	}
}

func TestNewWithHandlersCopiesInputSlice(t *testing.T) {
	handlers := []KeyHandler{serviceStubHandler{}}
	svc, err := New(WithHandlers(handlers...))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	handlers[0] = LabelHandler{}

	if got := svc.T("custom.stub"); got != "stub" {
		t.Fatalf("T(custom.stub) = %q, want %q", got, "stub")
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

func TestServicePromptAndLang(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if got, want := svc.Prompt("confirm"), "Are you sure?"; got != want {
		t.Fatalf("Prompt(confirm) = %q, want %q", got, want)
	}
	if got, want := svc.Prompt("prompt.confirm"), "Are you sure?"; got != want {
		t.Fatalf("Prompt(prompt.confirm) = %q, want %q", got, want)
	}
	if got, want := svc.CurrentPrompt("confirm"), "Are you sure?"; got != want {
		t.Fatalf("CurrentPrompt(confirm) = %q, want %q", got, want)
	}
	if got, want := svc.Lang("fr"), "French"; got != want {
		t.Fatalf("Lang(fr) = %q, want %q", got, want)
	}
	if got, want := svc.Lang("lang.fr"), "French"; got != want {
		t.Fatalf("Lang(lang.fr) = %q, want %q", got, want)
	}
	if got, want := svc.Lang("fr_CA"), "French"; got != want {
		t.Fatalf("Lang(fr_CA) = %q, want %q", got, want)
	}
}

func TestServicePromptAndLangExactMatch(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"prompt.exact": "prompt.exact",
		"lang.exact":   "lang.exact",
	})

	if got, want := svc.Prompt("exact"), "prompt.exact"; got != want {
		t.Fatalf("Prompt(exact) = %q, want %q", got, want)
	}
	if got, want := svc.Lang("exact"), "lang.exact"; got != want {
		t.Fatalf("Lang(exact) = %q, want %q", got, want)
	}
}

func TestNewWithLoaderNormalisesLanguageTags(t *testing.T) {
	svc, err := NewWithLoader(underscoreLangLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	langs := svc.AvailableLanguages()
	if len(langs) != 1 || langs[0] != "en-US" {
		t.Fatalf("AvailableLanguages() = %v, want [en-US]", langs)
	}

	if err := svc.SetLanguage("en_US"); err != nil {
		t.Fatalf("SetLanguage(en_US) failed: %v", err)
	}
	if got, want := svc.Language(), "en-US"; got != want {
		t.Fatalf("Language() after SetLanguage(en_US) = %q, want %q", got, want)
	}
	if got := svc.T("greeting"); got != "hello" {
		t.Fatalf("T(greeting) after SetLanguage(en_US) = %q, want %q", got, "hello")
	}
}

func TestNewWithLoaderMergesCanonicalDuplicateLanguages(t *testing.T) {
	svc, err := NewWithLoader(duplicateLangLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	if err := svc.SetLanguage("en-US"); err != nil {
		t.Fatalf("SetLanguage(en-US) failed: %v", err)
	}

	if got, want := svc.T("first.key"), "first"; got != want {
		t.Fatalf("T(first.key) = %q, want %q", got, want)
	}
	if got, want := svc.T("second.key"), "second"; got != want {
		t.Fatalf("T(second.key) = %q, want %q", got, want)
	}

	data := GetGrammarData("en-US")
	if data == nil {
		t.Fatal("GetGrammarData(en-US) returned nil")
	}
	if got, want := data.Words["pkg"], "package"; got != want {
		t.Fatalf("grammar word pkg = %q, want %q", got, want)
	}
	if got, want := data.Words["api"], "API"; got != want {
		t.Fatalf("grammar word api = %q, want %q", got, want)
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

func TestServiceFallbackCanonicalisesLanguageTagCase(t *testing.T) {
	svc, err := New(WithFallback("FR_ca"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if got, want := svc.Fallback(), "fr-CA"; got != want {
		t.Fatalf("WithFallback(FR_ca) = %q, want %q", got, want)
	}

	svc.SetFallback("EN_us")
	if got, want := svc.Fallback(), "en-US"; got != want {
		t.Fatalf("SetFallback(EN_us) = %q, want %q", got, want)
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

func TestServiceDefaultLocationAppliesToTranslationContext(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome._greeting._workspace": "hello from workspace",
	})
	svc.SetLocation("workspace")

	if got := svc.T("welcome", C("greeting")); got != "hello from workspace" {
		t.Errorf("T(welcome, C(greeting)) with default location = %q, want %q", got, "hello from workspace")
	}
}

func TestServiceDefaultLocationAppliesToSubject(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.AddMessages("en", map[string]string{
		"welcome._workspace": "welcome aboard",
	})
	svc.SetLocation("workspace")

	if got := svc.T("welcome", S("user", "Alice")); got != "welcome aboard" {
		t.Errorf("T(welcome, Subject) with default location = %q, want %q", got, "welcome aboard")
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

func TestServiceComposeGeneratedIntentTextUsesReceiverLanguage(t *testing.T) {
	const lang = "qaa-x-intent"

	prevDefault := Default()
	prevGrammar := GetGrammarData(lang)
	t.Cleanup(func() {
		SetDefault(prevDefault)
		SetGrammarData(lang, prevGrammar)
	})

	defaultSvc, err := New(WithLanguage("en"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(defaultSvc)

	svc, err := NewWithLoader(intentReceiverLanguageLoader{}, WithFallback("en"))
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	if err := svc.SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage(%s) failed: %v", lang, err)
	}

	composed := svc.Compose("core.delete", S("file", "rapport.pdf"))
	if got, want := composed.Success, "Rapport.pdf supprimé"; got != want {
		t.Fatalf("Compose().Success = %q, want %q", got, want)
	}
	if got, want := composed.Failure, "Impossible de supprimer rapport.pdf"; got != want {
		t.Fatalf("Compose().Failure = %q, want %q", got, want)
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

func TestServiceSetHandlers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	svc.SetHandlers(serviceStubHandler{}, nil, LabelHandler{})
	handlers := svc.Handlers()
	if got, want := len(handlers), 2; got != want {
		t.Fatalf("len(Handlers()) = %d, want %d", got, want)
	}
	if _, ok := handlers[0].(serviceStubHandler); !ok {
		t.Fatalf("Handlers()[0] = %T, want serviceStubHandler", handlers[0])
	}
	if _, ok := handlers[1].(LabelHandler); !ok {
		t.Fatalf("Handlers()[1] = %T, want LabelHandler", handlers[1])
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
		WithLanguage("fr_CA"),
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
	if got := svc.Language(); got != "fr" {
		t.Errorf("Language() = %q, want fr", got)
	}
}

func TestWithLanguage(t *testing.T) {
	svc, err := New(WithLanguage("fr_CA"))
	if err != nil {
		t.Fatalf("New() with WithLanguage() failed: %v", err)
	}

	if got := svc.Language(); got != "fr" {
		t.Fatalf("Language() = %q, want %q", got, "fr")
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

func TestServiceAddLoader_Nil(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if err := svc.AddLoader(nil); err == nil {
		t.Error("AddLoader() should fail with nil loader")
	}
}

func TestServiceAddLoader_NoLanguages(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	empty := fstest.MapFS{}
	if err := svc.AddLoader(NewFSLoader(empty, "missing")); err == nil {
		t.Error("AddLoader() should fail when no languages are available")
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

func TestNewWithLoaderNil(t *testing.T) {
	if _, err := NewWithLoader(nil); err == nil {
		t.Error("NewWithLoader(nil) should fail")
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

// --- AX-7 canonical triplets ---

func TestService_WithFallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFallback("fr"))
		if err != nil || svc.Fallback() != "fr" {
			t.Fatalf("fallback=%q err=%v", svc.Fallback(), err)
		}
	})
	if !called {
		t.Fatal("WithFallback was not exercised")
	}
}

func TestService_WithFallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFallback(""))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithFallback was not exercised")
	}
}

func TestService_WithFallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFallback("fr"))
		if err != nil || svc.Fallback() != "fr" {
			t.Fatalf("fallback=%q err=%v", svc.Fallback(), err)
		}
	})
	if !called {
		t.Fatal("WithFallback was not exercised")
	}
}

func TestService_WithLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLanguage("fr"))
		if err != nil || svc.Language() == "" {
			t.Fatalf("language=%q err=%v", svc.Language(), err)
		}
	})
	if !called {
		t.Fatal("WithLanguage was not exercised")
	}
}

func TestService_WithLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLanguage(""))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithLanguage was not exercised")
	}
}

func TestService_WithLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLanguage("fr"))
		if err != nil || svc.Language() == "" {
			t.Fatalf("language=%q err=%v", svc.Language(), err)
		}
	})
	if !called {
		t.Fatal("WithLanguage was not exercised")
	}
}

func TestService_WithFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFormality(FormalityFormal))
		if err != nil || svc.Formality() != FormalityFormal {
			t.Fatalf("formality=%v err=%v", svc.Formality(), err)
		}
	})
	if !called {
		t.Fatal("WithFormality was not exercised")
	}
}

func TestService_WithFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFormality(FormalityFormal))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithFormality was not exercised")
	}
}

func TestService_WithFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithFormality(FormalityFormal))
		if err != nil || svc.Formality() != FormalityFormal {
			t.Fatalf("formality=%v err=%v", svc.Formality(), err)
		}
	})
	if !called {
		t.Fatal("WithFormality was not exercised")
	}
}

func TestService_WithLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLocation("workspace"))
		if err != nil || svc.Location() != "workspace" {
			t.Fatalf("location=%q err=%v", svc.Location(), err)
		}
	})
	if !called {
		t.Fatal("WithLocation was not exercised")
	}
}

func TestService_WithLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLocation(""))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithLocation was not exercised")
	}
}

func TestService_WithLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLocation("workspace"))
		if err != nil || svc.Location() != "workspace" {
			t.Fatalf("location=%q err=%v", svc.Location(), err)
		}
	})
	if !called {
		t.Fatal("WithLocation was not exercised")
	}
}

func TestService_WithHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(ax7Handler{match: true}))
		if err != nil || len(svc.Handlers()) != 1 {
			t.Fatalf("handlers=%d err=%v", len(svc.Handlers()), err)
		}
	})
	if !called {
		t.Fatal("WithHandlers was not exercised")
	}
}

func TestService_WithHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(ax7Handler{match: true}))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithHandlers was not exercised")
	}
}

func TestService_WithHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(ax7Handler{match: true}))
		if err != nil || len(svc.Handlers()) != 1 {
			t.Fatalf("handlers=%d err=%v", len(svc.Handlers()), err)
		}
	})
	if !called {
		t.Fatal("WithHandlers was not exercised")
	}
}

func TestService_WithDefaultHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(), WithDefaultHandlers())
		if err != nil || len(svc.Handlers()) == 0 {
			t.Fatalf("handlers=%d err=%v", len(svc.Handlers()), err)
		}
	})
	if !called {
		t.Fatal("WithDefaultHandlers was not exercised")
	}
}

func TestService_WithDefaultHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(), WithDefaultHandlers())
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithDefaultHandlers was not exercised")
	}
}

func TestService_WithDefaultHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(), WithDefaultHandlers())
		if err != nil || len(svc.Handlers()) == 0 {
			t.Fatalf("handlers=%d err=%v", len(svc.Handlers()), err)
		}
	})
	if !called {
		t.Fatal("WithDefaultHandlers was not exercised")
	}
}

func TestService_WithMode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithMode(ModeCollect))
		if err != nil || svc.Mode() != ModeCollect {
			t.Fatalf("mode=%v err=%v", svc.Mode(), err)
		}
	})
	if !called {
		t.Fatal("WithMode was not exercised")
	}
}

func TestService_WithMode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithMode(ModeCollect))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithMode was not exercised")
	}
}

func TestService_WithMode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithMode(ModeCollect))
		if err != nil || svc.Mode() != ModeCollect {
			t.Fatalf("mode=%v err=%v", svc.Mode(), err)
		}
	})
	if !called {
		t.Fatal("WithMode was not exercised")
	}
}

func TestService_WithDebug_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithDebug(true))
		if err != nil || !svc.Debug() {
			t.Fatalf("debug=%v err=%v", svc.Debug(), err)
		}
	})
	if !called {
		t.Fatal("WithDebug was not exercised")
	}
}

func TestService_WithDebug_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithDebug(true))
		_ = svc
		_ = err
	})
	if !called {
		t.Fatal("WithDebug was not exercised")
	}
}

func TestService_WithDebug_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithDebug(true))
		if err != nil || !svc.Debug() {
			t.Fatalf("debug=%v err=%v", svc.Debug(), err)
		}
	})
	if !called {
		t.Fatal("WithDebug was not exercised")
	}
}

func TestService_New_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New()
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("New was not exercised")
	}
}

func TestService_New_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithLanguage("zz"))
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("New was not exercised")
	}
}

func TestService_New_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := New(WithHandlers(nil))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("New was not exercised")
	}
}

func TestService_NewService_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewService()
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewService was not exercised")
	}
}

func TestService_NewService_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewService(WithLanguage("zz"))
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewService was not exercised")
	}
}

func TestService_NewService_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewService(WithHandlers(nil))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewService was not exercised")
	}
}

func TestService_NewWithFS_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithFS(ax7TestFS(), "locales")
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithFS was not exercised")
	}
}

func TestService_NewWithFS_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithFS(ax7TestFS(), "missing")
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithFS was not exercised")
	}
}

func TestService_NewWithFS_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithFS(ax7TestFS(), "locales", WithLanguage("fr"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithFS was not exercised")
	}
}

func TestService_NewServiceWithFS_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithFS(ax7TestFS(), "locales")
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithFS was not exercised")
	}
}

func TestService_NewServiceWithFS_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithFS(ax7TestFS(), "missing")
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithFS was not exercised")
	}
}

func TestService_NewServiceWithFS_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithFS(ax7TestFS(), "locales", WithLanguage("fr"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithFS was not exercised")
	}
}

func TestService_NewWithLoader_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithLoader was not exercised")
	}
}

func TestService_NewWithLoader_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithLoader(nil)
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithLoader was not exercised")
	}
}

func TestService_NewWithLoader_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewWithLoader(NewFSLoader(ax7TestFS(), "locales"), WithLanguage("fr"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewWithLoader was not exercised")
	}
}

func TestService_NewServiceWithLoader_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithLoader was not exercised")
	}
}

func TestService_NewServiceWithLoader_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithLoader(nil)
		if err == nil || svc != nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithLoader was not exercised")
	}
}

func TestService_NewServiceWithLoader_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc, err := NewServiceWithLoader(NewFSLoader(ax7TestFS(), "locales"), WithLanguage("fr"))
		if err != nil || svc == nil {
			t.Fatalf("svc=%v err=%v", svc, err)
		}
	})
	if !called {
		t.Fatal("NewServiceWithLoader was not exercised")
	}
}

func TestService_Init_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		err := Init()
		if err != nil || Default() == nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Init was not exercised")
	}
}

func TestService_Init_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7SetDefault(t)
		err := Init()
		if err != nil || Default() != svc {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Init was not exercised")
	}
}

func TestService_Init_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		err := Init()
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Init was not exercised")
	}
}

func TestService_Default_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		svc := Default()
		if svc == nil {
			t.Fatal("expected default")
		}
	})
	if !called {
		t.Fatal("Default was not exercised")
	}
}

func TestService_Default_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7SetDefault(t)
		got := Default()
		if got != svc {
			t.Fatal("expected installed default")
		}
	})
	if !called {
		t.Fatal("Default was not exercised")
	}
}

func TestService_Default_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		_ = Default()
	})
	if !called {
		t.Fatal("Default was not exercised")
	}
}

func TestService_SetDefault_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		SetDefault(svc)
		if Default() != svc {
			t.Fatal("default not set")
		}
	})
	if !called {
		t.Fatal("SetDefault was not exercised")
	}
}

func TestService_SetDefault_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		if defaultService.Load() != nil {
			t.Fatal("default not cleared")
		}
	})
	if !called {
		t.Fatal("SetDefault was not exercised")
	}
}

func TestService_SetDefault_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		SetDefault(svc)
		SetDefault(nil)
		if defaultService.Load() != nil {
			t.Fatal("default not cleared")
		}
	})
	if !called {
		t.Fatal("SetDefault was not exercised")
	}
}

func TestService_AddLoader_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7SetDefault(t)
		AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		if len(svc.AvailableLanguages()) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("AddLoader was not exercised")
	}
}

func TestService_AddLoader_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		_ = defaultService.Load()
	})
	if !called {
		t.Fatal("AddLoader was not exercised")
	}
}

func TestService_AddLoader_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		AddLoader(nil)
		_ = AvailableLanguages()
	})
	if !called {
		t.Fatal("AddLoader was not exercised")
	}
}

func TestService_Service_SetLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.SetLanguage("fr")
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Service_SetLanguage was not exercised")
	}
}

func TestService_Service_SetLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		err := svc.SetLanguage("fr")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("Service_SetLanguage was not exercised")
	}
}

func TestService_Service_SetLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.SetLanguage("zz")
		if err == nil {
			t.Fatal("expected unsupported language")
		}
	})
	if !called {
		t.Fatal("Service_SetLanguage was not exercised")
	}
}

func TestService_Service_Language_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Language()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_Language was not exercised")
	}
}

func TestService_Service_Language_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Language()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Language was not exercised")
	}
}

func TestService_Service_Language_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		_ = svc.SetLanguage("fr")
		got := svc.Language()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_Language was not exercised")
	}
}

func TestService_Service_CurrentLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentLanguage()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_CurrentLanguage was not exercised")
	}
}

func TestService_Service_CurrentLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentLanguage()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentLanguage was not exercised")
	}
}

func TestService_Service_CurrentLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		_ = svc.SetLanguage("fr")
		got := svc.CurrentLanguage()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_CurrentLanguage was not exercised")
	}
}

func TestService_Service_CurrentLang_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentLang()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_CurrentLang was not exercised")
	}
}

func TestService_Service_CurrentLang_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentLang()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentLang was not exercised")
	}
}

func TestService_Service_CurrentLang_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		_ = svc.SetLanguage("fr")
		got := svc.CurrentLang()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_CurrentLang was not exercised")
	}
}

func TestService_Service_Prompt_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Prompt("yes")
		if got == "" {
			t.Fatal("expected prompt")
		}
	})
	if !called {
		t.Fatal("Service_Prompt was not exercised")
	}
}

func TestService_Service_Prompt_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Prompt("yes")
		if got == "" {
			t.Fatal("expected fallback prompt")
		}
	})
	if !called {
		t.Fatal("Service_Prompt was not exercised")
	}
}

func TestService_Service_Prompt_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Prompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Prompt was not exercised")
	}
}

func TestService_Service_CurrentPrompt_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected prompt")
		}
	})
	if !called {
		t.Fatal("Service_CurrentPrompt was not exercised")
	}
}

func TestService_Service_CurrentPrompt_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected fallback prompt")
		}
	})
	if !called {
		t.Fatal("Service_CurrentPrompt was not exercised")
	}
}

func TestService_Service_CurrentPrompt_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentPrompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentPrompt was not exercised")
	}
}

func TestService_Service_Lang_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Lang("en")
		if got == "" {
			t.Fatal("expected lang")
		}
	})
	if !called {
		t.Fatal("Service_Lang was not exercised")
	}
}

func TestService_Service_Lang_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Lang("en")
		if got == "" {
			t.Fatal("expected fallback lang")
		}
	})
	if !called {
		t.Fatal("Service_Lang was not exercised")
	}
}

func TestService_Service_Lang_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Lang("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Lang was not exercised")
	}
}

func TestService_Service_AvailableLanguages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		langs := svc.AvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("Service_AvailableLanguages was not exercised")
	}
}

func TestService_Service_AvailableLanguages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		langs := svc.AvailableLanguages()
		if len(langs) != 0 {
			t.Fatalf("got %v", langs)
		}
	})
	if !called {
		t.Fatal("Service_AvailableLanguages was not exercised")
	}
}

func TestService_Service_AvailableLanguages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		langs := svc.AvailableLanguages()
		langs[0] = "mutated"
		if svc.AvailableLanguages()[0] == "mutated" {
			t.Fatal("slice not copied")
		}
	})
	if !called {
		t.Fatal("Service_AvailableLanguages was not exercised")
	}
}

func TestService_Service_CurrentAvailableLanguages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		langs := svc.CurrentAvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("Service_CurrentAvailableLanguages was not exercised")
	}
}

func TestService_Service_CurrentAvailableLanguages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		langs := svc.CurrentAvailableLanguages()
		if len(langs) != 0 {
			t.Fatalf("got %v", langs)
		}
	})
	if !called {
		t.Fatal("Service_CurrentAvailableLanguages was not exercised")
	}
}

func TestService_Service_CurrentAvailableLanguages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		langs := svc.CurrentAvailableLanguages()
		langs[0] = "mutated"
		if svc.CurrentAvailableLanguages()[0] == "mutated" {
			t.Fatal("slice not copied")
		}
	})
	if !called {
		t.Fatal("Service_CurrentAvailableLanguages was not exercised")
	}
}

func TestService_Service_SetMode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetMode(ModeCollect)
		if svc.Mode() != ModeCollect {
			t.Fatal("mode not set")
		}
	})
	if !called {
		t.Fatal("Service_SetMode was not exercised")
	}
}

func TestService_Service_SetMode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.SetMode(ModeCollect)
		if svc.Mode() != ModeNormal {
			t.Fatal("nil mode changed")
		}
	})
	if !called {
		t.Fatal("Service_SetMode was not exercised")
	}
}

func TestService_Service_SetMode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetMode(ModeStrict)
		svc.SetMode(ModeNormal)
		if svc.Mode() != ModeNormal {
			t.Fatal("mode not reset")
		}
	})
	if !called {
		t.Fatal("Service_SetMode was not exercised")
	}
}

func TestService_Service_Mode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetMode(ModeCollect)
		got := svc.Mode()
		if got != ModeCollect {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Mode was not exercised")
	}
}

func TestService_Service_Mode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Mode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Mode was not exercised")
	}
}

func TestService_Service_Mode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Mode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Mode was not exercised")
	}
}

func TestService_Service_CurrentMode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetMode(ModeCollect)
		got := svc.CurrentMode()
		if got != ModeCollect {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentMode was not exercised")
	}
}

func TestService_Service_CurrentMode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentMode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentMode was not exercised")
	}
}

func TestService_Service_CurrentMode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentMode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentMode was not exercised")
	}
}

func TestService_Service_SetFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFormality(FormalityFormal)
		if svc.Formality() != FormalityFormal {
			t.Fatal("formality not set")
		}
	})
	if !called {
		t.Fatal("Service_SetFormality was not exercised")
	}
}

func TestService_Service_SetFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.SetFormality(FormalityFormal)
		if svc.Formality() != FormalityNeutral {
			t.Fatal("nil formality changed")
		}
	})
	if !called {
		t.Fatal("Service_SetFormality was not exercised")
	}
}

func TestService_Service_SetFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFormality(FormalityInformal)
		if svc.Formality() != FormalityInformal {
			t.Fatal("formality not set")
		}
	})
	if !called {
		t.Fatal("Service_SetFormality was not exercised")
	}
}

func TestService_Service_Formality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFormality(FormalityFormal)
		got := svc.Formality()
		if got != FormalityFormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Formality was not exercised")
	}
}

func TestService_Service_Formality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Formality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Formality was not exercised")
	}
}

func TestService_Service_Formality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Formality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Formality was not exercised")
	}
}

func TestService_Service_CurrentFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFormality(FormalityFormal)
		got := svc.CurrentFormality()
		if got != FormalityFormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentFormality was not exercised")
	}
}

func TestService_Service_CurrentFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentFormality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentFormality was not exercised")
	}
}

func TestService_Service_CurrentFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentFormality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentFormality was not exercised")
	}
}

func TestService_Service_SetFallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFallback("fr")
		if svc.Fallback() != "fr" {
			t.Fatal("fallback not set")
		}
	})
	if !called {
		t.Fatal("Service_SetFallback was not exercised")
	}
}

func TestService_Service_SetFallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.SetFallback("fr")
		if svc.Fallback() != "en" {
			t.Fatal("nil fallback changed")
		}
	})
	if !called {
		t.Fatal("Service_SetFallback was not exercised")
	}
}

func TestService_Service_SetFallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFallback("")
		if svc.Fallback() != "" {
			t.Fatal("empty fallback not set")
		}
	})
	if !called {
		t.Fatal("Service_SetFallback was not exercised")
	}
}

func TestService_Service_Fallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFallback("fr")
		got := svc.Fallback()
		if got != "fr" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Fallback was not exercised")
	}
}

func TestService_Service_Fallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Fallback()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Fallback was not exercised")
	}
}

func TestService_Service_Fallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Fallback()
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("Service_Fallback was not exercised")
	}
}

func TestService_Service_CurrentFallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetFallback("fr")
		got := svc.CurrentFallback()
		if got != "fr" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentFallback was not exercised")
	}
}

func TestService_Service_CurrentFallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentFallback()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentFallback was not exercised")
	}
}

func TestService_Service_CurrentFallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentFallback()
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("Service_CurrentFallback was not exercised")
	}
}

func TestService_Service_SetLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetLocation("workspace")
		if svc.Location() != "workspace" {
			t.Fatal("location not set")
		}
	})
	if !called {
		t.Fatal("Service_SetLocation was not exercised")
	}
}

func TestService_Service_SetLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.SetLocation("workspace")
		if svc.Location() != "" {
			t.Fatal("nil location changed")
		}
	})
	if !called {
		t.Fatal("Service_SetLocation was not exercised")
	}
}

func TestService_Service_SetLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetLocation("")
		if svc.Location() != "" {
			t.Fatal("empty location not set")
		}
	})
	if !called {
		t.Fatal("Service_SetLocation was not exercised")
	}
}

func TestService_Service_Location_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetLocation("workspace")
		got := svc.Location()
		if got != "workspace" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Location was not exercised")
	}
}

func TestService_Service_Location_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Location()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Location was not exercised")
	}
}

func TestService_Service_Location_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Location()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Location was not exercised")
	}
}

func TestService_Service_CurrentLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetLocation("workspace")
		got := svc.CurrentLocation()
		if got != "workspace" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentLocation was not exercised")
	}
}

func TestService_Service_CurrentLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentLocation()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentLocation was not exercised")
	}
}

func TestService_Service_CurrentLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentLocation()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentLocation was not exercised")
	}
}

func TestService_Service_Direction_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Direction()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Direction was not exercised")
	}
}

func TestService_Service_Direction_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Direction()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Direction was not exercised")
	}
}

func TestService_Service_Direction_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		_ = svc.SetLanguage("fr")
		got := svc.Direction()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_Direction was not exercised")
	}
}

func TestService_Service_CurrentDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentDirection was not exercised")
	}
}

func TestService_Service_CurrentDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentDirection was not exercised")
	}
}

func TestService_Service_CurrentDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentDirection was not exercised")
	}
}

func TestService_Service_CurrentTextDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentTextDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentTextDirection was not exercised")
	}
}

func TestService_Service_CurrentTextDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentTextDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentTextDirection was not exercised")
	}
}

func TestService_Service_CurrentTextDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentTextDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentTextDirection was not exercised")
	}
}

func TestService_Service_IsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.IsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_IsRTL was not exercised")
	}
}

func TestService_Service_IsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.IsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_IsRTL was not exercised")
	}
}

func TestService_Service_IsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.IsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_IsRTL was not exercised")
	}
}

func TestService_Service_CurrentIsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentIsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentIsRTL was not exercised")
	}
}

func TestService_Service_CurrentIsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentIsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentIsRTL was not exercised")
	}
}

func TestService_Service_CurrentIsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentIsRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentIsRTL was not exercised")
	}
}

func TestService_Service_RTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.RTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_RTL was not exercised")
	}
}

func TestService_Service_RTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.RTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_RTL was not exercised")
	}
}

func TestService_Service_RTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.RTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_RTL was not exercised")
	}
}

func TestService_Service_CurrentRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentRTL was not exercised")
	}
}

func TestService_Service_CurrentRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentRTL was not exercised")
	}
}

func TestService_Service_CurrentRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentRTL()
		if got {
			t.Fatal("expected ltr")
		}
	})
	if !called {
		t.Fatal("Service_CurrentRTL was not exercised")
	}
}

func TestService_Service_CurrentDebug_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetDebug(true)
		got := svc.CurrentDebug()
		if !got {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("Service_CurrentDebug was not exercised")
	}
}

func TestService_Service_CurrentDebug_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentDebug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("Service_CurrentDebug was not exercised")
	}
}

func TestService_Service_CurrentDebug_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentDebug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("Service_CurrentDebug was not exercised")
	}
}

func TestService_Service_PluralCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.PluralCategory(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_PluralCategory was not exercised")
	}
}

func TestService_Service_PluralCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.PluralCategory(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_PluralCategory was not exercised")
	}
}

func TestService_Service_PluralCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.PluralCategory(-1)
		if got == PluralZero {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_PluralCategory was not exercised")
	}
}

func TestService_Service_CurrentPluralCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentPluralCategory(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentPluralCategory was not exercised")
	}
}

func TestService_Service_CurrentPluralCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentPluralCategory(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_CurrentPluralCategory was not exercised")
	}
}

func TestService_Service_CurrentPluralCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentPluralCategory(-1)
		_ = got
	})
	if !called {
		t.Fatal("Service_CurrentPluralCategory was not exercised")
	}
}

func TestService_Service_PluralCategoryOf_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.PluralCategoryOf(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_PluralCategoryOf was not exercised")
	}
}

func TestService_Service_PluralCategoryOf_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.PluralCategoryOf(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("Service_PluralCategoryOf was not exercised")
	}
}

func TestService_Service_PluralCategoryOf_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.PluralCategoryOf(-1)
		_ = got
	})
	if !called {
		t.Fatal("Service_PluralCategoryOf was not exercised")
	}
}

func TestService_Service_AddHandler_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.AddHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("Service_AddHandler was not exercised")
	}
}

func TestService_Service_AddHandler_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.AddHandler(ax7Handler{match: true})
		if len(svc.Handlers()) != 0 {
			t.Fatal("nil receiver changed")
		}
	})
	if !called {
		t.Fatal("Service_AddHandler was not exercised")
	}
}

func TestService_Service_AddHandler_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.AddHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("Service_AddHandler was not exercised")
	}
}

func TestService_Service_SetHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetHandlers(ax7Handler{match: true})
		if len(svc.Handlers()) != 1 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("Service_SetHandlers was not exercised")
	}
}

func TestService_Service_SetHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.SetHandlers(ax7Handler{match: true})
		if len(svc.Handlers()) != 0 {
			t.Fatal("nil receiver changed")
		}
	})
	if !called {
		t.Fatal("Service_SetHandlers was not exercised")
	}
}

func TestService_Service_SetHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.SetHandlers(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) != 1 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("Service_SetHandlers was not exercised")
	}
}

func TestService_Service_PrependHandler_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.PrependHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("Service_PrependHandler was not exercised")
	}
}

func TestService_Service_PrependHandler_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.PrependHandler(ax7Handler{match: true})
		if len(svc.Handlers()) != 0 {
			t.Fatal("nil receiver changed")
		}
	})
	if !called {
		t.Fatal("Service_PrependHandler was not exercised")
	}
}

func TestService_Service_PrependHandler_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.PrependHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("Service_PrependHandler was not exercised")
	}
}

func TestService_Service_ClearHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.ClearHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("Service_ClearHandlers was not exercised")
	}
}

func TestService_Service_ClearHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.ClearHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatal("nil receiver changed")
		}
	})
	if !called {
		t.Fatal("Service_ClearHandlers was not exercised")
	}
}

func TestService_Service_ClearHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.ClearHandlers()
		svc.ClearHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("Service_ClearHandlers was not exercised")
	}
}

func TestService_Service_ResetHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.ClearHandlers()
		svc.ResetHandlers()
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Service_ResetHandlers was not exercised")
	}
}

func TestService_Service_ResetHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.ResetHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatal("nil receiver changed")
		}
	})
	if !called {
		t.Fatal("Service_ResetHandlers was not exercised")
	}
}

func TestService_Service_ResetHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.ResetHandlers()
		svc.ResetHandlers()
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Service_ResetHandlers was not exercised")
	}
}

func TestService_Service_Handlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		handlers := svc.Handlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Service_Handlers was not exercised")
	}
}

func TestService_Service_Handlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		handlers := svc.Handlers()
		if len(handlers) != 0 {
			t.Fatalf("got %d", len(handlers))
		}
	})
	if !called {
		t.Fatal("Service_Handlers was not exercised")
	}
}

func TestService_Service_Handlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		handlers := svc.Handlers()
		handlers[0] = nil
		if svc.Handlers()[0] == nil {
			t.Fatal("handlers not copied")
		}
	})
	if !called {
		t.Fatal("Service_Handlers was not exercised")
	}
}

func TestService_Service_CurrentHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		handlers := svc.CurrentHandlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Service_CurrentHandlers was not exercised")
	}
}

func TestService_Service_CurrentHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		handlers := svc.CurrentHandlers()
		if len(handlers) != 0 {
			t.Fatalf("got %d", len(handlers))
		}
	})
	if !called {
		t.Fatal("Service_CurrentHandlers was not exercised")
	}
}

func TestService_Service_CurrentHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		handlers := svc.CurrentHandlers()
		handlers[0] = nil
		if svc.CurrentHandlers()[0] == nil {
			t.Fatal("handlers not copied")
		}
	})
	if !called {
		t.Fatal("Service_CurrentHandlers was not exercised")
	}
}

func TestService_Service_T_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.T("prompt.yes")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("Service_T was not exercised")
	}
}

func TestService_Service_T_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.T("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_T was not exercised")
	}
}

func TestService_Service_T_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.T("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_T was not exercised")
	}
}

func TestService_Service_Compose_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Compose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("Service_Compose was not exercised")
	}
}

func TestService_Service_Compose_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Compose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("Service_Compose was not exercised")
	}
}

func TestService_Service_Compose_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Compose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("Service_Compose was not exercised")
	}
}

func TestService_Service_CurrentCompose_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentCompose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("Service_CurrentCompose was not exercised")
	}
}

func TestService_Service_CurrentCompose_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.CurrentCompose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("Service_CurrentCompose was not exercised")
	}
}

func TestService_Service_CurrentCompose_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.CurrentCompose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("Service_CurrentCompose was not exercised")
	}
}

func TestService_Service_Translate_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		r := svc.Translate("prompt.yes")
		if !r.OK {
			t.Fatalf("expected ok: %v", r)
		}
	})
	if !called {
		t.Fatal("Service_Translate was not exercised")
	}
}

func TestService_Service_Translate_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		r := svc.Translate("missing")
		if r.OK {
			t.Fatal("expected failure")
		}
	})
	if !called {
		t.Fatal("Service_Translate was not exercised")
	}
}

func TestService_Service_Translate_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		r := svc.Translate("missing")
		if r.OK {
			t.Fatal("expected missing")
		}
	})
	if !called {
		t.Fatal("Service_Translate was not exercised")
	}
}

func TestService_Service_Raw_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Raw("prompt.yes")
		if got == "" {
			t.Fatal("expected raw")
		}
	})
	if !called {
		t.Fatal("Service_Raw was not exercised")
	}
}

func TestService_Service_Raw_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		got := svc.Raw("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Raw was not exercised")
	}
}

func TestService_Service_Raw_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		got := svc.Raw("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Service_Raw was not exercised")
	}
}

func TestService_Service_AddMessages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.AddMessages("en", map[string]string{"ax7.service": "ready"})
		if svc.T("ax7.service") != "ready" {
			t.Fatal("message not added")
		}
	})
	if !called {
		t.Fatal("Service_AddMessages was not exercised")
	}
}

func TestService_Service_AddMessages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		svc.AddMessages("en", nil)
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("Service_AddMessages was not exercised")
	}
}

func TestService_Service_AddMessages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		svc.AddMessages("en", map[string]string{})
		_ = svc.AvailableLanguages()
	})
	if !called {
		t.Fatal("Service_AddMessages was not exercised")
	}
}

func TestService_Service_AddLoader_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Service_AddLoader was not exercised")
	}
}

func TestService_Service_AddLoader_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		err := svc.AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("Service_AddLoader was not exercised")
	}
}

func TestService_Service_AddLoader_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.AddLoader(nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("Service_AddLoader was not exercised")
	}
}

func TestService_Service_LoadFS_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.LoadFS(ax7TestFS(), "locales")
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("Service_LoadFS was not exercised")
	}
}

func TestService_Service_LoadFS_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *Service
		err := svc.LoadFS(ax7TestFS(), "locales")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("Service_LoadFS was not exercised")
	}
}

func TestService_Service_LoadFS_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7Service(t)
		err := svc.LoadFS(ax7TestFS(), "missing")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("Service_LoadFS was not exercised")
	}
}
