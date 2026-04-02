package i18n

// SetDebug enables or disables debug mode on the default service.
func SetDebug(enabled bool) {
	withDefaultService(func(svc *Service) { svc.SetDebug(enabled) })
}

func (s *Service) SetDebug(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debug = enabled
}

func (s *Service) Debug() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debug
}

func debugFormat(key, text string) string {
	return "[" + key + "] " + text
}
