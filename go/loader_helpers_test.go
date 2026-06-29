package i18n

import (
	"math"
	"testing"
)

// TestLoaderHelpers_float64Value_Good covers every numeric type the JSON/value
// coercion accepts and confirms the conversion is exact.
//
//	float64Value(42)        // 42, true
//	float64Value(3.5)       // 3.5, true
//	float64Value("nope")    // 0, false
func TestLoaderHelpers_float64Value_Good(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want float64
		ok   bool
	}{
		{"float64", float64(1.5), 1.5, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", int(3), 3, true},
		{"int64", int64(4), 4, true},
		{"int32", int32(5), 5, true},
		{"int16", int16(6), 6, true},
		{"int8", int8(7), 7, true},
		{"uint", uint(8), 8, true},
		{"uint64", uint64(9), 9, true},
		{"uint32", uint32(10), 10, true},
		{"uint16", uint16(11), 11, true},
		{"uint8", uint8(12), 12, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := float64Value(tt.in)
			if ok != tt.ok || got != tt.want {
				t.Errorf("float64Value(%v) = (%v, %v), want (%v, %v)", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// TestLoaderHelpers_float64Value_Bad confirms nil and non-numeric types are
// rejected with the zero value.
func TestLoaderHelpers_float64Value_Bad(t *testing.T) {
	for _, in := range []any{nil, "string", true, []int{1}, struct{}{}} {
		if got, ok := float64Value(in); ok || got != 0 {
			t.Errorf("float64Value(%v) = (%v, %v), want (0, false)", in, got, ok)
		}
	}
}

// TestLoaderHelpers_boolFromMap_Good covers native bools and every recognised
// string spelling of true/false, plus multi-key precedence.
func TestLoaderHelpers_boolFromMap_Good(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		keys   []string
		want   bool
		ok     bool
	}{
		{"native true", map[string]any{"x": true}, []string{"x"}, true, true},
		{"native false", map[string]any{"x": false}, []string{"x"}, false, true},
		{"string true", map[string]any{"x": "true"}, []string{"x"}, true, true},
		{"string yes", map[string]any{"x": "YES"}, []string{"x"}, true, true},
		{"string 1", map[string]any{"x": " 1 "}, []string{"x"}, true, true},
		{"string false", map[string]any{"x": "false"}, []string{"x"}, false, true},
		{"string no", map[string]any{"x": "no"}, []string{"x"}, false, true},
		{"string 0", map[string]any{"x": "0"}, []string{"x"}, false, true},
		{"second key", map[string]any{"y": "yes"}, []string{"x", "y"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := boolFromMap(tt.values, tt.keys...)
			if got != tt.want || ok != tt.ok {
				t.Errorf("boolFromMap(%v) = (%v, %v), want (%v, %v)", tt.values, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// TestLoaderHelpers_boolFromMap_Bad confirms missing keys and unrecognised
// strings report not-found.
func TestLoaderHelpers_boolFromMap_Bad(t *testing.T) {
	if _, ok := boolFromMap(map[string]any{"a": 1}, "x"); ok {
		t.Error("boolFromMap with missing key should report not-found")
	}
	if _, ok := boolFromMap(map[string]any{"x": "maybe"}, "x"); ok {
		t.Error("boolFromMap with unrecognised string should report not-found")
	}
	if _, ok := boolFromMap(map[string]any{"x": 3.14}, "x"); ok {
		t.Error("boolFromMap with numeric value should report not-found")
	}
}

// TestLoaderHelpers_stringSliceFromMap_Good covers []string and []any inputs,
// trimming, lowercasing and empty-entry skipping.
func TestLoaderHelpers_stringSliceFromMap_Good(t *testing.T) {
	gotStrings := stringSliceFromMap(map[string]any{"k": []string{" A ", "", "b"}}, "k")
	if len(gotStrings) != 2 || gotStrings[0] != "a" || gotStrings[1] != "b" {
		t.Errorf("stringSliceFromMap([]string) = %v, want [a b]", gotStrings)
	}

	gotAny := stringSliceFromMap(map[string]any{"k": []any{"X", 7, " y "}}, "k")
	if len(gotAny) != 2 || gotAny[0] != "x" || gotAny[1] != "y" {
		t.Errorf("stringSliceFromMap([]any) = %v, want [x y]", gotAny)
	}

	gotSecond := stringSliceFromMap(map[string]any{"second": []string{"z"}}, "first", "second")
	if len(gotSecond) != 1 || gotSecond[0] != "z" {
		t.Errorf("stringSliceFromMap(second key) = %v, want [z]", gotSecond)
	}
}

// TestLoaderHelpers_stringSliceFromMap_Bad confirms missing keys, empty slices
// and all-blank entries yield nil.
func TestLoaderHelpers_stringSliceFromMap_Bad(t *testing.T) {
	if got := stringSliceFromMap(map[string]any{"a": 1}, "x"); got != nil {
		t.Errorf("stringSliceFromMap(missing) = %v, want nil", got)
	}
	if got := stringSliceFromMap(map[string]any{"k": []string{"", "  "}}, "k"); got != nil {
		t.Errorf("stringSliceFromMap(all blank) = %v, want nil", got)
	}
	if got := stringSliceFromMap(map[string]any{"k": []any{1, 2}}, "k"); got != nil {
		t.Errorf("stringSliceFromMap(no strings) = %v, want nil", got)
	}
}

// TestLoaderHelpers_validSignalPriorScore covers the NaN/Inf/negative rejection
// guards used by the signal-prior loader.
func TestLoaderHelpers_validSignalPriorScore(t *testing.T) {
	tests := []struct {
		score float64
		want  bool
	}{
		{0, true},
		{1.5, true},
		{-0.1, false},
		{math.NaN(), false},
		{math.Inf(1), false},
		{math.Inf(-1), false},
	}
	for _, tt := range tests {
		if got := validSignalPriorScore(tt.score); got != tt.want {
			t.Errorf("validSignalPriorScore(%v) = %v, want %v", tt.score, got, tt.want)
		}
	}
}
