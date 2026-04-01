package i18n

// GetPluralRule returns the plural rule for a language code.
func GetPluralRule(lang string) PluralRule {
	if rule, ok := pluralRules[lang]; ok {
		return rule
	}
	if len(lang) > 2 {
		base := lang[:2]
		if rule, ok := pluralRules[base]; ok {
			return rule
		}
	}
	return pluralRuleEnglish
}

// GetPluralCategory returns the plural category for a count in the given language.
func GetPluralCategory(lang string, n int) PluralCategory {
	return GetPluralRule(lang)(n)
}

func pluralRuleEnglish(n int) PluralCategory {
	if n == 1 {
		return PluralOne
	}
	return PluralOther
}

func pluralRuleGerman(n int) PluralCategory  { return pluralRuleEnglish(n) }
func pluralRuleSpanish(n int) PluralCategory { return pluralRuleEnglish(n) }

func pluralRuleFrench(n int) PluralCategory {
	if n == 0 || n == 1 {
		return PluralOne
	}
	return PluralOther
}

func pluralRuleRussian(n int) PluralCategory {
	mod10 := n % 10
	mod100 := n % 100
	if mod10 == 1 && mod100 != 11 {
		return PluralOne
	}
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return PluralFew
	}
	return PluralMany
}

func pluralRulePolish(n int) PluralCategory {
	if n == 1 {
		return PluralOne
	}
	mod10 := n % 10
	mod100 := n % 100
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return PluralFew
	}
	return PluralMany
}

func pluralRuleArabic(n int) PluralCategory {
	if n == 0 {
		return PluralZero
	}
	if n == 1 {
		return PluralOne
	}
	if n == 2 {
		return PluralTwo
	}
	mod100 := n % 100
	if mod100 >= 3 && mod100 <= 10 {
		return PluralFew
	}
	if mod100 >= 11 && mod100 <= 99 {
		return PluralMany
	}
	return PluralOther
}

func pluralRuleChinese(n int) PluralCategory  { return PluralOther }
func pluralRuleJapanese(n int) PluralCategory { return PluralOther }
func pluralRuleKorean(n int) PluralCategory   { return PluralOther }

func pluralRuleWelsh(n int) PluralCategory {
	switch n {
	case 0:
		return PluralZero
	case 1:
		return PluralOne
	case 2:
		return PluralTwo
	case 3:
		return PluralFew
	case 6:
		return PluralMany
	default:
		return PluralOther
	}
}
