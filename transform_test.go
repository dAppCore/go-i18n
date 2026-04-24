package i18n

import (
	"math"
	"testing"
)

// --- getCount ---

func TestGetCount_Good(t *testing.T) {
	tests := []struct {
		name string
		data any
		want int
	}{
		{"nil", nil, 0},
		{"map_string_any", map[string]any{"Count": 5}, 5},
		{"map_string_any_float", map[string]any{"Count": 3.7}, 3},
		{"map_string_int", map[string]int{"Count": 42}, 42},
		{"map_string_string", map[string]string{"Count": "9"}, 9},
		{"no_count_key", map[string]any{"Name": "test"}, 0},
		{"wrong_type", "a string", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCount(tt.data)
			if (tt.want) != (got) {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestGetCount_Good_TranslationContextDefault(t *testing.T) {
	ctx := C("test")
	if (1) != (getCount(ctx)) {
		t.Fatalf("want %v, got %v", 1, getCount(ctx))
	}
}

func TestGetCount_Good_TranslationContextExtraCount(t *testing.T) {
	ctx := C("test").Set("Count", 3)
	if (3) != (getCount(ctx)) {
		t.Fatalf("want %v, got %v", 3, getCount(ctx))
	}
}

func TestToInt_Good(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want int
	}{
		{"nil", nil, 0},
		{"int", int(42), 42},
		{"int64", int64(100), 100},
		{"int32", int32(200), 200},
		{"int16", int16(300), 300},
		{"int8", int8(127), 127},
		{"uint", uint(10), 10},
		{"uint64", uint64(20), 20},
		{"uint32", uint32(30), 30},
		{"uint16", uint16(40), 40},
		{"uint8", uint8(50), 50},
		{"float64", float64(3.14), 3},
		{"float32", float32(2.71), 2},
		{"string_int", "123", 123},
		{"string", "not a number", 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInt(tt.val)
			if (tt.want) != (got) {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

// --- toInt64 ---

func TestToInt64_Good(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want int64
	}{
		{"nil", nil, 0},
		{"int", int(42), 42},
		{"int64", int64(100), 100},
		{"int32", int32(200), 200},
		{"int16", int16(300), 300},
		{"int8", int8(127), 127},
		{"uint", uint(10), 10},
		{"uint64", uint64(20), 20},
		{"uint32", uint32(30), 30},
		{"uint16", uint16(40), 40},
		{"uint8", uint8(50), 50},
		{"float64", float64(3.14), 3},
		{"float32", float32(2.71), 2},
		{"string_int64", "123", 123},
		{"string", "not a number", 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInt64(tt.val)
			if (tt.want) != (got) {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

// --- toFloat64 ---

func TestToFloat64_Good(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want float64
	}{
		{"nil", nil, 0},
		{"float64", float64(3.14), 3.14},
		{"float32", float32(2.5), 2.5},
		{"int", int(42), 42.0},
		{"int64", int64(100), 100.0},
		{"int32", int32(200), 200.0},
		{"int16", int16(300), 300.0},
		{"int8", int8(127), 127.0},
		{"uint", uint(10), 10.0},
		{"uint64", uint64(20), 20.0},
		{"uint32", uint32(30), 30.0},
		{"uint16", uint16(40), 40.0},
		{"uint8", uint8(50), 50.0},
		{"string_float", "3.5", 3.5},
		{"string", "not a number", 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFloat64(tt.val)
			if math.Abs((tt.want)-(got)) > 0.01 {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}
