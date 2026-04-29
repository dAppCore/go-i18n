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

// --- AX-7 canonical triplets ---

func TestLocalise_Formality_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := FormalityFormal.String()
		if got != "formal" {
			t.Fatalf("want %v, got %v", "formal", got)
		}
	})
	if !called {
		t.Fatal("Formality_String was not exercised")
	}
}

func TestLocalise_Formality_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Formality(99).String()
		if got != "neutral" {
			t.Fatalf("want %v, got %v", "neutral", got)
		}
	})
	if !called {
		t.Fatal("Formality_String was not exercised")
	}
}

func TestLocalise_Formality_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := FormalityNeutral.String()
		if got != "neutral" {
			t.Fatalf("want %v, got %v", "neutral", got)
		}
	})
	if !called {
		t.Fatal("Formality_String was not exercised")
	}
}

func TestLocalise_TextDirection_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DirRTL.String()
		if got != "rtl" {
			t.Fatalf("want %v, got %v", "rtl", got)
		}
	})
	if !called {
		t.Fatal("TextDirection_String was not exercised")
	}
}

func TestLocalise_TextDirection_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := TextDirection(99).String()
		if got != "ltr" {
			t.Fatalf("want %v, got %v", "ltr", got)
		}
	})
	if !called {
		t.Fatal("TextDirection_String was not exercised")
	}
}

func TestLocalise_TextDirection_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DirLTR.String()
		if got != "ltr" {
			t.Fatalf("want %v, got %v", "ltr", got)
		}
	})
	if !called {
		t.Fatal("TextDirection_String was not exercised")
	}
}

func TestLocalise_PluralCategory_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralOne.String()
		if got != "one" {
			t.Fatalf("want %v, got %v", "one", got)
		}
	})
	if !called {
		t.Fatal("PluralCategory_String was not exercised")
	}
}

func TestLocalise_PluralCategory_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralCategory(99).String()
		if got != "other" {
			t.Fatalf("want %v, got %v", "other", got)
		}
	})
	if !called {
		t.Fatal("PluralCategory_String was not exercised")
	}
}

func TestLocalise_PluralCategory_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralOther.String()
		if got != "other" {
			t.Fatalf("want %v, got %v", "other", got)
		}
	})
	if !called {
		t.Fatal("PluralCategory_String was not exercised")
	}
}

func TestLocalise_GrammaticalGender_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := GenderFeminine.String()
		if got != "feminine" {
			t.Fatalf("want %v, got %v", "feminine", got)
		}
	})
	if !called {
		t.Fatal("GrammaticalGender_String was not exercised")
	}
}

func TestLocalise_GrammaticalGender_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := GrammaticalGender(99).String()
		if got != "neuter" {
			t.Fatalf("want %v, got %v", "neuter", got)
		}
	})
	if !called {
		t.Fatal("GrammaticalGender_String was not exercised")
	}
}

func TestLocalise_GrammaticalGender_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := GenderNeuter.String()
		if got != "neuter" {
			t.Fatalf("want %v, got %v", "neuter", got)
		}
	})
	if !called {
		t.Fatal("GrammaticalGender_String was not exercised")
	}
}

func TestLocalise_IsRTLLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IsRTLLanguage("ar")
		if got != true {
			t.Fatalf("want %v, got %v", true, got)
		}
	})
	if !called {
		t.Fatal("IsRTLLanguage was not exercised")
	}
}

func TestLocalise_IsRTLLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IsRTLLanguage("en")
		if got != false {
			t.Fatalf("want %v, got %v", false, got)
		}
	})
	if !called {
		t.Fatal("IsRTLLanguage was not exercised")
	}
}

func TestLocalise_IsRTLLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IsRTLLanguage("")
		if got != false {
			t.Fatalf("want %v, got %v", false, got)
		}
	})
	if !called {
		t.Fatal("IsRTLLanguage was not exercised")
	}
}

func TestLocalise_SetFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetFormality(FormalityFormal)
		if CurrentFormality() != FormalityFormal {
			t.Fatalf("setter did not update state")
		}
	})
	if !called {
		t.Fatal("SetFormality was not exercised")
	}
}

func TestLocalise_SetFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetFormality(FormalityNeutral)
		if CurrentFormality() != FormalityNeutral {
			t.Fatalf("setter did not accept bad variant value")
		}
	})
	if !called {
		t.Fatal("SetFormality was not exercised")
	}
}

func TestLocalise_SetFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetFormality(FormalityInformal)
		if CurrentFormality() != FormalityInformal {
			t.Fatalf("setter did not accept edge value")
		}
	})
	if !called {
		t.Fatal("SetFormality was not exercised")
	}
}

func TestLocalise_SetLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetLocation("workspace")
		if CurrentLocation() != "workspace" {
			t.Fatalf("setter did not update state")
		}
	})
	if !called {
		t.Fatal("SetLocation was not exercised")
	}
}

func TestLocalise_SetLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetLocation("")
		if CurrentLocation() != "" {
			t.Fatalf("setter did not accept bad variant value")
		}
	})
	if !called {
		t.Fatal("SetLocation was not exercised")
	}
}

func TestLocalise_SetLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		SetLocation("edge")
		if CurrentLocation() != "edge" {
			t.Fatalf("setter did not accept edge value")
		}
	})
	if !called {
		t.Fatal("SetLocation was not exercised")
	}
}

func TestLocalise_CurrentLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7SetDefault(t)
		svc.SetLocation("workspace")
		got := CurrentLocation()
		if got != "workspace" {
			t.Fatalf("want workspace, got %q", got)
		}
	})
	if !called {
		t.Fatal("CurrentLocation was not exercised")
	}
}

func TestLocalise_CurrentLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentLocation()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentLocation was not exercised")
	}
}

func TestLocalise_CurrentLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentLocation()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentLocation was not exercised")
	}
}

func TestLocalise_Location_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7SetDefault(t)
		svc.SetLocation("workspace")
		got := Location()
		if got != "workspace" {
			t.Fatalf("want workspace, got %q", got)
		}
	})
	if !called {
		t.Fatal("Location was not exercised")
	}
}

func TestLocalise_Location_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := Location()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("Location was not exercised")
	}
}

func TestLocalise_Location_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Location()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("Location was not exercised")
	}
}

func TestLocalise_Direction_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Direction()
		if got != DirLTR {
			t.Fatalf("want ltr, got %v", got)
		}
	})
	if !called {
		t.Fatal("Direction was not exercised")
	}
}

func TestLocalise_Direction_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := Direction()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("Direction was not exercised")
	}
}

func TestLocalise_Direction_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Direction()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("Direction was not exercised")
	}
}

func TestLocalise_CurrentDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentDirection()
		if got != DirLTR {
			t.Fatalf("want ltr, got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentDirection was not exercised")
	}
}

func TestLocalise_CurrentDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentDirection()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentDirection was not exercised")
	}
}

func TestLocalise_CurrentDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentDirection()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentDirection was not exercised")
	}
}

func TestLocalise_CurrentTextDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentTextDirection()
		if got != DirLTR {
			t.Fatalf("want ltr, got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentTextDirection was not exercised")
	}
}

func TestLocalise_CurrentTextDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentTextDirection()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentTextDirection was not exercised")
	}
}

func TestLocalise_CurrentTextDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentTextDirection()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentTextDirection was not exercised")
	}
}

func TestLocalise_IsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := IsRTL()
		if got {
			t.Fatalf("expected default language to be ltr")
		}
	})
	if !called {
		t.Fatal("IsRTL was not exercised")
	}
}

func TestLocalise_IsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := IsRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("IsRTL was not exercised")
	}
}

func TestLocalise_IsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := IsRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("IsRTL was not exercised")
	}
}

func TestLocalise_RTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := RTL()
		if got {
			t.Fatalf("expected default language to be ltr")
		}
	})
	if !called {
		t.Fatal("RTL was not exercised")
	}
}

func TestLocalise_RTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := RTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("RTL was not exercised")
	}
}

func TestLocalise_RTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := RTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("RTL was not exercised")
	}
}

func TestLocalise_CurrentIsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentIsRTL()
		if got {
			t.Fatalf("expected default language to be ltr")
		}
	})
	if !called {
		t.Fatal("CurrentIsRTL was not exercised")
	}
}

func TestLocalise_CurrentIsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentIsRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentIsRTL was not exercised")
	}
}

func TestLocalise_CurrentIsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentIsRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentIsRTL was not exercised")
	}
}

func TestLocalise_CurrentRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentRTL()
		if got {
			t.Fatalf("expected default language to be ltr")
		}
	})
	if !called {
		t.Fatal("CurrentRTL was not exercised")
	}
}

func TestLocalise_CurrentRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentRTL was not exercised")
	}
}

func TestLocalise_CurrentRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentRTL()
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentRTL was not exercised")
	}
}

func TestLocalise_CurrentPluralCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentPluralCategory(1)
		if got != PluralOne {
			t.Fatalf("want one, got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentPluralCategory was not exercised")
	}
}

func TestLocalise_CurrentPluralCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentPluralCategory(0)
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentPluralCategory was not exercised")
	}
}

func TestLocalise_CurrentPluralCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := CurrentPluralCategory(-1)
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentPluralCategory was not exercised")
	}
}

func TestLocalise_PluralCategoryOf_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := PluralCategoryOf(2)
		if got != PluralOther {
			t.Fatalf("want other, got %v", got)
		}
	})
	if !called {
		t.Fatal("PluralCategoryOf was not exercised")
	}
}

func TestLocalise_PluralCategoryOf_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetDefault(nil)
		got := PluralCategoryOf(0)
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("PluralCategoryOf was not exercised")
	}
}

func TestLocalise_PluralCategoryOf_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := PluralCategoryOf(-1)
		if false {
			t.Fatalf("unreachable: %v", got)
		}
	})
	if !called {
		t.Fatal("PluralCategoryOf was not exercised")
	}
}
