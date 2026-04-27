package i18n

import "dappco.re/go/core"

// SetDebug enables or disables debug mode on the default service.
func SetDebug(enabled bool) {
	withDefaultService(func(svc *Service) { svc.SetDebug(enabled) })
}

// SetDebug toggles debug mode on the receiver service. When enabled, every
// translated value is wrapped with its key for easier QA inspection. No-op
// on nil receiver.
//
//	svc.SetDebug(true)
func (s *Service) SetDebug(enabled bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debug = enabled
}

// Debug reports whether debug mode is currently enabled on the service.
// Returns false on nil receiver.
//
//	if svc.Debug() { /* show key-wrapped strings */ }
func (s *Service) Debug() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debug
}

func debugFormat(key, text string) string {
	return core.Sprintf("[%s] %s", key, text)
}
