package i18n

import "testing"

func TestGetPluralCategory(t *testing.T) {
	tests := []struct {
		lang string
		n    int
		want PluralCategory
	}{
		// English
		{"en", 0, PluralOther},
		{"en", 1, PluralOne},
		{"en", 2, PluralOther},
		{"en_US", 1, PluralOne},

		// French (0 and 1 are singular)
		{"fr", 0, PluralOne},
		{"fr", 1, PluralOne},
		{"fr", 2, PluralOther},
		{"fr_CA", 2, PluralOther},

		// Russian
		{"ru", 1, PluralOne},
		{"ru", 2, PluralFew},
		{"ru", 5, PluralMany},
		{"ru", 11, PluralMany},
		{"ru", 21, PluralOne},
		{"ru", 22, PluralFew},

		// Polish
		{"pl", 1, PluralOne},
		{"pl", 2, PluralFew},
		{"pl", 5, PluralMany},

		// Arabic
		{"ar", 0, PluralZero},
		{"ar", 1, PluralOne},
		{"ar", 2, PluralTwo},
		{"ar", 5, PluralFew},
		{"ar", 11, PluralMany},
		{"ar", 100, PluralOther},

		// Welsh
		{"cy", 0, PluralZero},
		{"cy", 1, PluralOne},
		{"cy", 2, PluralTwo},
		{"cy", 3, PluralFew},
		{"cy", 6, PluralMany},
		{"cy", 7, PluralOther},

		// Chinese (always other)
		{"zh", 0, PluralOther},
		{"zh", 1, PluralOther},
		{"zh", 100, PluralOther},

		// Fallback for unknown language uses English rules
		{"xx", 1, PluralOne},
		{"xx", 5, PluralOther},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := GetPluralCategory(tt.lang, tt.n)
			if got != tt.want {
				t.Errorf("GetPluralCategory(%q, %d) = %v, want %v", tt.lang, tt.n, got, tt.want)
			}
		})
	}
}

func TestGetPluralRule(t *testing.T) {
	// Known language
	rule := GetPluralRule("en")
	if rule == nil {
		t.Fatal("GetPluralRule(en) returned nil")
	}
	if rule(1) != PluralOne {
		t.Error("English rule(1) should be PluralOne")
	}

	// Base language extraction
	rule = GetPluralRule("en-US")
	if rule(1) != PluralOne {
		t.Error("English-US rule(1) should be PluralOne")
	}

	rule = GetPluralRule("cy-GB")
	if rule(2) != PluralTwo {
		t.Error("Welsh-GB rule(2) should be PluralTwo")
	}

	rule = GetPluralRule("en_US")
	if rule(1) != PluralOne {
		t.Error("English_US rule(1) should be PluralOne")
	}

	// Unknown falls back to English
	rule = GetPluralRule("xx-YY")
	if rule(1) != PluralOne {
		t.Error("Unknown rule(1) should fallback to English PluralOne")
	}
}
