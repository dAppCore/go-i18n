package i18n

import "testing"

var benchContextSink *TranslationContext

func BenchmarkC(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchContextSink = C("navigation")
	}
}

func BenchmarkWithGender(b *testing.B) {
	ctx := C("greeting")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchContextSink = ctx.WithGender("f")
	}
}

func BenchmarkSet(b *testing.B) {
	ctx := C("notify")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchContextSink = ctx.Set("user_name", "Alice")
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := C("notify").Set("user_name", "Alice")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = ctx.Get("user_name")
	}
}

func BenchmarkContextString(b *testing.B) {
	ctx := C("navigation")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = ctx.ContextString()
	}
}

func BenchmarkFormalityValue(b *testing.B) {
	ctx := C("greeting").Formal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceFormalitySink = ctx.FormalityValue()
	}
}

func BenchmarkCountValue(b *testing.B) {
	ctx := C("status").Count(3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, ok := ctx.countValue()
		benchServiceIntSink = count
		benchServiceBoolSink = ok
	}
}
