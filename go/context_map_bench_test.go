package i18n

import "testing"

func BenchmarkMapValueString(b *testing.B) {
	anyValues := map[string]any{"Subject": " config.yaml ", "Count": 3}
	stringValues := map[string]string{"Subject": " config.yaml "}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink, benchServiceBoolSink = mapValueString(anyValues, "Subject")
		benchServiceStringSink, benchServiceBoolSink = mapValueString(anyValues, "Count")
		benchServiceStringSink, benchServiceBoolSink = mapValueString(stringValues, "Subject")
	}
}

func BenchmarkMapRawValueString(b *testing.B) {
	values := []any{" config.yaml ", 3, uint64(5), true, 1.25, []string{"fallback"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = mapRawValueString(value)
		}
	}
}

func BenchmarkContextMapValues(b *testing.B) {
	anyValues := map[string]any{"Context": "dashboard", "tier": "gold", "Extra": map[string]any{"role": "admin"}}
	stringValues := map[string]string{"Context": "dashboard", "tier": "gold"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = contextMapValues(anyValues)
		benchServiceAnySink = contextMapValues(stringValues)
		benchServiceAnySink = contextMapValues("invalid")
	}
}

func BenchmarkContextMapValuesAny(b *testing.B) {
	values := map[string]any{"Context": "dashboard", "tier": "gold", "Extra": map[string]string{"role": "admin"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = contextMapValuesAny(values)
	}
}

func BenchmarkContextMapValuesString(b *testing.B) {
	values := map[string]string{"Context": "dashboard", "tier": "gold"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = contextMapValuesString(values)
	}
}
