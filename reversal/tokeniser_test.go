package reversal

import (
	"testing"

	i18n "dappco.re/go/core/i18n"
)

func setup(t *testing.T) {
	t.Helper()
	svc, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)
}

func TestTokeniser_MatchVerb_Irregular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word    string
		wantOK  bool
		wantBase string
		wantTense string
	}{
		// Irregular past tense
		{"deleted", true, "delete", "past"},
		{"deleting", true, "delete", "gerund"},
		{"went", true, "go", "past"},
		{"going", true, "go", "gerund"},
		{"was", true, "be", "past"},
		{"being", true, "be", "gerund"},
		{"ran", true, "run", "past"},
		{"running", true, "run", "gerund"},
		{"wrote", true, "write", "past"},
		{"writing", true, "write", "gerund"},
		{"built", true, "build", "past"},
		{"building", true, "build", "gerund"},
		{"committed", true, "commit", "past"},
		{"committing", true, "commit", "gerund"},

		// Base forms
		{"delete", true, "delete", "base"},
		{"go", true, "go", "base"},

		// Unknown words return false
		{"xyzzy", false, "", ""},
		{"flurble", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchVerb(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchVerb(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchVerb(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Tense != tt.wantTense {
				t.Errorf("MatchVerb(%q).Tense = %q, want %q", tt.word, match.Tense, tt.wantTense)
			}
		})
	}
}

func TestTokeniser_MatchNoun_Irregular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word       string
		wantOK     bool
		wantBase   string
		wantPlural bool
	}{
		{"files", true, "file", true},
		{"file", true, "file", false},
		{"people", true, "person", true},
		{"person", true, "person", false},
		{"children", true, "child", true},
		{"child", true, "child", false},
		{"repositories", true, "repository", true},
		{"repository", true, "repository", false},
		{"branches", true, "branch", true},
		{"branch", true, "branch", false},
		{"xyzzy", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchNoun(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchNoun(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchNoun(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Plural != tt.wantPlural {
				t.Errorf("MatchNoun(%q).Plural = %v, want %v", tt.word, match.Plural, tt.wantPlural)
			}
		})
	}
}

func TestTokeniser_MatchNoun_Regular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word       string
		wantOK     bool
		wantBase   string
		wantPlural bool
	}{
		// Regular nouns NOT in grammar tables — detected by reverse morphology + round-trip
		{"servers", true, "server", true},
		{"processes", true, "process", true},
		{"entries", true, "entry", true},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchNoun(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchNoun(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchNoun(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Plural != tt.wantPlural {
				t.Errorf("MatchNoun(%q).Plural = %v, want %v", tt.word, match.Plural, tt.wantPlural)
			}
		})
	}
}

func TestTokeniser_MatchWord(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word    string
		wantCat string
		wantOK  bool
	}{
		{"URL", "url", true},
		{"url", "url", true},
		{"ID", "id", true},
		{"SSH", "ssh", true},
		{"PHP", "php", true},
		{"xyzzy", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			cat, ok := tok.MatchWord(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchWord(%q) ok=%v, want %v", tt.word, ok, tt.wantOK)
			}
			if ok && cat != tt.wantCat {
				t.Errorf("MatchWord(%q) = %q, want %q", tt.word, cat, tt.wantCat)
			}
		})
	}
}

func TestTokeniser_MatchArticle(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word     string
		wantType string
		wantOK   bool
	}{
		{"a", "indefinite", true},
		{"an", "indefinite", true},
		{"the", "definite", true},
		{"A", "indefinite", true},
		{"The", "definite", true},
		{"foo", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			artType, ok := tok.MatchArticle(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchArticle(%q) ok=%v, want %v", tt.word, ok, tt.wantOK)
			}
			if ok && artType != tt.wantType {
				t.Errorf("MatchArticle(%q) = %q, want %q", tt.word, artType, tt.wantType)
			}
		})
	}
}

func TestTokeniser_Tokenise(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("Deleted the configuration files")

	if len(tokens) != 4 {
		t.Fatalf("Tokenise() returned %d tokens, want 4", len(tokens))
	}

	// "Deleted" → verb, past tense
	if tokens[0].Type != TokenVerb {
		t.Errorf("tokens[0].Type = %v, want TokenVerb", tokens[0].Type)
	}
	if tokens[0].VerbInfo.Tense != "past" {
		t.Errorf("tokens[0].VerbInfo.Tense = %q, want %q", tokens[0].VerbInfo.Tense, "past")
	}

	// "the" → article
	if tokens[1].Type != TokenArticle {
		t.Errorf("tokens[1].Type = %v, want TokenArticle", tokens[1].Type)
	}

	// "configuration" → unknown
	if tokens[2].Type != TokenUnknown {
		t.Errorf("tokens[2].Type = %v, want TokenUnknown", tokens[2].Type)
	}

	// "files" → noun, plural
	if tokens[3].Type != TokenNoun {
		t.Errorf("tokens[3].Type = %v, want TokenNoun", tokens[3].Type)
	}
	if !tokens[3].NounInfo.Plural {
		t.Errorf("tokens[3].NounInfo.Plural = false, want true")
	}
}

func TestTokeniser_Tokenise_Punctuation(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("Building project...")
	hasPunct := false
	for _, tok := range tokens {
		if tok.Type == TokenPunctuation {
			hasPunct = true
		}
	}
	if !hasPunct {
		t.Error("did not detect punctuation in \"Building project...\"")
	}
}

func TestTokeniser_Tokenise_ClauseBoundarySentence(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("run tests. commit")
	hasSentenceEnd := false

	for _, token := range tokens {
		if token.Raw == "run" && token.Type != TokenVerb {
			t.Errorf("'run' should remain TokenVerb, got %v", token.Type)
		}
		if token.Type == TokenPunctuation && token.PunctType == "sentence_end" {
			hasSentenceEnd = true
		}
		if token.Lower == "commit" {
			// Without sentence-end boundary support, this can be demoted by verb saturation.
			// With boundary detection, it should still classify as a verb.
			if token.Type != TokenVerb {
				t.Errorf("'commit' after period should be TokenVerb, got %v", token.Type)
			}
		}
	}

	if !hasSentenceEnd {
		t.Error("did not detect sentence-end punctuation in \"run tests. commit\"")
	}
}

func TestTokeniser_Tokenise_Empty(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tokens := tok.Tokenise("")
	if len(tokens) != 0 {
		t.Errorf("Tokenise(\"\") returned %d tokens, want 0", len(tokens))
	}
}

func TestTokeniser_MatchVerb_Regular(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	tests := []struct {
		word      string
		wantOK    bool
		wantBase  string
		wantTense string
	}{
		// Regular verbs NOT in grammar tables — detected by reverse morphology + round-trip
		{"walked", true, "walk", "past"},
		{"walking", true, "walk", "gerund"},
		{"processed", true, "process", "past"},
		{"processing", true, "process", "gerund"},
		{"copied", true, "copy", "past"},
		{"copying", true, "copy", "gerund"},
		{"stopped", true, "stop", "past"},
		{"stopping", true, "stop", "gerund"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			match, ok := tok.MatchVerb(tt.word)
			if ok != tt.wantOK {
				t.Fatalf("MatchVerb(%q) ok = %v, want %v", tt.word, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if match.Base != tt.wantBase {
				t.Errorf("MatchVerb(%q).Base = %q, want %q", tt.word, match.Base, tt.wantBase)
			}
			if match.Tense != tt.wantTense {
				t.Errorf("MatchVerb(%q).Tense = %q, want %q", tt.word, match.Tense, tt.wantTense)
			}
		})
	}
}

func TestTokeniser_WithSignals(t *testing.T) {
	setup(t)
	tok := NewTokeniser(WithSignals())
	_ = tok // verify it compiles and accepts the option
}

func TestTokeniser_DualClassDetection(t *testing.T) {
	setup(t)
	tok := NewTokeniser()

	dualClass := []string{"commit", "run", "test", "check", "file", "build"}
	for _, word := range dualClass {
		if !tok.IsDualClass(word) {
			t.Errorf("%q should be dual-class", word)
		}
	}

	notDual := []string{"delete", "go", "push", "branch", "repo"}
	for _, word := range notDual {
		if tok.IsDualClass(word) {
			t.Errorf("%q should not be dual-class", word)
		}
	}
}

func TestToken_ConfidenceField(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Deleted the branch")

	for _, token := range tokens {
		if token.Type != TokenUnknown && token.Confidence == 0 {
			t.Errorf("token %q (type %d) has zero Confidence", token.Raw, token.Type)
		}
	}
}

func TestTokeniser_Disambiguate_NounAfterDeterminer(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("the commit was approved")
	if tokens[1].Type != TokenNoun {
		t.Errorf("'commit' after 'the': Type = %v, want TokenNoun", tokens[1].Type)
	}
	if tokens[1].Confidence < 0.8 {
		t.Errorf("'commit' Confidence = %f, want >= 0.8", tokens[1].Confidence)
	}
	if tokens[1].AltType != TokenVerb {
		t.Errorf("'commit' AltType = %v, want TokenVerb", tokens[1].AltType)
	}
}

func TestTokeniser_Disambiguate_VerbImperative(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Commit the changes")
	if tokens[0].Type != TokenVerb {
		t.Errorf("'Commit' imperative: Type = %v, want TokenVerb", tokens[0].Type)
	}
	if tokens[0].Confidence < 0.8 {
		t.Errorf("'Commit' Confidence = %f, want >= 0.8", tokens[0].Confidence)
	}
}

func TestTokeniser_Disambiguate_NounWithVerbSaturation(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("The test failed")
	if tokens[1].Type != TokenNoun {
		t.Errorf("'test' in 'The test failed': Type = %v, want TokenNoun", tokens[1].Type)
	}
}

func TestTokeniser_Disambiguate_VerbBeforeNoun(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Run tests")
	if tokens[0].Type != TokenVerb {
		t.Errorf("'Run' in 'Run tests': Type = %v, want TokenVerb", tokens[0].Type)
	}
}

func TestTokeniser_Disambiguate_InflectedSelfResolve(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("committed the branch")
	if tokens[0].Type != TokenVerb || tokens[0].Confidence != 1.0 {
		t.Errorf("'committed' should self-resolve as verb with confidence 1.0")
	}
	tokens = tok.Tokenise("the commits were reviewed")
	if tokens[1].Type != TokenNoun || tokens[1].Confidence != 1.0 {
		t.Errorf("'commits' should self-resolve as noun with confidence 1.0")
	}
}

func TestTokeniser_Disambiguate_VerbAfterAuxiliary(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("will commit the changes")
	if tokens[1].Type != TokenVerb {
		t.Errorf("'commit' after 'will': Type = %v, want TokenVerb", tokens[1].Type)
	}
}

func TestTokeniser_Disambiguate_ProseMultiple(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("The test failed because the commit introduced a regression")
	for _, token := range tokens {
		if token.Lower == "test" && token.Type != TokenNoun {
			t.Errorf("'test' in prose: Type = %v, want TokenNoun", token.Type)
		}
		if token.Lower == "commit" && token.Type != TokenNoun {
			t.Errorf("'commit' in prose: Type = %v, want TokenNoun", token.Type)
		}
	}
}

func TestTokeniser_Disambiguate_ClauseBoundary(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	// "passed" is a confident verb in clause 1, "commit" is a verb in clause 2
	tokens := tok.Tokenise("The test passed and we should commit the fix")
	for _, token := range tokens {
		if token.Lower == "test" && token.Type != TokenNoun {
			t.Errorf("'test' should be noun: got %v", token.Type)
		}
		if token.Lower == "commit" && token.Type != TokenVerb {
			t.Errorf("'commit' after 'should' should be verb: got %v", token.Type)
		}
	}
}

func TestTokeniser_Disambiguate_ContractionAux(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("don't run the tests")
	// "run" after "don't" (contraction auxiliary) should be verb
	for _, token := range tokens {
		if token.Lower == "run" && token.Type != TokenVerb {
			t.Errorf("'run' after \"don't\": Type = %v, want TokenVerb", token.Type)
		}
	}
}

func TestTokeniser_WithSignals_Breakdown(t *testing.T) {
	setup(t)
	tok := NewTokeniser(WithSignals())

	tokens := tok.Tokenise("the commit was approved")
	// "commit" should have a SignalBreakdown
	commitTok := tokens[1]
	if commitTok.Signals == nil {
		t.Fatal("WithSignals(): commit token has nil Signals")
	}
	if commitTok.Signals.NounScore <= commitTok.Signals.VerbScore {
		t.Errorf("NounScore (%f) should exceed VerbScore (%f) for 'the commit'",
			commitTok.Signals.NounScore, commitTok.Signals.VerbScore)
	}
	if len(commitTok.Signals.Components) == 0 {
		t.Error("Components should not be empty")
	}

	// Verify noun_determiner signal fired
	foundDet := false
	for _, c := range commitTok.Signals.Components {
		if c.Name == "noun_determiner" {
			foundDet = true
			if c.Contrib != 0.35 {
				t.Errorf("noun_determiner Contrib = %f, want 0.35", c.Contrib)
			}
		}
	}
	if !foundDet {
		t.Error("noun_determiner signal should have fired")
	}
}

func TestTokeniser_WithoutSignals_NilBreakdown(t *testing.T) {
	setup(t)
	tok := NewTokeniser() // no WithSignals

	tokens := tok.Tokenise("the commit was approved")
	if tokens[1].Signals != nil {
		t.Error("Without WithSignals(), Signals should be nil")
	}
}

func TestDisambiguationStats_WithAmbiguous(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("The commit passed the test")
	stats := DisambiguationStatsFromTokens(tokens)
	if stats.AmbiguousTokens == 0 {
		t.Error("expected ambiguous tokens for dual-class words")
	}
	if stats.TotalTokens != len(tokens) {
		t.Errorf("TotalTokens = %d, want %d", stats.TotalTokens, len(tokens))
	}
}

func TestDisambiguationStats_NoAmbiguous(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Deleted the files")
	stats := DisambiguationStatsFromTokens(tokens)
	if stats.AmbiguousTokens != 0 {
		t.Errorf("AmbiguousTokens = %d, want 0", stats.AmbiguousTokens)
	}
}

func TestWithWeights_Override(t *testing.T) {
	setup(t)
	// Override noun_determiner to 0 — "The commit" should no longer resolve as noun
	tok := NewTokeniser(WithWeights(map[string]float64{
		"noun_determiner":   0.0,
		"verb_auxiliary":    0.25,
		"following_class":   0.15,
		"sentence_position": 0.10,
		"verb_saturation":   0.10,
		"inflection_echo":   0.03,
		"default_prior":     0.02,
	}))
	tokens := tok.Tokenise("The commit")
	// With noun_determiner zeroed, default_prior (verb) should win
	if tokens[1].Type != TokenVerb {
		t.Errorf("with noun_determiner=0, 'commit' Type = %v, want TokenVerb", tokens[1].Type)
	}
}

// --- Benchmarks ---

func benchSetup(b *testing.B) {
	b.Helper()
	svc, err := i18n.New()
	if err != nil {
		b.Fatalf("i18n.New() failed: %v", err)
	}
	i18n.SetDefault(svc)
}

func BenchmarkTokenise_Short(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok.Tokenise("Delete the file")
	}
}

func BenchmarkTokenise_Medium(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	text := "The build failed because the test commit was not pushed to the branch"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok.Tokenise(text)
	}
}

func BenchmarkTokenise_DualClass(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	text := "Commit the changes and run the build test"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok.Tokenise(text)
	}
}

func BenchmarkTokenise_WithSignals(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser(WithSignals())
	text := "The commit was rebuilt and the test passed"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok.Tokenise(text)
	}
}

func BenchmarkNewImprint(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := tok.Tokenise("Delete the configuration file and rebuild the project")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewImprint(tokens)
	}
}

func BenchmarkImprint_Similar(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	imp1 := NewImprint(tok.Tokenise("Delete the configuration file"))
	imp2 := NewImprint(tok.Tokenise("Delete the old file"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imp1.Similar(imp2)
	}
}

func BenchmarkMultiplier_Expand(b *testing.B) {
	benchSetup(b)
	m := NewMultiplier()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Expand("Delete the configuration file")
	}
}
