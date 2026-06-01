package i18n

import "testing"

// TestLanguageRules_PluralRuleGerman_Good confirms German resolves through the
// pluralRules map and mirrors the English one/other split.
//
//	GetPluralCategory("de", 1) // PluralOne
//	GetPluralCategory("de", 2) // PluralOther
func TestLanguageRules_PluralRuleGerman_Good(t *testing.T) {
	tests := []struct {
		lang string
		n    int
		want PluralCategory
	}{
		{"de", 1, PluralOne},
		{"de", 2, PluralOther},
		{"de", 0, PluralOther},
		{"de-DE", 1, PluralOne},
		{"de-AT", 5, PluralOther},
		{"de-CH", 1, PluralOne},
	}
	for _, tt := range tests {
		if got := GetPluralCategory(tt.lang, tt.n); got != tt.want {
			t.Errorf("GetPluralCategory(%q, %d) = %v, want %v", tt.lang, tt.n, got, tt.want)
		}
	}
}

// TestLanguageRules_PluralRuleSpanish_Good confirms Spanish mirrors the English
// one/other split through every registered region tag.
func TestLanguageRules_PluralRuleSpanish_Good(t *testing.T) {
	tests := []struct {
		lang string
		n    int
		want PluralCategory
	}{
		{"es", 1, PluralOne},
		{"es", 2, PluralOther},
		{"es", 0, PluralOther},
		{"es-ES", 1, PluralOne},
		{"es-MX", 3, PluralOther},
	}
	for _, tt := range tests {
		if got := GetPluralCategory(tt.lang, tt.n); got != tt.want {
			t.Errorf("GetPluralCategory(%q, %d) = %v, want %v", tt.lang, tt.n, got, tt.want)
		}
	}
}

// TestLanguageRules_PluralRuleJapanese_Good confirms Japanese is invariant —
// every count maps to PluralOther.
func TestLanguageRules_PluralRuleJapanese_Good(t *testing.T) {
	for _, n := range []int{0, 1, 2, 5, 100} {
		if got := GetPluralCategory("ja", n); got != PluralOther {
			t.Errorf("GetPluralCategory(ja, %d) = %v, want PluralOther", n, got)
		}
	}
	if got := GetPluralCategory("ja-JP", 1); got != PluralOther {
		t.Errorf("GetPluralCategory(ja-JP, 1) = %v, want PluralOther", got)
	}
}

// TestLanguageRules_PluralRuleKorean_Good confirms Korean is invariant — every
// count maps to PluralOther.
func TestLanguageRules_PluralRuleKorean_Good(t *testing.T) {
	for _, n := range []int{0, 1, 2, 5, 100} {
		if got := GetPluralCategory("ko", n); got != PluralOther {
			t.Errorf("GetPluralCategory(ko, %d) = %v, want PluralOther", n, got)
		}
	}
	if got := GetPluralCategory("ko-KR", 2); got != PluralOther {
		t.Errorf("GetPluralCategory(ko-KR, 2) = %v, want PluralOther", got)
	}
}

// TestLanguageRules_PluralRuleChinese_Good confirms every Chinese region tag is
// invariant — count never changes the category.
func TestLanguageRules_PluralRuleChinese_Good(t *testing.T) {
	for _, lang := range []string{"zh", "zh-CN", "zh-TW"} {
		for _, n := range []int{0, 1, 2, 100} {
			if got := GetPluralCategory(lang, n); got != PluralOther {
				t.Errorf("GetPluralCategory(%q, %d) = %v, want PluralOther", lang, n, got)
			}
		}
	}
}

// TestLanguageRules_PluralRule_Ugly drives the invariant-language rules through
// negative and extreme counts; they must never panic and always return
// PluralOther.
func TestLanguageRules_PluralRule_Ugly(t *testing.T) {
	for _, lang := range []string{"ja", "ko", "zh"} {
		for _, n := range []int{-1, -100, 1 << 30} {
			noPanicForAudit(t, func() {
				if got := GetPluralCategory(lang, n); got != PluralOther {
					t.Errorf("GetPluralCategory(%q, %d) = %v, want PluralOther", lang, n, got)
				}
			})
		}
	}
}
