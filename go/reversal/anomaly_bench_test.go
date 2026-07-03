package reversal

import "testing"

func benchAnomalyReferenceSet(b *testing.B) (*Tokeniser, *ReferenceSet, []ClassifiedText) {
	b.Helper()
	benchSetup(b)
	tok := NewTokeniser()
	refSamples := []ClassifiedText{
		{Text: "Delete the configuration file", Domain: "technical"},
		{Text: "Build the project from source", Domain: "technical"},
		{Text: "Update the dependencies now", Domain: "technical"},
		{Text: "Format the source files", Domain: "technical"},
		{Text: "She wrote the story by candlelight", Domain: "creative"},
		{Text: "He drew a map of forgotten places", Domain: "creative"},
		{Text: "The river froze under the winter moon", Domain: "creative"},
		{Text: "They sang the old songs by the fire", Domain: "creative"},
	}
	rs, err := valueFromResult[*ReferenceSet](BuildReferences(tok, refSamples))
	if err != nil {
		b.Fatalf("BuildReferences: %v", err)
	}
	samples := []ClassifiedText{
		{Text: "Push the changes to the branch", Domain: "technical"},
		{Text: "She painted the sunset over the mountains", Domain: "technical"},
		{Text: "The old story ended under the winter moon", Domain: "creative"},
		{Text: "Build and format the source package", Domain: "creative"},
		{Text: "Some text without a model domain"},
	}
	return tok, rs, samples
}

func BenchmarkDetectAnomalies(b *testing.B) {
	tok, rs, samples := benchAnomalyReferenceSet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rs.DetectAnomalies(tok, samples)
	}
}
