package i18n

import (
	"path/filepath"
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
