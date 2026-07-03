package i18n

import "testing"

func BenchmarkDebugFormat(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = debugFormat("app.hello", "Hello")
	}
}
