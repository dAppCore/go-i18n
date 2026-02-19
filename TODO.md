# TODO.md — go-i18n Task Queue

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 1: Harden the Engine

- [ ] **Add CLAUDE.md** — Document the grammar engine contract: what it is (grammar primitives + reversal), what it isn't (translation file manager). Include build/test commands, the gram.* sacred rule, and the agent-flattening prohibition.
- [ ] **Ambiguity resolution for dual-class words** — Words like "run", "file", "test", "check", "build" are both verb and noun. Tokeniser currently picks first match. Need context-aware disambiguation (look at surrounding tokens: article before = noun, after subject = verb).
- [ ] **Extend irregular verb coverage** — Audit against common dev/ops vocabulary. Missing forms cause silent fallback to regular rules which may produce wrong output (e.g. "builded" instead of "built").
- [ ] **Add benchmarks** — `grammar_test.go` and `reversal/tokeniser_test.go` need `Benchmark*` functions. The engine will run in hot paths (TIM, Poindexter) — need baseline numbers.

## Phase 2: Reference Distribution

- [ ] **Reference distribution builder** — Process the 88K scored seeds from LEM Phase 0 through the tokeniser + imprint pipeline. Output: per-category (ethical, technical, harmful) reference distributions stored as JSON. This calibrates what "normal" grammar looks like.
- [ ] **Imprint comparator** — Given a new text and reference distributions, compute distance metrics (cosine, KL divergence, Mahalanobis). Return a classification signal with confidence score. This is the Poindexter integration point.

## Phase 3: Multi-Language

- [ ] **Grammar table format spec** — Document the exact JSON schema for `gram.*` keys so new languages can be added. Currently only inferred from `en.json`.
- [ ] **French grammar tables** — First non-English language. French has gendered nouns, complex verb conjugation, elision rules. Good stress test for the grammar engine's language-agnostic design.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
4. New discoveries → add tasks, flag in FINDINGS.md
