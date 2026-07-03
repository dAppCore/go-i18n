package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

func benchLocaliseSupportedTags() []language.Tag {
	return []language.Tag{
		language.English,
		language.French,
		language.Arabic,
	}
}

func BenchmarkIsRTLLanguage(b *testing.B) {
	langs := []string{"ar", "ar-EG", "fa", "en-US"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lang := range langs {
			benchServiceBoolSink = IsRTLLanguage(lang)
		}
	}
}

func BenchmarkDetectLanguage(b *testing.B) {
	b.Setenv("LC_ALL", "")
	b.Setenv("LANGUAGE", "fr_FR:en_US")
	b.Setenv("LC_MESSAGES", "")
	b.Setenv("LANG", "")
	supported := benchLocaliseSupportedTags()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = detectLanguage(supported)
	}
}

func BenchmarkDetectLanguageFromEnv(b *testing.B) {
	supported := benchLocaliseSupportedTags()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = detectLanguageFromEnv("fr_FR.UTF-8", supported)
		benchServiceStringSink = detectLanguageFromEnv("ar_EG.UTF-8", supported)
	}
}

func BenchmarkFirstLocaleFromList(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = firstLocaleFromList(" : fr_FR : en_US")
	}
}

func BenchmarkNormalizeLanguageTag(b *testing.B) {
	langs := []string{" en_US ", "fr-FR", "ar_EG", "bad tag"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lang := range langs {
			benchServiceStringSink = normalizeLanguageTag(lang)
		}
	}
}
