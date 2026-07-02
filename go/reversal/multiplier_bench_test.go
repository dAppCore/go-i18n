package reversal

import "testing"

func benchMultiplierTokens(tb testing.TB) (*Multiplier, []Token, int, int) {
	tb.Helper()
	m := NewMultiplier()
	tokens := m.tokeniser.Tokenise("Delete the configuration branch.")
	verbIndex, nounIndex := -1, -1
	for i, tok := range tokens {
		switch tok.Type {
		case TokenVerb:
			if verbIndex < 0 {
				verbIndex = i
			}
		case TokenNoun:
			if nounIndex < 0 {
				nounIndex = i
			}
		}
	}
	if verbIndex < 0 || nounIndex < 0 {
		tb.Fatalf("benchmark fixture missing verb or noun token: %#v", tokens)
	}
	return m, tokens, verbIndex, nounIndex
}

func BenchmarkNewMultiplier(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMultiplier()
	}
}

func BenchmarkNewMultiplierForLang(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMultiplierForLang("fr")
	}
}

func BenchmarkExpand(b *testing.B) {
	benchSetup(b)
	m := NewMultiplier()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Expand("Delete the configuration branch.")
	}
}

func BenchmarkApplyVerbTransform(b *testing.B) {
	benchSetup(b)
	m, tokens, vi, _ := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.applyVerbTransform(tokens, vi, "past")
	}
}

func BenchmarkApplyNounTransform(b *testing.B) {
	benchSetup(b)
	m, tokens, _, ni := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.applyNounTransform(tokens, ni)
	}
}

func BenchmarkApplyNounTransformOnTokens(b *testing.B) {
	benchSetup(b)
	m, tokens, _, ni := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.applyNounTransformOnTokens(tokens, ni)
	}
}

func BenchmarkReconstructWithVerbTransform(b *testing.B) {
	benchSetup(b)
	m, tokens, vi, _ := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.reconstructWithVerbTransform(tokens, vi, "past")
	}
}

func BenchmarkReconstructWithNounTransform(b *testing.B) {
	benchSetup(b)
	m, tokens, _, ni := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.reconstructWithNounTransform(tokens, ni)
	}
}

func BenchmarkReconstructWithVerbAndNounTransform(b *testing.B) {
	benchSetup(b)
	m, tokens, vi, ni := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.reconstructWithVerbAndNounTransform(tokens, vi, "gerund", ni)
	}
}

func BenchmarkReconstructWithTransforms(b *testing.B) {
	benchSetup(b)
	m, tokens, vi, ni := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.reconstructWithTransforms(tokens, vi, "base", ni)
	}
}

func BenchmarkTransformedVerbRaw(b *testing.B) {
	benchSetup(b)
	_, tokens, vi, _ := benchMultiplierTokens(b)
	tok := tokens[vi]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transformedVerbRaw(tok, "past")
	}
}

func BenchmarkTransformedNounRaw(b *testing.B) {
	benchSetup(b)
	_, tokens, _, ni := benchMultiplierTokens(b)
	tok := tokens[ni]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = transformedNounRaw(tok)
	}
}

func BenchmarkReconstruct(b *testing.B) {
	benchSetup(b)
	_, tokens, _, _ := benchMultiplierTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reconstruct(tokens)
	}
}

func BenchmarkPreserveCase(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = preserveCase("Delete", "deleted")
	}
}

func BenchmarkIsASCIIOnly(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isASCIIOnly("Delete the configuration branch")
	}
}

func BenchmarkPreserveCaseASCII(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = preserveCaseASCII("DELETE", "deleted")
	}
}

func BenchmarkIsAllUpperASCII(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isAllUpperASCII("DELETE")
	}
}

func BenchmarkUpperASCII(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = upperASCII("delete")
	}
}

func BenchmarkIsAllUpper(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isAllUpper("DELETE")
	}
}
