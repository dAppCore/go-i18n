# 1B Pre-Sort Pipeline — Completion Summary

**Completed:** 20 February 2026
**Module:** `forge.lthn.ai/core/go-i18n`
**Status:** Complete — batch classification pipeline shipping at 80 prompts/sec

## What Was Built

Batch classification pipeline that runs a JSONL corpus through Gemma3-1B to
add `domain_1b` labels (`{technical, creative, ethical, casual}`). Pre-sorts
the 88K Phase 0 seeds for reference distribution building (Phase 2b).

### Key components

- `ClassifyCorpus()` — reads JSONL, batches prompts, classifies via MLX
  inference, writes labelled output
- Batch processing at 80 prompts/sec on M3 Ultra (native MLX backend)
- Integration tests against live Gemma3-1B model
- Output feeds into go-i18n reference distribution calibration

### Integration

Uses `go-mlx` batch inference (BatchGenerate with prefill-only mode) for
high-throughput single-token classification. The 1B model runs entirely
on Metal with no CPU fallback needed.
