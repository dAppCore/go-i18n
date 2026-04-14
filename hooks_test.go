package i18n

import (
	"io/fs"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, 1, count, "should have 1 registered locale")
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
	require.NoError(t, err)
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
	assert.Equal(t, "loaded from provider", got)
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
	require.NoError(t, err)
	SetDefault(svc)

	RegisterLocaleProvider(testByteLocaleProvider{
		langs: map[string][]byte{
			"en": []byte(`{"provider.bytes.loaded": "loaded from bytes"}`),
		},
	})

	got := svc.T("provider.bytes.loaded")
	assert.Equal(t, "loaded from bytes", got)
}

func TestRegisterLocales_Good_AfterLocalesLoaded(t *testing.T) {
	// When localesLoaded is true, RegisterLocales should also call LoadFS immediately
	svc, err := New()
	require.NoError(t, err)
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
	assert.Equal(t, "arrived late", got)
}

func TestRegisterLocales_Good_WithInitializedDefaultService(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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
	assert.Equal(t, "loaded immediately", got)
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
	require.NoError(t, err)
	SetDefault(svc)

	got := svc.T("queued.registration")
	assert.Equal(t, "loaded via setdefault", got)
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
	require.NoError(t, err)
	SetDefault(first)
	require.Equal(t, "fresh value", first.T("fresh.registration"))

	second, err := New()
	require.NoError(t, err)
	SetDefault(second)

	got := second.T("fresh.registration")
	assert.Equal(t, "fresh value", got)
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

	require.NoError(t, Init())

	svc := Default()
	require.NotNil(t, svc)

	got := svc.T("init.registered")
	assert.Equal(t, "loaded on init", got)
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
	require.NoError(t, err)

	svc := Default()
	require.NotNil(t, svc)
	got := svc.T("core.registered")
	assert.Equal(t, "loaded on core bootstrap", got)
}

func TestNewCoreService_InvalidLanguagePreservesSetLanguageError(t *testing.T) {
	factory := NewCoreService(ServiceOptions{Language: "es"})

	_, err := factory(nil)
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "unsupported language: es")
	assert.Contains(t, msg, "available:")
	assert.NotContains(t, msg, "invalid language")
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
	require.NoError(t, err)

	svc := Default()
	require.NotNil(t, svc)
	assert.Equal(t, "en", svc.Language())
	assert.Equal(t, "fr", svc.Fallback())
	assert.Equal(t, FormalityFormal, svc.Formality())
	assert.Equal(t, "workspace", svc.Location())
	assert.Equal(t, ModeCollect, svc.Mode())
	assert.True(t, svc.Debug())
}

func TestCoreService_DelegatesToWrappedService(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	coreSvc := &CoreService{svc: svc}

	assert.Equal(t, svc.T("i18n.label.status"), coreSvc.T("i18n.label.status"))
	assert.Equal(t, svc.Raw("i18n.label.status"), coreSvc.Raw("i18n.label.status"))
	assert.Equal(t, svc.Translate("i18n.label.status"), coreSvc.Translate("i18n.label.status"))
	assert.Equal(t, svc.AvailableLanguages(), coreSvc.AvailableLanguages())
	assert.Equal(t, svc.AvailableLanguages(), coreSvc.CurrentAvailableLanguages())
	assert.Equal(t, svc.Direction(), coreSvc.Direction())
	assert.Equal(t, svc.Direction(), coreSvc.CurrentDirection())
	assert.Equal(t, svc.Direction(), coreSvc.CurrentTextDirection())
	assert.Equal(t, svc.IsRTL(), coreSvc.IsRTL())
	assert.Equal(t, svc.IsRTL(), coreSvc.CurrentIsRTL())
	assert.Equal(t, svc.IsRTL(), coreSvc.RTL())
	assert.Equal(t, svc.IsRTL(), coreSvc.CurrentRTL())
	assert.Equal(t, svc.PluralCategory(2), coreSvc.PluralCategory(2))
	assert.Equal(t, svc.PluralCategory(2), coreSvc.CurrentPluralCategory(2))
	assert.Equal(t, svc.PluralCategory(2), coreSvc.PluralCategoryOf(2))
	assert.Equal(t, svc.Mode(), coreSvc.CurrentMode())
	assert.Equal(t, svc.Language(), coreSvc.CurrentLanguage())
	assert.Equal(t, svc.Language(), coreSvc.CurrentLang())
	assert.Equal(t, svc.Prompt("confirm"), coreSvc.Prompt("confirm"))
	assert.Equal(t, svc.Prompt("confirm"), coreSvc.CurrentPrompt("confirm"))
	assert.Equal(t, svc.Lang("fr"), coreSvc.Lang("fr"))
	assert.Equal(t, svc.Fallback(), coreSvc.CurrentFallback())
	assert.Equal(t, svc.Formality(), coreSvc.CurrentFormality())
	assert.Equal(t, svc.Location(), coreSvc.CurrentLocation())
	assert.Equal(t, svc.Debug(), coreSvc.CurrentDebug())

	require.NoError(t, coreSvc.SetLanguage("en"))
	assert.Equal(t, "en", coreSvc.Language())

	coreSvc.SetFallback("fr")
	assert.Equal(t, "fr", coreSvc.Fallback())

	coreSvc.SetFormality(FormalityFormal)
	assert.Equal(t, FormalityFormal, coreSvc.Formality())

	coreSvc.SetLocation("workspace")
	assert.Equal(t, "workspace", coreSvc.Location())

	coreSvc.SetDebug(true)
	assert.True(t, coreSvc.Debug())
	coreSvc.SetDebug(false)
	assert.False(t, coreSvc.Debug())

	handlers := coreSvc.Handlers()
	assert.Equal(t, svc.Handlers(), handlers)
	assert.Equal(t, svc.Handlers(), coreSvc.CurrentHandlers())

	coreSvc.SetHandlers(LabelHandler{})
	require.Len(t, coreSvc.Handlers(), 1)
	assert.IsType(t, LabelHandler{}, coreSvc.Handlers()[0])

	coreSvc.AddHandler(ProgressHandler{})
	require.Len(t, coreSvc.Handlers(), 2)
	assert.IsType(t, ProgressHandler{}, coreSvc.Handlers()[1])

	coreSvc.PrependHandler(CountHandler{})
	require.Len(t, coreSvc.Handlers(), 3)
	assert.IsType(t, CountHandler{}, coreSvc.Handlers()[0])

	coreSvc.ClearHandlers()
	assert.Empty(t, coreSvc.Handlers())

	coreSvc.ResetHandlers()
	require.NotEmpty(t, coreSvc.Handlers())
	assert.IsType(t, LabelHandler{}, coreSvc.Handlers()[0])

	require.NoError(t, coreSvc.AddLoader(NewFSLoader(fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"core.service.loaded": "loaded"}`)},
	}, "locales")))
	assert.Equal(t, "loaded", coreSvc.T("core.service.loaded"))

	require.NoError(t, coreSvc.LoadFS(fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"core.service.loaded.fs": "loaded via fs"}`)},
	}, "locales"))
	assert.Equal(t, "loaded via fs", coreSvc.T("core.service.loaded.fs"))

	coreSvc.AddMessages("en", map[string]string{
		"core.service.add.messages": "loaded via add messages",
	})
	assert.Equal(t, "loaded via add messages", coreSvc.T("core.service.add.messages"))
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

	require.NoError(t, Init())

	svc := Default()
	require.NotNil(t, svc)
	assert.Contains(t, svc.Language(), "de")
	assert.Equal(t, "hallo", svc.T("hello"))
}

func TestDefault_ReinitialisesAfterClear(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	SetDefault(nil)

	require.NoError(t, Init())

	svc := Default()
	require.NotNil(t, svc)
	assert.Equal(t, "y", svc.T("prompt.yes"))
}

func TestLoadRegisteredLocales_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

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
	assert.True(t, loaded, "localesLoaded should be true after loadRegisteredLocales")

	got := svc.T("extra.key")
	assert.Equal(t, "extra value", got)
}

func TestOnMissingKey_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, "missing.test.key", captured.Key)
	assert.Equal(t, "bar", captured.Args["foo"])
	assert.Equal(t, "hooks_test.go", filepath.Base(captured.CallerFile))
}

func TestOnMissingKey_Good_AppendsHandlers(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, 1, first)
	assert.Equal(t, 1, second)
}

func TestAddMissingKeyHandler_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, 1, first)
	assert.Equal(t, 1, second)
}

func TestSetMissingKeyHandlers_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, 1, first)
	assert.Equal(t, 1, second)
	assert.Len(t, missingKeyHandlers().handlers, 2)
}

func TestSetMissingKeyHandlers_Good_Clear(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, 0, called)
	assert.Empty(t, missingKeyHandlers().handlers)
}

func TestAddMissingKeyHandler_Good_Concurrent(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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
	assert.Len(t, state.handlers, handlers)
}

func TestClearMissingKeyHandlers_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, 0, called)
}

func TestOnMissingKey_Good_SubjectArgs(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, "missing.subject.key", captured.Key)
	assert.Equal(t, "config.yaml", captured.Args["Subject"])
	assert.Equal(t, "file", captured.Args["Noun"])
	assert.Equal(t, 3, captured.Args["Count"])
	assert.Equal(t, "workspace", captured.Args["Location"])
	assert.Equal(t, FormalityFormal, captured.Args["Formality"])
}

func TestOnMissingKey_Good_TranslationContextArgs(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, "missing.context.key", captured.Key)
	assert.Equal(t, "navigation", captured.Args["Context"])
	assert.Equal(t, "feminine", captured.Args["Gender"])
	assert.Equal(t, "workspace", captured.Args["Location"])
	assert.Equal(t, FormalityFormal, captured.Args["Formality"])
}

func TestOnMissingKey_Good_MergesAdditionalArgs(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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

	assert.Equal(t, "missing.extra.args", captured.Key)
	assert.Equal(t, "config.yaml", captured.Args["Subject"])
	assert.Equal(t, "abc123", captured.Args["trace"])
}

func TestDispatchMissingKey_Good_NoHandler(t *testing.T) {
	// Reset to the empty handler set.
	OnMissingKey(nil)

	// Should not panic when dispatching with nil handler
	dispatchMissingKey("test.key", nil)
}

func TestCoreServiceSetMode_Good_PreservesMissingKeyHandlers(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

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
