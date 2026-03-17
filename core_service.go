// SPDX-License-Identifier: EUPL-1.2

package i18n

import (
	"context"
	"io/fs"
	"sync"

	"forge.lthn.ai/core/go/pkg/core"
)

// CoreService wraps the i18n Service as a Core framework service.
// Register with: core.WithName("i18n", i18n.NewCoreService(i18n.ServiceOptions{}))
type CoreService struct {
	*core.ServiceRuntime[ServiceOptions]
	svc *Service

	missingKeys   []MissingKey
	missingKeysMu sync.Mutex
}

// ServiceOptions configures the i18n Core service.
type ServiceOptions struct {
	// Language overrides auto-detection (e.g., "en-GB", "de")
	Language string
	// Mode sets the translation mode (Normal, Strict, Collect)
	Mode Mode
	// ExtraFS loads additional translation files on top of the embedded defaults.
	// Each entry is an fs.FS + directory path within it.
	ExtraFS []FSSource
}

// FSSource pairs a filesystem with a directory path for loading translations.
type FSSource struct {
	FS  fs.FS
	Dir string
}

// NewCoreService creates an i18n Core service factory.
// Automatically loads locale filesystems from:
// 1. Embedded go-i18n base translations (grammar, verbs, nouns)
// 2. ExtraFS sources passed via ServiceOptions
// 3. Locales collected from Core services implementing LocaleProvider
func NewCoreService(opts ServiceOptions) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		svc, err := New()
		if err != nil {
			return nil, err
		}

		// Load additional translation sources from options
		for _, src := range opts.ExtraFS {
			if loadErr := svc.LoadFS(src.FS, src.Dir); loadErr != nil {
				return nil, loadErr
			}
		}

		// Load locale filesystems collected from Core services
		// (services that implement core.LocaleProvider)
		for _, localeFS := range c.Locales() {
			loader := NewFSLoader(localeFS, ".")
			if addErr := svc.AddLoader(loader); addErr != nil {
				// Non-fatal — log and continue
				continue
			}
		}

		if opts.Language != "" {
			_ = svc.SetLanguage(opts.Language)
		}

		svc.SetMode(opts.Mode)
		SetDefault(svc)

		return &CoreService{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
			svc:            svc,
			missingKeys:    make([]MissingKey, 0),
		}, nil
	}
}

// OnStartup initialises the i18n service.
func (s *CoreService) OnStartup(_ context.Context) error {
	if s.svc.Mode() == ModeCollect {
		OnMissingKey(s.handleMissingKey)
	}
	return nil
}

func (s *CoreService) handleMissingKey(mk MissingKey) {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = append(s.missingKeys, mk)
}

// MissingKeys returns all missing keys collected in collect mode.
func (s *CoreService) MissingKeys() []MissingKey {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	result := make([]MissingKey, len(s.missingKeys))
	copy(result, s.missingKeys)
	return result
}

// ClearMissingKeys resets the collected missing keys.
func (s *CoreService) ClearMissingKeys() {
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = s.missingKeys[:0]
}

// SetMode changes the translation mode.
func (s *CoreService) SetMode(mode Mode) {
	s.svc.SetMode(mode)
	if mode == ModeCollect {
		OnMissingKey(s.handleMissingKey)
	} else {
		OnMissingKey(nil)
	}
}

// Mode returns the current translation mode.
func (s *CoreService) Mode() Mode {
	return s.svc.Mode()
}
