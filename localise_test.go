package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.want, tt.f.String())
		})
	}
}

// --- TextDirection.String() ---

func TestTextDirection_String_Good(t *testing.T) {
	assert.Equal(t, "ltr", DirLTR.String())
	assert.Equal(t, "rtl", DirRTL.String())
}

// --- PluralCategory.String() ---

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
			assert.Equal(t, tt.want, tt.cat.String())
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
			assert.Equal(t, tt.want, tt.g.String())
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
			assert.Equal(t, tt.want, IsRTLLanguage(tt.lang))
		})
	}
}

// --- Package-level SetFormality ---

func TestSetFormality_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	SetFormality(FormalityFormal)
	assert.Equal(t, FormalityFormal, svc.Formality())

	SetFormality(FormalityNeutral)
	assert.Equal(t, FormalityNeutral, svc.Formality())
}

// --- Package-level SetFallback ---

func TestSetFallback_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	SetFallback("fr")
	assert.Equal(t, "fr", svc.Fallback())

	SetFallback("en")
	assert.Equal(t, "en", svc.Fallback())
}

// --- Package-level CurrentFormality ---

func TestCurrentFormality_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	assert.Equal(t, FormalityNeutral, CurrentFormality())

	SetFormality(FormalityFormal)
	assert.Equal(t, FormalityFormal, CurrentFormality())
}

// --- Package-level CurrentFallback ---

func TestCurrentFallback_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	assert.Equal(t, "en", CurrentFallback())

	SetFallback("fr")
	assert.Equal(t, "fr", CurrentFallback())
}

// --- Package-level SetLocation ---

func TestSetLocation_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	SetLocation("workspace")
	assert.Equal(t, "workspace", svc.Location())

	SetLocation("")
	assert.Equal(t, "", svc.Location())
}

// --- Package-level CurrentLocation ---

func TestCurrentLocation_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	assert.Equal(t, "", CurrentLocation())

	SetLocation("workspace")
	assert.Equal(t, "workspace", CurrentLocation())
}

// --- Package-level Direction ---

func TestDirection_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	dir := Direction()
	assert.Equal(t, DirLTR, dir)
}

// --- Package-level CurrentDirection ---

func TestCurrentDirection_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	assert.Equal(t, DirLTR, CurrentDirection())
}

// --- Package-level IsRTL ---

func TestIsRTL_Good(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	assert.False(t, IsRTL(), "English should not be RTL")
}

// --- detectLanguage ---

func TestDetectLanguage_Good(t *testing.T) {
	// detectLanguage relies on env vars, which we can't easily set in tests
	// but we can test with no supported languages
	result := detectLanguage(nil)
	assert.Equal(t, "", result, "should return empty with no supported languages")
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
			assert.Equal(t, tt.want, tt.m.String())
		})
	}
}
