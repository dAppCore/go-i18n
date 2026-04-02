package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- C() constructor ---

func TestC_Good(t *testing.T) {
	ctx := C("navigation")
	require.NotNil(t, ctx)
	assert.Equal(t, "navigation", ctx.Context)
	assert.Equal(t, "navigation", ctx.ContextString())
	assert.Equal(t, "navigation", ctx.String())
	assert.Equal(t, 1, ctx.CountInt())
	assert.Equal(t, "1", ctx.CountString())
	assert.False(t, ctx.IsPlural())
}

func TestC_Good_EmptyContext(t *testing.T) {
	ctx := C("")
	require.NotNil(t, ctx)
	assert.Equal(t, "", ctx.ContextString())
}

// --- Nil receiver safety ---

func TestTranslationContext_NilReceiver_Good(t *testing.T) {
	var ctx *TranslationContext

	assert.Nil(t, ctx.Count(2))
	assert.Nil(t, ctx.WithGender("masculine"))
	assert.Nil(t, ctx.In("workspace"))
	assert.Nil(t, ctx.Formal())
	assert.Nil(t, ctx.Informal())
	assert.Nil(t, ctx.WithFormality(FormalityFormal))
	assert.Nil(t, ctx.Set("key", "value"))
	assert.Nil(t, ctx.Get("key"))
	assert.Equal(t, "", ctx.ContextString())
	assert.Equal(t, "", ctx.GenderString())
	assert.Equal(t, "", ctx.LocationString())
	assert.Equal(t, FormalityNeutral, ctx.FormalityValue())
	assert.Equal(t, 1, ctx.CountInt())
	assert.Equal(t, "1", ctx.CountString())
	assert.False(t, ctx.IsPlural())
}

// --- WithGender ---

func TestTranslationContext_WithGender_Good(t *testing.T) {
	ctx := C("test").WithGender("feminine")
	assert.Equal(t, "feminine", ctx.Gender)
	assert.Equal(t, "feminine", ctx.GenderString())
}

func TestTranslationContext_In_Good(t *testing.T) {
	ctx := C("test").In("workspace")
	assert.Equal(t, "workspace", ctx.Location)
	assert.Equal(t, "workspace", ctx.LocationString())
}

func TestTranslationContext_In_Bad_NilReceiver(t *testing.T) {
	var ctx *TranslationContext
	assert.Nil(t, ctx.In("workspace"))
}

// --- Formal / Informal ---

func TestTranslationContext_Formal_Good(t *testing.T) {
	ctx := C("greeting").Formal()
	assert.Equal(t, FormalityFormal, ctx.Formality)
	assert.Equal(t, FormalityFormal, ctx.FormalityValue())
}

func TestTranslationContext_Informal_Good(t *testing.T) {
	ctx := C("greeting").Informal()
	assert.Equal(t, FormalityInformal, ctx.Formality)
	assert.Equal(t, FormalityInformal, ctx.FormalityValue())
}

// --- WithFormality ---

func TestTranslationContext_WithFormality_Good(t *testing.T) {
	tests := []struct {
		name string
		f    Formality
		want Formality
	}{
		{"neutral", FormalityNeutral, FormalityNeutral},
		{"formal", FormalityFormal, FormalityFormal},
		{"informal", FormalityInformal, FormalityInformal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := C("test").WithFormality(tt.f)
			assert.Equal(t, tt.want, ctx.FormalityValue())
		})
	}
}

// --- Set / Get ---

func TestTranslationContext_SetGet_Good(t *testing.T) {
	ctx := C("test").
		Set("region", "europe").
		Set("audience", "developers")

	assert.Equal(t, "europe", ctx.Get("region"))
	assert.Equal(t, "developers", ctx.Get("audience"))
}

func TestTranslationContext_Get_Bad_MissingKey(t *testing.T) {
	ctx := C("test")
	assert.Nil(t, ctx.Get("nonexistent"), "Get on empty Extra should return nil")
}

func TestTranslationContext_Get_Bad_NilExtra(t *testing.T) {
	ctx := &TranslationContext{Context: "test"}
	assert.Nil(t, ctx.Get("anything"), "Get on nil Extra should return nil")
}

// --- Full chaining ---

func TestTranslationContext_FullChain_Good(t *testing.T) {
	ctx := C("medical").
		Count(3).
		WithGender("feminine").
		In("clinic").
		Formal().
		Set("speciality", "cardiology")

	assert.Equal(t, "medical", ctx.ContextString())
	assert.Equal(t, 3, ctx.CountInt())
	assert.Equal(t, "3", ctx.CountString())
	assert.True(t, ctx.IsPlural())
	assert.Equal(t, "feminine", ctx.GenderString())
	assert.Equal(t, "clinic", ctx.LocationString())
	assert.Equal(t, FormalityFormal, ctx.FormalityValue())
	assert.Equal(t, "cardiology", ctx.Get("speciality"))
}
