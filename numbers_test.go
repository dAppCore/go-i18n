package i18n

import (
	"math"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	// Ensure service is initialised for English locale
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{100, "100"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{-1234567, "-1,234,567"},
		{1000000000, "1,000,000,000"},
	}

	for _, tt := range tests {
		got := FormatNumber(tt.n)
		if got != tt.want {
			t.Errorf("FormatNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatNumber_MinInt64(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	got := FormatNumber(math.MinInt64)
	want := "-9,223,372,036,854,775,808"
	if got != want {
		t.Fatalf("FormatNumber(math.MinInt64) = %q, want %q", got, want)
	}
}

func TestFormatDecimal(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		f    float64
		want string
	}{
		{1.5, "1.5"},
		{1.0, "1"},
		{1.995, "2"},
		{9.999, "10"},
		{1234.56, "1,234.56"},
		{0.1, "0.1"},
		{-0.1, "-0.1"},
		{-1234.56, "-1,234.56"},
	}

	for _, tt := range tests {
		got := FormatDecimal(tt.f)
		if got != tt.want {
			t.Errorf("FormatDecimal(%v) = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatDecimalN_RoundsCarry(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		f        float64
		decimals int
		want     string
	}{
		{1.995, 2, "2"},
		{9.999, 2, "10"},
		{999.999, 2, "1,000"},
	}

	for _, tt := range tests {
		got := FormatDecimalN(tt.f, tt.decimals)
		if got != tt.want {
			t.Errorf("FormatDecimalN(%v, %d) = %q, want %q", tt.f, tt.decimals, got, tt.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		f    float64
		want string
	}{
		{0.85, "85%"},
		{1.0, "100%"},
		{0.0, "0%"},
		{0.333, "33.3%"},
		{-0.1, "-10%"},
	}

	for _, tt := range tests {
		got := FormatPercent(tt.f)
		if got != tt.want {
			t.Errorf("FormatPercent(%v) = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{1536000, "1.46 MB"},
		{1073741824, "1 GB"},
		{1099511627776, "1 TB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatOrdinal(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		n    int
		want string
	}{
		{1, "1st"},
		{2, "2nd"},
		{3, "3rd"},
		{4, "4th"},
		{11, "11th"},
		{12, "12th"},
		{13, "13th"},
		{21, "21st"},
		{22, "22nd"},
		{23, "23rd"},
		{100, "100th"},
		{101, "101st"},
		{111, "111th"},
	}

	for _, tt := range tests {
		got := FormatOrdinal(tt.n)
		if got != tt.want {
			t.Errorf("FormatOrdinal(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatOrdinalFromLocale(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	tests := []struct {
		n    int
		want string
	}{
		{1, "1er"},
		{2, "2e"},
		{3, "3e"},
		{11, "11e"},
	}

	for _, tt := range tests {
		got := FormatOrdinal(tt.n)
		if got != tt.want {
			t.Errorf("FormatOrdinal(fr, %d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatNumberFromLocale(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	if got := FormatNumber(1234567); got != "1 234 567" {
		t.Errorf("FormatNumber(fr) = %q, want %q", got, "1 234 567")
	}
	if got := FormatDecimal(1234.56); got != "1 234,56" {
		t.Errorf("FormatDecimal(fr) = %q, want %q", got, "1 234,56")
	}
	if got := FormatPercent(0.85); got != "85 %" {
		t.Errorf("FormatPercent(fr) = %q, want %q", got, "85 %")
	}
	if got := FormatDecimal(-0.1); got != "-0,1" {
		t.Errorf("FormatDecimal(fr, negative) = %q, want %q", got, "-0,1")
	}
}
