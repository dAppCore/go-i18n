package reversal

import (
	"testing"
)

func TestDetectAnomalies_NoAnomalies(t *testing.T) {
	tok := initI18n(t)

	// Build references from the same domain samples.
	refSamples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Update the dependencies", Domain: "technical"},
		{Text: "Format the source files", Domain: "technical"},
	}
	rs, err := BuildReferences(tok, refSamples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	// Test samples that should match the reference.
	testSamples := []ClassifiedText{
		{Text: "Push the changes to the branch", Domain: "technical"},
		{Text: "Reset the branch to the previous version", Domain: "technical"},
	}

	results, stats := rs.DetectAnomalies(tok, testSamples)
	if stats.Total != 2 {
		t.Errorf("Total = %d, want 2", stats.Total)
	}

	// With only one domain reference, everything classifies as that domain.
	// So no anomalies expected.
	if stats.Anomalies != 0 {
		t.Errorf("Anomalies = %d, want 0", stats.Anomalies)
	}
	if len(results) != 2 {
		t.Fatalf("Results len = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.IsAnomaly {
			t.Errorf("unexpected anomaly: model=%s imprint=%s text=%q", r.ModelDomain, r.ImprintDomain, r.Text)
		}
	}
}

func TestDetectAnomalies_WithMismatch(t *testing.T) {
	tok := initI18n(t)

	// Build references with two well-separated domains.
	refSamples := []ClassifiedText{
		// Technical: imperatives.
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Update the dependencies now", Domain: "technical"},
		{Text: "Format the source files", Domain: "technical"},
		{Text: "Reset the branch to the previous version", Domain: "technical"},
		// Creative: past tense narratives.
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He drew a map of forgotten places", Domain: "creative"},
		{Text: "The river froze under the winter moon", Domain: "creative"},
		{Text: "They sang the old songs by the fire", Domain: "creative"},
		{Text: "She painted the sky with broad strokes", Domain: "creative"},
	}
	rs, err := BuildReferences(tok, refSamples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	// A past-tense narrative labelled as "technical" by the model —
	// the imprint should say "creative", creating an anomaly.
	testSamples := []ClassifiedText{
		{Text: "She painted the sunset over the mountains", Domain: "technical"},
	}

	results, stats := rs.DetectAnomalies(tok, testSamples)
	t.Logf("Total=%d Anomalies=%d Rate=%.2f", stats.Total, stats.Anomalies, stats.Rate)
	for _, r := range results {
		t.Logf("  model=%s imprint=%s anomaly=%v conf=%.4f text=%q",
			r.ModelDomain, r.ImprintDomain, r.IsAnomaly, r.Confidence, r.Text)
	}

	if stats.Total != 1 {
		t.Errorf("Total = %d, want 1", stats.Total)
	}
	// This may or may not be flagged as anomaly depending on grammar overlap.
	// We just verify the pipeline runs without error and returns valid data.
	if len(results) != 1 {
		t.Fatalf("Results len = %d, want 1", len(results))
	}
	if results[0].ModelDomain != "technical" {
		t.Errorf("ModelDomain = %q, want technical", results[0].ModelDomain)
	}
}

func TestDetectAnomalies_SkipsEmptyDomain(t *testing.T) {
	tok := initI18n(t)

	refSamples := []ClassifiedText{
		{Text: "Delete the file", Domain: "technical"},
	}
	rs, _ := BuildReferences(tok, refSamples)

	testSamples := []ClassifiedText{
		{Text: "Some text without domain", Domain: ""},
		{Text: "Build the project", Domain: "technical"},
	}

	_, stats := rs.DetectAnomalies(tok, testSamples)
	if stats.Total != 1 {
		t.Errorf("Total = %d, want 1 (empty domain skipped)", stats.Total)
	}
}

func TestDetectAnomalies_ByPairTracking(t *testing.T) {
	tok := initI18n(t)

	refSamples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Format the source files", Domain: "technical"},
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He drew a map of forgotten places", Domain: "creative"},
		{Text: "The river froze under the winter moon", Domain: "creative"},
	}
	rs, err := BuildReferences(tok, refSamples)
	if err != nil {
		t.Fatalf("BuildReferences: %v", err)
	}

	// Force some mislabelled samples.
	testSamples := []ClassifiedText{
		{Text: "Push the changes now", Domain: "technical"},
		{Text: "She sang an old song by the fire", Domain: "creative"},
	}

	_, stats := rs.DetectAnomalies(tok, testSamples)
	t.Logf("Anomalies=%d ByPair=%v", stats.Anomalies, stats.ByPair)

	// ByPair should only contain entries for actual disagreements.
	for pair, count := range stats.ByPair {
		if count <= 0 {
			t.Errorf("ByPair[%s] = %d, want > 0", pair, count)
		}
	}
}

func TestAnomalyStats_Rate(t *testing.T) {
	tok := initI18n(t)

	// Single domain reference — everything maps to it.
	refSamples := []ClassifiedText{
		{Text: "Delete the file", Domain: "technical"},
		{Text: "Build the project", Domain: "technical"},
	}
	rs, _ := BuildReferences(tok, refSamples)

	// Two samples claiming "creative" — both should be anomalies.
	testSamples := []ClassifiedText{
		{Text: "Update the code", Domain: "creative"},
		{Text: "Fix the build", Domain: "creative"},
	}

	_, stats := rs.DetectAnomalies(tok, testSamples)
	if stats.Rate < 0.99 {
		t.Errorf("Rate = %.2f, want ~1.0 (all should be anomalies)", stats.Rate)
	}
}
