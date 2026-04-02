package i18n

import (
	"strconv"

	"dappco.re/go/core"
)

func getCount(data any) int {
	if data == nil {
		return 0
	}
	switch d := data.(type) {
	case *Subject:
		if d == nil {
			return 0
		}
		return d.CountInt()
	case *TranslationContext:
		if d == nil {
			return 0
		}
		if count, ok := d.countValue(); ok {
			return count
		}
		if d.Extra != nil {
			if c, ok := d.Extra["Count"]; ok {
				return toInt(c)
			}
			if c, ok := d.Extra["count"]; ok {
				return toInt(c)
			}
		}
		return d.count
	case map[string]any:
		if c, ok := d["Count"]; ok {
			return toInt(c)
		}
		if c, ok := d["count"]; ok {
			return toInt(c)
		}
	case map[string]int:
		if c, ok := d["Count"]; ok {
			return c
		}
		if c, ok := d["count"]; ok {
			return c
		}
	case map[string]string:
		if c, ok := d["Count"]; ok {
			return toInt(c)
		}
		if c, ok := d["count"]; ok {
			return toInt(c)
		}
	}
	return toInt(data)
}

func toInt(v any) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case int16:
		return int(n)
	case int8:
		return int(n)
	case uint:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	case uint16:
		return int(n)
	case uint8:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	case string:
		if n == "" {
			return 0
		}
		if parsed, err := strconv.Atoi(core.Trim(n)); err == nil {
			return parsed
		}
	}
	return 0
}

func toInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case int32:
		return int64(n)
	case int16:
		return int64(n)
	case int8:
		return int64(n)
	case uint:
		return int64(n)
	case uint64:
		return int64(n)
	case uint32:
		return int64(n)
	case uint16:
		return int64(n)
	case uint8:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	case string:
		if n == "" {
			return 0
		}
		if parsed, err := strconv.ParseInt(core.Trim(n), 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func toFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case int16:
		return float64(n)
	case int8:
		return float64(n)
	case uint:
		return float64(n)
	case uint64:
		return float64(n)
	case uint32:
		return float64(n)
	case uint16:
		return float64(n)
	case uint8:
		return float64(n)
	case string:
		if n == "" {
			return 0
		}
		if parsed, err := strconv.ParseFloat(core.Trim(n), 64); err == nil {
			return parsed
		}
	}
	return 0
}
