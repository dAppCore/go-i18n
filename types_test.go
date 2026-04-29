package i18n

import "testing"

func TestMessageForCategory(t *testing.T) {
	msg := Message{
		One:   "1 item",
		Other: "{{.Count}} items",
		Zero:  "no items",
	}

	if got := msg.ForCategory(PluralOne); got != "1 item" {
		t.Errorf("ForCategory(One) = %q, want '1 item'", got)
	}
	if got := msg.ForCategory(PluralOther); got != "{{.Count}} items" {
		t.Errorf("ForCategory(Other) = %q, want template", got)
	}
	if got := msg.ForCategory(PluralZero); got != "no items" {
		t.Errorf("ForCategory(Zero) = %q, want 'no items'", got)
	}

	// Falls back to Other when category not set
	msg2 := Message{Other: "many"}
	if got := msg2.ForCategory(PluralFew); got != "many" {
		t.Errorf("ForCategory(Few) fallback = %q, want 'many'", got)
	}

	// Falls back to One if Other not set
	msg3 := Message{One: "single"}
	if got := msg3.ForCategory(PluralFew); got != "single" {
		t.Errorf("ForCategory(Few) fallback to One = %q, want 'single'", got)
	}

	// Falls back to Text
	msg4 := Message{Text: "default"}
	if got := msg4.ForCategory(PluralOther); got != "default" {
		t.Errorf("ForCategory fallback to Text = %q, want 'default'", got)
	}
}

func TestMessageIsPlural(t *testing.T) {
	if (Message{Text: "hello"}).IsPlural() {
		t.Error("simple message should not be plural")
	}
	if !(Message{One: "a", Other: "b"}).IsPlural() {
		t.Error("message with One+Other should be plural")
	}
	if !(Message{Zero: "none"}).IsPlural() {
		t.Error("message with Zero should be plural")
	}
}

func TestSubjectFluent(t *testing.T) {
	s := S("file", "config.yaml")
	if s.Noun != "file" {
		t.Errorf("Noun = %q, want 'file'", s.Noun)
	}
	if s.String() != "config.yaml" {
		t.Errorf("String() = %q, want 'config.yaml'", s.String())
	}
	if s.CountInt() != 1 {
		t.Errorf("CountInt() = %d, want 1", s.CountInt())
	}
	if s.IsPlural() {
		t.Error("count=1 should not be plural")
	}

	// Chain
	s.Count(3).Gender("neuter").In("workspace").Formal()
	if s.CountInt() != 3 {
		t.Errorf("CountInt() = %d, want 3", s.CountInt())
	}
	if !s.IsPlural() {
		t.Error("count=3 should be plural")
	}
	if s.GenderString() != "neuter" {
		t.Errorf("GenderString() = %q, want 'neuter'", s.GenderString())
	}
	if s.LocationString() != "workspace" {
		t.Errorf("LocationString() = %q, want 'workspace'", s.LocationString())
	}
	if !s.IsFormal() {
		t.Error("should be formal")
	}
}

func TestSubjectNil(t *testing.T) {
	var s *Subject
	if s.Count(3) != nil {
		t.Error("nil.Count() should return nil")
	}
	if s.Gender("m") != nil {
		t.Error("nil.Gender() should return nil")
	}
	if s.In("x") != nil {
		t.Error("nil.In() should return nil")
	}
	if s.Formal() != nil {
		t.Error("nil.Formal() should return nil")
	}
	if s.Informal() != nil {
		t.Error("nil.Informal() should return nil")
	}
	if s.String() != "" {
		t.Error("nil.String() should be empty")
	}
	if s.IsPlural() {
		t.Error("nil.IsPlural() should be false")
	}
}

func TestTranslationContext(t *testing.T) {
	c := C("navigation")
	if c.ContextString() != "navigation" {
		t.Errorf("Context = %q, want 'navigation'", c.ContextString())
	}

	c.WithGender("masculine").Formal()
	if c.GenderString() != "masculine" {
		t.Errorf("Gender = %q, want 'masculine'", c.GenderString())
	}
	if c.FormalityValue() != FormalityFormal {
		t.Error("should be formal")
	}

	c.Set("key", "value")
	if c.Get("key") != "value" {
		t.Errorf("Get(key) = %v, want 'value'", c.Get("key"))
	}
}

func TestTranslationContextNil(t *testing.T) {
	var c *TranslationContext
	if c.WithGender("m") != nil {
		t.Error("nil.WithGender() should return nil")
	}
	if c.Formal() != nil {
		t.Error("nil.Formal() should return nil")
	}
	if c.Informal() != nil {
		t.Error("nil.Informal() should return nil")
	}
	if c.WithFormality(FormalityFormal) != nil {
		t.Error("nil.WithFormality() should return nil")
	}
	if c.Set("k", "v") != nil {
		t.Error("nil.Set() should return nil")
	}
	if c.Get("k") != nil {
		t.Error("nil.Get() should return nil")
	}
	if c.ContextString() != "" {
		t.Error("nil.ContextString() should be empty")
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		m    Mode
		want string
	}{
		{ModeNormal, "normal"},
		{ModeStrict, "strict"},
		{ModeCollect, "collect"},
		{Mode(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.m, got, tt.want)
		}
	}
}

func TestFormalityString(t *testing.T) {
	tests := []struct {
		f    Formality
		want string
	}{
		{FormalityNeutral, "neutral"},
		{FormalityInformal, "informal"},
		{FormalityFormal, "formal"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Formality(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestTextDirectionString(t *testing.T) {
	if DirLTR.String() != "ltr" {
		t.Errorf("DirLTR.String() = %q", DirLTR.String())
	}
	if DirRTL.String() != "rtl" {
		t.Errorf("DirRTL.String() = %q", DirRTL.String())
	}
}

func TestPluralCategoryString(t *testing.T) {
	tests := []struct {
		p    PluralCategory
		want string
	}{
		{PluralOther, "other"},
		{PluralZero, "zero"},
		{PluralOne, "one"},
		{PluralTwo, "two"},
		{PluralFew, "few"},
		{PluralMany, "many"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("PluralCategory(%d).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestGrammaticalGenderString(t *testing.T) {
	tests := []struct {
		g    GrammaticalGender
		want string
	}{
		{GenderNeuter, "neuter"},
		{GenderMasculine, "masculine"},
		{GenderFeminine, "feminine"},
		{GenderCommon, "common"},
	}
	for _, tt := range tests {
		if got := tt.g.String(); got != tt.want {
			t.Errorf("GrammaticalGender(%d).String() = %q, want %q", tt.g, got, tt.want)
		}
	}
}

func TestIsRTLLanguage(t *testing.T) {
	tests := []struct {
		lang string
		want bool
	}{
		{"en", false},
		{"de", false},
		{"ar", true},
		{"ar-SA", true},
		{"he", true},
		{"fa", true},
		{"ur-PK", true},
	}
	for _, tt := range tests {
		if got := IsRTLLanguage(tt.lang); got != tt.want {
			t.Errorf("IsRTLLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
		}
	}
}

// --- AX-7 canonical triplets ---

func TestTypes_Mode_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ModeStrict.String()
		if got != "strict" {
			t.Fatalf("want %v, got %v", "strict", got)
		}
	})
	if !called {
		t.Fatal("Mode_String was not exercised")
	}
}

func TestTypes_Mode_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Mode(99).String()
		if got != "unknown" {
			t.Fatalf("want %v, got %v", "unknown", got)
		}
	})
	if !called {
		t.Fatal("Mode_String was not exercised")
	}
}

func TestTypes_Mode_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Mode(-1).String()
		if got != "unknown" {
			t.Fatalf("want %v, got %v", "unknown", got)
		}
	})
	if !called {
		t.Fatal("Mode_String was not exercised")
	}
}

func TestTypes_Message_ForCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		msg := Message{One: "one", Other: "other"}
		got := msg.ForCategory(PluralOne)
		if got != "one" {
			t.Fatalf("want one, got %q", got)
		}
	})
	if !called {
		t.Fatal("Message_ForCategory was not exercised")
	}
}

func TestTypes_Message_ForCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		msg := Message{Text: "plain"}
		got := msg.ForCategory(PluralMany)
		if got != "plain" {
			t.Fatalf("want plain, got %q", got)
		}
	})
	if !called {
		t.Fatal("Message_ForCategory was not exercised")
	}
}

func TestTypes_Message_ForCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		msg := Message{}
		got := msg.ForCategory(PluralZero)
		if got != "" {
			t.Fatalf("want empty, got %q", got)
		}
	})
	if !called {
		t.Fatal("Message_ForCategory was not exercised")
	}
}

func TestTypes_Message_IsPlural_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Message{One: "one"}.IsPlural()
		if got != true {
			t.Fatalf("want %v, got %v", true, got)
		}
	})
	if !called {
		t.Fatal("Message_IsPlural was not exercised")
	}
}

func TestTypes_Message_IsPlural_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Message{Text: "plain"}.IsPlural()
		if got != false {
			t.Fatalf("want %v, got %v", false, got)
		}
	})
	if !called {
		t.Fatal("Message_IsPlural was not exercised")
	}
}

func TestTypes_Message_IsPlural_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Message{}.IsPlural()
		if got != false {
			t.Fatalf("want %v, got %v", false, got)
		}
	})
	if !called {
		t.Fatal("Message_IsPlural was not exercised")
	}
}
