# Project History

## Initial Assessment (2026-02-19)

**State at assessment**: 5,800 lines across 32 files (14 test files). All tests passing. One external dependency (`golang.org/x/text`). Grammar engine solid: forward composition, reversal, imprint, and multiplier all functional.

**Key gaps identified**:

| Gap | Impact |
|-----|--------|
| No CLAUDE.md — agents would flatten locale files | High |
| Dual-class word ambiguity (e.g. "file" as verb vs noun) | Medium |
| No benchmark baselines for hot-path usage | Medium |
| No reference distributions for imprint calibration | High |
| English-only grammar tables | Medium |

---

## Phase 1: Engine Hardening

**Commit d5b3eac** — Added CLAUDE.md with the critical rule against flattening `gram.*` locale JSON. Established the grammar engine's identity as a primitive provider, not a translation file manager.

**Commit 3848297** — Two-pass probabilistic disambiguation for dual-class words.

Words in both `gram.verb` and `gram.noun`: `{commit, run, test, check, file, build}`. Previously resolved verb-first without context. Now uses a two-pass algorithm with seven weighted signals.

Pass 1 classifies unambiguous tokens and marks base forms of dual-class words as `tokenAmbiguous`. Inflected forms self-resolve (e.g. `"committed"` → verb past, `"commits"` → noun plural).

Pass 2 evaluates signals:

| Signal | Weight |
|--------|--------|
| `noun_determiner` | 0.35 |
| `verb_auxiliary` | 0.25 |
| `following_class` | 0.15 |
| `sentence_position` | 0.10 |
| `verb_saturation` | 0.10 |
| `inflection_echo` | 0.03 |
| `default_prior` | 0.02 |

Design decisions recorded during implementation:
- Confidence floor of 0.55/0.45 when only the default prior fires (total < 0.10), preventing misleading 1.0 confidence from a single weak signal (fix B3)
- Contractions (`don't`, `can't`, `won't`, etc.) added to `verb_auxiliary` signal list (fix D1)
- Clause boundary isolation for `verb_saturation` — scans within punctuation and coordinating conjunctions only (fix D2)
- `WithWeights()` option for configurable signal weights without code changes (fix F3)
- `DisambiguationStats` for aggregate Phase 2 calibration (fix F1)
- `WithSignals()` opt-in for per-token signal diagnostics (kept out of hot path)
- `buildSignalIndex()` guards each signal list independently, allowing partial locale data to fall back per-field (fix R3)
- Removed `"passed"`, `"failed"`, `"skipped"` from `gram.noun` and `gram.word` — these are past participles, not nouns (fix R1)

Test coverage for this commit: 9 disambiguation scenario tests, 12 dual-class round-trip tests, imprint convergence test, `DisambiguationStats` tests, `WithWeights` override test. Race detector clean.

**Same session** — Extended irregular verb coverage.

Added 44 irregular verbs:
- 17 compound irregular (prefix + base): `undo`, `redo`, `rerun`, `rewrite`, `rebuild`, `resend`, `override`, `rethink`, `remake`, `undergo`, `overcome`, `withdraw`, `uphold`, `withhold`, `outgrow`, `outrun`, `overshoot`
- 22 simple irregular (dev/ops): `become`, `come`, `give`, `fall`, `understand`, `arise`, `bind`, `spin`, `quit`, `cast`, `broadcast`, `burst`, `cost`, `shed`, `rid`, `shrink`, `shoot`, `forbid`, `offset`, `upset`, `input`, `output`
- 5 CVC doubling overrides: `debug`, `embed`, `unzip`, `remap`, `unpin`, `unwrap` — words with stressed final syllable that `shouldDoubleConsonant()` misses because they exceed four characters

Total irregular verb count: approximately 140 (from approximately 96).

**Same session** — Added benchmarks.

8 forward composition benchmarks, 7 reversal benchmarks. Baselines on M3 Ultra (arm64):

Forward composition:

| Benchmark | ns/op | allocs/op |
|-----------|-------|-----------|
| PastTense (irregular) | 25.67 | 0 |
| PastTense (regular) | 48.52 | 1 |
| PastTense (compound) | 26.15 | 0 |
| Gerund | 25.87 | 0 |
| Pluralize | 67.97 | 1 |
| Article | 177.4 | 0 |
| Progress | 107.1 | 2 |
| ActionResult | 115.3 | 3 |

Reversal engine:

| Benchmark | ns/op | allocs/op |
|-----------|-------|-----------|
| Tokenise (3 words) | 639 | 8 |
| Tokenise (12 words) | 2859 | 14 |
| Tokenise (dual-class) | 1657 | 9 |
| Tokenise (WithSignals) | 2255 | 28 |
| NewImprint | 648 | 10 |
| Imprint.Similar | 516 | 0 |
| Multiplier.Expand | 3609 | 63 |

Key observations:
- `Similar` is zero-alloc at 516 ns/op — hot-path safe for high-volume imprint comparison
- Tokenise scales linearly at approximately 200–240 ns/word
- `WithSignals` adds 36% latency and 3x allocs — keep opt-in

---

## Phase 2a: 1B Pre-Classification

**Classification benchmark results** (220 domain-tagged sentences, 55 per domain, leave-one-out imprint similarity):

| Domain | Accuracy | Tense signature |
|--------|----------|-----------------|
| Technical | 78.2% | base 46%, gerund 30%, past 24% |
| Creative | 81.8% | past 80%, gerund 16%, base 4% |
| Ethical | 45.5% | base 55%, past 25%, gerund 20% |
| Casual | 10.9% | past 70%, base 17%, gerund 14% |

Overall: 54.1% (versus 25% random chance).

Confusion axes:
- Ethical → Technical: both use base-form verbs heavily (prescriptive vs imperative register share the same grammar profile)
- Casual → Creative: both use past tense heavily (anecdotal vs narrative register share the same grammar profile)

Conclusion: grammar-based classification is a strong first pass for technical (78%) and creative (82%). The 1B model is specifically needed for the ethical/technical and casual/creative axes.

**LEK-Gemma3-1B-v2 benchmark** (M3 Ultra, temp=0.05):
- Domain classification: 75% across three evaluation rounds, consistent
- Article correctness T/F: 100% (three cases)
- Irregular base forms A/B: 100% (two cases)
- Dead zones: pattern fill (0%), tense detection (50%), generative output (unreliable)

At 0.17s per classification, a single M3 can pre-sort approximately 5,000 sentences per second. The 88K Phase 0 seed corpus would take approximately 15–18 seconds.

**`ClassifyCorpus()` added to `classify.go`**

Streaming JSONL input → batch classification via `inference.TextModel` → JSONL output with `domain_1b` field. Configurable batch size, prompt field, prompt template. Mock-testable via `inference.TextModel` interface.

Integration test results: 50 prompts classified in 625ms (80 prompts/second), all 50 technical prompts correctly labelled as `"technical"`.

**`CalibrateDomains()` added to `calibrate.go`** — commit 3b7ef9d

Accepts two `TextModel` instances (model A = 1B, model B = 27B), classifies the full corpus sequentially with each model (A then B, to manage memory for large models), and computes agreement metrics, confusion pairs, and accuracy against ground truth.

Integration corpus: 500 samples (220 ground-truth + 280 unlabelled). Soft assertion: agreement rate greater than 50%.

**Virgil review fixes applied**: go.mod cleanup, prefix collision fix in `mapTokenToDomain()`, short-mode skip in integration tests, accuracy assertion on 5 items minimum.

**Article/irregular validator** — single-token classification via `m.Generate()` at temp=0.05 for use as a lightweight grammar validator in the forward composition path.

---

## Phase 2b: Reference Distributions

**Commit c3e9153** — Reference distribution builder, imprint comparator, anomaly detection.

**`BuildReferences()` in `reversal/reference.go`**

Tokenises classified samples, builds imprints, groups by domain, computes centroid and per-key variance for each domain. Centroid is computed by accumulating all map fields then normalising (L1 norm). Variance is sample variance, prefixed by component.

**`Compare()` and `Classify()`**

Three distance metrics between a query imprint and each reference centroid:
- Cosine similarity via `Similar()` (weighted, same component weights as imprint comparison)
- Symmetric KL divergence (Jensen-Shannon style, epsilon-smoothed at 1e-10)
- Mahalanobis distance (variance-normalised Euclidean, falls back to unit variance when variance map is absent)

`Classify()` ranks by cosine similarity and returns the best domain plus the confidence margin (gap between 1st and 2nd similarity scores).

**`DetectAnomalies()` in `reversal/anomaly.go`**

Compares model-assigned domain labels against imprint-based classification. Flags mismatches. Returns per-sample `AnomalyResult` and aggregate `AnomalyStats`.

Validated: a creative sentence tagged as technical by the model is correctly flagged as an anomaly with confidence 0.37.

Key findings from this phase:
- Grammar alone separates technical from creative (cosine similarity gap of 0.21)
- Ethical/technical and casual/creative overlap persists — confirms the 1B model is required for those axes
- Anomaly detection is a training signal for human review or 27B spot-checking
- `ReferenceSet` API is the bridge to the Poindexter trust verification pipeline

Test coverage: technical vs creative centroid distance test, creative-mislabelled-as-technical anomaly test, single-domain degenerate case, KL identity test (symmetric KL of identical distributions ≈ 0.0), Mahalanobis unit-variance fallback test.

---

## Phase 3: Multi-Language

**Grammar table format specification** — `docs/grammar-table-spec.md`

Full JSON schema documenting all `gram.*` sections: verb, noun, article, word, punct, signal, number. Includes detection rules (how the loader identifies verb vs noun vs plural objects), fallback chain documentation, dual-class word guidance, and step-by-step instructions for adding a new language.

**French grammar tables** — `locales/fr.json`

- 50 verb conjugations (passé composé participle form for `past`, présent participe for `gerund`)
- 24 gendered nouns with `"m"` or `"f"` gender
- Gendered articles: `by_gender: {"m": "le", "f": "la"}`, indefinite `"un"` (vowel and default identical in French)
- Punctuation: `label: " :"` (French typographic convention: space before colon)
- Full `gram.signal` lists with French determiners and auxiliaries

---

## Known Limitations

**Grammar-based classification ceiling**

At 54.1% overall accuracy, the imprint alone cannot distinguish ethical from technical or casual from creative. These axes require semantic understanding that grammar features cannot provide. The 1B model addresses this, but with its own ceiling at 75% domain accuracy.

**Tier 2 and tier 3 are English-only**

The `irregularVerbs` and `irregularNouns` Go maps and the regular morphology rules encode English patterns. For French, German, Spanish, or other languages, all irregular forms must be in the JSON grammar tables. A French word not in `locales/fr.json` will fall through to the English irregular maps (unlikely to match) and then to English morphology rules (will produce wrong output).

**French reversal**

Elision (`l'`) and plural articles (`les`, `des`) are not handled by the current `Article()` function or the reversal tokeniser. The `by_gender` article map supports gendered articles for composition, but the reversal tokeniser's `MatchArticle()` only checks `IndefiniteDefault`, `IndefiniteVowel`, and `Definite`. French reversal is therefore incomplete.

**Dual-class expansion candidates not yet measured**

Twenty additional words are candidates for the dual-class set: `patch`, `release`, `update`, `change`, `merge`, `push`, `pull`, `tag`, `log`, `watch`, `link`, `host`, `import`, `export`, `process`, `function`, `handle`, `trigger`, `stream`, `queue`. The decision to add any of them should be based on measured imprint drift in the 88K seed corpus rather than intuition.

**Multiplier allocation cost**

`Multiplier.Expand` allocates 63 objects for a four-word sentence. This is acceptable at current usage volumes but would become a bottleneck if called at high frequency. Token slice pooling is the obvious mitigation.

**88K seed corpus not yet processed**

The Phase 0 LEM seed corpus (88K sentences) has not been run through the pre-sort pipeline or used to build reference distributions. The current reference distributions and classification benchmarks are based on 220–500 manually curated sentences.

---

## Future Considerations

These are not tasked. They represent the natural next work given what has been built.

**Expanded dual-class words**

Measure imprint drift on the 88K seeds for the 20 candidate words listed above. Add only those that show statistically meaningful drift between verb and noun roles — adding words that do not cause imprint changes has no benefit and increases disambiguation overhead.

**French reversal**

Extend `Article()` to handle elision (`l'` before vowel-initial nouns) and plural forms (`les`, `des`). Update `MatchArticle()` in the reversal tokeniser to recognise the full French article set including gendered and plural variants.

**88K seed corpus processing**

Run `ClassifyCorpus()` against the full Phase 0 seed corpus to produce domain-tagged JSONL. Use the output to call `BuildReferences()` and produce reference distributions grounded in real data rather than the 220-sentence hand-curated set. This would make anomaly detection and imprint-based classification significantly more reliable.

**Corpus-derived word priors**

`SignalData.Priors` (`map[string]map[string]float64`) is reserved for per-word priors derived from corpus frequencies. Currently unused. A corpus-derived prior would allow, for example, `"run"` to carry a higher verb prior in technical contexts based on observed frequencies rather than the fixed 0.02 default prior.
