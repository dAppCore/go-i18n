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
	// NilReceiver covers chain methods on a nil *TranslationContext.
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
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	if err := errorFromResult(SetLanguage("fr")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := C("test").Count(1234)
	if ("1 234") != (ctx.CountString()) {
		t.Fatalf("want %v, got %v", "1 234", ctx.CountString())
	}
}

func TestTranslationContext_SetGet_Good(t *testing.T) {
	// SetGet verifies that Set values are observable through Get.
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
	// FullChain exercises a complete context-building chain.
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

// --- AX-7 canonical triplets ---

func TestContext_C_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("navigation")
		if ctx.Context != "navigation" {
			t.Fatalf("got %q", ctx.Context)
		}
	})
	if !called {
		t.Fatal("C was not exercised")
	}
}

func TestContext_C_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("")
		if ctx.Context != "" {
			t.Fatalf("got %q", ctx.Context)
		}
	})
	if !called {
		t.Fatal("C was not exercised")
	}
}

func TestContext_C_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(0)
		if ctx.CountInt() != 0 {
			t.Fatalf("want 0")
		}
	})
	if !called {
		t.Fatal("C was not exercised")
	}
}

func TestContext_TranslationContext_WithGender_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").WithGender("f")
		if ctx.GenderString() != "f" {
			t.Fatalf("want f")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithGender was not exercised")
	}
}

func TestContext_TranslationContext_WithGender_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.WithGender("f")
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithGender was not exercised")
	}
}

func TestContext_TranslationContext_WithGender_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").WithGender("")
		if ctx.GenderString() != "" {
			t.Fatalf("want empty")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithGender was not exercised")
	}
}

func TestContext_TranslationContext_In_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").In("workspace")
		if ctx.LocationString() != "workspace" {
			t.Fatalf("want workspace")
		}
	})
	if !called {
		t.Fatal("TranslationContext_In was not exercised")
	}
}

func TestContext_TranslationContext_In_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.In("workspace")
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_In was not exercised")
	}
}

func TestContext_TranslationContext_In_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").In("")
		if ctx.LocationString() != "" {
			t.Fatalf("want empty")
		}
	})
	if !called {
		t.Fatal("TranslationContext_In was not exercised")
	}
}

func TestContext_TranslationContext_Formal_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Formal()
		if ctx.FormalityValue() != FormalityFormal {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Formal was not exercised")
	}
}

func TestContext_TranslationContext_Formal_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.Formal()
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Formal was not exercised")
	}
}

func TestContext_TranslationContext_Formal_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Formal().Informal()
		if ctx.FormalityValue() != FormalityInformal {
			t.Fatalf("want informal")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Formal was not exercised")
	}
}

func TestContext_TranslationContext_Informal_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Informal()
		if ctx.FormalityValue() != FormalityInformal {
			t.Fatalf("want informal")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Informal was not exercised")
	}
}

func TestContext_TranslationContext_Informal_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.Informal()
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Informal was not exercised")
	}
}

func TestContext_TranslationContext_Informal_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Informal().Formal()
		if ctx.FormalityValue() != FormalityFormal {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Informal was not exercised")
	}
}

func TestContext_TranslationContext_WithFormality_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").WithFormality(FormalityFormal)
		if ctx.FormalityValue() != FormalityFormal {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithFormality was not exercised")
	}
}

func TestContext_TranslationContext_WithFormality_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.WithFormality(FormalityFormal)
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithFormality was not exercised")
	}
}

func TestContext_TranslationContext_WithFormality_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").WithFormality(FormalityNeutral)
		if ctx.FormalityValue() != FormalityNeutral {
			t.Fatalf("want neutral")
		}
	})
	if !called {
		t.Fatal("TranslationContext_WithFormality was not exercised")
	}
}

func TestContext_TranslationContext_Count_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(3)
		if ctx.CountInt() != 3 {
			t.Fatalf("want 3")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Count was not exercised")
	}
}

func TestContext_TranslationContext_Count_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.Count(3)
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Count was not exercised")
	}
}

func TestContext_TranslationContext_Count_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(0)
		if ctx.CountInt() != 0 {
			t.Fatalf("want 0")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Count was not exercised")
	}
}

func TestContext_TranslationContext_Set_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Set("k", "v")
		if ctx.Get("k") != "v" {
			t.Fatalf("want v")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Set was not exercised")
	}
}

func TestContext_TranslationContext_Set_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var ctx *TranslationContext
		got := ctx.Set("k", "v")
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Set was not exercised")
	}
}

func TestContext_TranslationContext_Set_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Set("", 1)
		if ctx.Get("") != 1 {
			t.Fatalf("want 1")
		}
	})
	if !called {
		t.Fatal("TranslationContext_Set was not exercised")
	}
}

func TestContext_TranslationContext_Get_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Set("k", "v")
		got := ctx.Get("k")
		if got != "v" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_Get was not exercised")
	}
}

func TestContext_TranslationContext_Get_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_Get Bad edge case")
		var ctx *TranslationContext
		got := ctx.Get("k")
		if got != nil {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_Get was not exercised")
	}
}

func TestContext_TranslationContext_Get_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_Get Ugly edge case")
		var ctx *TranslationContext
		got := ctx.Get("k")
		if got != nil {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_Get was not exercised")
	}
}

func TestContext_TranslationContext_ContextString_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("navigation")
		got := ctx.ContextString()
		if got != "navigation" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_ContextString was not exercised")
	}
}

func TestContext_TranslationContext_ContextString_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_ContextString Bad edge case")
		var ctx *TranslationContext
		got := ctx.ContextString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_ContextString was not exercised")
	}
}

func TestContext_TranslationContext_ContextString_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_ContextString Ugly edge case")
		var ctx *TranslationContext
		got := ctx.ContextString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_ContextString was not exercised")
	}
}

func TestContext_TranslationContext_String_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("navigation")
		got := ctx.String()
		if got != "navigation" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_String was not exercised")
	}
}

func TestContext_TranslationContext_String_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_String Bad edge case")
		var ctx *TranslationContext
		got := ctx.String()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_String was not exercised")
	}
}

func TestContext_TranslationContext_String_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_String Ugly edge case")
		var ctx *TranslationContext
		got := ctx.String()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_String was not exercised")
	}
}

func TestContext_TranslationContext_GenderString_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").WithGender("f")
		got := ctx.GenderString()
		if got != "f" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_GenderString was not exercised")
	}
}

func TestContext_TranslationContext_GenderString_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_GenderString Bad edge case")
		var ctx *TranslationContext
		got := ctx.GenderString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_GenderString was not exercised")
	}
}

func TestContext_TranslationContext_GenderString_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_GenderString Ugly edge case")
		var ctx *TranslationContext
		got := ctx.GenderString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_GenderString was not exercised")
	}
}

func TestContext_TranslationContext_LocationString_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").In("workspace")
		got := ctx.LocationString()
		if got != "workspace" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_LocationString was not exercised")
	}
}

func TestContext_TranslationContext_LocationString_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_LocationString Bad edge case")
		var ctx *TranslationContext
		got := ctx.LocationString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_LocationString was not exercised")
	}
}

func TestContext_TranslationContext_LocationString_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_LocationString Ugly edge case")
		var ctx *TranslationContext
		got := ctx.LocationString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_LocationString was not exercised")
	}
}

func TestContext_TranslationContext_FormalityValue_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Formal()
		got := ctx.FormalityValue()
		if got != FormalityFormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_FormalityValue was not exercised")
	}
}

func TestContext_TranslationContext_FormalityValue_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_FormalityValue Bad edge case")
		var ctx *TranslationContext
		got := ctx.FormalityValue()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_FormalityValue was not exercised")
	}
}

func TestContext_TranslationContext_FormalityValue_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_FormalityValue Ugly edge case")
		var ctx *TranslationContext
		got := ctx.FormalityValue()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_FormalityValue was not exercised")
	}
}

func TestContext_TranslationContext_CountInt_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(3)
		got := ctx.CountInt()
		if got != 3 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountInt was not exercised")
	}
}

func TestContext_TranslationContext_CountInt_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_CountInt Bad edge case")
		var ctx *TranslationContext
		got := ctx.CountInt()
		if got != 1 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountInt was not exercised")
	}
}

func TestContext_TranslationContext_CountInt_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_CountInt Ugly edge case")
		var ctx *TranslationContext
		got := ctx.CountInt()
		if got != 1 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountInt was not exercised")
	}
}

func TestContext_TranslationContext_CountString_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(1234)
		got := ctx.CountString()
		if got != "1,234" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountString was not exercised")
	}
}

func TestContext_TranslationContext_CountString_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_CountString Bad edge case")
		var ctx *TranslationContext
		got := ctx.CountString()
		if got != "1" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountString was not exercised")
	}
}

func TestContext_TranslationContext_CountString_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_CountString Ugly edge case")
		var ctx *TranslationContext
		got := ctx.CountString()
		if got != "1" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("TranslationContext_CountString was not exercised")
	}
}

func TestContext_TranslationContext_IsPlural_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		ctx := C("x").Count(2)
		got := ctx.IsPlural()
		if !got {
			t.Fatalf("want plural")
		}
	})
	if !called {
		t.Fatal("TranslationContext_IsPlural was not exercised")
	}
}

func TestContext_TranslationContext_IsPlural_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_IsPlural Bad edge case")
		var ctx *TranslationContext
		got := ctx.IsPlural()
		if got {
			t.Fatalf("nil is not plural")
		}
	})
	if !called {
		t.Fatal("TranslationContext_IsPlural was not exercised")
	}
}

func TestContext_TranslationContext_IsPlural_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		t.Log("TranslationContext_IsPlural Ugly edge case")
		var ctx *TranslationContext
		got := ctx.IsPlural()
		if got {
			t.Fatalf("nil is not plural")
		}
	})
	if !called {
		t.Fatal("TranslationContext_IsPlural was not exercised")
	}
}
