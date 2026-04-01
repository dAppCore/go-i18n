package i18n

import (
	"time"

	"dappco.re/go/core"
)

// TimeAgo returns a localised relative time string.
//
//	TimeAgo(time.Now().Add(-4 * time.Second)) // "just now"
//	TimeAgo(time.Now().Add(-5 * time.Minute)) // "5 minutes ago"
func TimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < 0 {
		duration = 0
	}
	switch {
	case duration < 5*time.Second:
		return T("time.just_now")
	case duration < time.Minute:
		return FormatAgo(int(duration/time.Second), "second")
	case duration < time.Hour:
		return FormatAgo(int(duration.Minutes()), "minute")
	case duration < 24*time.Hour:
		return FormatAgo(int(duration.Hours()), "hour")
	case duration < 7*24*time.Hour:
		return FormatAgo(int(duration.Hours()/24), "day")
	case duration < 30*24*time.Hour:
		return FormatAgo(int(duration.Hours()/(24*7)), "week")
	case duration < 365*24*time.Hour:
		return FormatAgo(int(duration.Hours()/(24*30)), "month")
	default:
		return FormatAgo(int(duration.Hours()/(24*365)), "year")
	}
}

// FormatAgo formats "N unit ago" with proper pluralisation.
func FormatAgo(count int, unit string) string {
	svc := Default()
	unit = normalizeAgoUnit(unit)
	if svc == nil {
		return core.Sprintf("%d %ss ago", count, unit)
	}
	key := "time.ago." + unit
	result := svc.T(key, map[string]any{"Count": count})
	if result == key {
		return core.Sprintf("%d %s ago", count, Pluralize(unit, count))
	}
	return result
}

func normalizeAgoUnit(unit string) string {
	switch unit {
	case "seconds":
		return "second"
	case "minutes":
		return "minute"
	case "hours":
		return "hour"
	case "days":
		return "day"
	case "weeks":
		return "week"
	case "months":
		return "month"
	case "years":
		return "year"
	default:
		return unit
	}
}
