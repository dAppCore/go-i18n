package i18n

// ServiceState captures the current configuration of a service in one
// copy-safe snapshot.
type ServiceState struct {
	Language           string
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

func (s *Service) State() ServiceState {
	if s == nil {
		return ServiceState{
			Language:           "en",
			AvailableLanguages: []string{},
			Mode:               ModeNormal,
			Fallback:           "en",
			Formality:          FormalityNeutral,
			Direction:          DirLTR,
			IsRTL:              false,
			Debug:              false,
			Handlers:           []KeyHandler{},
		}
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
