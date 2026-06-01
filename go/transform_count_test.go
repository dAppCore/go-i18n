package i18n

import "testing"

// TestTransformCount_getCount drives every data shape the count extractor
// inspects: Subject, TranslationContext (countValue / Extra / raw count),
// map[string]any, map[string]int, map[string]string and the bare-value
// fallback.
func TestTransformCount_getCount(t *testing.T) {
	if got := getCount(nil); got != 0 {
		t.Errorf("getCount(nil) = %d, want 0", got)
	}

	// Subject with an explicit count.
	if got := getCount(S("file", "x").Count(3)); got != 3 {
		t.Errorf("getCount(Subject) = %d, want 3", got)
	}

	// TranslationContext with countSet via Count().
	if got := getCount(C("greeting").Count(5)); got != 5 {
		t.Errorf("getCount(Context.Count) = %d, want 5", got)
	}

	// TranslationContext whose count comes from Extra["Count"].
	if got := getCount(C("greeting").Set("Count", 7)); got != 7 {
		t.Errorf("getCount(Context.Extra Count) = %d, want 7", got)
	}
	if got := getCount(C("greeting").Set("count", 9)); got != 9 {
		t.Errorf("getCount(Context.Extra count) = %d, want 9", got)
	}

	// map[string]any.
	if got := getCount(map[string]any{"Count": 11}); got != 11 {
		t.Errorf("getCount(map[any] Count) = %d, want 11", got)
	}
	if got := getCount(map[string]any{"count": 12}); got != 12 {
		t.Errorf("getCount(map[any] count) = %d, want 12", got)
	}

	// map[string]int.
	if got := getCount(map[string]int{"Count": 13}); got != 13 {
		t.Errorf("getCount(map[int] Count) = %d, want 13", got)
	}
	if got := getCount(map[string]int{"count": 14}); got != 14 {
		t.Errorf("getCount(map[int] count) = %d, want 14", got)
	}

	// map[string]string.
	if got := getCount(map[string]string{"Count": "15"}); got != 15 {
		t.Errorf("getCount(map[string] Count) = %d, want 15", got)
	}

	// Bare value fallback.
	if got := getCount(42); got != 42 {
		t.Errorf("getCount(bare int) = %d, want 42", got)
	}
}

// TestTransformCount_getCount_NilTypedPointers confirms typed-nil Subject and
// TranslationContext pointers are handled without panic.
func TestTransformCount_getCount_NilTypedPointers(t *testing.T) {
	var s *Subject
	if got := getCount(s); got != 0 {
		t.Errorf("getCount(nil *Subject) = %d, want 0", got)
	}
	var c *TranslationContext
	if got := getCount(c); got != 0 {
		t.Errorf("getCount(nil *TranslationContext) = %d, want 0", got)
	}
}

// TestTransformCount_countValue covers the nil-receiver and set/unset return
// shapes of TranslationContext.countValue.
func TestTransformCount_countValue(t *testing.T) {
	var nilCtx *TranslationContext
	if n, ok := nilCtx.countValue(); n != 1 || ok {
		t.Errorf("nil.countValue() = (%d, %v), want (1, false)", n, ok)
	}

	unset := C("greeting")
	if _, ok := unset.countValue(); ok {
		t.Error("unset context countValue should report not-set")
	}

	set := C("greeting").Count(4)
	if n, ok := set.countValue(); n != 4 || !ok {
		t.Errorf("set.countValue() = (%d, %v), want (4, true)", n, ok)
	}
}

// TestTransformCount_toInt covers every numeric type, the string-parse paths
// (valid, empty, unparseable) and the nil/unsupported fallbacks.
func TestTransformCount_toInt(t *testing.T) {
	tests := []struct {
		in   any
		want int
	}{
		{nil, 0},
		{int(1), 1},
		{int64(2), 2},
		{int32(3), 3},
		{int16(4), 4},
		{int8(5), 5},
		{uint(6), 6},
		{uint64(7), 7},
		{uint32(8), 8},
		{uint16(9), 9},
		{uint8(10), 10},
		{float64(11.9), 11},
		{float32(12.9), 12},
		{" 13 ", 13},
		{"", 0},
		{"notnum", 0},
		{struct{}{}, 0},
	}
	for _, tt := range tests {
		if got := toInt(tt.in); got != tt.want {
			t.Errorf("toInt(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestTransformCount_subjectArgText drives every arg shape subjectArgText
// resolves: string, Subject, TranslationContext (String + Extra fallback), both
// map flavours, a generic Stringer, and the unrecognised default.
func TestTransformCount_subjectArgText(t *testing.T) {
	if got := subjectArgText("plain"); got != "plain" {
		t.Errorf("subjectArgText(string) = %q, want %q", got, "plain")
	}

	if got := subjectArgText(S("file", "config.yaml")); got != "config.yaml" {
		t.Errorf("subjectArgText(Subject) = %q, want %q", got, "config.yaml")
	}

	// TranslationContext with a non-empty name returns the name via String().
	if got := subjectArgText(C("greeting")); got != "greeting" {
		t.Errorf("subjectArgText(Context name) = %q, want %q", got, "greeting")
	}

	// TranslationContext with an empty name falls back to the Extra subject.
	if got := subjectArgText(C("").Set("Subject", "fromExtra")); got != "fromExtra" {
		t.Errorf("subjectArgText(Context Extra) = %q, want %q", got, "fromExtra")
	}

	if got := subjectArgText(map[string]any{"value": "mapval"}); got != "mapval" {
		t.Errorf("subjectArgText(map[any]) = %q, want %q", got, "mapval")
	}
	if got := subjectArgText(map[string]string{"noun": "mapnoun"}); got != "mapnoun" {
		t.Errorf("subjectArgText(map[string]) = %q, want %q", got, "mapnoun")
	}

	// Generic Stringer.
	if got := subjectArgText(stringerArg("stringy")); got != "stringy" {
		t.Errorf("subjectArgText(Stringer) = %q, want %q", got, "stringy")
	}

	// Unrecognised types and typed nils → empty.
	if got := subjectArgText(42); got != "" {
		t.Errorf("subjectArgText(int) = %q, want empty", got)
	}
	var nilSubj *Subject
	if got := subjectArgText(nilSubj); got != "" {
		t.Errorf("subjectArgText(nil Subject) = %q, want empty", got)
	}
	var nilCtx *TranslationContext
	if got := subjectArgText(nilCtx); got != "" {
		t.Errorf("subjectArgText(nil Context) = %q, want empty", got)
	}
}

// stringerArg is a minimal fmt.Stringer used to drive subjectArgText's generic
// stringer branch.
type stringerArg string

func (s stringerArg) String() string { return string(s) }
