package i18n

import "dappco.re/go/core"

func mapValueString(values any, key string) (string, bool) {
	switch m := values.(type) {
	case map[string]any:
		raw, ok := m[key]
		if !ok {
			return "", false
		}
		text := core.Trim(core.Sprintf("%v", raw))
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
	extra := make(map[string]any, len(values))
	for key, value := range values {
		switch key {
		case "Context", "Gender", "Location", "Formality":
			continue
		case "Extra", "extra", "Extras", "extras":
			mergeContextExtra(extra, value)
			continue
		default:
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
	extra := make(map[string]any, len(values))
	for key, value := range values {
		switch key {
		case "Context", "Gender", "Location", "Formality", "Extra", "extra", "Extras", "extras":
			continue
		default:
			extra[key] = value
		}
	}
	if len(extra) == 0 {
		return nil
	}
	return extra
}
