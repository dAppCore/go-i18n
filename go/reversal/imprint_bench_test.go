package reversal

import "testing"

// benchImprintTokens builds two realistic classified-token streams for the
// GrammarImprint hot path (NewImprint runs on *any* tokenised text, not just
// translation labels), so the benchmarks below charge the real per-text cost.
func benchImprintTokens(tb testing.TB) (a, b []Token) {
	tb.Helper()
	tok := NewTokeniser()
	a = tok.Tokenise("Deleted the configuration files successfully")
	b = tok.Tokenise("Building the updated server packages now")
	return a, b
}

// BenchmarkNewImprint measures the per-text feature-vector projection. The six
// result maps are inherent to GrammarImprint; the figure to watch is anything
// above that floor (temporary sets, regrows).
func BenchmarkNewImprint(b *testing.B) {
	benchSetup(b)
	toks, _ := benchImprintTokens(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewImprint(toks)
	}
}

// BenchmarkSimilar measures weighted cosine similarity between two imprints —
// the scoring hot loop when matching a text against reference distributions.
func BenchmarkSimilar(b *testing.B) {
	benchSetup(b)
	ta, tb := benchImprintTokens(b)
	ia, ib := NewImprint(ta), NewImprint(tb)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ia.Similar(ib)
	}
}

// BenchmarkMapSimilarity isolates the cosine kernel that Similar invokes once
// per distribution component (up to 5×/call).
func BenchmarkMapSimilarity(b *testing.B) {
	benchSetup(b)
	ta, tb := benchImprintTokens(b)
	ia, ib := NewImprint(ta), NewImprint(tb)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapSimilarity(ia.NounDistribution, ib.NounDistribution)
	}
}

// BenchmarkNormaliseMap guards the in-place normaliser (must stay 0-alloc).
// normaliseMap is idempotent once a map sums to 1.0, so looping the same map
// is safe and measures steady state.
func BenchmarkNormaliseMap(b *testing.B) {
	m := map[string]float64{"past": 3, "gerund": 2, "base": 1, "participle": 4}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normaliseMap(m)
	}
}

// TestImprintZeroAlloc is the hard guard (runs under `go test ./...`, not just
// `-bench`) for the GrammarImprint comparison path. NewImprint allocates the
// six inherent result maps by design — they are fields of the returned struct.
// Similar and its mapSimilarity kernel, by contrast, must stay 0-alloc: the
// key union is built on the stack (escape analysis keeps it off the heap), so
// any allocation here means a change pushed that union onto the heap.
func TestImprintZeroAlloc(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	ia := NewImprint(tok.Tokenise("Deleted the configuration files successfully"))
	ib := NewImprint(tok.Tokenise("Building the updated server packages now"))

	// The guard must exercise the populated cosine path, not the empty
	// early-return — otherwise a regression to the heap would slip through.
	if len(ia.NounDistribution) == 0 || len(ib.NounDistribution) == 0 {
		t.Fatal("benchmark text produced empty noun distributions; 0-alloc guard would be vacuous")
	}

	if avg := testing.AllocsPerRun(100, func() { _ = ia.Similar(ib) }); avg != 0 {
		t.Fatalf("GrammarImprint.Similar allocates %.1f/run, want 0", avg)
	}
	if avg := testing.AllocsPerRun(100, func() { _ = mapSimilarity(ia.NounDistribution, ib.NounDistribution) }); avg != 0 {
		t.Fatalf("mapSimilarity allocates %.1f/run, want 0", avg)
	}
}
