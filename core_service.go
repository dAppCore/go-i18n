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

func (s *CoreService) wrapped() *Service {
	if s == nil {
		return nil
	}
	return s.svc
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
	if svc := s.wrapped(); svc != nil && svc.Mode() == ModeCollect {
		s.ensureMissingKeyCollector()
	}
	return nil
}

func (s *CoreService) ensureMissingKeyCollector() {
	if s == nil || s.svc == nil || s.hookInstalled {
		return
	}
	AddMissingKeyHandler(s.handleMissingKey)
	s.hookInstalled = true
}

func (s *CoreService) handleMissingKey(mk MissingKey) {
	if s == nil {
		return
	}
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = append(s.missingKeys, cloneMissingKey(mk))
}

// MissingKeys returns all missing keys collected in collect mode.
func (s *CoreService) MissingKeys() []MissingKey {
	if s == nil {
		return []MissingKey{}
	}
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	result := make([]MissingKey, len(s.missingKeys))
	for i, mk := range s.missingKeys {
		result[i] = cloneMissingKey(mk)
	}
	return result
}

// ClearMissingKeys resets the collected missing keys.
func (s *CoreService) ClearMissingKeys() {
	if s == nil {
		return
	}
	s.missingKeysMu.Lock()
	defer s.missingKeysMu.Unlock()
	s.missingKeys = s.missingKeys[:0]
}

// SetMode changes the translation mode.
func (s *CoreService) SetMode(mode Mode) {
	if svc := s.wrapped(); svc != nil {
		svc.SetMode(mode)
	}
	if s != nil && s.svc != nil && mode == ModeCollect {
		s.ensureMissingKeyCollector()
	}
}

// Mode returns the current translation mode.
func (s *CoreService) Mode() Mode {
	if svc := s.wrapped(); svc != nil {
		return svc.Mode()
	}
	return ModeNormal
}

// CurrentMode returns the current translation mode.
func (s *CoreService) CurrentMode() Mode {
	return s.Mode()
}

// T translates a message through the wrapped i18n service.
func (s *CoreService) T(messageID string, args ...any) string {
	if svc := s.wrapped(); svc != nil {
		return svc.T(messageID, args...)
	}
	return messageID
}

// Translate translates a message through the wrapped i18n service.
func (s *CoreService) Translate(messageID string, args ...any) core.Result {
	if svc := s.wrapped(); svc != nil {
		return svc.Translate(messageID, args...)
	}
	return core.Result{Value: messageID, OK: false}
}

// Raw translates without namespace handler magic.
func (s *CoreService) Raw(messageID string, args ...any) string {
	if svc := s.wrapped(); svc != nil {
		return svc.Raw(messageID, args...)
	}
	return messageID
}

// AddMessages adds message strings to the wrapped service.
func (s *CoreService) AddMessages(lang string, messages map[string]string) {
	if svc := s.wrapped(); svc != nil {
		svc.AddMessages(lang, messages)
	}
}

// SetLanguage changes the wrapped service language.
func (s *CoreService) SetLanguage(lang string) error {
	if svc := s.wrapped(); svc != nil {
		return svc.SetLanguage(lang)
	}
	return ErrServiceNotInitialised
}

// Language returns the wrapped service language.
func (s *CoreService) Language() string {
	if svc := s.wrapped(); svc != nil {
		return svc.Language()
	}
	return "en"
}

// CurrentLanguage returns the wrapped service language.
func (s *CoreService) CurrentLanguage() string {
	return s.Language()
}

// CurrentLang is a short alias for CurrentLanguage.
func (s *CoreService) CurrentLang() string {
	return s.CurrentLanguage()
}

// Prompt translates a prompt key from the prompt namespace using the wrapped service.
func (s *CoreService) Prompt(key string) string {
	if svc := s.wrapped(); svc != nil {
		return svc.Prompt(key)
	}
	return namespaceLookupKey("prompt", key)
}

// CurrentPrompt is a short alias for Prompt.
func (s *CoreService) CurrentPrompt(key string) string {
	return s.Prompt(key)
}

// Lang translates a language label from the lang namespace using the wrapped service.
func (s *CoreService) Lang(key string) string {
	if svc := s.wrapped(); svc != nil {
		return svc.Lang(key)
	}
	return namespaceLookupKey("lang", key)
}

// SetFallback changes the wrapped service fallback language.
func (s *CoreService) SetFallback(lang string) {
	if svc := s.wrapped(); svc != nil {
		svc.SetFallback(lang)
	}
}

// Fallback returns the wrapped service fallback language.
func (s *CoreService) Fallback() string {
	if svc := s.wrapped(); svc != nil {
		return svc.Fallback()
	}
	return "en"
}

// CurrentFallback returns the wrapped service fallback language.
func (s *CoreService) CurrentFallback() string {
	return s.Fallback()
}

// SetFormality changes the wrapped service default formality.
func (s *CoreService) SetFormality(f Formality) {
	if svc := s.wrapped(); svc != nil {
		svc.SetFormality(f)
	}
}

// Formality returns the wrapped service default formality.
func (s *CoreService) Formality() Formality {
	if svc := s.wrapped(); svc != nil {
		return svc.Formality()
	}
	return FormalityNeutral
}

// CurrentFormality returns the wrapped service default formality.
func (s *CoreService) CurrentFormality() Formality {
	return s.Formality()
}

// SetLocation changes the wrapped service default location.
func (s *CoreService) SetLocation(location string) {
	if svc := s.wrapped(); svc != nil {
		svc.SetLocation(location)
	}
}

// Location returns the wrapped service default location.
func (s *CoreService) Location() string {
	if svc := s.wrapped(); svc != nil {
		return svc.Location()
	}
	return ""
}

// CurrentLocation returns the wrapped service default location.
func (s *CoreService) CurrentLocation() string {
	return s.Location()
}

// SetDebug changes the wrapped service debug mode.
func (s *CoreService) SetDebug(enabled bool) {
	if svc := s.wrapped(); svc != nil {
		svc.SetDebug(enabled)
	}
}

// Debug reports whether wrapped service debug mode is enabled.
func (s *CoreService) Debug() bool {
	if svc := s.wrapped(); svc != nil {
		return svc.Debug()
	}
	return false
}

// CurrentDebug reports whether wrapped service debug mode is enabled.
func (s *CoreService) CurrentDebug() bool {
	return s.Debug()
}

// State returns a copy-safe snapshot of the wrapped service configuration.
func (s *CoreService) State() ServiceState {
	if s == nil || s.svc == nil {
		return defaultServiceStateSnapshot()
	}
	return s.svc.State()
}

// CurrentState is a more explicit alias for State.
func (s *CoreService) CurrentState() ServiceState {
	return s.State()
}

// AddHandler appends handlers to the wrapped service's chain.
func (s *CoreService) AddHandler(handlers ...KeyHandler) {
	if svc := s.wrapped(); svc != nil {
		svc.AddHandler(handlers...)
	}
}

// SetHandlers replaces the wrapped service's handler chain.
func (s *CoreService) SetHandlers(handlers ...KeyHandler) {
	if svc := s.wrapped(); svc != nil {
		svc.SetHandlers(handlers...)
	}
}

// PrependHandler inserts handlers at the front of the wrapped service's chain.
func (s *CoreService) PrependHandler(handlers ...KeyHandler) {
	if svc := s.wrapped(); svc != nil {
		svc.PrependHandler(handlers...)
	}
}

// ClearHandlers removes all handlers from the wrapped service.
func (s *CoreService) ClearHandlers() {
	if svc := s.wrapped(); svc != nil {
		svc.ClearHandlers()
	}
}

// ResetHandlers restores the wrapped service's default handler chain.
func (s *CoreService) ResetHandlers() {
	if svc := s.wrapped(); svc != nil {
		svc.ResetHandlers()
	}
}

// Handlers returns a copy of the wrapped service's handler chain.
func (s *CoreService) Handlers() []KeyHandler {
	if svc := s.wrapped(); svc != nil {
		return svc.Handlers()
	}
	return []KeyHandler{}
}

// CurrentHandlers returns a copy of the wrapped service's handler chain.
func (s *CoreService) CurrentHandlers() []KeyHandler {
	return s.Handlers()
}

// AddLoader loads extra locale data into the wrapped service.
func (s *CoreService) AddLoader(loader Loader) error {
	if svc := s.wrapped(); svc != nil {
		return svc.AddLoader(loader)
	}
	return ErrServiceNotInitialised
}

// LoadFS loads locale data from a filesystem into the wrapped service.
func (s *CoreService) LoadFS(fsys fs.FS, dir string) error {
	if svc := s.wrapped(); svc != nil {
		return svc.LoadFS(fsys, dir)
	}
	return ErrServiceNotInitialised
}

// AvailableLanguages returns the wrapped service languages.
func (s *CoreService) AvailableLanguages() []string {
	if svc := s.wrapped(); svc != nil {
		return svc.AvailableLanguages()
	}
	return []string{}
}

// CurrentAvailableLanguages returns the wrapped service languages.
func (s *CoreService) CurrentAvailableLanguages() []string {
	return s.AvailableLanguages()
}

// Direction returns the wrapped service text direction.
func (s *CoreService) Direction() TextDirection {
	if svc := s.wrapped(); svc != nil {
		return svc.Direction()
	}
	return DirLTR
}

// CurrentDirection returns the wrapped service text direction.
func (s *CoreService) CurrentDirection() TextDirection {
	return s.Direction()
}

// CurrentTextDirection is a more explicit alias for CurrentDirection.
func (s *CoreService) CurrentTextDirection() TextDirection {
	return s.CurrentDirection()
}

// IsRTL reports whether the wrapped service language is right-to-left.
func (s *CoreService) IsRTL() bool {
	if svc := s.wrapped(); svc != nil {
		return svc.IsRTL()
	}
	return false
}

// RTL reports whether the wrapped service language is right-to-left.
func (s *CoreService) RTL() bool {
	return s.IsRTL()
}

// CurrentIsRTL reports whether the wrapped service language is right-to-left.
func (s *CoreService) CurrentIsRTL() bool {
	return s.IsRTL()
}

// CurrentRTL reports whether the wrapped service language is right-to-left.
func (s *CoreService) CurrentRTL() bool {
	return s.CurrentIsRTL()
}

// PluralCategory returns the plural category for the wrapped service language.
func (s *CoreService) PluralCategory(n int) PluralCategory {
	if svc := s.wrapped(); svc != nil {
		return svc.PluralCategory(n)
	}
	return PluralOther
}

// CurrentPluralCategory returns the plural category for the wrapped service language.
func (s *CoreService) CurrentPluralCategory(n int) PluralCategory {
	return s.PluralCategory(n)
}

// PluralCategoryOf is a short alias for CurrentPluralCategory.
func (s *CoreService) PluralCategoryOf(n int) PluralCategory {
	return s.CurrentPluralCategory(n)
}
