package i18n

import (
	"testing"

	"dappco.re/go/core"
)

// --- S() constructor ---

func TestS_Good(t *testing.T) {
	subj := S("file", "config.yaml")
	if (subj) == (nil) {
		t.Fatalf("expected non-nil")
	}
	if ("file") != (subj.Noun) {
		t.Fatalf("want %v, got %v", "file", subj.Noun)
	}
	if ("config.yaml") != (subj.Value) {
		t.Fatalf("want %v, got %v", "config.yaml", subj.Value)
	}
	if (1) != (subj.CountInt()) {
		t.Fatalf("want %v, got %v", 1, subj.CountInt())
	}
}

func TestS_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if ("") != (s.String()) {
		t.Fatalf("want %v, got %v", "", s.String())
	}
	if s.IsPlural() {
		t.Fatal("expected false")
	}
	if (1) != (s.CountInt()) {
		t.Fatalf("want %v, got %v", 1, s.CountInt())
	}
	if ("1") != (s.CountString()) {
		t.Fatalf("want %v, got %v", "1", s.CountString())
	}
	if ("") != (s.GenderString()) {
		t.Fatalf("want %v, got %v", "", s.GenderString())
	}
	if ("") != (s.LocationString()) {
		t.Fatalf("want %v, got %v", "", s.LocationString())
	}
	if ("") != (s.NounString()) {
		t.Fatalf("want %v, got %v", "", s.NounString())
	}
	if ("neutral") != (s.FormalityString()) {
		t.Fatalf("want %v, got %v", "neutral", s.FormalityString())
	}
	if s.IsFormal() {
		t.Fatal("expected false")
	}
	if s.IsInformal() {
		t.Fatal("expected false")
	}
}

func TestSubject_Count_Good(t *testing.T) {
	subj := S("file", "test.txt").Count(5)
	if (5) != (subj.CountInt()) {
		t.Fatalf("want %v, got %v", 5, subj.CountInt())
	}
	if ("5") != (subj.CountString()) {
		t.Fatalf("want %v, got %v", "5", subj.CountString())
	}
	if !(subj.IsPlural()) {
		t.Fatal("expected true")
	}
}

func TestSubject_CountString_UsesLocaleFormatting(t *testing.T) {
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

	subj := S("file", "test.txt").Count(1234)
	if ("1 234") != (subj.CountString()) {
		t.Fatalf("want %v, got %v", "1 234", subj.CountString())
	}
}

func TestSubject_Count_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	result := s.Count(5)
	if (result) != (nil) {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestSubject_Gender_Good(t *testing.T) {
	subj := S("user", "Alice").Gender("feminine")
	if ("feminine") != (subj.GenderString()) {
		t.Fatalf("want %v, got %v", "feminine", subj.GenderString())
	}
}

func TestSubject_Gender_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if (s.Gender("masculine")) != (nil) {
		t.Fatalf("expected nil, got %v", s.Gender("masculine"))
	}
}

func TestSubject_In_Good(t *testing.T) {
	subj := S("file", "config.yaml").In("workspace")
	if ("workspace") != (subj.LocationString()) {
		t.Fatalf("want %v, got %v", "workspace", subj.LocationString())
	}
}

func TestSubject_In_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if (s.In("workspace")) != (nil) {
		t.Fatalf("expected nil, got %v", s.In("workspace"))
	}
}

func TestSubject_Formal_Good(t *testing.T) {
	subj := S("document", "report").Formal()
	if !(subj.IsFormal()) {
		t.Fatal("expected true")
	}
	if subj.IsInformal() {
		t.Fatal("expected false")
	}
	if ("formal") != (subj.FormalityString()) {
		t.Fatalf("want %v, got %v", "formal", subj.FormalityString())
	}
}

func TestSubject_Formal_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if (s.Formal()) != (nil) {
		t.Fatalf("expected nil, got %v", s.Formal())
	}
}

func TestSubject_Informal_Good(t *testing.T) {
	subj := S("message", "hello").Informal()
	if !(subj.IsInformal()) {
		t.Fatal("expected true")
	}
	if subj.IsFormal() {
		t.Fatal("expected false")
	}
	if ("informal") != (subj.FormalityString()) {
		t.Fatalf("want %v, got %v", "informal", subj.FormalityString())
	}
}

func TestSubject_Informal_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if (s.Informal()) != (nil) {
		t.Fatalf("expected nil, got %v", s.Informal())
	}
}

func TestSubject_SetFormality_Good(t *testing.T) {
	tests := []struct {
		name      string
		formality Formality
		want      string
	}{
		{"neutral", FormalityNeutral, "neutral"},
		{"formal", FormalityFormal, "formal"},
		{"informal", FormalityInformal, "informal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subj := S("item", "x").SetFormality(tt.formality)
			if (tt.want) != (subj.FormalityString()) {
				t.Fatalf("want %v, got %v", tt.want, subj.FormalityString())
			}
		})
	}
}

func TestSubject_SetFormality_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	if (s.SetFormality(FormalityFormal)) != (nil) {
		t.Fatalf("expected nil, got %v", s.SetFormality(FormalityFormal))
	}
}

// --- String() ---

func TestSubject_String_Good(t *testing.T) {
	subj := S("file", "config.yaml")
	if ("config.yaml") != (subj.String()) {
		t.Fatalf("want %v, got %v", "config.yaml", subj.String())
	}
}

func TestSubject_String_Good_Stringer(t *testing.T) {
	// Use a type that implements fmt.Stringer
	subj := S("error", core.NewError("something broke"))
	if ("something broke") != (subj.String()) {
		t.Fatalf("want %v, got %v", "something broke", subj.String())
	}
}

func TestSubject_String_Good_IntValue(t *testing.T) {
	subj := S("count", 42)
	if ("42") != (subj.String()) {
		t.Fatalf("want %v, got %v", "42", subj.String())
	}
}

func TestSubject_NounString_Good(t *testing.T) {
	subj := S("repository", "go-i18n")
	if ("repository") != (subj.NounString()) {
		t.Fatalf("want %v, got %v", "repository", subj.NounString())
	}
}

func TestSubject_IsPlural_Good(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  bool
	}{
		{"zero_is_plural", 0, true},
		{"one_is_singular", 1, false},
		{"two_is_plural", 2, true},
		{"negative_is_plural", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subj := S("item", "x").Count(tt.count)
			if (tt.want) != (subj.IsPlural()) {
				t.Fatalf("want %v, got %v", tt.want, subj.IsPlural())
			}
		})
	}
}

// --- newTemplateData ---

func TestNewTemplateData_Good(t *testing.T) {
	subj := S("file", "test.txt").Count(3).Gender("neuter").In("workspace").Formal()
	data := newTemplateData(subj)
	if ("test.txt") != (data.Subject) {
		t.Fatalf("want %v, got %v", "test.txt", data.Subject)
	}
	if ("file") != (data.Noun) {
		t.Fatalf("want %v, got %v", "file", data.Noun)
	}
	if (3) != (data.Count) {
		t.Fatalf("want %v, got %v", 3, data.Count)
	}
	if ("neuter") != (data.Gender) {
		t.Fatalf("want %v, got %v", "neuter", data.Gender)
	}
	if ("workspace") != (data.Location) {
		t.Fatalf("want %v, got %v", "workspace", data.Location)
	}
	if (FormalityFormal) != (data.Formality) {
		t.Fatalf("want %v, got %v", FormalityFormal, data.Formality)
	}
	if !(data.IsFormal) {
		t.Fatal("expected true")
	}
	if !(data.IsPlural) {
		t.Fatal("expected true")
	}
	if ("test.txt") != (data.Value) {
		t.Fatalf("want %v, got %v", "test.txt", data.Value)
	}
}

func TestNewTemplateData_Bad_NilSubject(t *testing.T) {
	data := newTemplateData(nil)
	if (1) != (data.Count) {
		t.Fatalf("want %v, got %v", 1, data.Count)
	}
	if ("") != (data.Subject) {
		t.Fatalf("want %v, got %v", "", data.Subject)
	}
	if ("") != (data.Noun) {
		t.Fatalf("want %v, got %v", "", data.Noun)
	}
}

func TestNewTemplateData_Good_Singular(t *testing.T) {
	subj := S("item", "widget")
	data := newTemplateData(subj)
	if data.IsPlural {
		t.Fatal("expected false")
	}
	if data.IsFormal {
		t.Fatal("expected false")
	}
}

func TestSubject_FullChain_Good(t *testing.T) {
	subj := S("file", "readme.md").
		Count(2).
		Gender("neuter").
		In("project").
		Formal()
	if ("file") != (subj.NounString()) {
		t.Fatalf("want %v, got %v", "file", subj.NounString())
	}
	if ("readme.md") != (subj.String()) {
		t.Fatalf("want %v, got %v", "readme.md", subj.String())
	}
	if (2) != (subj.CountInt()) {
		t.Fatalf("want %v, got %v", 2, subj.CountInt())
	}
	if ("2") != (subj.CountString()) {
		t.Fatalf("want %v, got %v", "2", subj.CountString())
	}
	if ("neuter") != (subj.GenderString()) {
		t.Fatalf("want %v, got %v", "neuter", subj.GenderString())
	}
	if ("project") != (subj.LocationString()) {
		t.Fatalf("want %v, got %v", "project", subj.LocationString())
	}
	if !(subj.IsFormal()) {
		t.Fatal("expected true")
	}
	if !(subj.IsPlural()) {
		t.Fatal("expected true")
	}
}
