package reversal

import "testing"

// TestGrammarDataLookupZeroAlloc is the hard guard for the property the
// benchmarks below merely measure: the per-token grammar-data lookup must
// allocate nothing. MatchArticle is the pure surface — it resolves grammar
// data via t.grammarData() -> GetGrammarData -> normalizeLanguageTag and does
// no morphology, so any allocation here means the BCP-47 tag is being
// re-parsed per call (i.e. normalizedLangCache, commit 2f9be3c, was removed
// or bypassed). Unlike a benchmark, this assertion runs under `go test ./...`
// and fails the suite on regression.
//
// Asserted on MatchArticle only — MatchNoun's Tier-3 reverse morphology
// allocates candidate slices unrelated to the grammar/tag cache.
func TestGrammarDataLookupZeroAlloc(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	avg := testing.AllocsPerRun(100, func() {
		for _, w := range matchWords {
			tok.MatchArticle(w)
		}
	})
	if avg != 0 {
		t.Fatalf("MatchArticle allocates %.1f/run; per-Tokeniser grammar lookup regressed (language-tag cache removed?)", avg)
	}
}

// These benchmarks mimic the downstream scorer's per-token access pattern:
// a single Tokeniser/language is reused while MatchArticle/MatchNoun/MatchVerb
// (and full Tokenise scoring) are driven many times. The grammar data is
// stable for a Tokeniser's lifetime, so any per-call grammarData() /
// GetGrammarData(t.lang) resolution is redundant work charged to every token.
//
// Per [[ax-11-benchmarks]] — used to confirm (via -memprofile + pprof
// -alloc_objects -list grammarData) whether the per-token grammar lookup
// re-parses the language tag or otherwise allocs, and to guard byte-identity
// after the per-Tokeniser memoisation lands.

// matchWords spans articles, nouns (base + plural + morphology), and verbs
// (base + past + gerund + morphology) so every Match* tier is exercised.
var matchWords = []string{
	// articles
	"the", "a", "an",
	// base nouns
	"file", "server", "branch", "package", "commit",
	// plural nouns (inverse-map + morphology tiers)
	"files", "servers", "branches", "entries", "wolves",
	// base verbs
	"delete", "build", "run", "push", "update",
	// past-tense verbs (inverse-map + round-trip tiers)
	"deleted", "built", "pushed", "updated", "committed",
	// gerunds (inverse-map + round-trip tiers)
	"deleting", "building", "running", "pushing", "updating",
	// unknown / fall-through words
	"quux", "frobnicate", "wibble",
}

// BenchmarkMatchArticle drives MatchArticle for one Tokeniser, the most direct
// per-token caller of t.grammarData().
func BenchmarkMatchArticle(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range matchWords {
			tok.MatchArticle(w)
		}
	}
}

// BenchmarkMatchNoun drives MatchNoun for one Tokeniser. Tier 3 re-enters the
// forward engine (i18n.PluralForm), which resolves grammar data per call.
func BenchmarkMatchNoun(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range matchWords {
			tok.MatchNoun(w)
		}
	}
}

// BenchmarkMatchVerb drives MatchVerb for one Tokeniser. Tier 3 round-trips via
// i18n.PastTense / i18n.Gerund, which resolve grammar data per call.
func BenchmarkMatchVerb(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range matchWords {
			tok.MatchVerb(w)
		}
	}
}

// BenchmarkMatchAll mirrors the scorer most closely: every word probed against
// all three Match* surfaces for a single reused Tokeniser/language.
func BenchmarkMatchAll(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range matchWords {
			tok.MatchArticle(w)
			tok.MatchNoun(w)
			tok.MatchVerb(w)
		}
	}
}

// BenchmarkTokeniseSignals drives full Tokenise scoring with WithSignals on
// dual-class input, the only path that reaches corpusPrior() — which calls
// i18n.GetGrammarData(t.lang) directly, bypassing grammarData().
func BenchmarkTokeniseSignals(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser(WithSignals())
	// Dual-class words ("commit", "branch", "build", "push", "set", "run")
	// force the ambiguity-scoring path where corpusPrior is consulted.
	sentences := []string{
		"The commit broke the build on the branch",
		"Push the set of changes and run the build",
		"They branch the build and commit the run",
		"Set the branch and push the commit",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range sentences {
			tok.Tokenise(s)
		}
	}
}
