package i18n

import "testing"

var benchLanguageRuleSink PluralRule

func BenchmarkGetPluralRule(b *testing.B) {
	langs := []string{"en", "en-US", "fr", "ru", "pl", "ar", "cy", "unknown"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lang := range langs {
			benchLanguageRuleSink = GetPluralRule(lang)
		}
	}
}

func BenchmarkGetPluralCategory(b *testing.B) {
	cases := []struct {
		lang  string
		count int
	}{
		{lang: "en", count: 1},
		{lang: "fr", count: 0},
		{lang: "ru", count: 2},
		{lang: "pl", count: 12},
		{lang: "ar", count: 11},
		{lang: "cy", count: 6},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchServicePluralSink = GetPluralCategory(tc.lang, tc.count)
		}
	}
}

func BenchmarkPluralRuleEnglish(b *testing.B) {
	counts := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleEnglish(count)
		}
	}
}

func BenchmarkPluralRuleGerman(b *testing.B) {
	counts := []int{1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleGerman(count)
		}
	}
}

func BenchmarkPluralRuleSpanish(b *testing.B) {
	counts := []int{1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleSpanish(count)
		}
	}
}

func BenchmarkPluralRuleFrench(b *testing.B) {
	counts := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleFrench(count)
		}
	}
}

func BenchmarkPluralRuleRussian(b *testing.B) {
	counts := []int{1, 2, 5, 11, 22}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleRussian(count)
		}
	}
}

func BenchmarkPluralRulePolish(b *testing.B) {
	counts := []int{1, 2, 5, 12, 24}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRulePolish(count)
		}
	}
}

func BenchmarkPluralRuleArabic(b *testing.B) {
	counts := []int{0, 1, 2, 3, 11, 100}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleArabic(count)
		}
	}
}

func BenchmarkPluralRuleChinese(b *testing.B) {
	counts := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleChinese(count)
		}
	}
}

func BenchmarkPluralRuleJapanese(b *testing.B) {
	counts := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleJapanese(count)
		}
	}
}

func BenchmarkPluralRuleKorean(b *testing.B) {
	counts := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleKorean(count)
		}
	}
}

func BenchmarkPluralRuleWelsh(b *testing.B) {
	counts := []int{0, 1, 2, 3, 6, 7}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, count := range counts {
			benchServicePluralSink = pluralRuleWelsh(count)
		}
	}
}
