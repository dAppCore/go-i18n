package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

// --- Formality.String() ---

func TestFormality_String_Good(t *testing.T) {
	tests := []struct {
		name string
		f    Formality
		want string
	}{
		{"neutral", FormalityNeutral, "neutral"},
		{"informal", FormalityInformal, "informal"},
		{"formal", FormalityFormal, "formal"},
		{"unknown", Formality(99), "neutral"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.want) != (tt.f.String()) {
				t.Fatalf("want %v, got %v", tt.want, tt.f.String())
			}
		})
	}
}

// --- TextDirection.String() ---

func TestTextDirection_String_Good(t *testing.T) {
	if ("ltr") != (DirLTR.String()) {
		t.Fatalf("want %v, got %v", "ltr", DirLTR.String())
	}
	if ("rtl") != (DirRTL.String()) {
		t.Fatalf("want %v, got %v", "rtl", DirRTL.String())
	}
}

func TestPluralCategory_String_Good(t *testing.T) {
	tests := []struct {
		name string
		cat  PluralCategory
		want string
	}{
		{"zero", PluralZero, "zero"},
		{"one", PluralOne, "one"},
		{"two", PluralTwo, "two"},
		{"few", PluralFew, "few"},
		{"many", PluralMany, "many"},
		{"other", PluralOther, "other"},
		{"unknown", PluralCategory(99), "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.want) != (tt.cat.String()) {
				t.Fatalf("want %v, got %v", tt.want, tt.cat.String())
			}
		})
	}
}

// --- GrammaticalGender.String() ---

func TestGrammaticalGender_String_Good(t *testing.T) {
	tests := []struct {
		name string
		g    GrammaticalGender
		want string
	}{
		{"neuter", GenderNeuter, "neuter"},
		{"masculine", GenderMasculine, "masculine"},
		{"feminine", GenderFeminine, "feminine"},
		{"common", GenderCommon, "common"},
		{"unknown", GrammaticalGender(99), "neuter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.want) != (tt.g.String()) {
				t.Fatalf("want %v, got %v", tt.want, tt.g.String())
			}
		})
	}
}

// --- IsRTLLanguage ---

func TestIsRTLLanguage_Good(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		{"arabic", "ar", true},
		{"arabic_sa", "ar-SA", true},
		{"arabic_sa_underscore", "ar_EG", true},
		{"hebrew", "he", true},
		{"farsi", "fa", true},
		{"urdu", "ur", true},
		{"english", "en", false},
		{"german", "de", false},
		{"french", "fr", false},
		{"unknown", "xx", false},
		{"arabic_variant", "ar-EG-extra", true},   // len > 2 prefix check
		{"english_variant", "en-US-extra", false}, // len > 2, not RTL
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.want) != (IsRTLLanguage(tt.lang)) {
				t.Fatalf("want %v, got %v", tt.want, IsRTLLanguage(tt.lang))
			}
		})
	}
}

// --- Package-level SetFormality ---

func TestSetFormality_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	SetFormality(FormalityFormal)
	if (FormalityFormal) != (svc.Formality()) {
		t.Fatalf("want %v, got %v", FormalityFormal, svc.Formality())
	}

	SetFormality(FormalityNeutral)
	if (FormalityNeutral) != (svc.Formality()) {
		t.Fatalf("want %v, got %v", FormalityNeutral, svc.Formality())
	}
}

func TestSetFallback_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	SetFallback("fr")
	if ("fr") != (svc.Fallback()) {
		t.Fatalf("want %v, got %v", "fr", svc.Fallback())
	}

	SetFallback("en")
	if ("en") != (svc.Fallback()) {
		t.Fatalf("want %v, got %v", "en", svc.Fallback())
	}
}

func TestCurrentFormality_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (FormalityNeutral) != (CurrentFormality()) {
		t.Fatalf("want %v, got %v", FormalityNeutral, CurrentFormality())
	}

	SetFormality(FormalityFormal)
	if (FormalityFormal) != (CurrentFormality()) {
		t.Fatalf("want %v, got %v", FormalityFormal, CurrentFormality())
	}
}

func TestCurrentFallback_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if ("en") != (CurrentFallback()) {
		t.Fatalf("want %v, got %v", "en", CurrentFallback())
	}

	SetFallback("fr")
	if ("fr") != (CurrentFallback()) {
		t.Fatalf("want %v, got %v", "fr", CurrentFallback())
	}
}

func TestSetLocation_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	SetLocation("workspace")
	if ("workspace") != (svc.Location()) {
		t.Fatalf("want %v, got %v", "workspace", svc.Location())
	}

	SetLocation("")
	if ("") != (svc.Location()) {
		t.Fatalf("want %v, got %v", "", svc.Location())
	}
}

func TestCurrentLocation_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if ("") != (CurrentLocation()) {
		t.Fatalf("want %v, got %v", "", CurrentLocation())
	}

	SetLocation("workspace")
	if ("workspace") != (CurrentLocation()) {
		t.Fatalf("want %v, got %v", "workspace", CurrentLocation())
	}
}

func TestLocation_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (CurrentLocation()) != (Location()) {
		t.Fatalf("want %v, got %v", CurrentLocation(), Location())
	}

	SetLocation("workspace")
	if (CurrentLocation()) != (Location()) {
		t.Fatalf("want %v, got %v", CurrentLocation(), Location())
	}
}

func TestDirection_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	dir := Direction()
	if (DirLTR) != (dir) {
		t.Fatalf("want %v, got %v", DirLTR, dir)
	}
}

func TestCurrentDirection_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (DirLTR) != (CurrentDirection()) {
		t.Fatalf("want %v, got %v", DirLTR, CurrentDirection())
	}
}

func TestCurrentTextDirection_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (CurrentDirection()) != (CurrentTextDirection()) {
		t.Fatalf("want %v, got %v", CurrentDirection(), CurrentTextDirection())
	}
}

func TestIsRTL_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if IsRTL() {
		t.Fatal("expected false")
	}
}

// --- Package-level RTL ---

func TestRTL_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (IsRTL()) != (RTL()) {
		t.Fatalf("want %v, got %v", IsRTL(), RTL())
	}
}

func TestCurrentIsRTL_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if CurrentIsRTL() {
		t.Fatal("expected false")
	}
}

// --- Package-level CurrentRTL ---

func TestCurrentRTL_Good(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (CurrentIsRTL()) != (CurrentRTL()) {
		t.Fatalf("want %v, got %v", CurrentIsRTL(), CurrentRTL())
	}
}

func TestCurrentPluralCategory_Good(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (PluralOther) != (CurrentPluralCategory(0)) {
		t.Fatalf("want %v, got %v", PluralOther, CurrentPluralCategory(0))
	}
	if (PluralOne) != (CurrentPluralCategory(1)) {
		t.Fatalf("want %v, got %v", PluralOne, CurrentPluralCategory(1))
	}
	if (PluralOther) != (CurrentPluralCategory(2)) {
		t.Fatalf("want %v, got %v", PluralOther, CurrentPluralCategory(2))
	}
	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (PluralOne) != (CurrentPluralCategory(0)) {
		t.Fatalf("want %v, got %v", PluralOne, CurrentPluralCategory(0))
	}
	if (PluralOne) != (CurrentPluralCategory(1)) {
		t.Fatalf("want %v, got %v", PluralOne, CurrentPluralCategory(1))
	}
	if (PluralOther) != (CurrentPluralCategory(2)) {
		t.Fatalf("want %v, got %v", PluralOther, CurrentPluralCategory(2))
	}
}

func TestPluralCategoryOf_Good(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if (PluralOther) != (PluralCategoryOf(0)) {
		t.Fatalf("want %v, got %v", PluralOther, PluralCategoryOf(0))
	}
	if (PluralOne) != (PluralCategoryOf(1)) {
		t.Fatalf("want %v, got %v", PluralOne, PluralCategoryOf(1))
	}
	if (PluralOther) != (PluralCategoryOf(2)) {
		t.Fatalf("want %v, got %v", PluralOther, PluralCategoryOf(2))
	}
	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (PluralOne) != (PluralCategoryOf(0)) {
		t.Fatalf("want %v, got %v", PluralOne, PluralCategoryOf(0))
	}
	if (PluralOne) != (PluralCategoryOf(1)) {
		t.Fatalf("want %v, got %v", PluralOne, PluralCategoryOf(1))
	}
	if (PluralOther) != (PluralCategoryOf(2)) {
		t.Fatalf("want %v, got %v", PluralOther, PluralCategoryOf(2))
	}
}

func TestCurrentPluralCategory_NoDefaultService(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	SetDefault(nil)
	if (PluralOther) != (CurrentPluralCategory(2)) {
		t.Fatalf("want %v, got %v", PluralOther, CurrentPluralCategory(2))
	}
}

func TestDetectLanguage_Good(t *testing.T) {
	// detectLanguage relies on env vars, which we can't easily set in tests
	// but we can test with no supported languages
	result := detectLanguage(nil)
	if ("") != (result) {
		t.Fatalf("want %v, got %v", "", result)
	}
}

func TestDetectLanguage_PrefersLocaleOverrides(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_MESSAGES", "fr_FR.UTF-8")
	t.Setenv("LC_ALL", "de_DE.UTF-8")

	supported := []language.Tag{
		language.AmericanEnglish,
		language.French,
		language.German,
	}

	result := detectLanguage(supported)
	if ("de") != (result) {
		t.Fatalf("want %v, got %v", "de", result)
	}
}

func TestDetectLanguage_SkipsInvalidHigherPriorityLocale(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_MESSAGES", "fr_FR.UTF-8")
	t.Setenv("LC_ALL", "not-a-locale")

	supported := []language.Tag{
		language.AmericanEnglish,
		language.French,
	}

	result := detectLanguage(supported)
	if ("fr") != (result) {
		t.Fatalf("want %v, got %v", "fr", result)
	}
}

func TestDetectLanguage_PrefersLanguageList(t *testing.T) {
	t.Setenv("LANGUAGE", "fr_FR.UTF-8:de_DE.UTF-8")
	t.Setenv("LANG", "en_US.UTF-8")

	supported := []language.Tag{
		language.AmericanEnglish,
		language.French,
		language.German,
	}

	result := detectLanguage(supported)
	if ("fr") != (result) {
		t.Fatalf("want %v, got %v", "fr", result)
	}
}

// --- Mode.String() ---

func TestMode_String_Good(t *testing.T) {
	tests := []struct {
		name string
		m    Mode
		want string
	}{
		{"normal", ModeNormal, "normal"},
		{"strict", ModeStrict, "strict"},
		{"collect", ModeCollect, "collect"},
		{"unknown", Mode(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.want) != (tt.m.String()) {
				t.Fatalf("want %v, got %v", tt.want, tt.m.String())
			}
		})
	}
}
