package i18n

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- TimeAgo ---

func TestTimeAgo_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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
			assert.Contains(t, got, tt.contains)
		})
	}
}

func TestTimeAgo_Good_EdgeCases(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	// Just under 1 minute
	got := TimeAgo(time.Now().Add(-59 * time.Second))
	assert.Contains(t, got, "seconds ago")

	// Exactly 1 minute
	got = TimeAgo(time.Now().Add(-60 * time.Second))
	assert.Contains(t, got, "minute")

	// Just under 1 hour
	got = TimeAgo(time.Now().Add(-59 * time.Minute))
	assert.Contains(t, got, "minutes ago")

	// Just under 1 day
	got = TimeAgo(time.Now().Add(-23 * time.Hour))
	assert.Contains(t, got, "hours ago")

	// Just under 1 week
	got = TimeAgo(time.Now().Add(-6 * 24 * time.Hour))
	assert.Contains(t, got, "days ago")

	// Just over 4 weeks
	got = TimeAgo(time.Now().Add(-31 * 24 * time.Hour))
	assert.Contains(t, got, "month ago")

	// Well over a year
	got = TimeAgo(time.Now().Add(-800 * 24 * time.Hour))
	assert.Contains(t, got, "years ago")
}

func TestTimeAgo_Good_SingleUnits(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	// 1 minute ago
	got := TimeAgo(time.Now().Add(-1 * time.Minute))
	assert.Contains(t, got, "1 minute ago")

	// 1 hour ago
	got = TimeAgo(time.Now().Add(-1 * time.Hour))
	assert.Contains(t, got, "1 hour ago")

	// 1 day ago
	got = TimeAgo(time.Now().Add(-24 * time.Hour))
	assert.Contains(t, got, "1 day ago")

	// 1 week ago
	got = TimeAgo(time.Now().Add(-7 * 24 * time.Hour))
	assert.Contains(t, got, "1 week ago")
}

// --- FormatAgo ---

func TestFormatAgo_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatAgo_Good_PluralUnitAlias(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	got := FormatAgo(5, "minutes")
	assert.Equal(t, "5 minutes ago", got)
}

func TestFormatAgo_Good_MorePluralUnitAliases(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatAgo_Bad_UnknownUnit(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	// Unknown unit should fallback to programmatic format
	got := FormatAgo(5, "fortnight")
	assert.Equal(t, "5 fortnights ago", got)
}

func TestFormatAgo_Good_SingularUnit(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	got := FormatAgo(1, "fortnight")
	assert.Equal(t, "1 fortnight ago", got)
}

func TestFormatAgo_Good_FrenchRelativeTime(t *testing.T) {
	prev := Default()
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	require.NoError(t, SetLanguage("fr"))

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
			assert.Equal(t, tt.want, got)
		})
	}
}
