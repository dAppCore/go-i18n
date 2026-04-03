package i18n

import (
	"io/fs"
	"runtime"
	"sync"
	"sync/atomic"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
)

var missingKeyHandler atomic.Value
var missingKeyHandlerMu sync.Mutex

type missingKeyHandlersState struct {
	handlers []MissingKeyHandler
}

type localeRegistration struct {
	fsys fs.FS
	dir  string
	id   int
}

type localeProviderRegistration struct {
	provider LocaleProvider
	id       int
}

// LocaleProvider supplies one or more locale filesystems to the default service.
//
//	i18n.RegisterLocaleProvider(myProvider)
type LocaleProvider interface {
	LocaleSources() []FSSource
}

var (
	registeredLocales         []localeRegistration
	registeredLocaleProviders []localeProviderRegistration
	registeredLocalesMu       sync.Mutex
	localesLoaded             bool
	nextLocaleRegistrationID  int
	nextLocaleProviderID      int
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
	reg := localeRegistration{fsys: fsys, dir: dir}
	registeredLocalesMu.Lock()
	nextLocaleRegistrationID++
	reg.id = nextLocaleRegistrationID
	registeredLocales = append(registeredLocales, reg)
	svc := defaultService.Load()
	registeredLocalesMu.Unlock()
	if svc != nil {
		if err := svc.LoadFS(fsys, dir); err != nil {
			log.Error("i18n: RegisterLocales failed to load", "dir", dir, "err", err)
		} else {
			svc.markLocaleRegistrationLoaded(reg.id)
			markLocalesLoaded()
		}
	}
}

// RegisterLocaleProvider registers a provider that can contribute locale files.
// This is useful for packages that need to expose multiple locale sources as a
// single unit.
//
//	i18n.RegisterLocaleProvider(myProvider)
func RegisterLocaleProvider(provider LocaleProvider) {
	if provider == nil {
		return
	}
	reg := localeProviderRegistration{provider: provider}
	registeredLocalesMu.Lock()
	nextLocaleProviderID++
	reg.id = nextLocaleProviderID
	registeredLocaleProviders = append(registeredLocaleProviders, reg)
	svc := defaultService.Load()
	registeredLocalesMu.Unlock()
	if svc != nil {
		loadLocaleProvider(svc, reg)
	}
}

func loadRegisteredLocales(svc *Service) {
	if svc == nil {
		return
	}
	registeredLocalesMu.Lock()
	locales := append([]localeRegistration(nil), registeredLocales...)
	providers := append([]localeProviderRegistration(nil), registeredLocaleProviders...)
	registeredLocalesMu.Unlock()

	for _, reg := range locales {
		if svc != nil && svc.hasLocaleRegistrationLoaded(reg.id) {
			continue
		}
		if err := svc.LoadFS(reg.fsys, reg.dir); err != nil {
			log.Error("i18n: loadRegisteredLocales failed to load", "dir", reg.dir, "err", err)
			continue
		}
		svc.markLocaleRegistrationLoaded(reg.id)
	}
	for _, provider := range providers {
		if svc != nil && svc.hasLocaleProviderLoaded(provider.id) {
			continue
		}
		loadLocaleProvider(svc, provider)
	}

	markLocalesLoaded()
}

func loadLocaleProvider(svc *Service, provider localeProviderRegistration) {
	if svc == nil || provider.provider == nil {
		return
	}
	for _, src := range provider.provider.LocaleSources() {
		if err := svc.LoadFS(src.FS, src.Dir); err != nil {
			log.Error("i18n: loadLocaleProvider failed to load", "dir", src.Dir, "err", err)
		}
	}
	svc.markLocaleProviderLoaded(provider.id)
	markLocalesLoaded()
}

func markLocalesLoaded() {
	registeredLocalesMu.Lock()
	localesLoaded = true
	registeredLocalesMu.Unlock()
}

// OnMissingKey registers a handler for missing translation keys.
func OnMissingKey(h MissingKeyHandler) {
	if h == nil {
		ClearMissingKeyHandlers()
		return
	}
	AddMissingKeyHandler(h)
}

// SetMissingKeyHandlers replaces the full missing-key handler chain.
func SetMissingKeyHandlers(handlers ...MissingKeyHandler) {
	missingKeyHandlerMu.Lock()
	defer missingKeyHandlerMu.Unlock()
	handlers = filterNilMissingKeyHandlers(handlers)
	if len(handlers) == 0 {
		missingKeyHandler.Store(missingKeyHandlersState{})
		return
	}
	missingKeyHandler.Store(missingKeyHandlersState{handlers: handlers})
}

// ClearMissingKeyHandlers removes all registered missing-key handlers.
func ClearMissingKeyHandlers() {
	missingKeyHandlerMu.Lock()
	defer missingKeyHandlerMu.Unlock()
	missingKeyHandler.Store(missingKeyHandlersState{})
}

// AddMissingKeyHandler appends a missing-key handler without replacing any
// existing handlers.
func AddMissingKeyHandler(h MissingKeyHandler) {
	if h == nil {
		return
	}
	missingKeyHandlerMu.Lock()
	defer missingKeyHandlerMu.Unlock()
	current := missingKeyHandlers()
	current.handlers = append(current.handlers, h)
	missingKeyHandler.Store(current)
}

func filterNilMissingKeyHandlers(handlers []MissingKeyHandler) []MissingKeyHandler {
	if len(handlers) == 0 {
		return nil
	}
	filtered := make([]MissingKeyHandler, 0, len(handlers))
	for _, h := range handlers {
		if h != nil {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
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
	mk := cloneMissingKey(MissingKey{Key: key, Args: args, CallerFile: file, CallerLine: line})
	for _, h := range state.handlers {
		if h != nil {
			h(mk)
		}
	}
}

func cloneMissingKey(mk MissingKey) MissingKey {
	if len(mk.Args) == 0 {
		mk.Args = nil
		return mk
	}
	args := make(map[string]any, len(mk.Args))
	for key, value := range mk.Args {
		args[key] = value
	}
	mk.Args = args
	return mk
}

func missingKeyCaller() (string, int) {
	const packagePrefix = "dappco.re/go/core/i18n."

	pcs := make([]uintptr, 16)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !core.HasPrefix(frame.Function, packagePrefix) || core.HasSuffix(frame.File, "_test.go") {
			return frame.File, frame.Line
		}
		if !more {
			break
		}
	}
	return "unknown", 0
}
