package i18n

import (
	"testing"
)

// --- C() constructor ---

func TestC_Good(t *testing.T) {
	ctx := C("navigation")
	if (ctx) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if ("navigation") != (ctx.Context) {
		t.Fatalf("want %v, got %v", "navigation", ctx.Context)
	}
	if ("navigation") != (ctx.ContextString()) {
		t.Fatalf("want %v, got %v", "navigation", ctx.ContextString())
	}
	if ("navigation") != (ctx.String()) {
		t.Fatalf("want %v, got %v", "navigation", ctx.String())
	}
	if (1) != (ctx.CountInt()) {
		t.Fatalf("want %v, got %v", 1, ctx.CountInt())
	}
	if ("1") != (ctx.CountString()) {
		t.Fatalf("want %v, got %v", "1", ctx.CountString())
	}
	if ctx.IsPlural() {
		t.Fatal("expected false")
	}
}

func TestC_Good_EmptyContext(t *testing.T) {
	ctx := C("")
	if (ctx) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if ("") != (ctx.ContextString()) {
		t.Fatalf("want %v, got %v", "", ctx.ContextString())
	}
}

func TestTranslationContext_NilReceiver_Good(t *testing.T) {
	var ctx *TranslationContext
	if (ctx.Count(2)) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Count(2))
	}
	if (ctx.WithGender("masculine")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.WithGender("masculine"))
	}
	if (ctx.In("workspace")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.In("workspace"))
	}
	if (ctx.Formal()) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Formal())
	}
	if (ctx.Informal()) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Informal())
	}
	if (ctx.WithFormality(FormalityFormal)) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.WithFormality(FormalityFormal))
	}
	if (ctx.Set("key", "value")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Set("key", "value"))
	}
	if (ctx.Get("key")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Get("key"))
	}
	if ("") != (ctx.ContextString()) {
		t.Fatalf("want %v, got %v", "", ctx.ContextString())
	}
	if ("") != (ctx.GenderString()) {
		t.Fatalf("want %v, got %v", "", ctx.GenderString())
	}
	if ("") != (ctx.LocationString()) {
		t.Fatalf("want %v, got %v", "", ctx.LocationString())
	}
	if (FormalityNeutral) != (ctx.FormalityValue()) {
		t.Fatalf("want %v, got %v", FormalityNeutral, ctx.FormalityValue())
	}
	if (1) != (ctx.CountInt()) {
		t.Fatalf("want %v, got %v", 1, ctx.CountInt())
	}
	if ("1") != (ctx.CountString()) {
		t.Fatalf("want %v, got %v", "1", ctx.CountString())
	}
	if ctx.IsPlural() {
		t.Fatal("expected false")
	}
}

func TestTranslationContext_WithGender_Good(t *testing.T) {
	ctx := C("test").WithGender("feminine")
	if ("feminine") != (ctx.Gender) {
		t.Fatalf("want %v, got %v", "feminine", ctx.Gender)
	}
	if ("feminine") != (ctx.GenderString()) {
		t.Fatalf("want %v, got %v", "feminine", ctx.GenderString())
	}
}

func TestTranslationContext_In_Good(t *testing.T) {
	ctx := C("test").In("workspace")
	if ("workspace") != (ctx.Location) {
		t.Fatalf("want %v, got %v", "workspace", ctx.Location)
	}
	if ("workspace") != (ctx.LocationString()) {
		t.Fatalf("want %v, got %v", "workspace", ctx.LocationString())
	}
}

func TestTranslationContext_In_Bad_NilReceiver(t *testing.T) {
	var ctx *TranslationContext
	if (ctx.In("workspace")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.In("workspace"))
	}
}

// --- Formal / Informal ---

func TestTranslationContext_Formal_Good(t *testing.T) {
	ctx := C("greeting").Formal()
	if (FormalityFormal) != (ctx.Formality) {
		t.Fatalf("want %v, got %v", FormalityFormal, ctx.Formality)
	}
	if (FormalityFormal) != (ctx.FormalityValue()) {
		t.Fatalf("want %v, got %v", FormalityFormal, ctx.FormalityValue())
	}
}

func TestTranslationContext_Informal_Good(t *testing.T) {
	ctx := C("greeting").Informal()
	if (FormalityInformal) != (ctx.Formality) {
		t.Fatalf("want %v, got %v", FormalityInformal, ctx.Formality)
	}
	if (FormalityInformal) != (ctx.FormalityValue()) {
		t.Fatalf("want %v, got %v", FormalityInformal, ctx.FormalityValue())
	}
}

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
			if (tt.want) != (ctx.FormalityValue()) {
				t.Fatalf("want %v, got %v", tt.want, ctx.FormalityValue())
			}
		})
	}
}

func TestTranslationContext_CountString_UsesLocaleFormatting(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := C("test").Count(1234)
	if ("1 234") != (ctx.CountString()) {
		t.Fatalf("want %v, got %v", "1 234", ctx.CountString())
	}
}

func TestTranslationContext_SetGet_Good(t *testing.T) {
	ctx := C("test").
		Set("region", "europe").
		Set("audience", "developers")
	if ("europe") != (ctx.Get("region")) {
		t.Fatalf("want %v, got %v", "europe", ctx.Get("region"))
	}
	if ("developers") != (ctx.Get("audience")) {
		t.Fatalf("want %v, got %v", "developers", ctx.Get("audience"))
	}
}

func TestTranslationContext_Get_Bad_MissingKey(t *testing.T) {
	ctx := C("test")
	if (ctx.Get("nonexistent")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Get("nonexistent"))
	}
}

func TestTranslationContext_Get_Bad_NilExtra(t *testing.T) {
	ctx := &TranslationContext{Context: "test"}
	if (ctx.Get("anything")) != (nil) {
		t.Fatalf("expected nil, got %v", ctx.Get("anything"))
	}
}

// --- Full chaining ---

func TestTranslationContext_FullChain_Good(t *testing.T) {
	ctx := C("medical").
		Count(3).
		WithGender("feminine").
		In("clinic").
		Formal().
		Set("speciality", "cardiology")
	if ("medical") != (ctx.ContextString()) {
		t.Fatalf("want %v, got %v", "medical", ctx.ContextString())
	}
	if (3) != (ctx.CountInt()) {
		t.Fatalf("want %v, got %v", 3, ctx.CountInt())
	}
	if ("3") != (ctx.CountString()) {
		t.Fatalf("want %v, got %v", "3", ctx.CountString())
	}
	if !(ctx.IsPlural()) {
		t.Fatal("expected true")
	}
	if ("feminine") != (ctx.GenderString()) {
		t.Fatalf("want %v, got %v", "feminine", ctx.GenderString())
	}
	if ("clinic") != (ctx.LocationString()) {
		t.Fatalf("want %v, got %v", "clinic", ctx.LocationString())
	}
	if (FormalityFormal) != (ctx.FormalityValue()) {
		t.Fatalf("want %v, got %v", FormalityFormal, ctx.FormalityValue())
	}
	if ("cardiology") != (ctx.Get("speciality")) {
		t.Fatalf("want %v, got %v", "cardiology", ctx.Get("speciality"))
	}
}
