package i18n

import (
	"os"

	"dappco.re/go/core"
	"golang.org/x/text/language"
)

func (f Formality) String() string {
	switch f {
	case FormalityInformal:
		return "informal"
	case FormalityFormal:
		return "formal"
	default:
		return "neutral"
	}
}

func (d TextDirection) String() string {
	if d == DirRTL {
		return "rtl"
	}
	return "ltr"
}

func (p PluralCategory) String() string {
	switch p {
	case PluralZero:
		return "zero"
	case PluralOne:
		return "one"
	case PluralTwo:
		return "two"
	case PluralFew:
		return "few"
	case PluralMany:
		return "many"
	default:
		return "other"
	}
}

func (g GrammaticalGender) String() string {
	switch g {
	case GenderMasculine:
		return "masculine"
	case GenderFeminine:
		return "feminine"
	case GenderCommon:
		return "common"
	default:
		return "neuter"
	}
}

// IsRTLLanguage returns true if the language code uses right-to-left text.
func IsRTLLanguage(lang string) bool {
	lang = normalizeLanguageTag(lang)
	if rtlLanguages[lang] {
		return true
	}
	if len(lang) > 2 {
		return rtlLanguages[lang[:2]]
	}
	return false
}

// SetFormality sets the default formality level on the default service.
func SetFormality(f Formality) {
	if svc := Default(); svc != nil {
		svc.SetFormality(f)
	}
}

// SetLocation sets the default location context on the default service.
func SetLocation(location string) {
	if svc := Default(); svc != nil {
		svc.SetLocation(location)
	}
}

// CurrentLocation returns the current default location context.
func CurrentLocation() string {
	if svc := Default(); svc != nil {
		return svc.Location()
	}
	return ""
}

// Direction returns the text direction for the current language.
func Direction() TextDirection {
	if svc := Default(); svc != nil {
		return svc.Direction()
	}
	return DirLTR
}

// CurrentDirection returns the current default text direction.
func CurrentDirection() TextDirection {
	return Direction()
}

// IsRTL returns true if the current language uses right-to-left text.
func IsRTL() bool { return Direction() == DirRTL }

// CurrentPluralCategory returns the plural category for the current default language.
func CurrentPluralCategory(n int) PluralCategory {
	if svc := Default(); svc != nil {
		return svc.PluralCategory(n)
	}
	return GetPluralCategory("en", n)
}

func detectLanguage(supported []language.Tag) string {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = os.Getenv("LC_ALL")
		if langEnv == "" {
			langEnv = os.Getenv("LC_MESSAGES")
		}
	}
	if langEnv == "" {
		return ""
	}
	baseLang := normalizeLanguageTag(core.Split(langEnv, ".")[0])
	parsedLang, err := language.Parse(baseLang)
	if err != nil {
		return ""
	}
	if len(supported) == 0 {
		return ""
	}
	matcher := language.NewMatcher(supported)
	bestMatch, bestIndex, confidence := matcher.Match(parsedLang)
	if confidence >= language.Low {
		if bestIndex >= 0 && bestIndex < len(supported) {
			return supported[bestIndex].String()
		}
		return bestMatch.String()
	}
	return ""
}

func normalizeLanguageTag(lang string) string {
	lang = core.Trim(lang)
	if lang == "" {
		return ""
	}
	return core.Replace(lang, "_", "-")
}
