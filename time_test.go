package i18n

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// --- TimeAgo ---

func TestTimeAgo_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"just_now", 4 * time.Second, "just now"},
		{"seconds_ago", 5 * time.Second, "5 seconds ago"},
		{"minutes_ago", 5 * time.Minute, "5 minutes ago"},
		{"hours_ago", 3 * time.Hour, "3 hours ago"},
		{"days_ago", 2 * 24 * time.Hour, "2 days ago"},
		{"weeks_ago", 3 * 7 * 24 * time.Hour, "3 weeks ago"},
		{"months_ago", 40 * 24 * time.Hour, "1 month ago"},
		{"years_ago", 400 * 24 * time.Hour, "1 year ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeAgo(time.Now().Add(-tt.duration))
			if !strings.Contains(got, tt.contains) {
				t.Fatalf("expected %q to contain %q", got, tt.contains)
			}
		})
	}
}

func TestTimeAgo_Good_EdgeCases(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	// Just under 1 minute
	got := TimeAgo(time.Now().Add(-59 * time.Second))
	if !strings.Contains(got, "seconds ago") {
		t.Fatalf("expected %q to contain %q", got, "seconds ago")
	}

	got = TimeAgo(time.Now().Add(-60 * time.Second))
	if !strings.Contains(got, "minute") {
		t.Fatalf("expected %q to contain %q", got, "minute")
	}

	got = TimeAgo(time.Now().Add(-59 * time.Minute))
	if !strings.Contains(got, "minutes ago") {
		t.Fatalf("expected %q to contain %q", got, "minutes ago")
	}

	got = TimeAgo(time.Now().Add(-23 * time.Hour))
	if !strings.Contains(got, "hours ago") {
		t.Fatalf("expected %q to contain %q", got, "hours ago")
	}

	got = TimeAgo(time.Now().Add(-6 * 24 * time.Hour))
	if !strings.Contains(got, "days ago") {
		t.Fatalf("expected %q to contain %q", got, "days ago")
	}

	got = TimeAgo(time.Now().Add(-31 * 24 * time.Hour))
	if !strings.Contains(got, "month ago") {
		t.Fatalf("expected %q to contain %q", got, "month ago")
	}

	got = TimeAgo(time.Now().Add(-800 * 24 * time.Hour))
	if !strings.Contains(got, "years ago") {
		t.Fatalf("expected %q to contain %q", got, "years ago")
	}
}

func TestTimeAgo_Good_SingleUnits(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	// 1 minute ago
	got := TimeAgo(time.Now().Add(-1 * time.Minute))
	if !strings.Contains(got, "1 minute ago") {
		t.Fatalf("expected %q to contain %q", got, "1 minute ago")
	}

	got = TimeAgo(time.Now().Add(-1 * time.Hour))
	if !strings.Contains(got, "1 hour ago") {
		t.Fatalf("expected %q to contain %q", got, "1 hour ago")
	}

	got = TimeAgo(time.Now().Add(-24 * time.Hour))
	if !strings.Contains(got, "1 day ago") {
		t.Fatalf("expected %q to contain %q", got, "1 day ago")
	}

	got = TimeAgo(time.Now().Add(-7 * 24 * time.Hour))
	if !strings.Contains(got, "1 week ago") {
		t.Fatalf("expected %q to contain %q", got, "1 week ago")
	}
}

func TestTimeAgo_Good_MissingJustNowKeyFallback(t *testing.T) {
	svc, err := NewWithFS(fstest.MapFS{
		"xx.json": &fstest.MapFile{
			Data: []byte(`{}`),
		},
	}, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	got := TimeAgo(time.Now().Add(-4 * time.Second))
	if "just now" != got {
		t.Fatalf("want %v, got %v", "just now", got)
	}
}

func TestFormatAgo_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	tests := []struct {
		name  string
		count int
		unit  string
		want  string
	}{
		{"1_second", 1, "second", "1 second ago"},
		{"5_seconds", 5, "second", "5 seconds ago"},
		{"1_minute", 1, "minute", "1 minute ago"},
		{"30_minutes", 30, "minute", "30 minutes ago"},
		{"1_hour", 1, "hour", "1 hour ago"},
		{"12_hours", 12, "hour", "12 hours ago"},
		{"1_day", 1, "day", "1 day ago"},
		{"7_days", 7, "day", "7 days ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAgo(tt.count, tt.unit)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatAgo_Good_PluralUnitAlias(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	got := FormatAgo(5, "minutes")
	if "5 minutes ago" != got {
		t.Fatalf("want %v, got %v", "5 minutes ago", got)
	}
}

func TestFormatAgo_Good_MorePluralUnitAliases(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	tests := []struct {
		name  string
		count int
		unit  string
		want  string
	}{
		{"months", 3, "months", "3 months ago"},
		{"year", 1, "years", "1 year ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAgo(tt.count, tt.unit)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatAgo_Good_NormalisesUnitInput(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	got := FormatAgo(2, " Hours ")
	if "2 hours ago" != got {
		t.Fatalf("want %v, got %v", "2 hours ago", got)
	}
}

func TestFormatAgo_Bad_UnknownUnit(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	// Unknown unit should fallback to programmatic format
	got := FormatAgo(5, "fortnight")
	if "5 fortnights ago" != got {
		t.Fatalf("want %v, got %v", "5 fortnights ago", got)
	}
}

func TestFormatAgo_Good_SingularUnit(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	got := FormatAgo(1, "fortnight")
	if "1 fortnight ago" != got {
		t.Fatalf("want %v, got %v", "1 fortnight ago", got)
	}
}

func TestFormatAgo_Good_NoDefaultService(t *testing.T) {
	prev := Default()
	SetDefault(nil)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	got := FormatAgo(1, "second")
	if "1 second ago" != got {
		t.Fatalf("want %v, got %v", "1 second ago", got)
	}

	got = FormatAgo(5, "second")
	if "5 seconds ago" != got {
		t.Fatalf("want %v, got %v", "5 seconds ago", got)
	}
}

func TestFormatAgo_Good_FrenchRelativeTime(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		count int
		unit  string
		want  string
	}{
		{"month", 1, "month", "il y a 1 mois"},
		{"months", 3, "month", "il y a 3 mois"},
		{"year", 1, "year", "il y a 1 an"},
		{"years", 4, "year", "il y a 4 ans"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAgo(tt.count, tt.unit)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFormatAgo_FallsBackToLocaleWordMap(t *testing.T) {
	prev := Default()
	svc, err := NewWithFS(fstest.MapFS{
		"en.json": &fstest.MapFile{
			Data: []byte(`{
				"gram": {
					"word": {
						"month": "mois"
					}
				}
			}`),
		},
	}, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	if err := SetLanguage("en"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := FormatAgo(2, "month")
	if "2 mois ago" != got {
		t.Fatalf("want %v, got %v", "2 mois ago", got)
	}
}
