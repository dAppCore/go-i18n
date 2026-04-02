package i18n

import (
	"dappco.re/go/core"
)

func defaultServiceStateSnapshot() ServiceState {
	// Keep the nil/default snapshot aligned with Service.State() so callers get
	// the same shape regardless of whether a Service has been initialised.
	return ServiceState{
		Language:           "en",
		RequestedLanguage:  "",
		LanguageExplicit:   false,
		AvailableLanguages: []string{},
		Mode:               ModeNormal,
		Fallback:           "en",
		Formality:          FormalityNeutral,
		Location:           "",
		Direction:          DirLTR,
		IsRTL:              false,
		Debug:              false,
		Handlers:           []KeyHandler{},
	}
}

// ServiceState captures the current configuration of a service in one
// copy-safe snapshot.
type ServiceState struct {
	Language           string
	RequestedLanguage  string
	LanguageExplicit   bool
	AvailableLanguages []string
	Mode               Mode
	Fallback           string
	Formality          Formality
	Location           string
	Direction          TextDirection
	IsRTL              bool
	Debug              bool
	Handlers           []KeyHandler
}

// String returns a concise, stable summary of the service snapshot.
func (s ServiceState) String() string {
	langs := "[]"
	if len(s.AvailableLanguages) > 0 {
		langs = "[" + core.Join(", ", s.AvailableLanguages...) + "]"
	}
	handlers := "[]"
	if len(s.Handlers) > 0 {
		names := make([]string, 0, len(s.Handlers))
		for _, handler := range s.Handlers {
			if handler == nil {
				names = append(names, "<nil>")
				continue
			}
			names = append(names, core.Sprintf("%T", handler))
		}
		handlers = "[" + core.Join(", ", names...) + "]"
	}
	return core.Sprintf(
		"ServiceState{language=%q requested=%q explicit=%t fallback=%q mode=%s formality=%s location=%q direction=%s rtl=%t debug=%t available=%s handlers=%d types=%s}",
		s.Language,
		s.RequestedLanguage,
		s.LanguageExplicit,
		s.Fallback,
		s.Mode,
		s.Formality,
		s.Location,
		s.Direction,
		s.IsRTL,
		s.Debug,
		langs,
		len(s.Handlers),
		handlers,
	)
}

func (s *Service) State() ServiceState {
	if s == nil {
		return defaultServiceStateSnapshot()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	langs := make([]string, len(s.availableLangs))
	for i, tag := range s.availableLangs {
		langs[i] = tag.String()
	}

	handlers := make([]KeyHandler, len(s.handlers))
	copy(handlers, s.handlers)

	dir := DirLTR
	if IsRTLLanguage(s.currentLang) {
		dir = DirRTL
	}

	return ServiceState{
		Language:           s.currentLang,
		RequestedLanguage:  s.requestedLang,
		LanguageExplicit:   s.languageExplicit,
		AvailableLanguages: langs,
		Mode:               s.mode,
		Fallback:           s.fallbackLang,
		Formality:          s.formality,
		Location:           s.location,
		Direction:          dir,
		IsRTL:              dir == DirRTL,
		Debug:              s.debug,
		Handlers:           handlers,
	}
}

// CurrentState is a more explicit alias for State.
func (s *Service) CurrentState() ServiceState {
	return s.State()
}
