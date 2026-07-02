package i18n

import "testing"

func BenchmarkNewServiceStateSnapshot(b *testing.B) {
	handlers := DefaultHandlers()
	langs := []string{"en", "fr"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCoreStateSink = newServiceStateSnapshot(
			"en",
			"en",
			true,
			langs,
			ModeCollect,
			"fr",
			FormalityFormal,
			"dashboard",
			DirLTR,
			true,
			handlers,
		)
	}
}

func BenchmarkDefaultServiceStateSnapshot(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchCoreStateSink = defaultServiceStateSnapshot()
	}
}

func BenchmarkHandlerTypeNames(b *testing.B) {
	state := newServiceStateSnapshot(
		"en",
		"en",
		true,
		[]string{"en"},
		ModeNormal,
		"en",
		FormalityNeutral,
		"",
		DirLTR,
		false,
		[]KeyHandler{nil, LabelHandler{}, NumericHandler{}},
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = state.HandlerTypeNames()
	}
}

func BenchmarkShortHandlerTypeName(b *testing.B) {
	handler := LabelHandler{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = shortHandlerTypeName(handler)
	}
}
