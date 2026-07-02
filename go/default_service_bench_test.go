package i18n

import "testing"

func BenchmarkWithDefaultService(b *testing.B) {
	svc := benchServiceFixture()
	benchSetDefaultService(b, svc)
	fn := func(svc *Service) {
		benchServiceServiceSink = svc
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		withDefaultService(fn)
	}
}

func BenchmarkDefaultServiceValue(b *testing.B) {
	svc := benchServiceFixture()
	benchSetDefaultService(b, svc)
	fn := func(svc *Service) string {
		return svc.Language()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = defaultServiceValue("fallback", fn)
	}
}

func BenchmarkDefaultServiceNamespaceValue(b *testing.B) {
	svc := benchServiceFixture()
	benchSetDefaultService(b, svc)
	lookup := func(svc *Service, key string) string {
		return svc.T(key)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = defaultServiceNamespaceValue("app", "hello", lookup)
	}
}
