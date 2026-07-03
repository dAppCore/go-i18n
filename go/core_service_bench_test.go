package i18n

import (
	"context"
	"testing"

	core "dappco.re/go"
)

var (
	benchCoreFactorySink     func(*core.Core) core.Result
	benchCoreMissingKeysSink []MissingKey
	benchCoreStateSink       ServiceState
)

func benchCoreServiceFixture() *CoreService {
	return &CoreService{
		svc:           benchServiceFixture(),
		missingKeys:   []MissingKey{{Key: "missing.key", Args: map[string]any{"Context": "dashboard"}}},
		hookInstalled: true,
	}
}

func BenchmarkWrapped(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceServiceSink = coreSvc.wrapped()
	}
}

func BenchmarkString(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	src := FSSource{FS: benchServiceFS(), Dir: "locales"}
	subject := S("file", benchComposeStringer("config.yaml"))
	stringerSubject := S("file", benchComposeStringer("from-stringer"))
	ctx := C("navigation")
	opts := ServiceOptions{
		Language:  "en",
		Fallback:  "fr",
		Formality: FormalityFormal,
		Location:  "dashboard",
		Mode:      ModeCollect,
		Debug:     true,
		ExtraFS:   []FSSource{src},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = src.String()
		benchServiceStringSink = opts.String()
		benchServiceStringSink = coreSvc.String()
		benchServiceStringSink = subject.String()
		benchServiceStringSink = stringerSubject.String()
		benchServiceStringSink = ctx.String()
	}
}

func BenchmarkNewCoreService(b *testing.B) {
	previous := defaultService.Load()
	b.Cleanup(func() {
		SetDefault(previous)
	})
	opts := ServiceOptions{
		Language: "en",
		Fallback: "fr",
		ExtraFS:  []FSSource{{FS: benchServiceFS(), Dir: "locales"}},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := NewCoreService(opts)
		benchCoreFactorySink = factory
		benchServiceResultSink = factory(nil)
	}
}

func BenchmarkOnStartup(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	coreSvc.svc.SetMode(ModeCollect)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = coreSvc.OnStartup(ctx)
	}
}

func BenchmarkOnShutdown(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = coreSvc.OnShutdown(ctx)
	}
}

func BenchmarkEnsureMissingKeyCollector(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coreSvc.ensureMissingKeyCollector()
	}
	benchServiceBoolSink = coreSvc.hookInstalled
}

func BenchmarkMissingKeys(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCoreMissingKeysSink = coreSvc.MissingKeys()
	}
}

func BenchmarkClearMissingKeys(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	missing := MissingKey{Key: "missing.key", Args: map[string]any{"Context": "dashboard"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coreSvc.missingKeys = append(coreSvc.missingKeys[:0], missing)
		coreSvc.ClearMissingKeys()
	}
	benchCoreMissingKeysSink = coreSvc.missingKeys
}

func BenchmarkSetDebug(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coreSvc.SetDebug(true)
	}
	benchServiceBoolSink = coreSvc.svc.Debug()
}

func BenchmarkDebug(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	coreSvc.svc.SetDebug(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = coreSvc.Debug()
	}
}

func BenchmarkState(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCoreStateSink = coreSvc.State()
	}
}

func BenchmarkCurrentState(b *testing.B) {
	coreSvc := benchCoreServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCoreStateSink = coreSvc.CurrentState()
	}
}
