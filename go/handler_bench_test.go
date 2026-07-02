package i18n

import "testing"

type benchHandlerCase struct {
	handler KeyHandler
	key     string
	args    []any
}

func benchHandlerFallback() string {
	return "fallback"
}

func benchHandlerCases() []benchHandlerCase {
	return []benchHandlerCase{
		{handler: LabelHandler{}, key: "i18n.label.status"},
		{handler: ProgressHandler{}, key: "i18n.progress.build", args: []any{"file"}},
		{handler: CountHandler{}, key: "i18n.count.file", args: []any{3}},
		{handler: DoneHandler{}, key: "i18n.done.delete", args: []any{"file"}},
		{handler: FailHandler{}, key: "i18n.fail.delete", args: []any{S("file", "config.yaml")}},
		{handler: NumericHandler{}, key: "i18n.numeric.number", args: []any{1234}},
		{handler: NumericHandler{}, key: "i18n.numeric.decimal", args: []any{1234.5}},
		{handler: NumericHandler{}, key: "i18n.numeric.percent", args: []any{0.42}},
		{handler: NumericHandler{}, key: "i18n.numeric.size", args: []any{2048}},
		{handler: NumericHandler{}, key: "i18n.numeric.ordinal", args: []any{21}},
		{handler: NumericHandler{}, key: "i18n.numeric.ago", args: []any{2, "hour"}},
	}
}

func BenchmarkMatch(b *testing.B) {
	cases := benchHandlerCases()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchServiceBoolSink = tc.handler.Match(tc.key)
		}
	}
}

func BenchmarkHandle(b *testing.B) {
	benchInstallGrammarData(b, "zz-handler")
	benchSetDefaultLanguage(b, "zz-handler")
	cases := benchHandlerCases()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			benchServiceStringSink = tc.handler.Handle(tc.key, tc.args, benchHandlerFallback)
		}
	}
}

func BenchmarkDefaultHandlers(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceHandlersSink = DefaultHandlers()
	}
}

func BenchmarkCountWordForm(b *testing.B) {
	benchInstallGrammarData(b, "zz-handler")
	benchSetDefaultLanguage(b, "zz-handler")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = countWordForm("zz-handler", "file", 3)
		benchServiceStringSink = countWordForm("zz-handler", "status", 2)
		benchServiceStringSink = countWordForm("zz-handler", "API", 2)
	}
}

func BenchmarkHasGrammarCountForms(b *testing.B) {
	benchInstallGrammarData(b, "zz-handler")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasGrammarCountForms("zz-handler", "file")
		benchServiceBoolSink = hasGrammarCountForms("zz-handler", "status")
	}
}

func BenchmarkIsPluralisableWordDisplay(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isPluralisableWordDisplay("file")
		benchServiceBoolSink = isPluralisableWordDisplay("file name")
	}
}

func BenchmarkIsUpperAcronymPlural(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isUpperAcronymPlural("APIs")
		benchServiceBoolSink = isUpperAcronymPlural("files")
	}
}

func BenchmarkIsAllUpper(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isAllUpper("API")
		benchServiceBoolSink = isAllUpper("Api")
	}
}

func BenchmarkSubjectArgText(b *testing.B) {
	subject := S("file", benchComposeStringer("config.yaml"))
	ctx := C("").Set("Subject", "archive.zip")
	values := map[string]any{"Context": "dashboard"}
	stringValues := map[string]string{"Text": "done"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subjectArgText("literal")
		benchServiceStringSink = subjectArgText(subject)
		benchServiceStringSink = subjectArgText(ctx)
		benchServiceStringSink = subjectArgText(values)
		benchServiceStringSink = subjectArgText(stringValues)
		benchServiceStringSink = subjectArgText(benchComposeStringer("stringer"))
	}
}

func BenchmarkContextArgText(b *testing.B) {
	values := map[string]any{"Context": "dashboard"}
	stringValues := map[string]string{"Text": "done"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = contextArgText(values)
		benchServiceStringSink = contextArgText(stringValues)
	}
}

func BenchmarkRunHandlerChain(b *testing.B) {
	benchInstallGrammarData(b, "zz-handler")
	benchSetDefaultLanguage(b, "zz-handler")
	handlers := DefaultHandlers()
	args := []any{S("file", "config.yaml")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = RunHandlerChain(handlers, "i18n.fail.delete", args, benchHandlerFallback)
	}
}

func BenchmarkFilterNilHandlers(b *testing.B) {
	handlers := []KeyHandler{nil, LabelHandler{}, nil, NumericHandler{}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceHandlersSink = filterNilHandlers(handlers)
	}
}
