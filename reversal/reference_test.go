package reversal

import (
	"math"
	"testing"

	i18n "dappco.re/go/core/i18n"
)

func initI18n(t *testing.T) *Tokeniser {
	t.Helper()
	svc, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n.New(): %v", err)
	}
	i18n.SetDefault(svc)
	return NewTokeniser()
}

func TestBuildReferences_Basic(t *testing.T) {
	tok := initI18n(t)

	samples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He drew a map of forgotten places", Domain: "creative"},
	}

	rs, err := BuildReferences(tok, samples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	if len(rs.Domains) != 2 {
		t.Errorf("Domains = %d, want 2", len(rs.Domains))
	}
	if rs.Domains["technical"] == nil {
		t.Error("missing technical domain")
	}
	if rs.Domains["creative"] == nil {
		t.Error("missing creative domain")
	}
	if rs.Domains["technical"].SampleCount != 2 {
		t.Errorf("technical SampleCount = %d, want 2", rs.Domains["technical"].SampleCount)
	}
}

func TestBuildReferences_Empty(t *testing.T) {
	tok := initI18n(t)
	_, err := BuildReferences(tok, nil)
	if err == nil {
		t.Error("expected error for empty samples")
	}
}

func TestBuildReferences_NoDomainLabels(t *testing.T) {
	tok := initI18n(t)
	samples := []ClassifiedText{
		{Text: "Hello world", Domain: ""},
	}
	_, err := BuildReferences(tok, samples)
	if err == nil {
		t.Error("expected error for no domain labels")
	}
}

func TestReferenceSet_Compare(t *testing.T) {
	tok := initI18n(t)

	samples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Run the tests before committing", Domain: "technical"},
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He painted the sky with broad strokes", Domain: "creative"},
		{Text: "They sang the old songs by the fire", Domain: "creative"},
	}

	rs, err := BuildReferences(tok, samples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	// Compare a technical sentence — should be closer to technical centroid.
	tokens := tok.Tokenise("Push the changes to the branch")
	imp := NewImprint(tokens)
	distances := rs.Compare(imp)

	techSim := distances["technical"].CosineSimilarity
	creativeSim := distances["creative"].CosineSimilarity

	t.Logf("Technical sentence: tech_sim=%.4f creative_sim=%.4f", techSim, creativeSim)
	// We don't hard-assert ordering because grammar similarity is coarse,
	// but both should be valid numbers.
	if math.IsNaN(techSim) || math.IsNaN(creativeSim) {
		t.Error("NaN in similarity scores")
	}

	// KL divergence should be non-negative.
	if distances["technical"].KLDivergence < 0 {
		t.Errorf("KLDivergence = %f, want >= 0", distances["technical"].KLDivergence)
	}
	if distances["technical"].Mahalanobis < 0 {
		t.Errorf("Mahalanobis = %f, want >= 0", distances["technical"].Mahalanobis)
	}
}

func TestReferenceSet_Classify(t *testing.T) {
	tok := initI18n(t)

	// Build references with clear domain separation.
	samples := []ClassifiedText{
		// Technical: imperative, base-form verbs.
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Update the dependencies", Domain: "technical"},
		{Text: "Format the source files", Domain: "technical"},
		{Text: "Reset the branch to the previous version", Domain: "technical"},
		// Creative: past tense, literary nouns.
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He drew a map of forgotten places", Domain: "creative"},
		{Text: "The river froze under the winter moon", Domain: "creative"},
		{Text: "They sang the old songs by the fire", Domain: "creative"},
		{Text: "She painted the sky with broad strokes", Domain: "creative"},
	}

	rs, err := BuildReferences(tok, samples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	// Classify returns a result with domain and confidence.
	tokens := tok.Tokenise("Stop the running process")
	imp := NewImprint(tokens)
	cls := rs.Classify(imp)

	t.Logf("Classified as %q with confidence %.4f", cls.Domain, cls.Confidence)
	if cls.Domain == "" {
		t.Error("empty classification domain")
	}
	if len(cls.Distances) != 2 {
		t.Errorf("Distances map has %d entries, want 2", len(cls.Distances))
	}
}

func TestReferenceSet_Classify_SingleDomainConfidence(t *testing.T) {
	tok := initI18n(t)

	samples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
	}

	rs, err := BuildReferences(tok, samples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	imp := NewImprint(tok.Tokenise("Run the tests before committing"))
	cls := rs.Classify(imp)

	if cls.Domain == "" {
		t.Fatal("empty classification domain")
	}
	if cls.Confidence != 0 {
		t.Errorf("Confidence = %f, want 0 when only one domain is available", cls.Confidence)
	}
}

func TestReferenceSet_DomainNames(t *testing.T) {
	tok := initI18n(t)
	samples := []ClassifiedText{
		{Text: "Delete the file", Domain: "technical"},
		{Text: "She wrote a poem", Domain: "creative"},
		{Text: "We should be fair", Domain: "ethical"},
	}
	rs, _ := BuildReferences(tok, samples)
	names := rs.DomainNames()
	want := []string{"creative", "ethical", "technical"}
	if len(names) != len(want) {
		t.Fatalf("DomainNames = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("DomainNames[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

func TestKLDivergence_Identical(t *testing.T) {
	a := GrammarImprint{
		TenseDistribution: map[string]float64{"base": 0.5, "past": 0.3, "gerund": 0.2},
	}
	kl := klDivergence(a, a)
	if kl > 0.001 {
		t.Errorf("KL divergence of identical distributions = %f, want ~0", kl)
	}
}

func TestKLDivergence_Different(t *testing.T) {
	a := GrammarImprint{
		TenseDistribution: map[string]float64{"base": 0.9, "past": 0.05, "gerund": 0.05},
	}
	b := GrammarImprint{
		TenseDistribution: map[string]float64{"base": 0.1, "past": 0.8, "gerund": 0.1},
	}
	kl := klDivergence(a, b)
	if kl < 0.01 {
		t.Errorf("KL divergence of different distributions = %f, want > 0.01", kl)
	}
}

func TestMapKL_Empty(t *testing.T) {
	kl := mapKL(nil, nil)
	if kl != 0 {
		t.Errorf("KL of two empty maps = %f, want 0", kl)
	}
}

func TestMahalanobis_NoVariance(t *testing.T) {
	// Without variance data, should fall back to Euclidean-like distance.
	a := GrammarImprint{
		TenseDistribution: map[string]float64{"base": 0.8, "past": 0.2},
	}
	b := GrammarImprint{
		TenseDistribution: map[string]float64{"base": 0.2, "past": 0.8},
	}
	dist := mahalanobis(a, b, nil)
	if dist <= 0 {
		t.Errorf("Mahalanobis without variance = %f, want > 0", dist)
	}
}

func TestComputeCentroid_SingleSample(t *testing.T) {
	tok := initI18n(t)
	tokens := tok.Tokenise("Delete the file")
	imp := NewImprint(tokens)

	centroid := computeCentroid([]GrammarImprint{imp})
	// Single sample centroid should be very similar to the original.
	sim := imp.Similar(centroid)
	if sim < 0.99 {
		t.Errorf("Single-sample centroid similarity = %f, want ~1.0", sim)
	}
}

func TestComputeVariance_SingleSample(t *testing.T) {
	tok := initI18n(t)
	tokens := tok.Tokenise("Delete the file")
	imp := NewImprint(tokens)
	centroid := computeCentroid([]GrammarImprint{imp})

	// Single sample: variance should be nil (n < 2).
	v := computeVariance([]GrammarImprint{imp}, centroid)
	if v != nil {
		t.Errorf("Single-sample variance should be nil, got %v", v)
	}
}
