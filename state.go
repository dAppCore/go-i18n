package i18n

import (
	"dappco.re/go/core"
)

func newServiceStateSnapshot(
	language string,
	requestedLanguage string,
	languageExplicit bool,
	availableLanguages []string,
	mode Mode,
	fallback string,
	formality Formality,
	location string,
	direction TextDirection,
	debug bool,
	handlers []KeyHandler,
) ServiceState {
	return ServiceState{
		Language:           language,
		RequestedLanguage:  requestedLanguage,
		LanguageExplicit:   languageExplicit,
		AvailableLanguages: availableLanguages,
		Mode:               mode,
		Fallback:           fallback,
		Formality:          formality,
		Location:           location,
		Direction:          direction,
		IsRTL:              direction == DirRTL,
		Debug:              debug,
		Handlers:           handlers,
	}
}

func defaultServiceStateSnapshot() ServiceState {
	// Keep the nil/default snapshot aligned with Service.State() so callers get
	// the same shape regardless of whether a Service has been initialised.
	return newServiceStateSnapshot(
		"en",
		"",
		false,
		[]string{},
		ModeNormal,
		"en",
		FormalityNeutral,
		"",
		DirLTR,
		false,
		[]KeyHandler{},
	)
}

// ServiceState captures the current configuration of a service in one
// copy-safe snapshot.
//
//	state := i18n.CurrentState()
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

// HandlerTypeNames returns the short type names of the snapshot's handlers.
//
//	names := i18n.CurrentState().HandlerTypeNames()
//
// The returned slice is a fresh copy, so callers can inspect or mutate it
// without affecting the snapshot.
func (s ServiceState) HandlerTypeNames() []string {
	if len(s.Handlers) == 0 {
		return []string{}
	}
	names := make([]string, 0, len(s.Handlers))
	for _, handler := range s.Handlers {
		if handler == nil {
			names = append(names, "<nil>")
			continue
		}
		names = append(names, shortHandlerTypeName(handler))
	}
	return names
}

// String returns a concise, stable summary of the service snapshot.
//
//	fmt.Println(i18n.CurrentState().String())
func (s ServiceState) String() string {
	langs := "[]"
	if len(s.AvailableLanguages) > 0 {
		langs = "[" + core.Join(", ", s.AvailableLanguages...) + "]"
	}
	handlerNames := s.HandlerTypeNames()
	handlers := "[]"
	if len(handlerNames) > 0 {
		handlers = "[" + core.Join(", ", handlerNames...) + "]"
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

func shortHandlerTypeName(handler KeyHandler) string {
	name := core.Sprintf("%T", handler)
	parts := core.Split(name, ".")
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}
	return core.TrimPrefix(name, "*")
}

// State returns a snapshot of the service's current configuration —
// language, available languages, mode, formality, fallback, location,
// direction, debug flag, and handler chain — captured atomically while
// holding the read lock. Safe to call on nil receiver (returns the
// uninitialised-service default snapshot).
//
//	state := svc.State()
//	fmt.Println(state.Language, state.Mode)
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

	return newServiceStateSnapshot(
		s.currentLang,
		s.requestedLang,
		s.languageExplicit,
		langs,
		s.mode,
		s.fallbackLang,
		s.formality,
		s.location,
		dir,
		s.debug,
		handlers,
	)
}

// String returns a concise snapshot of the service state.
func (s *Service) String() string {
	return s.State().String()
}

// CurrentState is a more explicit alias for State.
//
//	state := i18n.CurrentState()
func (s *Service) CurrentState() ServiceState {
	return s.State()
}
