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

var (
	registeredLocales   []localeRegistration
	registeredLocalesMu sync.Mutex
	localesLoaded       bool
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

func loadRegisteredLocales(svc *Service) {
	registeredLocalesMu.Lock()
	defer registeredLocalesMu.Unlock()
	for _, reg := range registeredLocales {
		if err := svc.LoadFS(reg.fsys, reg.dir); err != nil {
			log.Printf("i18n: loadRegisteredLocales failed to load %q: %v", reg.dir, err)
		}
	}
	localesLoaded = true
}

// OnMissingKey registers a handler for missing translation keys.
func OnMissingKey(h MissingKeyHandler) {
	if h == nil {
		missingKeyHandler.Store(missingKeyHandlersState{})
		return
	}
	missingKeyHandler.Store(missingKeyHandlersState{handlers: []MissingKeyHandler{h}})
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
