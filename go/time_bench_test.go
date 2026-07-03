package i18n

import (
	"testing"
	"time"
)

func BenchmarkTimeAgo(b *testing.B) {
	benchSetDefaultLanguage(b, "en")
	now := time.Now()
	values := []time.Time{
		now.Add(time.Hour),
		now.Add(-10 * time.Second),
		now.Add(-5 * time.Minute),
		now.Add(-3 * time.Hour),
		now.Add(-2 * 24 * time.Hour),
		now.Add(-14 * 24 * time.Hour),
		now.Add(-90 * 24 * time.Hour),
		now.Add(-2 * 365 * 24 * time.Hour),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceStringSink = TimeAgo(value)
		}
	}
}

func BenchmarkFormatAgo(b *testing.B) {
	benchSetDefaultLanguage(b, "en")
	cases := []struct {
		count int
		unit  string
	}{
		{count: 1, unit: "second"},
		{count: 2, unit: "minutes"},
		{count: 3, unit: "hours"},
		{count: 4, unit: "day"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchServiceStringSink = FormatAgo(tc.count, tc.unit)
		}
	}
}

func BenchmarkFallbackAgoUnit(b *testing.B) {
	benchSetDefaultLanguage(b, "en")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = fallbackAgoUnit("second", 1)
		benchServiceStringSink = fallbackAgoUnit("minute", 2)
	}
}

func BenchmarkNormalizeAgoUnit(b *testing.B) {
	units := []string{" Seconds ", "minutes", "hours", "days", "weeks", "months", "years", "fortnight"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, unit := range units {
			benchServiceStringSink = normalizeAgoUnit(unit)
		}
	}
}
