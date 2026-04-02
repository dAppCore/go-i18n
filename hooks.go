package i18n

import (
	"io/fs"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var missingKeyHandler atomic.Value

type missingKeyHandlersState struct {
	handlers []MissingKeyHandler
}

type localeRegistration struct {
	fsys fs.FS
	dir  string
}

// LocaleProvider supplies one or more locale filesystems to the default service.
type LocaleProvider interface {
	LocaleSources() []FSSource
}

var (
	registeredLocales        []localeRegistration
	registeredLocaleProviders []LocaleProvider
	registeredLocalesMu      sync.Mutex
	localesLoaded            bool
)

// RegisterLocales registers a filesystem containing locale files.
// Call this in your package's init() to register translations.
//
//	//go:embed locales/*.json
//	var localeFS embed.FS
//
//	func init() {
//	    i18n.RegisterLocales(localeFS, "locales")
//	}
func RegisterLocales(fsys fs.FS, dir string) {
	registeredLocalesMu.Lock()
	defer registeredLocalesMu.Unlock()
	registeredLocales = append(registeredLocales, localeRegistration{fsys: fsys, dir: dir})
	if svc := defaultService.Load(); svc != nil {
		if err := svc.LoadFS(fsys, dir); err != nil {
			log.Printf("i18n: RegisterLocales failed to load %q: %v", dir, err)
		}
	}
}

// RegisterLocaleProvider registers a provider that can contribute locale files.
// This is useful for packages that need to expose multiple locale sources as a
// single unit.
func RegisterLocaleProvider(provider LocaleProvider) {
	if provider == nil {
		return
	}
	registeredLocalesMu.Lock()
	registeredLocaleProviders = append(registeredLocaleProviders, provider)
	registeredLocalesMu.Unlock()
	if svc := defaultService.Load(); svc != nil {
		loadLocaleProvider(svc, provider)
	}
}

func loadRegisteredLocales(svc *Service) {
	registeredLocalesMu.Lock()
	locales := append([]localeRegistration(nil), registeredLocales...)
	providers := append([]LocaleProvider(nil), registeredLocaleProviders...)
	registeredLocalesMu.Unlock()

	for _, reg := range locales {
		if err := svc.LoadFS(reg.fsys, reg.dir); err != nil {
			log.Printf("i18n: loadRegisteredLocales failed to load %q: %v", reg.dir, err)
		}
	}
	for _, provider := range providers {
		loadLocaleProvider(svc, provider)
	}

	registeredLocalesMu.Lock()
	localesLoaded = true
	registeredLocalesMu.Unlock()
}

func loadLocaleProvider(svc *Service, provider LocaleProvider) {
	if svc == nil || provider == nil {
		return
	}
	for _, src := range provider.LocaleSources() {
		if err := svc.LoadFS(src.FS, src.Dir); err != nil {
			log.Printf("i18n: loadLocaleProvider failed to load %q: %v", src.Dir, err)
		}
	}
}

// OnMissingKey registers a handler for missing translation keys.
func OnMissingKey(h MissingKeyHandler) {
	if h == nil {
		missingKeyHandler.Store(missingKeyHandlersState{})
		return
	}
	missingKeyHandler.Store(missingKeyHandlersState{handlers: []MissingKeyHandler{h}})
}

// ClearMissingKeyHandlers removes all registered missing-key handlers.
func ClearMissingKeyHandlers() {
	missingKeyHandler.Store(missingKeyHandlersState{})
}

// AddMissingKeyHandler appends a missing-key handler without replacing any
// existing handlers.
func AddMissingKeyHandler(h MissingKeyHandler) {
	if h == nil {
		return
	}
	current := missingKeyHandlers()
	current.handlers = append(current.handlers, h)
	missingKeyHandler.Store(current)
}

func appendMissingKeyHandler(h MissingKeyHandler) {
	AddMissingKeyHandler(h)
}

func missingKeyHandlers() missingKeyHandlersState {
	v := missingKeyHandler.Load()
	if v == nil {
		return missingKeyHandlersState{}
	}
	state, ok := v.(missingKeyHandlersState)
	if !ok {
		return missingKeyHandlersState{}
	}
	return state
}

func dispatchMissingKey(key string, args map[string]any) {
	state := missingKeyHandlers()
	if len(state.handlers) == 0 {
		return
	}
	file, line := missingKeyCaller()
	mk := MissingKey{Key: key, Args: args, CallerFile: file, CallerLine: line}
	for _, h := range state.handlers {
		if h != nil {
			h(mk)
		}
	}
}

func missingKeyCaller() (string, int) {
	const packagePrefix = "dappco.re/go/core/i18n."

	pcs := make([]uintptr, 16)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, packagePrefix) || strings.HasSuffix(frame.File, "_test.go") {
			return frame.File, frame.Line
		}
		if !more {
			break
		}
	}
	return "unknown", 0
}
