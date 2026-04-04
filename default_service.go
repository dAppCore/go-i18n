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

// defaultServiceNamespaceValue resolves a namespace key against the default
// service when available, or returns the namespace-qualified key otherwise.
func defaultServiceNamespaceValue(namespace, key string, lookup func(*Service, string) string) string {
	return defaultServiceValue(namespaceLookupKey(namespace, key), func(svc *Service) string {
		return lookup(svc, key)
	})
}
