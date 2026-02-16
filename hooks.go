package i18n

import (
	"io/fs"
	"runtime"
	"sync"
	"sync/atomic"
)

var missingKeyHandler atomic.Value

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
	if localesLoaded {
		if svc := Default(); svc != nil {
			_ = svc.LoadFS(fsys, dir)
		}
	}
}

func loadRegisteredLocales(svc *Service) {
	registeredLocalesMu.Lock()
	defer registeredLocalesMu.Unlock()
	for _, reg := range registeredLocales {
		_ = svc.LoadFS(reg.fsys, reg.dir)
	}
	localesLoaded = true
}

// OnMissingKey registers a handler for missing translation keys.
func OnMissingKey(h MissingKeyHandler) {
	missingKeyHandler.Store(h)
}

func dispatchMissingKey(key string, args map[string]any) {
	v := missingKeyHandler.Load()
	if v == nil {
		return
	}
	h, ok := v.(MissingKeyHandler)
	if !ok || h == nil {
		return
	}
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	h(MissingKey{Key: key, Args: args, CallerFile: file, CallerLine: line})
}
