package i18n

import "testing"

var benchNumberFormatSink NumberFormat

const benchMinInt64 = int64(-1 << 63)

func benchNumbersLocale(b *testing.B) {
	b.Helper()
	benchInstallGrammarData(b, "zz-numbers")
	benchSetDefaultLanguage(b, "zz-numbers")
}

func BenchmarkGetNumberFormat(b *testing.B) {
	benchNumbersLocale(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchNumberFormatSink = getNumberFormat()
	}
}

func BenchmarkGetLocaleNumberFormat(b *testing.B) {
	benchInstallGrammarData(b, "zz-numbers")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchNumberFormatSink, benchServiceBoolSink = getLocaleNumberFormat("zz-numbers")
		benchNumberFormatSink, benchServiceBoolSink = getLocaleNumberFormat("en")
	}
}

func BenchmarkFormatNumber(b *testing.B) {
	benchNumbersLocale(b)
	values := []int64{0, 12, 1234, -9876543, benchMinInt64}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = FormatNumber(value)
		}
	}
}

func BenchmarkFormatDecimal(b *testing.B) {
	benchNumbersLocale(b)
	values := []float64{1234, 1234.5, -9876.543}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = FormatDecimal(value)
		}
	}
}

func BenchmarkFormatDecimalN(b *testing.B) {
	benchNumbersLocale(b)
	cases := []struct {
		value    float64
		decimals int
	}{
		{value: 1234, decimals: 2},
		{value: 1234.5, decimals: 3},
		{value: -9876.543, decimals: 2},
		{value: 9.999, decimals: 2},
		{value: 42.1, decimals: 0},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchServiceStringSink = FormatDecimalN(tc.value, tc.decimals)
		}
	}
}

func BenchmarkFormatPercent(b *testing.B) {
	benchNumbersLocale(b)
	values := []float64{0.42, 0.125, 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = FormatPercent(value)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	benchNumbersLocale(b)
	values := []int64{512, 2048, 5 << 20, 7 << 30, 2 << 40}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = FormatBytes(value)
		}
	}
}

func BenchmarkFormatOrdinal(b *testing.B) {
	benchSetDefaultLanguage(b, "en")
	values := []int{1, 2, 3, 4, 11, 21, -1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = FormatOrdinal(value)
		}
	}
}

func BenchmarkFormatFrenchOrdinal(b *testing.B) {
	values := []int{1, 2, -1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = formatFrenchOrdinal(value)
		}
	}
}

func BenchmarkFormatEnglishOrdinal(b *testing.B) {
	values := []int{1, 2, 3, 4, 11, 21, -1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = formatEnglishOrdinal(value)
		}
	}
}

func BenchmarkFormatIntWithSep(b *testing.B) {
	values := []int64{12, 1234, -9876543, benchMinInt64}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = formatIntWithSep(value, ",")
			benchServiceStringSink = formatIntWithSep(value, "")
		}
	}
}

func BenchmarkIndexAny(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceIntSink = indexAny("en-US", "-_")
		benchServiceIntSink = indexAny("english", "-_")
	}
}

func BenchmarkTrimRight(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = trimRight("1200", "0")
		benchServiceStringSink = trimRight("1234", "0")
	}
}
