package i18n

import "testing"

var (
	benchTransformFloatSink float64
	benchTransformInt64Sink int64
)

func BenchmarkGetCount(b *testing.B) {
	values := []any{
		S("file", "config.yaml").Count(3),
		C("status").Count(4),
		C("status").Set("Count", "5"),
		map[string]any{"Count": "6"},
		map[string]int{"count": 7},
		map[string]string{"Count": "8"},
		9,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceIntSink = getCount(value)
		}
	}
}

func BenchmarkToInt(b *testing.B) {
	values := []any{int(1), int64(2), int32(3), int16(4), int8(5), uint(6), uint64(7), uint32(8), uint16(9), uint8(10), float64(11), float32(12), " 13 ", "bad"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceIntSink = toInt(value)
		}
	}
}

func BenchmarkToInt64(b *testing.B) {
	values := []any{int(1), int64(2), int32(3), int16(4), int8(5), uint(6), uint64(7), uint32(8), uint16(9), uint8(10), float64(11), float32(12), " 13 ", "bad"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchTransformInt64Sink = toInt64(value)
		}
	}
}

func BenchmarkToFloat64(b *testing.B) {
	values := []any{float64(1), float32(2), int(3), int64(4), int32(5), int16(6), int8(7), uint(8), uint64(9), uint32(10), uint16(11), uint8(12), " 13.5 ", "bad"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchTransformFloatSink = toFloat64(value)
		}
	}
}
