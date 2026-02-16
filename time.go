package i18n

import (
	"fmt"
	"time"
)

// TimeAgo returns a localised relative time string.
//
//	TimeAgo(time.Now().Add(-5 * time.Minute)) // "5 minutes ago"
func TimeAgo(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return T("time.just_now")
	case duration < time.Hour:
		return FormatAgo(int(duration.Minutes()), "minute")
	case duration < 24*time.Hour:
		return FormatAgo(int(duration.Hours()), "hour")
	case duration < 7*24*time.Hour:
		return FormatAgo(int(duration.Hours()/24), "day")
	default:
		return FormatAgo(int(duration.Hours()/(24*7)), "week")
	}
}

// FormatAgo formats "N unit ago" with proper pluralisation.
func FormatAgo(count int, unit string) string {
	svc := Default()
	if svc == nil {
		return fmt.Sprintf("%d %ss ago", count, unit)
	}
	key := "time.ago." + unit
	result := svc.T(key, map[string]any{"Count": count})
	if result == key {
		return fmt.Sprintf("%d %s ago", count, Pluralize(unit, count))
	}
	return result
}
