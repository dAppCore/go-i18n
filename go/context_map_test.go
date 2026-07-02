package i18n

import "testing"

func TestMapValueString_MapAnyStringValueDoesNotAllocate(t *testing.T) {
	values := map[string]any{"Subject": " config.yaml "}

	allocs := testing.AllocsPerRun(1000, func() {
		text, ok := mapValueString(values, "Subject")
		if !ok || text != "config.yaml" {
			t.Fatalf("mapValueString() = %q, %v; want %q, true", text, ok, "config.yaml")
		}
	})
	if allocs != 0 {
		t.Fatalf("mapValueString() allocated %.0f times for a string value; want zero", allocs)
	}
}

func TestMapValueStringBranches(t *testing.T) {
	if text, ok := mapValueString(map[string]any{"Subject": 42}, "Subject"); !ok || text != "42" {
		t.Fatalf("mapValueString(map[string]any int) = %q, %v; want 42, true", text, ok)
	}
	if text, ok := mapValueString(map[string]string{"Subject": " config.yaml "}, "Subject"); !ok || text != "config.yaml" {
		t.Fatalf("mapValueString(map[string]string) = %q, %v; want config.yaml, true", text, ok)
	}
	if text, ok := mapValueString(map[string]any{"Subject": "  "}, "Subject"); ok || text != "" {
		t.Fatalf("mapValueString(empty map[string]any value) = %q, %v; want empty, false", text, ok)
	}
	if text, ok := mapValueString(map[string]string{"Subject": "  "}, "Subject"); ok || text != "" {
		t.Fatalf("mapValueString(empty map[string]string value) = %q, %v; want empty, false", text, ok)
	}
	if text, ok := mapValueString(map[string]any{}, "Subject"); ok || text != "" {
		t.Fatalf("mapValueString(missing map[string]any value) = %q, %v; want empty, false", text, ok)
	}
	if text, ok := mapValueString(map[string]string{}, "Subject"); ok || text != "" {
		t.Fatalf("mapValueString(missing map[string]string value) = %q, %v; want empty, false", text, ok)
	}
	if text, ok := mapValueString("invalid", "Subject"); ok || text != "" {
		t.Fatalf("mapValueString(invalid values) = %q, %v; want empty, false", text, ok)
	}
}

func TestMapRawValueStringScalars(t *testing.T) {
	tests := []struct {
		name string
		raw  any
		want string
	}{
		{"string", " value ", "value"},
		{"int", int(-3), "-3"},
		{"int8", int8(-4), "-4"},
		{"int16", int16(-5), "-5"},
		{"int32", int32(-6), "-6"},
		{"int64", int64(-7), "-7"},
		{"uint", uint(3), "3"},
		{"uint8", uint8(4), "4"},
		{"uint16", uint16(5), "5"},
		{"uint32", uint32(6), "6"},
		{"uint64", uint64(7), "7"},
		{"bool", true, "true"},
		{"float32", float32(1.25), "1.25"},
		{"float64", float64(2.5), "2.5"},
		{"fallback", []string{"value"}, "[value]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapRawValueString(tt.raw); got != tt.want {
				t.Fatalf("mapRawValueString(%T) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestContextMapValuesBranches(t *testing.T) {
	if got := contextMapValues("invalid"); got != nil {
		t.Fatalf("contextMapValues(invalid) = %#v; want nil", got)
	}
	if got := contextMapValues(map[string]any{"Context": "dashboard", "tier": "gold"}); got["tier"] != "gold" {
		t.Fatalf("contextMapValues(map[string]any) = %#v; want tier extra", got)
	}
	if got := contextMapValues(map[string]string{"Context": "dashboard", "tier": "gold"}); got["tier"] != "gold" {
		t.Fatalf("contextMapValues(map[string]string) = %#v; want tier extra", got)
	}
	if got := contextMapValuesAny(nil); got != nil {
		t.Fatalf("contextMapValuesAny(nil) = %#v; want nil", got)
	}
	if got := contextMapValuesString(nil); got != nil {
		t.Fatalf("contextMapValuesString(nil) = %#v; want nil", got)
	}
}

func TestContextMapValuesReservedOnlyDoesNotAllocate(t *testing.T) {
	anyValues := map[string]any{"Count": 2, "Context": "dashboard"}
	stringValues := map[string]string{"Count": "2", "Context": "dashboard"}

	allocs := testing.AllocsPerRun(1000, func() {
		if got := contextMapValuesAny(anyValues); got != nil {
			t.Fatalf("contextMapValuesAny() = %#v; want nil", got)
		}
		if got := contextMapValuesString(stringValues); got != nil {
			t.Fatalf("contextMapValuesString() = %#v; want nil", got)
		}
	})
	if allocs != 0 {
		t.Fatalf("contextMapValues reserved-only path allocated %.0f times; want zero", allocs)
	}
}
