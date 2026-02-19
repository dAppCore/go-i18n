# FINDINGS.md — go-i18n Research & Discovery

Record findings, gaps, and architectural decisions here as work progresses.

---

## 2026-02-19: Initial Assessment (Virgil)

### Current State

- 5,800 lines across 32 files (14 test files)
- All tests pass
- Only dependency: `golang.org/x/text`
- Grammar engine is solid: forward composition + reversal + imprint + multiplier

### Architecture

go-i18n is a **grammar engine**, not a translation file manager. Consumers bring their own translations. The library provides:

1. **Forward composition** — Grammar primitives that compose grammatically correct text
2. **Reverse grammar** — Tokeniser reads grammar tables backwards to extract structure
3. **GrammarImprint** — Lossy feature vector projection (content to grammar fingerprint)
4. **Multiplier** — Deterministic training data augmentation (no LLM calls)

### Key Gaps Identified

| Gap | Impact | Notes |
|-----|--------|-------|
| No CLAUDE.md | High | Agents don't know the rules, will flatten locale files |
| Dual-class word ambiguity | Medium | "file" as verb vs noun, "run" as verb vs noun |
| No benchmarks | Medium | No perf baselines for hot-path usage (TIM, Poindexter) |
| No reference distributions | High | Can't calibrate imprints without scored seed data |
| Only English grammar tables | Medium | Reversal only works with loaded GrammarData |

### Sacred Rules

- `gram.*` keys in locale JSON MUST remain nested — flattening breaks the grammar engine
- Irregular forms in grammar tables override regular morphological rules
- Round-trip property must hold: forward(base) then reverse must recover base
- go-i18n does NOT ship or manage consumer translation files
