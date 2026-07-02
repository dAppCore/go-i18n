package i18n

import (
	"testing"
	"text/template"
)

func BenchmarkN(b *testing.B) {
	benchSetDefaultService(b, benchServiceFixture())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = N("percent", 0.42)
	}
}

func BenchmarkNormalizeLookupKey(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = normalizeLookupKey(" Prompt.Delete ")
	}
}

func BenchmarkNamespaceLookupKey(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = namespaceLookupKey("prompt", "Delete")
	}
}

func BenchmarkExecuteIntentTemplate(b *testing.B) {
	data := templateData{
		Subject: "config.yaml",
		Noun:    "file",
		Count:   1,
	}
	tmpl := "Delete {{.Subject}}?"
	_ = executeIntentTemplate(tmpl, data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = executeIntentTemplate(tmpl, data)
	}
}

func BenchmarkApplyTemplate(b *testing.B) {
	data := map[string]any{"Name": "Alex"}
	tmpl := "Hello {{.Name}}"
	_ = applyTemplate(tmpl, data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = applyTemplate(tmpl, data)
	}
}

func BenchmarkExecuteParsedTemplate(b *testing.B) {
	tmpl := template.Must(template.New("").Parse("Hello {{.Name}}"))
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = executeParsedTemplate(tmpl, data, "fallback")
	}
}

func BenchmarkTemplateDataForRendering(b *testing.B) {
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = templateDataForRendering(ctx)
	}
}
