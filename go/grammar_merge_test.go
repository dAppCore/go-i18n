package i18n

import "testing"

// TestGrammarMerge_MergeGrammarData_IntoExisting drives the merge-into-existing
// branch of MergeGrammarData, exercising every sub-merger: articles,
// punctuation, signals (append-unique + priors), intents (cloned) and the
// number-format field overrides. The simpler word-merge and nil guards are
// covered elsewhere; this targets the second-merge path where an existing entry
// already exists.
func TestGrammarMerge_MergeGrammarData_IntoExisting(t *testing.T) {
	const lang = "qaa-x-merge"
	t.Cleanup(func() { SetGrammarData(lang, nil) })

	// First merge seeds the cache entry.
	MergeGrammarData(lang, &GrammarData{
		Words:    map[string]string{"a": "one"},
		Articles: ArticleForms{IndefiniteDefault: "un"},
		Punct:    PunctuationRules{LabelSuffix: ":"},
		Signals:  SignalData{NounDeterminers: []string{"le"}},
		Intents:  map[string]Intent{"first": {Question: "Q1"}},
		Number:   NumberFormat{ThousandsSep: " "},
	})

	// Second merge layers on top of the existing entry.
	MergeGrammarData(lang, &GrammarData{
		Words: map[string]string{"b": "two"},
		Articles: ArticleForms{
			IndefiniteVowel: "un'",
			Definite:        "le",
			ByGender:        map[string]string{"f": "la"},
		},
		Punct: PunctuationRules{ProgressSuffix: "..."},
		Signals: SignalData{
			NounDeterminers: []string{"le", "la"}, // "le" already present → dedup
			Priors:          map[string]map[string]float64{"set": {"verb": 0.7}},
		},
		Intents: map[string]Intent{"second": {Meta: IntentMeta{Supports: []string{"x"}}}},
		Number:  NumberFormat{DecimalSep: ",", PercentFmt: "%s %%"},
	})

	got := GetGrammarData(lang)
	if got == nil {
		t.Fatal("GetGrammarData returned nil after merge")
	}

	// Words merged from both.
	if got.Words["a"] != "one" || got.Words["b"] != "two" {
		t.Errorf("Words = %v, want both a and b", got.Words)
	}

	// Article fields: seeded default kept, vowel/definite/ByGender added.
	if got.Articles.IndefiniteDefault != "un" {
		t.Errorf("IndefiniteDefault = %q, want %q", got.Articles.IndefiniteDefault, "un")
	}
	if got.Articles.IndefiniteVowel != "un'" || got.Articles.Definite != "le" {
		t.Errorf("article merge = %+v", got.Articles)
	}
	if got.Articles.ByGender["f"] != "la" {
		t.Errorf("ByGender = %v, want f:la", got.Articles.ByGender)
	}

	// Punct: both suffixes present.
	if got.Punct.LabelSuffix != ":" || got.Punct.ProgressSuffix != "..." {
		t.Errorf("punct merge = %+v", got.Punct)
	}

	// Signals: append-unique kept "le" once and added "la"; priors merged.
	if len(got.Signals.NounDeterminers) != 2 {
		t.Errorf("NounDeterminers = %v, want 2 unique", got.Signals.NounDeterminers)
	}
	if got.Signals.Priors["set"]["verb"] != 0.7 {
		t.Errorf("Signals.Priors = %v, want set/verb 0.7", got.Signals.Priors)
	}

	// Intents: both present, Supports slice cloned.
	if _, ok := got.Intents["first"]; !ok {
		t.Error("intent 'first' lost after merge")
	}
	if intent, ok := got.Intents["second"]; !ok || len(intent.Meta.Supports) != 1 {
		t.Errorf("intent 'second' = %+v, want Supports len 1", intent)
	}

	// Number: seeded thousands kept, decimal/percent added.
	if got.Number.ThousandsSep != " " || got.Number.DecimalSep != "," || got.Number.PercentFmt != "%s %%" {
		t.Errorf("number merge = %+v", got.Number)
	}
}

// TestGrammarMerge_mergeIntentData_Direct covers mergeIntentData's three
// branches: empty source (returns dst untouched), nil destination (allocates),
// and merge-into-existing.
func TestGrammarMerge_mergeIntentData_Direct(t *testing.T) {
	// Empty source returns dst unchanged.
	dst := map[string]Intent{"a": {Question: "Q"}}
	if got := mergeIntentData(dst, nil); len(got) != 1 || got["a"].Question != "Q" {
		t.Errorf("mergeIntentData(empty src) = %v, want unchanged", got)
	}

	// Nil destination allocates a new map.
	got := mergeIntentData(nil, map[string]Intent{"b": {Confirm: "C"}})
	if got == nil || got["b"].Confirm != "C" {
		t.Errorf("mergeIntentData(nil dst) = %v, want allocated with b", got)
	}

	// Merge into existing; Supports slice is cloned (not aliased).
	src := map[string]Intent{"c": {Meta: IntentMeta{Supports: []string{"s1"}}}}
	merged := mergeIntentData(map[string]Intent{"a": {Question: "Q"}}, src)
	if len(merged) != 2 {
		t.Fatalf("mergeIntentData merged len = %d, want 2", len(merged))
	}
	merged["c"].Meta.Supports[0] = "mutated"
	if src["c"].Meta.Supports[0] != "s1" {
		t.Error("mergeIntentData did not clone the Supports slice (source mutated)")
	}
}

// TestGrammarMerge_appendUniqueStrings covers de-duplication, empty-skip and the
// empty-values short-circuit.
func TestGrammarMerge_appendUniqueStrings(t *testing.T) {
	if got := appendUniqueStrings([]string{"a"}); len(got) != 1 {
		t.Errorf("appendUniqueStrings(no values) = %v, want unchanged", got)
	}
	got := appendUniqueStrings([]string{"a", "b"}, "b", "", "c", "a")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("appendUniqueStrings = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("appendUniqueStrings[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
