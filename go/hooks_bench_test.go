package i18n

import (
	"testing"
	"testing/fstest"
)

var benchMissingKeySink MissingKey

type benchHooksSourceProvider struct {
	sources []FSSource
}

func (p benchHooksSourceProvider) LocaleSources() []FSSource {
	return p.sources
}

type benchHooksByteProvider struct {
	langs map[string][]byte
}

func (p benchHooksByteProvider) Available() []string {
	langs := make([]string, 0, len(p.langs))
	for lang := range p.langs {
		langs = append(langs, lang)
	}
	return langs
}

func (p benchHooksByteProvider) Load(lang string) ([]byte, error) {
	return p.langs[lang], nil
}

func benchHooksLocaleFS() fstest.MapFS {
	return fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"hooks.loaded":"loaded"}`)},
	}
}

func benchHooksByteProviderFixture() benchHooksByteProvider {
	return benchHooksByteProvider{
		langs: map[string][]byte{
			"en_US": []byte(`{"hooks.bytes":"loaded from bytes"}`),
		},
	}
}

func benchResetLocaleHooks(b *testing.B) {
	b.Helper()
	previousDefault := defaultService.Load()
	registeredLocalesMu.Lock()
	savedLocales := registeredLocales
	savedProviders := registeredLocaleProviders
	savedLoaded := localesLoaded
	savedLocaleID := nextLocaleRegistrationID
	savedProviderID := nextLocaleProviderID
	registeredLocales = nil
	registeredLocaleProviders = nil
	localesLoaded = false
	nextLocaleRegistrationID = 0
	nextLocaleProviderID = 0
	registeredLocalesMu.Unlock()
	defaultService.Store((*Service)(nil))
	b.Cleanup(func() {
		registeredLocalesMu.Lock()
		registeredLocales = savedLocales
		registeredLocaleProviders = savedProviders
		localesLoaded = savedLoaded
		nextLocaleRegistrationID = savedLocaleID
		nextLocaleProviderID = savedProviderID
		registeredLocalesMu.Unlock()
		defaultService.Store(previousDefault)
	})
}

func benchResetMissingKeyHooks(b *testing.B) {
	b.Helper()
	previous := missingKeyHandlers()
	missingKeyHandler.Store(missingKeyHandlersState{})
	b.Cleanup(func() {
		missingKeyHandler.Store(previous)
	})
}

func benchMissingKeyHandler(missing MissingKey) {
	benchMissingKeySink = missing
}

func BenchmarkRegisterLocales(b *testing.B) {
	benchResetLocaleHooks(b)
	fsys := benchHooksLocaleFS()
	registeredLocalesMu.Lock()
	registeredLocales = make([]localeRegistration, 0, b.N)
	registeredLocalesMu.Unlock()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterLocales(fsys, "locales")
	}
}

func BenchmarkRegisterLocaleProvider(b *testing.B) {
	benchResetLocaleHooks(b)
	fsys := benchHooksLocaleFS()
	provider := benchHooksSourceProvider{sources: []FSSource{{FS: fsys, Dir: "locales"}}}
	registeredLocalesMu.Lock()
	registeredLocaleProviders = make([]localeProviderRegistration, 0, b.N)
	registeredLocalesMu.Unlock()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterLocaleProvider(provider)
	}
}

func BenchmarkLoadRegisteredLocales(b *testing.B) {
	benchResetLocaleHooks(b)
	fsys := benchHooksLocaleFS()
	provider := benchHooksByteProviderFixture()
	registeredLocalesMu.Lock()
	registeredLocales = []localeRegistration{{fsys: fsys, dir: "locales", id: 1}}
	registeredLocaleProviders = []localeProviderRegistration{{provider: provider, id: 1}}
	registeredLocalesMu.Unlock()
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.loadedLocales = make(map[int]struct{})
		svc.loadedProviders = make(map[int]struct{})
		loadRegisteredLocales(svc)
	}
}

func BenchmarkLoadLocaleProvider(b *testing.B) {
	fsys := benchHooksLocaleFS()
	providers := []localeProviderRegistration{
		{provider: NewFSLoader(fsys, "locales"), id: 1},
		{provider: benchHooksSourceProvider{sources: []FSSource{{FS: fsys, Dir: "locales"}}}, id: 2},
		{provider: benchHooksByteProviderFixture(), id: 3},
	}
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, provider := range providers {
			loadLocaleProvider(svc, provider)
		}
	}
}

func BenchmarkLoadLocaleProviderBytes(b *testing.B) {
	svc := benchServiceFixture()
	provider := benchHooksByteProviderFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadLocaleProviderBytes(svc, provider)
	}
}

func BenchmarkMarkLocalesLoaded(b *testing.B) {
	benchResetLocaleHooks(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		markLocalesLoaded()
	}
}

func BenchmarkOnMissingKey(b *testing.B) {
	benchResetMissingKeyHooks(b)
	missingKeyHandler.Store(missingKeyHandlersState{handlers: make([]MissingKeyHandler, 0, b.N)})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OnMissingKey(benchMissingKeyHandler)
	}
}

func BenchmarkSetMissingKeyHandlers(b *testing.B) {
	benchResetMissingKeyHooks(b)
	handlers := []MissingKeyHandler{benchMissingKeyHandler, nil, benchMissingKeyHandler}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetMissingKeyHandlers(handlers...)
	}
}

func BenchmarkClearMissingKeyHandlers(b *testing.B) {
	benchResetMissingKeyHooks(b)
	missingKeyHandler.Store(missingKeyHandlersState{handlers: []MissingKeyHandler{benchMissingKeyHandler}})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClearMissingKeyHandlers()
	}
}

func BenchmarkAddMissingKeyHandler(b *testing.B) {
	benchResetMissingKeyHooks(b)
	missingKeyHandler.Store(missingKeyHandlersState{handlers: make([]MissingKeyHandler, 0, b.N)})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddMissingKeyHandler(benchMissingKeyHandler)
	}
}

func BenchmarkFilterNilMissingKeyHandlers(b *testing.B) {
	handlers := []MissingKeyHandler{nil, benchMissingKeyHandler, nil, benchMissingKeyHandler}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = filterNilMissingKeyHandlers(handlers)
	}
}

func BenchmarkMissingKeyHandlers(b *testing.B) {
	benchResetMissingKeyHooks(b)
	missingKeyHandler.Store(missingKeyHandlersState{handlers: []MissingKeyHandler{benchMissingKeyHandler}})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = missingKeyHandlers()
	}
}

func BenchmarkDispatchMissingKey(b *testing.B) {
	benchResetMissingKeyHooks(b)
	SetMissingKeyHandlers(benchMissingKeyHandler)
	args := map[string]any{"Context": "dashboard"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dispatchMissingKey("missing.key", args)
	}
}

func BenchmarkCloneMissingKey(b *testing.B) {
	missing := MissingKey{
		Key:        "missing.key",
		Args:       map[string]any{"Context": "dashboard"},
		CallerFile: "hooks_test.go",
		CallerLine: 42,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchMissingKeySink = cloneMissingKey(missing)
	}
}

func BenchmarkMissingKeyCaller(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, line := missingKeyCaller()
		benchServiceStringSink = file
		benchServiceIntSink = line
	}
}
