// SPDX-License-Identifier: EUPL-1.2

package i18n

import (
	"context"
	"io/fs"
	"sync"

	"dappco.re/go/core"
)

// CoreService wraps the i18n Service as a Core framework service.
// Register with: core.WithName("i18n", i18n.NewCoreService(i18n.ServiceOptions{}))
type CoreService struct {
	*core.ServiceRuntime[ServiceOptions]
	svc *Service

	missingKeys   []MissingKey
	missingKeysMu sync.Mutex
	hookInstalled bool
}

// ServiceOptions configures the i18n Core service.
type ServiceOptions struct {
	// Language overrides auto-detection (e.g., "en-GB", "de")
	Language string
	// Fallback sets the fallback language for missing translations.
	Fallback string
	// Formality sets the default formality level.
	Formality Formality
	// Location sets the default location context.
	Location string
	// Mode sets the translation mode (Normal, Strict, Collect)
	Mode Mode
	// Debug prefixes translated output with the message key.
	Debug bool
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
func NewCoreService(opts ServiceOptions) func(*core.Core) (any, error) {
	return func(c *core.Core) (any, error) {
		svc, err := New()
		if err != nil {
			return nil, err
		}

		for _, src := range opts.ExtraFS {
			loader := NewFSLoader(src.FS, src.Dir)
			if addErr := svc.AddLoader(loader); addErr != nil {
				// Non-fatal — skip sources that fail (e.g. missing language files)
				continue
			}
		}

		// Preserve the same init-time locale registration behaviour used by Init().
		// Core bootstrap should not bypass packages that registered locale files
		// before the service was constructed.
		loadRegisteredLocales(svc)

		if opts.Language != "" {
			if langErr := svc.SetLanguage(opts.Language); langErr != nil {
				return nil, langErr
			}
		}
		if opts.Fallback != "" {
			svc.SetFallback(opts.Fallback)
		}
		if opts.Formality != FormalityNeutral {
			svc.SetFormality(opts.Formality)
		}
		if opts.Location != "" {
			svc.SetLocation(opts.Location)
		}

		svc.SetMode(opts.Mode)
		svc.SetDebug(opts.Debug)
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
		s.ensureMissingKeyCollector()
	}
	return nil
}

func (s *CoreService) ensureMissingKeyCollector() {
	if s.hookInstalled {
		return
	}
	AddMissingKeyHandler(s.handleMissingKey)
	s.hookInstalled = true
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
		s.ensureMissingKeyCollector()
	}
}

// Mode returns the current translation mode.
func (s *CoreService) Mode() Mode {
	return s.svc.Mode()
}

// CurrentMode returns the current translation mode.
func (s *CoreService) CurrentMode() Mode {
	return s.Mode()
}

// T translates a message through the wrapped i18n service.
func (s *CoreService) T(messageID string, args ...any) string {
	return s.svc.T(messageID, args...)
}

// Translate translates a message through the wrapped i18n service.
func (s *CoreService) Translate(messageID string, args ...any) core.Result {
	return s.svc.Translate(messageID, args...)
}

// Raw translates without namespace handler magic.
func (s *CoreService) Raw(messageID string, args ...any) string {
	return s.svc.Raw(messageID, args...)
}

// AddMessages adds message strings to the wrapped service.
func (s *CoreService) AddMessages(lang string, messages map[string]string) {
	s.svc.AddMessages(lang, messages)
}

// SetLanguage changes the wrapped service language.
func (s *CoreService) SetLanguage(lang string) error {
	return s.svc.SetLanguage(lang)
}

// Language returns the wrapped service language.
func (s *CoreService) Language() string {
	return s.svc.Language()
}

// CurrentLanguage returns the wrapped service language.
func (s *CoreService) CurrentLanguage() string {
	return s.Language()
}

// SetFallback changes the wrapped service fallback language.
func (s *CoreService) SetFallback(lang string) {
	s.svc.SetFallback(lang)
}

// Fallback returns the wrapped service fallback language.
func (s *CoreService) Fallback() string {
	return s.svc.Fallback()
}

// CurrentFallback returns the wrapped service fallback language.
func (s *CoreService) CurrentFallback() string {
	return s.Fallback()
}

// SetFormality changes the wrapped service default formality.
func (s *CoreService) SetFormality(f Formality) {
	s.svc.SetFormality(f)
}

// Formality returns the wrapped service default formality.
func (s *CoreService) Formality() Formality {
	return s.svc.Formality()
}

// CurrentFormality returns the wrapped service default formality.
func (s *CoreService) CurrentFormality() Formality {
	return s.Formality()
}

// SetLocation changes the wrapped service default location.
func (s *CoreService) SetLocation(location string) {
	s.svc.SetLocation(location)
}

// Location returns the wrapped service default location.
func (s *CoreService) Location() string {
	return s.svc.Location()
}

// CurrentLocation returns the wrapped service default location.
func (s *CoreService) CurrentLocation() string {
	return s.Location()
}

// SetDebug changes the wrapped service debug mode.
func (s *CoreService) SetDebug(enabled bool) {
	s.svc.SetDebug(enabled)
}

// Debug reports whether wrapped service debug mode is enabled.
func (s *CoreService) Debug() bool {
	return s.svc.Debug()
}

// CurrentDebug reports whether wrapped service debug mode is enabled.
func (s *CoreService) CurrentDebug() bool {
	return s.Debug()
}

// AddHandler appends handlers to the wrapped service's chain.
func (s *CoreService) AddHandler(handlers ...KeyHandler) {
	s.svc.AddHandler(handlers...)
}

// SetHandlers replaces the wrapped service's handler chain.
func (s *CoreService) SetHandlers(handlers ...KeyHandler) {
	s.svc.SetHandlers(handlers...)
}

// PrependHandler inserts handlers at the front of the wrapped service's chain.
func (s *CoreService) PrependHandler(handlers ...KeyHandler) {
	s.svc.PrependHandler(handlers...)
}

// ClearHandlers removes all handlers from the wrapped service.
func (s *CoreService) ClearHandlers() {
	s.svc.ClearHandlers()
}

// ResetHandlers restores the wrapped service's default handler chain.
func (s *CoreService) ResetHandlers() {
	s.svc.ResetHandlers()
}

// Handlers returns a copy of the wrapped service's handler chain.
func (s *CoreService) Handlers() []KeyHandler {
	return s.svc.Handlers()
}

// CurrentHandlers returns a copy of the wrapped service's handler chain.
func (s *CoreService) CurrentHandlers() []KeyHandler {
	return s.Handlers()
}

// AddLoader loads extra locale data into the wrapped service.
func (s *CoreService) AddLoader(loader Loader) error {
	return s.svc.AddLoader(loader)
}

// LoadFS loads locale data from a filesystem into the wrapped service.
func (s *CoreService) LoadFS(fsys fs.FS, dir string) error {
	return s.svc.LoadFS(fsys, dir)
}

// AvailableLanguages returns the wrapped service languages.
func (s *CoreService) AvailableLanguages() []string {
	return s.svc.AvailableLanguages()
}

// CurrentAvailableLanguages returns the wrapped service languages.
func (s *CoreService) CurrentAvailableLanguages() []string {
	return s.AvailableLanguages()
}

// Direction returns the wrapped service text direction.
func (s *CoreService) Direction() TextDirection {
	return s.svc.Direction()
}

// CurrentDirection returns the wrapped service text direction.
func (s *CoreService) CurrentDirection() TextDirection {
	return s.Direction()
}

// IsRTL reports whether the wrapped service language is right-to-left.
func (s *CoreService) IsRTL() bool {
	return s.svc.IsRTL()
}

// PluralCategory returns the plural category for the wrapped service language.
func (s *CoreService) PluralCategory(n int) PluralCategory {
	return s.svc.PluralCategory(n)
}

// CurrentPluralCategory returns the plural category for the wrapped service language.
func (s *CoreService) CurrentPluralCategory(n int) PluralCategory {
	return s.PluralCategory(n)
}
