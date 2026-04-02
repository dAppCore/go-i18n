package i18n

// withDefaultService runs fn when the default service is available.
func withDefaultService(fn func(*Service)) {
	if svc := Default(); svc != nil {
		fn(svc)
	}
}

// defaultServiceValue returns the value produced by fn when the default
// service exists, or fallback otherwise.
func defaultServiceValue[T any](fallback T, fn func(*Service) T) T {
	if svc := Default(); svc != nil {
		return fn(svc)
	}
	return fallback
}
