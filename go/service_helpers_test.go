package i18n

import (
	"testing"

	"dappco.re/go"
)

// TestServiceHelpers_failResult covers each branch of the result-coercion
// helper: a failed core.Result is passed through, an OK result carrying an
// error value is unwrapped, an OK result with a non-error value is wrapped, a
// bare error is wrapped, and a plain value is stringified.
func TestServiceHelpers_failResult(t *testing.T) {
	// Already-failed core.Result passes through unchanged.
	orig := core.Fail(core.NewError("boom"))
	if got := failResult(orig); got.OK {
		t.Error("failResult(failed result) should stay failed")
	}

	// OK result whose value is an error gets unwrapped to a failure.
	if got := failResult(core.Ok(core.NewError("inner"))); got.OK {
		t.Error("failResult(ok-with-error) should fail")
	}

	// OK result with a non-error value is wrapped as a failure.
	if got := failResult(core.Ok("not-an-error")); got.OK {
		t.Error("failResult(ok-with-value) should fail")
	}

	// Bare error.
	if got := failResult(core.NewError("bare")); got.OK {
		t.Error("failResult(error) should fail")
	}

	// Plain value gets stringified.
	if got := failResult(42); got.OK {
		t.Error("failResult(value) should fail")
	}
}

// TestServiceHelpers_localeFromResult covers the loadedLocale, []any and error
// branches of the loader-result decoder.
func TestServiceHelpers_localeFromResult(t *testing.T) {
	// Failed result is propagated.
	if _, _, status := localeFromResult(core.Fail(core.NewError("nope"))); status.OK {
		t.Error("localeFromResult(failed) should propagate failure")
	}

	// loadedLocale value unwraps to messages + grammar.
	msgs := map[string]Message{"k": {Text: "v"}}
	gd := &GrammarData{Words: map[string]string{"a": "b"}}
	gotMsgs, gotGrammar, status := localeFromResult(core.Ok(loadedLocale{Messages: msgs, Grammar: gd}))
	if !status.OK || gotMsgs["k"].Text != "v" || gotGrammar.Words["a"] != "b" {
		t.Errorf("localeFromResult(loadedLocale) = %v/%v (ok=%v)", gotMsgs, gotGrammar, status.OK)
	}

	// []any with messages + grammar.
	gotMsgs, gotGrammar, status = localeFromResult(core.Ok([]any{msgs, gd}))
	if !status.OK || gotMsgs["k"].Text != "v" || gotGrammar.Words["a"] != "b" {
		t.Errorf("localeFromResult([]any) = %v/%v (ok=%v)", gotMsgs, gotGrammar, status.OK)
	}

	// []any too short → failure.
	if _, _, status := localeFromResult(core.Ok([]any{msgs})); status.OK {
		t.Error("localeFromResult(short []any) should fail")
	}

	// []any with wrong message type → failure.
	if _, _, status := localeFromResult(core.Ok([]any{"bad", gd})); status.OK {
		t.Error("localeFromResult([]any wrong msg type) should fail")
	}

	// Unexpected value type → failure.
	if _, _, status := localeFromResult(core.Ok(123)); status.OK {
		t.Error("localeFromResult(unexpected) should fail")
	}
}

// TestServiceHelpers_promptLookupKeys covers the prompt., common.prompt. and
// default namespacing branches plus the empty-suffix special cases.
func TestServiceHelpers_promptLookupKeys(t *testing.T) {
	tests := []struct {
		key   string
		first string
	}{
		{"prompt.yes", "prompt.yes"},
		{"prompt.", "prompt."},
		{"common.prompt.yes", "common.prompt.yes"},
		{"common.prompt.", "common.prompt."},
		{"yes", "prompt.yes"},
	}
	for _, tt := range tests {
		got := promptLookupKeys(tt.key)
		if len(got) != 2 {
			t.Errorf("promptLookupKeys(%q) returned %d keys, want 2", tt.key, len(got))
			continue
		}
		if got[0] != tt.first {
			t.Errorf("promptLookupKeys(%q)[0] = %q, want %q", tt.key, got[0], tt.first)
		}
	}
}

// TestServiceHelpers_lookupSegment normalises arbitrary text into a key segment:
// letters/digits kept, separators collapsed to single underscores and trimmed
// from both ends.
//
//	lookupSegment("  Hello, World!  ") // "hello_world"
func TestServiceHelpers_lookupSegment(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Hello", "hello"},
		{"  Hello, World!  ", "hello_world"},
		{"a..b--c__d", "a_b_c_d"},
		{"___trim___", "trim"},
		{"", ""},
		{"!!!", ""},
	}
	for _, tt := range tests {
		if got := lookupSegment(tt.in); got != tt.want {
			t.Errorf("lookupSegment(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestServiceHelpers_lookupVariantsAllocationBudget(t *testing.T) {
	extra := map[string]any{"tier": "gold"}
	var variants []string

	allocs := testing.AllocsPerRun(1000, func() {
		variants = lookupVariants("app.context", "dashboard", "f", "eu", FormalityFormal, extra)
	})
	if len(variants) != 31 {
		t.Fatalf("lookupVariants() returned %d variants; want 31", len(variants))
	}
	if variants[0] != "app.context._dashboard._f._eu._formal._tier._gold" {
		t.Fatalf("lookupVariants()[0] = %q; want extra-specific first variant", variants[0])
	}
	if variants[len(variants)-1] != "app.context" {
		t.Fatalf("lookupVariants() last = %q; want base key", variants[len(variants)-1])
	}
	if allocs > 40 {
		t.Fatalf("lookupVariants() allocated %.0f times; want <= 40", allocs)
	}
}

// TestServiceHelpers_runeHelpers covers firstRuneOf/lastRuneOf/trimRuneRun on
// ASCII, multibyte and empty inputs, plus compareStrings ordering.
func TestServiceHelpers_runeHelpers(t *testing.T) {
	if r, n := firstRuneOf("héllo"); r != 'h' || n != 1 {
		t.Errorf("firstRuneOf(héllo) = (%c, %d), want (h, 1)", r, n)
	}
	if r, n := firstRuneOf("éclair"); r != 'é' || n != 2 {
		t.Errorf("firstRuneOf(éclair) = (%c, %d), want (é, 2)", r, n)
	}
	if r, n := firstRuneOf(""); r != 0 || n != 0 {
		t.Errorf("firstRuneOf(empty) = (%d, %d), want (0, 0)", r, n)
	}
	if r, n := lastRuneOf("café"); r != 'é' || n != 2 {
		t.Errorf("lastRuneOf(café) = (%c, %d), want (é, 2)", r, n)
	}

	if got := trimRuneRun("__foo__", '_'); got != "foo" {
		t.Errorf("trimRuneRun(__foo__) = %q, want %q", got, "foo")
	}
	if got := trimRuneRun("nochange", '_'); got != "nochange" {
		t.Errorf("trimRuneRun(nochange) = %q, want unchanged", got)
	}

	if compareStrings("en", "fr") >= 0 {
		t.Error("compareStrings(en, fr) should be negative")
	}
	if compareStrings("fr", "fr") != 0 {
		t.Error("compareStrings(fr, fr) should be 0")
	}
}
