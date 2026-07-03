package i18n

import (
	"strconv"

	"dappco.re/go"
)

func mapValueString(values any, key string) (string, bool) {
	switch m := values.(type) {
	case map[string]any:
		raw, ok := m[key]
		if !ok {
			return "", false
		}
		text := mapRawValueString(raw)
		if text == "" {
			return "", false
		}
		return text, true
	case map[string]string:
		text, ok := m[key]
		if !ok {
			return "", false
		}
		text = core.Trim(text)
		if text == "" {
			return "", false
		}
		return text, true
	default:
		return "", false
	}
}

func mapRawValueString(raw any) string {
	switch v := raw.(type) {
	case string:
		return core.Trim(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	default:
		return core.Trim(core.Sprintf("%v", raw))
	}
}

func contextMapValues(values any) map[string]any {
	switch m := values.(type) {
	case map[string]any:
		return contextMapValuesAny(m)
	case map[string]string:
		return contextMapValuesString(m)
	default:
		return nil
	}
}

func contextMapValuesAny(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	var extra map[string]any
	for key, value := range values {
		switch key {
		case "Context", "Gender", "Location", "Formality", "Count", "IsPlural":
			continue
		case "Extra", "extra", "Extras", "extras":
			if extra == nil {
				extra = make(map[string]any, len(values))
			}
			mergeContextExtra(extra, value)
			continue
		default:
			if extra == nil {
				extra = make(map[string]any, len(values))
			}
			extra[key] = value
		}
	}
	if len(extra) == 0 {
		return nil
	}
	return extra
}

func contextMapValuesString(values map[string]string) map[string]any {
	if len(values) == 0 {
		return nil
	}
	var extra map[string]any
	for key, value := range values {
		switch key {
		case "Context", "Gender", "Location", "Formality", "Count", "IsPlural",
			"Extra", "extra", "Extras", "extras":
			continue
		default:
			if extra == nil {
				extra = make(map[string]any, len(values))
			}
			extra[key] = value
		}
	}
	if len(extra) == 0 {
		return nil
	}
	return extra
}
