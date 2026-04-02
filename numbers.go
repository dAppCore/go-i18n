package i18n

import (
	"math"
	"strconv"

	"dappco.re/go/core"
)

func getNumberFormat() NumberFormat {
	lang := currentLangForGrammar()
	if fmt, ok := getLocaleNumberFormat(lang); ok {
		return fmt
	}
	if idx := indexAny(lang, "-_"); idx > 0 {
		if fmt, ok := getLocaleNumberFormat(lang[:idx]); ok {
			return fmt
		}
	}
	return numberFormats["en"]
}

func getLocaleNumberFormat(lang string) (NumberFormat, bool) {
	if data := GetGrammarData(lang); data != nil && data.Number != (NumberFormat{}) {
		return data.Number, true
	}
	if fmt, ok := numberFormats[lang]; ok {
		return fmt, true
	}
	return NumberFormat{}, false
}

// FormatNumber formats an integer with locale-specific thousands separators.
func FormatNumber(n int64) string {
	return formatIntWithSep(n, getNumberFormat().ThousandsSep)
}

// FormatDecimal formats a float with locale-specific separators.
func FormatDecimal(f float64) string {
	return FormatDecimalN(f, 2)
}

// FormatDecimalN formats a float with N decimal places.
func FormatDecimalN(f float64, decimals int) string {
	nf := getNumberFormat()
	negative := f < 0
	absVal := math.Abs(f)
	intPart := int64(absVal)
	fracPart := absVal - float64(intPart)
	intStr := formatIntWithSep(intPart, nf.ThousandsSep)
	if decimals <= 0 || fracPart == 0 {
		if negative {
			return "-" + intStr
		}
		return intStr
	}
	multiplier := math.Pow(10, float64(decimals))
	fracInt := int64(math.Round(fracPart * multiplier))
	if fracInt == 0 {
		if negative {
			return "-" + intStr
		}
		return intStr
	}
	fracStr := core.Sprintf("%0*d", decimals, fracInt)
	fracStr = trimRight(fracStr, "0")
	if negative {
		return "-" + intStr + nf.DecimalSep + fracStr
	}
	return intStr + nf.DecimalSep + fracStr
}

// FormatPercent formats a decimal as a percentage.
func FormatPercent(f float64) string {
	nf := getNumberFormat()
	pct := f * 100
	var numStr string
	if pct == float64(int64(pct)) {
		numStr = strconv.FormatInt(int64(pct), 10)
	} else {
		numStr = FormatDecimalN(pct, 1)
	}
	return core.Sprintf(nf.PercentFmt, numStr)
}

// FormatBytes formats bytes as human-readable size.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	nf := getNumberFormat()
	var value float64
	var unit string
	switch {
	case bytes >= TB:
		value = float64(bytes) / TB
		unit = "TB"
	case bytes >= GB:
		value = float64(bytes) / GB
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / MB
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / KB
		unit = "KB"
	default:
		return core.Sprintf("%d B", bytes)
	}
	intPart := int64(value)
	fracPart := value - float64(intPart)
	if fracPart < 0.05 {
		return core.Sprintf("%d %s", intPart, unit)
	}
	fracDigit := int(math.Round(fracPart * 10))
	if fracDigit == 10 {
		return core.Sprintf("%d %s", intPart+1, unit)
	}
	return core.Sprintf("%d%s%d %s", intPart, nf.DecimalSep, fracDigit, unit)
}

// FormatOrdinal formats a number as an ordinal.
func FormatOrdinal(n int) string {
	lang := currentLangForGrammar()
	if idx := indexAny(lang, "-_"); idx > 0 {
		lang = lang[:idx]
	}
	switch lang {
	case "en":
		return formatEnglishOrdinal(n)
	default:
		return core.Sprintf("%d.", n)
	}
}

func formatEnglishOrdinal(n int) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	if abs%100 >= 11 && abs%100 <= 13 {
		return core.Sprintf("%dth", n)
	}
	switch abs % 10 {
	case 1:
		return core.Sprintf("%dst", n)
	case 2:
		return core.Sprintf("%dnd", n)
	case 3:
		return core.Sprintf("%drd", n)
	default:
		return core.Sprintf("%dth", n)
	}
}

func formatIntWithSep(n int64, sep string) string {
	if sep == "" {
		return strconv.FormatInt(n, 10)
	}
	negative := n < 0
	if negative {
		n = -n
	}
	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		if negative {
			return "-" + str
		}
		return str
	}
	result := core.NewBuilder()
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(sep)
		}
		result.WriteRune(c)
	}
	if negative {
		return "-" + result.String()
	}
	return result.String()
}

// indexAny returns the index of the first occurrence of any char in chars, or -1.
func indexAny(s, chars string) int {
	for i, c := range s {
		for _, ch := range chars {
			if c == ch {
				return i
			}
		}
	}
	return -1
}

// trimRight returns s with all trailing occurrences of cutset removed.
func trimRight(s, cutset string) string {
	for len(s) > 0 {
		found := false
		r := rune(s[len(s)-1])
		for _, c := range cutset {
			if r == c {
				found = true
				break
			}
		}
		if !found {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}
