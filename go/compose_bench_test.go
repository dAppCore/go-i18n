package i18n

import "testing"

var (
	benchComposeSubjectSink      *Subject
	benchComposeTemplateDataSink templateData
)

type benchComposeStringer string

func (s benchComposeStringer) String() string {
	return string(s)
}

func benchComposeIntent() Intent {
	return Intent{
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "delete",
			Dangerous: true,
			Default:   "no",
			Supports:  []string{"confirm", "cancel"},
		},
		Question: "Delete {{.Subject}}?",
		Confirm:  "Really delete {{.Subject}} from {{.Location}}?",
		Success:  "{{.Count}} {{.Noun}} deleted",
		Failure:  "Failed to delete {{.Subject}}",
	}
}

func BenchmarkS(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = S("file", "config.yaml")
	}
}

func BenchmarkComposeIntent(b *testing.B) {
	intent := benchComposeIntent()
	subject := S("file", benchComposeStringer("config.yaml")).Count(3).In("workspace")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceComposedSink = ComposeIntent(intent, subject)
	}
}

func BenchmarkCount(b *testing.B) {
	subject := S("file", "config.yaml")
	ctx := C("status")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = subject.Count(3)
		benchContextSink = ctx.Count(3)
	}
}

func BenchmarkGender(b *testing.B) {
	subject := S("client", "Alex")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = subject.Gender("f")
	}
}

func BenchmarkIn(b *testing.B) {
	subject := S("file", "config.yaml")
	ctx := C("status")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = subject.In("workspace")
		benchContextSink = ctx.In("dashboard")
	}
}

func BenchmarkFormal(b *testing.B) {
	subject := S("user", "Alex")
	ctx := C("greeting")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = subject.Formal()
		benchContextSink = ctx.Formal()
	}
}

func BenchmarkInformal(b *testing.B) {
	subject := S("user", "Alex")
	ctx := C("greeting")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeSubjectSink = subject.Informal()
		benchContextSink = ctx.Informal()
	}
}

func BenchmarkIsPlural(b *testing.B) {
	subject := S("file", "config.yaml").Count(3)
	ctx := C("status").Count(3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = subject.IsPlural()
		benchServiceBoolSink = ctx.IsPlural()
	}
}

func BenchmarkCountInt(b *testing.B) {
	subject := S("file", "config.yaml").Count(3)
	ctx := C("status").Count(3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceIntSink = subject.CountInt()
		benchServiceIntSink = ctx.CountInt()
	}
}

func BenchmarkCountString(b *testing.B) {
	subject := S("file", "config.yaml").Count(1234)
	ctx := C("status").Count(1234)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subject.CountString()
		benchServiceStringSink = ctx.CountString()
	}
}

func BenchmarkGenderString(b *testing.B) {
	subject := S("client", "Alex").Gender("f")
	ctx := C("greeting").WithGender("f")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subject.GenderString()
		benchServiceStringSink = ctx.GenderString()
	}
}

func BenchmarkLocationString(b *testing.B) {
	subject := S("file", "config.yaml").In("workspace")
	ctx := C("status").In("dashboard")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subject.LocationString()
		benchServiceStringSink = ctx.LocationString()
	}
}

func BenchmarkNounString(b *testing.B) {
	subject := S("file", "config.yaml")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subject.NounString()
	}
}

func BenchmarkFormalityString(b *testing.B) {
	subject := S("user", "Alex").Formal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = subject.FormalityString()
	}
}

func BenchmarkIsFormal(b *testing.B) {
	subject := S("user", "Alex").Formal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = subject.IsFormal()
	}
}

func BenchmarkIsInformal(b *testing.B) {
	subject := S("user", "Alex").Informal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = subject.IsInformal()
	}
}

func BenchmarkNewTemplateData(b *testing.B) {
	subject := S("file", benchComposeStringer("config.yaml")).Count(3).Gender("n").In("workspace").Formal()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchComposeTemplateDataSink = newTemplateData(subject)
	}
}
