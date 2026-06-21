package reversal

import (
	"testing"

	"dappco.re/go"
)

// benchReferenceSamples is a small, realistic classified-text corpus spanning
// two domains. The reference/classify path always runs on tokenised text, so
// the figures below charge the real per-text cost (tokenise -> imprint ->
// centroid/variance/distance) rather than synthetic maps.
//
// "Push the api url to the ssh server" carries gram.word tokens (api, url,
// ssh), which populates DomainVocabulary — the centroid-accumulation path that
// distribution-only samples would otherwise leave empty.
var benchReferenceSamples = []ClassifiedText{
	{Text: "Delete the configuration file", Domain: "technical"},
	{Text: "Build the project from source", Domain: "technical"},
	{Text: "Push the api url to the ssh server", Domain: "technical"},
	{Text: "She wrote the story by candlelight", Domain: "creative"},
	{Text: "He drew a map of forgotten places", Domain: "creative"},
	{Text: "The river froze under the winter moon", Domain: "creative"},
}

// benchReferenceImprints projects the corpus into per-domain imprint slices and
// a flat slice, plus a built ReferenceSet and a probe imprint — the shared
// fixtures every reference benchmark needs. It is charged outside the timed
// loop via b.ResetTimer at each call site.
func benchReferenceImprints(tb testing.TB) (perDomain [][]GrammarImprint, flat []GrammarImprint, rs *ReferenceSet, probe GrammarImprint) {
	tb.Helper()
	tok := NewTokeniser()

	tech := []GrammarImprint{
		NewImprint(tok.Tokenise(benchReferenceSamples[0].Text)),
		NewImprint(tok.Tokenise(benchReferenceSamples[1].Text)),
		NewImprint(tok.Tokenise(benchReferenceSamples[2].Text)),
	}
	creative := []GrammarImprint{
		NewImprint(tok.Tokenise(benchReferenceSamples[3].Text)),
		NewImprint(tok.Tokenise(benchReferenceSamples[4].Text)),
		NewImprint(tok.Tokenise(benchReferenceSamples[5].Text)),
	}
	perDomain = [][]GrammarImprint{tech, creative}
	flat = append(append(flat, tech...), creative...)

	built, err := valueFromResult[*ReferenceSet](BuildReferences(tok, benchReferenceSamples))
	if err != nil {
		tb.Fatalf("BuildReferences: %v", err)
	}
	rs = built
	probe = NewImprint(tok.Tokenise("Run the tests before committing"))
	return perDomain, flat, rs, probe
}

// TestReferenceDistanceZeroAlloc is the hard guard (runs under `go test ./...`,
// not just `-bench`) for the query-side distance path. Compare/Classify call
// klDivergence and mahalanobis once per domain; both, and their per-component
// kernels mapKL/mapMahalanobis, must stay 0-alloc. Each builds its key union in
// a stack-resident map[string]bool (escape analysis keeps the header off the
// heap; the bucket array never reaches mallocgc for these small unions), and
// mapMahalanobis's prefix+k is a transient lookup key the compiler does not
// materialise — so any allocation here means that property regressed.
//
// addMap is included on its steady-state path: dst already contains src's keys,
// so the in-place sum grows no buckets. The guards assert populated inputs so a
// regression cannot hide behind a len==0 early return.
func TestReferenceDistanceZeroAlloc(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	a := NewImprint(tok.Tokenise("Deleted the configuration files successfully"))
	c := NewImprint(tok.Tokenise("Building the updated server packages now"))
	variance := computeVariance(
		[]GrammarImprint{a, c},
		computeCentroid([]GrammarImprint{a, c}),
	)

	// Non-vacuity: the guarded kernels must run their populated path, not the
	// empty-map early return, and the Mahalanobis guard must hit the populated
	// variance (the variance != nil hot path through Compare).
	if len(a.TenseDistribution) == 0 || len(c.TenseDistribution) == 0 {
		t.Fatal("benchmark text produced empty tense distributions; 0-alloc guards would be vacuous")
	}
	if len(variance) == 0 {
		t.Fatal("expected populated variance; Mahalanobis 0-alloc guard would not exercise the hot path")
	}

	if avg := testing.AllocsPerRun(100, func() { _ = klDivergence(a, c) }); avg != 0 {
		t.Fatalf("klDivergence allocates %.1f/run, want 0", avg)
	}
	if avg := testing.AllocsPerRun(100, func() { _ = mapKL(a.TenseDistribution, c.TenseDistribution) }); avg != 0 {
		t.Fatalf("mapKL allocates %.1f/run, want 0", avg)
	}
	if avg := testing.AllocsPerRun(100, func() { _ = mahalanobis(a, c, variance) }); avg != 0 {
		t.Fatalf("mahalanobis allocates %.1f/run, want 0", avg)
	}
	if avg := testing.AllocsPerRun(100, func() {
		_ = mapMahalanobis("tense:", a.TenseDistribution, c.TenseDistribution, variance)
	}); avg != 0 {
		t.Fatalf("mapMahalanobis allocates %.1f/run, want 0", avg)
	}

	// addMap steady state: dst pre-populated with src's keys -> no bucket growth.
	src := a.TenseDistribution
	dst := make(map[string]float64, len(src))
	addMap(dst, src)
	if avg := testing.AllocsPerRun(100, func() { addMap(dst, src) }); avg != 0 {
		t.Fatalf("addMap (steady state) allocates %.1f/run, want 0", avg)
	}
}

// BenchmarkFailResult charges the result-coercion helper on its most common
// branch: an already-failed Result passing straight through.
func BenchmarkFailResult(b *testing.B) {
	benchSetup(b)
	bad := failResult(core.NewError("boom"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = failResult(bad)
	}
}

// BenchmarkBuildReferences charges the full build: tokenise every sample,
// project imprints, then aggregate centroid + variance per domain. The
// Tokeniser is reused (the realistic caller pattern); only the per-call
// grouping/centroid/variance work is timed.
func BenchmarkBuildReferences(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildReferences(tok, benchReferenceSamples)
	}
}

// BenchmarkCompare charges the per-imprint scoring against every domain
// reference: cosine + symmetric KL + Mahalanobis, once per domain.
func BenchmarkCompare(b *testing.B) {
	benchSetup(b)
	_, _, rs, probe := benchReferenceImprints(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.Compare(probe)
	}
}

// BenchmarkClassify charges Compare plus the descending cosine ranking and
// best/second-best confidence margin.
func BenchmarkClassify(b *testing.B) {
	benchSetup(b)
	_, _, rs, probe := benchReferenceImprints(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.Classify(probe)
	}
}

// BenchmarkDomainNames charges the sorted domain-name extraction (maps.Keys +
// slices.Sorted), the cheap metadata surface.
func BenchmarkDomainNames(b *testing.B) {
	benchSetup(b)
	_, _, rs, _ := benchReferenceImprints(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.DomainNames()
	}
}

// BenchmarkComputeCentroid charges the averaging of a domain's imprints into a
// single normalised centroid (six map fields + four scalars).
func BenchmarkComputeCentroid(b *testing.B) {
	benchSetup(b)
	perDomain, _, _, _ := benchReferenceImprints(b)
	tech := perDomain[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computeCentroid(tech)
	}
}

// BenchmarkComputeVariance charges per-key sample-variance accumulation across
// a domain's imprints relative to its centroid.
func BenchmarkComputeVariance(b *testing.B) {
	benchSetup(b)
	perDomain, _, _, _ := benchReferenceImprints(b)
	tech := perDomain[0]
	centroid := computeCentroid(tech)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computeVariance(tech, centroid)
	}
}

// BenchmarkAccumVariance charges the inner squared-deviation accumulator for a
// single distribution component (the work computeVariance does 5×/imprint).
func BenchmarkAccumVariance(b *testing.B) {
	benchSetup(b)
	perDomain, _, _, _ := benchReferenceImprints(b)
	tech := perDomain[0]
	centroid := computeCentroid(tech)
	sample := tech[0]
	variance := make(map[string]float64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accumVariance(variance, "verb:", sample.VerbDistribution, centroid.VerbDistribution)
	}
}

// BenchmarkAddMap charges the in-place frequency accumulator. dst is
// pre-populated with src's keys so the steady state (no bucket growth) is
// measured, mirroring how the centroid loop reuses one destination map.
func BenchmarkAddMap(b *testing.B) {
	benchSetup(b)
	perDomain, _, _, _ := benchReferenceImprints(b)
	src := perDomain[0][0].VerbDistribution
	dst := make(map[string]float64, len(src))
	addMap(dst, src) // pre-populate keys so the timed loop only sums
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addMap(dst, src)
	}
}

// BenchmarkKLDivergence charges the weighted symmetric KL across all five
// distribution components of two imprints.
func BenchmarkKLDivergence(b *testing.B) {
	benchSetup(b)
	_, flat, _, _ := benchReferenceImprints(b)
	a, c := flat[0], flat[3]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = klDivergence(a, c)
	}
}

// BenchmarkMapKL isolates the symmetric-KL kernel klDivergence invokes once per
// distribution component (5×/call).
func BenchmarkMapKL(b *testing.B) {
	benchSetup(b)
	_, flat, _, _ := benchReferenceImprints(b)
	a, c := flat[0], flat[3]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapKL(a.TenseDistribution, c.TenseDistribution)
	}
}

// BenchmarkMahalanobis charges the variance-normalised distance across all five
// components, using a populated variance map (the hot Compare path, variance
// != nil).
func BenchmarkMahalanobis(b *testing.B) {
	benchSetup(b)
	perDomain, flat, _, _ := benchReferenceImprints(b)
	variance := computeVariance(perDomain[0], computeCentroid(perDomain[0]))
	a, c := flat[0], flat[3]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mahalanobis(a, c, variance)
	}
}

// BenchmarkMapMahalanobis isolates the variance-normalised squared-distance
// kernel for one component on the populated-variance path.
func BenchmarkMapMahalanobis(b *testing.B) {
	benchSetup(b)
	perDomain, flat, _, _ := benchReferenceImprints(b)
	variance := computeVariance(perDomain[0], computeCentroid(perDomain[0]))
	a, c := flat[0], flat[3]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapMahalanobis("tense:", a.TenseDistribution, c.TenseDistribution, variance)
	}
}
