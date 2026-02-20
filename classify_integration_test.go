//go:build integration

package i18n

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"forge.lthn.ai/core/go-inference"
	_ "forge.lthn.ai/core/go-mlx" // registers Metal backend
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
		lines = append(lines, fmt.Sprintf(`{"id":%d,"prompt":"Delete the configuration file and rebuild the project"}`, i))
	}
	input := strings.NewReader(strings.Join(lines, "\n") + "\n")

	var output bytes.Buffer
	start := time.Now()
	stats, err := ClassifyCorpus(context.Background(), model, input, &output, WithBatchSize(8))
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
}
