package integration

import (
	"bytes"
	"context"
	"testing"
	"time"

	"dappco.re/go/core"
	i18n "dappco.re/go/i18n"
	"dappco.re/go/inference"
	_ "dappco.re/go/mlx" // registers Metal backend
)

func TestClassifyCorpus_Integration(t *testing.T) {
	model, err := inference.LoadModel("/Volumes/Data/lem/LEM-Gemma3-1B-layered-v2")
	if err != nil {
		t.Skipf("model not available: %v", err)
	}
	defer model.Close()

	// Build 50 technical prompts for throughput measurement
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, core.Sprintf(`{"id":%d,"prompt":"Delete the configuration file and rebuild the project"}`, i))
	}
	input := core.NewReader(core.Join("\n", lines...) + "\n")

	var output bytes.Buffer
	start := time.Now()
	stats, err := i18n.ClassifyCorpus(context.Background(), model, input, &output, i18n.WithBatchSize(8))
	if err != nil {
		t.Fatalf("ClassifyCorpus: %v", err)
	}
	elapsed := time.Since(start)

	t.Logf("Classified %d prompts in %v (%.1f prompts/sec)", stats.Total, elapsed, stats.PromptsPerSec)
	t.Logf("By domain: %v", stats.ByDomain)

	if stats.Total != 50 {
		t.Errorf("Total = %d, want 50", stats.Total)
	}
	if stats.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", stats.Skipped)
	}

	// Fix 5: Assert minimum accuracy — at least 80% of the 50 technical prompts
	// should be classified as "technical". The FINDINGS data shows 100% accuracy
	// on controlled technical input, so 80% is a conservative floor.
	technicalCount := stats.ByDomain["technical"]
	minRequired := 40 // 80% of 50
	if technicalCount < minRequired {
		// Log full domain breakdown for debugging
		for domain, count := range stats.ByDomain {
			t.Logf("  domain %q: %d (%.0f%%)", domain, count, float64(count)/float64(stats.Total)*100)
		}

		// Also inspect the output JSONL for misclassified entries
		outLines := core.Split(core.Trim(output.String()), "\n")
		for _, line := range outLines {
			var record map[string]any
			if r := core.JSONUnmarshal([]byte(line), &record); r.OK {
				if record["domain_1b"] != "technical" {
					t.Logf("  misclassified: id=%v domain_1b=%v", record["id"], record["domain_1b"])
				}
			}
		}

		t.Errorf("accuracy too low: %d/%d (%.0f%%) classified as technical, want >= %d (80%%)",
			technicalCount, stats.Total, float64(technicalCount)/float64(stats.Total)*100, minRequired)
	}
}
