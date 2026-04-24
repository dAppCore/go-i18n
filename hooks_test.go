package i18n

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

type testLocaleProvider struct {
	sources []FSSource
}

func (p testLocaleProvider) LocaleSources() []FSSource {
	return p.sources
}

type testByteLocaleProvider struct {
	langs map[string][]byte
}

func (p testByteLocaleProvider) Available() []string {
	if len(p.langs) == 0 {
		return nil
	}
	langs := make([]string, 0, len(p.langs))
	for lang := range p.langs {
		langs = append(langs, lang)
	}
	return langs
}

func (p testByteLocaleProvider) Load(lang string) ([]byte, error) {
	data, ok := p.langs[lang]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func TestRegisterLocales_Good(t *testing.T) {
	// Save and restore registered locales state
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	fs := fstest.MapFS{
		"locales/test.json": &fstest.MapFile{
			Data: []byte(`{"custom.hook": "hooked"}`),
		},
	}

	RegisterLocales(fs, "locales")

	registeredLocalesMu.Lock()
	count := len(registeredLocales)
	registeredLocalesMu.Unlock()
	if (1) != (count) {
		t.Fatalf("want %v, got %v", 1, count)
	}
}

func TestRegisterLocaleProvider_Good(t *testing.T) {
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"provider.loaded": "loaded from provider"}`),
		},
	}

	RegisterLocaleProvider(testLocaleProvider{
		sources: []FSSource{{FS: fs, Dir: "locales"}},
	})

	got := svc.T("provider.loaded")
	if ("loaded from provider") != (got) {
		t.Fatalf("want %v, got %v", "loaded from provider", got)
	}
}

func TestRegisterLocaleProvider_Good_ByteProvider(t *testing.T) {
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	RegisterLocaleProvider(testByteLocaleProvider{
		langs: map[string][]byte{
			"en": []byte(`{"provider.bytes.loaded": "loaded from bytes"}`),
		},
	})

	got := svc.T("provider.bytes.loaded")
	if ("loaded from bytes") != (got) {
		t.Fatalf("want %v, got %v", "loaded from bytes", got)
	}
}

func TestRegisterLocales_Good_AfterLocalesLoaded(t *testing.T) {
	// When localesLoaded is true, RegisterLocales should also call LoadFS immediately
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	// Save and restore state
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = true // Simulate already loaded
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	// Use "en.json" as filename so language matches fallback
	fs := fstest.MapFS{
		"i18n/en.json": &fstest.MapFile{
			Data: []byte(`{"late.registration": "arrived late"}`),
		},
	}

	RegisterLocales(fs, "i18n")

	// Should be able to resolve the newly registered key
	got := svc.T("late.registration")
	if ("arrived late") != (got) {
		t.Fatalf("want %v, got %v", "arrived late", got)
	}
}

func TestRegisterLocales_Good_WithInitializedDefaultService(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"eager.registration": "loaded immediately"}`),
		},
	}

	RegisterLocales(fs, "locales")

	got := svc.T("eager.registration")
	if ("loaded immediately") != (got) {
		t.Fatalf("want %v, got %v", "loaded immediately", got)
	}
}

func TestSetDefault_Good_LoadsQueuedRegisteredLocales(t *testing.T) {
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"queued.registration": "loaded via setdefault"}`),
		},
	}
	RegisterLocales(fs, "locales")

	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	got := svc.T("queued.registration")
	if ("loaded via setdefault") != (got) {
		t.Fatalf("want %v, got %v", "loaded via setdefault", got)
	}
}

func TestSetDefault_Good_LoadsRegisteredLocalesIntoFreshService(t *testing.T) {
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"fresh.registration": "fresh value"}`),
		},
	}
	RegisterLocales(fs, "locales")

	first, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(first)
	if ("fresh value") != (first.T("fresh.registration")) {
		t.Fatalf("want %v, got %v", "fresh value", first.T("fresh.registration"))
	}

	second, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(second)

	got := second.T("fresh.registration")
	if ("fresh value") != (got) {
		t.Fatalf("want %v, got %v", "fresh value", got)
	}
}

func TestInit_LoadsRegisteredLocales(t *testing.T) {
	// Save and restore global service state.
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()

	defaultService.Store(nil)

	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
		defaultService.Store(nil)
	}()

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"init.registered": "loaded on init"}`),
		},
	}
	RegisterLocales(fs, "locales")
	if err := Init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := Default()
	if (svc) == (nil) {
		t.Fatalf("expected non-nil")
	}

	got := svc.T("init.registered")
	if ("loaded on init") != (got) {
		t.Fatalf("want %v, got %v", "loaded on init", got)
	}
}

func TestNewCoreService_LoadsRegisteredLocales(t *testing.T) {
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedLoaded := localesLoaded
	registeredLocales = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()

	prev := defaultService.Load()
	SetDefault(nil)

	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
		SetDefault(prev)
	}()

	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"core.registered": "loaded on core bootstrap"}`),
		},
	}
	RegisterLocales(fs, "locales")

	factory := NewCoreService(ServiceOptions{})
	_, err := factory(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := Default()
	if (svc) == (nil) {
		t.Fatalf("expected non-nil")
	}
	got := svc.T("core.registered")
	if ("loaded on core bootstrap") != (got) {
		t.Fatalf("want %v, got %v", "loaded on core bootstrap", got)
	}
}

func TestNewCoreService_InvalidLanguagePreservesSetLanguageError(t *testing.T) {
	factory := NewCoreService(ServiceOptions{Language: "es"})

	_, err := factory(nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "unsupported language: es") {
		t.Fatalf("expected %q to contain %q", msg, "unsupported language: es")
	}
	if !strings.Contains(msg, "available:") {
		t.Fatalf("expected %q to contain %q", msg, "available:")
	}
	if strings.Contains(msg, "invalid language") {
		t.Fatalf("did not expect %q to contain %q", msg, "invalid language")
	}
}

func TestNewCoreService_AppliesOptions(t *testing.T) {
	prev := Default()
	SetDefault(nil)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	factory := NewCoreService(ServiceOptions{
		Language:  "en",
		Fallback:  "fr",
		Formality: FormalityFormal,
		Location:  "workspace",
		Mode:      ModeCollect,
		Debug:     true,
	})

	_, err := factory(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := Default()
	if (svc) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if ("en") != (svc.Language()) {
		t.Fatalf("want %v, got %v", "en", svc.Language())
	}
	if ("fr") != (svc.Fallback()) {
		t.Fatalf("want %v, got %v", "fr", svc.Fallback())
	}
	if (FormalityFormal) != (svc.Formality()) {
		t.Fatalf("want %v, got %v", FormalityFormal, svc.Formality())
	}
	if ("workspace") != (svc.Location()) {
		t.Fatalf("want %v, got %v", "workspace", svc.Location())
	}
	if (ModeCollect) != (svc.Mode()) {
		t.Fatalf("want %v, got %v", ModeCollect, svc.Mode())
	}
	if !(svc.Debug()) {
		t.Fatal("expected true")
	}
}

func TestCoreService_DelegatesToWrappedService(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	coreSvc := &CoreService{svc: svc}
	if (svc.T("i18n.label.status")) != (coreSvc.T("i18n.label.status")) {
		t.Fatalf("want %v, got %v", svc.T("i18n.label.status"), coreSvc.T("i18n.label.status"))
	}
	if (svc.Raw("i18n.label.status")) != (coreSvc.Raw("i18n.label.status")) {
		t.Fatalf("want %v, got %v", svc.Raw("i18n.label.status"), coreSvc.Raw("i18n.label.status"))
	}
	if (svc.Translate("i18n.label.status")) != (coreSvc.Translate("i18n.label.status")) {
		t.Fatalf("want %v, got %v", svc.Translate("i18n.label.status"), coreSvc.Translate("i18n.label.status"))
	}
	if !reflect.DeepEqual(svc.AvailableLanguages(), coreSvc.AvailableLanguages()) {
		t.Fatalf("want %v, got %v", svc.AvailableLanguages(), coreSvc.AvailableLanguages())
	}
	if !reflect.DeepEqual(svc.AvailableLanguages(), coreSvc.CurrentAvailableLanguages()) {
		t.Fatalf("want %v, got %v", svc.AvailableLanguages(), coreSvc.CurrentAvailableLanguages())
	}
	if (svc.Direction()) != (coreSvc.Direction()) {
		t.Fatalf("want %v, got %v", svc.Direction(), coreSvc.Direction())
	}
	if (svc.Direction()) != (coreSvc.CurrentDirection()) {
		t.Fatalf("want %v, got %v", svc.Direction(), coreSvc.CurrentDirection())
	}
	if (svc.Direction()) != (coreSvc.CurrentTextDirection()) {
		t.Fatalf("want %v, got %v", svc.Direction(), coreSvc.CurrentTextDirection())
	}
	if (svc.IsRTL()) != (coreSvc.IsRTL()) {
		t.Fatalf("want %v, got %v", svc.IsRTL(), coreSvc.IsRTL())
	}
	if (svc.IsRTL()) != (coreSvc.CurrentIsRTL()) {
		t.Fatalf("want %v, got %v", svc.IsRTL(), coreSvc.CurrentIsRTL())
	}
	if (svc.IsRTL()) != (coreSvc.RTL()) {
		t.Fatalf("want %v, got %v", svc.IsRTL(), coreSvc.RTL())
	}
	if (svc.IsRTL()) != (coreSvc.CurrentRTL()) {
		t.Fatalf("want %v, got %v", svc.IsRTL(), coreSvc.CurrentRTL())
	}
	if (svc.PluralCategory(2)) != (coreSvc.PluralCategory(2)) {
		t.Fatalf("want %v, got %v", svc.PluralCategory(2), coreSvc.PluralCategory(2))
	}
	if (svc.PluralCategory(2)) != (coreSvc.CurrentPluralCategory(2)) {
		t.Fatalf("want %v, got %v", svc.PluralCategory(2), coreSvc.CurrentPluralCategory(2))
	}
	if (svc.PluralCategory(2)) != (coreSvc.PluralCategoryOf(2)) {
		t.Fatalf("want %v, got %v", svc.PluralCategory(2), coreSvc.PluralCategoryOf(2))
	}
	if (svc.Mode()) != (coreSvc.CurrentMode()) {
		t.Fatalf("want %v, got %v", svc.Mode(), coreSvc.CurrentMode())
	}
	if (svc.Language()) != (coreSvc.CurrentLanguage()) {
		t.Fatalf("want %v, got %v", svc.Language(), coreSvc.CurrentLanguage())
	}
	if (svc.Language()) != (coreSvc.CurrentLang()) {
		t.Fatalf("want %v, got %v", svc.Language(), coreSvc.CurrentLang())
	}
	if (svc.Prompt("confirm")) != (coreSvc.Prompt("confirm")) {
		t.Fatalf("want %v, got %v", svc.Prompt("confirm"), coreSvc.Prompt("confirm"))
	}
	if (svc.Prompt("confirm")) != (coreSvc.CurrentPrompt("confirm")) {
		t.Fatalf("want %v, got %v", svc.Prompt("confirm"), coreSvc.CurrentPrompt("confirm"))
	}
	if (svc.Lang("fr")) != (coreSvc.Lang("fr")) {
		t.Fatalf("want %v, got %v", svc.Lang("fr"), coreSvc.Lang("fr"))
	}
	if (svc.Fallback()) != (coreSvc.CurrentFallback()) {
		t.Fatalf("want %v, got %v", svc.Fallback(), coreSvc.CurrentFallback())
	}
	if (svc.Formality()) != (coreSvc.CurrentFormality()) {
		t.Fatalf("want %v, got %v", svc.Formality(), coreSvc.CurrentFormality())
	}
	if (svc.Location()) != (coreSvc.CurrentLocation()) {
		t.Fatalf("want %v, got %v", svc.Location(), coreSvc.CurrentLocation())
	}
	if (svc.Debug()) != (coreSvc.CurrentDebug()) {
		t.Fatalf("want %v, got %v", svc.Debug(), coreSvc.CurrentDebug())
	}
	if err := coreSvc.SetLanguage("en"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ("en") != (coreSvc.Language()) {
		t.Fatalf("want %v, got %v", "en", coreSvc.Language())
	}

	coreSvc.SetFallback("fr")
	if ("fr") != (coreSvc.Fallback()) {
		t.Fatalf("want %v, got %v", "fr", coreSvc.Fallback())
	}

	coreSvc.SetFormality(FormalityFormal)
	if (FormalityFormal) != (coreSvc.Formality()) {
		t.Fatalf("want %v, got %v", FormalityFormal, coreSvc.Formality())
	}

	coreSvc.SetLocation("workspace")
	if ("workspace") != (coreSvc.Location()) {
		t.Fatalf("want %v, got %v", "workspace", coreSvc.Location())
	}

	coreSvc.SetDebug(true)
	if !(coreSvc.Debug()) {
		t.Fatal("expected true")
	}
	coreSvc.SetDebug(false)
	if coreSvc.Debug() {
		t.Fatal("expected false")
	}

	handlers := coreSvc.Handlers()
	if !reflect.DeepEqual(svc.Handlers(), handlers) {
		t.Fatalf("want %v, got %v", svc.Handlers(), handlers)
	}
	if !reflect.DeepEqual(svc.Handlers(), coreSvc.CurrentHandlers()) {
		t.Fatalf("want %v, got %v", svc.Handlers(), coreSvc.CurrentHandlers())
	}

	coreSvc.SetHandlers(LabelHandler{})
	if len(coreSvc.Handlers()) != 1 {
		t.Fatalf("expected length %v, got %v", 1, coreSvc.Handlers())
	}
	if reflect.TypeOf(coreSvc.Handlers()[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, coreSvc.Handlers()[0])
	}

	coreSvc.AddHandler(ProgressHandler{})
	if len(coreSvc.Handlers()) != 2 {
		t.Fatalf("expected length %v, got %v", 2, coreSvc.Handlers())
	}
	if reflect.TypeOf(coreSvc.Handlers()[1]) != reflect.TypeOf(ProgressHandler{}) {
		t.Fatalf("expected type %T, got %T", ProgressHandler{}, coreSvc.Handlers()[1])
	}

	coreSvc.PrependHandler(CountHandler{})
	if len(coreSvc.Handlers()) != 3 {
		t.Fatalf("expected length %v, got %v", 3, coreSvc.Handlers())
	}
	if reflect.TypeOf(coreSvc.Handlers()[0]) != reflect.TypeOf(CountHandler{}) {
		t.Fatalf("expected type %T, got %T", CountHandler{}, coreSvc.Handlers()[0])
	}

	coreSvc.ClearHandlers()
	if len(coreSvc.Handlers()) != 0 {
		t.Fatalf("expected empty, got %v", coreSvc.Handlers())
	}

	coreSvc.ResetHandlers()
	if len(coreSvc.Handlers()) == 0 {
		t.Fatalf("expected non-empty")
	}
	if reflect.TypeOf(coreSvc.Handlers()[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, coreSvc.Handlers()[0])
	}
	if err := coreSvc.AddLoader(NewFSLoader(fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"core.service.loaded": "loaded"}`)},
	}, "locales")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ("loaded") != (coreSvc.T("core.service.loaded")) {
		t.Fatalf("want %v, got %v", "loaded", coreSvc.T("core.service.loaded"))
	}
	if err := coreSvc.LoadFS(fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"core.service.loaded.fs": "loaded via fs"}`)},
	}, "locales"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ("loaded via fs") != (coreSvc.T("core.service.loaded.fs")) {
		t.Fatalf("want %v, got %v", "loaded via fs", coreSvc.T("core.service.loaded.fs"))
	}

	coreSvc.AddMessages("en", map[string]string{
		"core.service.add.messages": "loaded via add messages",
	})
	if ("loaded via add messages") != (coreSvc.T("core.service.add.messages")) {
		t.Fatalf("want %v, got %v", "loaded via add messages", coreSvc.T("core.service.add.messages"))
	}
}

func TestInit_ReDetectsRegisteredLocales(t *testing.T) {
	t.Setenv("LANG", "de_DE.UTF-8")

	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()

	defaultService.Store(nil)

	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
		defaultService.Store(nil)
	}()

	fs := fstest.MapFS{
		"locales/de.json": &fstest.MapFile{
			Data: []byte(`{"hello": "hallo"}`),
		},
	}
	RegisterLocales(fs, "locales")
	if err := Init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := Default()
	if (svc) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if !strings.Contains(svc.Language(), "de") {
		t.Fatalf("expected %q to contain %q", svc.Language(), "de")
	}
	if ("hallo") != (svc.T("hello")) {
		t.Fatalf("want %v, got %v", "hallo", svc.T("hello"))
	}
}

func TestDefault_ReinitialisesAfterClear(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	SetDefault(nil)
	if err := Init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc := Default()
	if (svc) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if ("y") != (svc.T("prompt.yes")) {
		t.Fatalf("want %v, got %v", "y", svc.T("prompt.yes"))
	}
}

func TestLoadRegisteredLocales_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Save and restore state
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	registeredLocales = []localeRegistration{
		{
			fsys: fstest.MapFS{
				"loc/en.json": &fstest.MapFile{
					Data: []byte(`{"extra.key": "extra value"}`),
				},
			},
			dir: "loc",
		},
	}
	registeredLocaleProviders = nil
	localesLoaded = false
	registeredLocalesMu.Unlock()
	defer func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		registeredLocalesMu.Unlock()
	}()

	loadRegisteredLocales(svc)

	registeredLocalesMu.Lock()
	loaded := localesLoaded
	registeredLocalesMu.Unlock()
	if !(loaded) {
		t.Fatal("expected true")
	}

	got := svc.T("extra.key")
	if ("extra value") != (got) {
		t.Fatalf("want %v, got %v", "extra value", got)
	}
}

func TestOnMissingKey_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var captured MissingKey
	OnMissingKey(func(m MissingKey) {
		captured = m
	})

	_ = T("missing.test.key", map[string]any{"foo": "bar"})
	if ("missing.test.key") != (captured.Key) {
		t.Fatalf("want %v, got %v", "missing.test.key", captured.Key)
	}
	if ("bar") != (captured.Args["foo"]) {
		t.Fatalf("want %v, got %v", "bar", captured.Args["foo"])
	}
	if ("hooks_test.go") != (filepath.Base(captured.CallerFile)) {
		t.Fatalf("want %v, got %v", "hooks_test.go", filepath.Base(captured.CallerFile))
	}
}

func TestOnMissingKey_Good_AppendsHandlers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})

	svc.SetMode(ModeCollect)
	ClearMissingKeyHandlers()
	t.Cleanup(func() {
		ClearMissingKeyHandlers()
	})

	var first, second int
	OnMissingKey(func(MissingKey) { first++ })
	OnMissingKey(func(MissingKey) { second++ })

	_ = T("missing.on.handler.append")
	if (1) != (first) {
		t.Fatalf("want %v, got %v", 1, first)
	}
	if (1) != (second) {
		t.Fatalf("want %v, got %v", 1, second)
	}
}

func TestAddMissingKeyHandler_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	ClearMissingKeyHandlers()
	t.Cleanup(func() {
		ClearMissingKeyHandlers()
	})

	var first, second int
	AddMissingKeyHandler(func(MissingKey) {
		first++
	})
	AddMissingKeyHandler(func(MissingKey) {
		second++
	})

	_ = T("missing.multiple.handlers")
	if (1) != (first) {
		t.Fatalf("want %v, got %v", 1, first)
	}
	if (1) != (second) {
		t.Fatalf("want %v, got %v", 1, second)
	}
}

func TestSetMissingKeyHandlers_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var first, second int
	SetMissingKeyHandlers(
		nil,
		func(MissingKey) { first++ },
		func(MissingKey) { second++ },
	)

	_ = T("missing.set.handlers")
	if (1) != (first) {
		t.Fatalf("want %v, got %v", 1, first)
	}
	if (1) != (second) {
		t.Fatalf("want %v, got %v", 1, second)
	}
	if len(missingKeyHandlers().handlers) != 2 {
		t.Fatalf("expected length %v, got %v", 2, missingKeyHandlers().handlers)
	}
}

func TestSetMissingKeyHandlers_Good_Clear(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var called int
	SetMissingKeyHandlers(func(MissingKey) { called++ })
	SetMissingKeyHandlers(nil)

	_ = T("missing.set.handlers.clear")
	if (0) != (called) {
		t.Fatalf("want %v, got %v", 0, called)
	}
	if len(missingKeyHandlers().handlers) != 0 {
		t.Fatalf("expected empty, got %v", missingKeyHandlers().handlers)
	}
}

func TestAddMissingKeyHandler_Good_Concurrent(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	ClearMissingKeyHandlers()
	t.Cleanup(func() {
		ClearMissingKeyHandlers()
	})

	const handlers = 32
	var wg sync.WaitGroup
	wg.Add(handlers)
	for i := 0; i < handlers; i++ {
		go func() {
			defer wg.Done()
			AddMissingKeyHandler(func(MissingKey) {})
		}()
	}
	wg.Wait()

	state := missingKeyHandlers()
	if len(state.handlers) != handlers {
		t.Fatalf("expected length %v, got %v", handlers, state.handlers)
	}
}

func TestClearMissingKeyHandlers_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	prevHandlers := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prevHandlers)
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var called int
	AddMissingKeyHandler(func(MissingKey) {
		called++
	})

	ClearMissingKeyHandlers()

	_ = T("missing.after.clear")
	if (0) != (called) {
		t.Fatalf("want %v, got %v", 0, called)
	}
}

func TestOnMissingKey_Good_SubjectArgs(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var captured MissingKey
	OnMissingKey(func(m MissingKey) {
		captured = m
	})

	_ = T("missing.subject.key", S("file", "config.yaml").Count(3).In("workspace").Formal())
	if ("missing.subject.key") != (captured.Key) {
		t.Fatalf("want %v, got %v", "missing.subject.key", captured.Key)
	}
	if ("config.yaml") != (captured.Args["Subject"]) {
		t.Fatalf("want %v, got %v", "config.yaml", captured.Args["Subject"])
	}
	if ("file") != (captured.Args["Noun"]) {
		t.Fatalf("want %v, got %v", "file", captured.Args["Noun"])
	}
	if (3) != (captured.Args["Count"]) {
		t.Fatalf("want %v, got %v", 3, captured.Args["Count"])
	}
	if ("workspace") != (captured.Args["Location"]) {
		t.Fatalf("want %v, got %v", "workspace", captured.Args["Location"])
	}
	if (FormalityFormal) != (captured.Args["Formality"]) {
		t.Fatalf("want %v, got %v", FormalityFormal, captured.Args["Formality"])
	}
}

func TestOnMissingKey_Good_TranslationContextArgs(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var captured MissingKey
	OnMissingKey(func(m MissingKey) {
		captured = m
	})

	_ = T("missing.context.key", C("navigation").WithGender("feminine").In("workspace").Formal())
	if ("missing.context.key") != (captured.Key) {
		t.Fatalf("want %v, got %v", "missing.context.key", captured.Key)
	}
	if ("navigation") != (captured.Args["Context"]) {
		t.Fatalf("want %v, got %v", "navigation", captured.Args["Context"])
	}
	if ("feminine") != (captured.Args["Gender"]) {
		t.Fatalf("want %v, got %v", "feminine", captured.Args["Gender"])
	}
	if ("workspace") != (captured.Args["Location"]) {
		t.Fatalf("want %v, got %v", "workspace", captured.Args["Location"])
	}
	if (FormalityFormal) != (captured.Args["Formality"]) {
		t.Fatalf("want %v, got %v", FormalityFormal, captured.Args["Formality"])
	}
}

func TestOnMissingKey_Good_MergesAdditionalArgs(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	svc.SetMode(ModeCollect)

	var captured MissingKey
	OnMissingKey(func(m MissingKey) {
		captured = m
	})

	_ = T("missing.extra.args", S("file", "config.yaml"), map[string]any{"trace": "abc123"})
	if ("missing.extra.args") != (captured.Key) {
		t.Fatalf("want %v, got %v", "missing.extra.args", captured.Key)
	}
	if ("config.yaml") != (captured.Args["Subject"]) {
		t.Fatalf("want %v, got %v", "config.yaml", captured.Args["Subject"])
	}
	if ("abc123") != (captured.Args["trace"]) {
		t.Fatalf("want %v, got %v", "abc123", captured.Args["trace"])
	}
}

func TestDispatchMissingKey_Good_NoHandler(t *testing.T) {
	// Reset to the empty handler set.
	OnMissingKey(nil)

	// Should not panic when dispatching with nil handler
	dispatchMissingKey("test.key", nil)
}

func TestCoreServiceSetMode_Good_PreservesMissingKeyHandlers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := missingKeyHandlers()
	t.Cleanup(func() {
		missingKeyHandler.Store(prev)
	})

	var observed int
	OnMissingKey(func(MissingKey) {
		observed++
	})
	t.Cleanup(func() {
		OnMissingKey(nil)
	})

	coreSvc := &CoreService{svc: svc}
	coreSvc.SetMode(ModeCollect)

	_ = svc.T("missing.core.service.key")

	if observed != 1 {
		t.Fatalf("custom missing key handler called %d times, want 1", observed)
	}

	missing := coreSvc.MissingKeys()
	if len(missing) != 1 {
		t.Fatalf("CoreService captured %d missing keys, want 1", len(missing))
	}
	if missing[0].Key != "missing.core.service.key" {
		t.Fatalf("captured missing key = %q, want %q", missing[0].Key, "missing.core.service.key")
	}
}
