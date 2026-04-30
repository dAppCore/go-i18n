package i18n

import (
	"testing"

	"dappco.re/go"
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

// --- AX-7 canonical triplets ---

func TestCompose_S_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "config.yaml")
		if subj.Noun != "file" {
			t.Fatalf("want file, got %q", subj.Noun)
		}
	})
	if !called {
		t.Fatal("S was not exercised")
	}
}

func TestCompose_S_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("", nil)
		if subj.CountInt() != 1 {
			t.Fatalf("want default count")
		}
	})
	if !called {
		t.Fatal("S was not exercised")
	}
}

func TestCompose_S_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", 42).Count(0)
		if subj.CountInt() != 0 {
			t.Fatalf("want zero count")
		}
	})
	if !called {
		t.Fatal("S was not exercised")
	}
}

func TestCompose_ComposeIntent_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ComposeIntent(Intent{Question: "Delete {{.Subject}}?"}, S("file", "config.yaml"))
		if got.Question != "Delete config.yaml?" {
			t.Fatalf("got %q", got.Question)
		}
	})
	if !called {
		t.Fatal("ComposeIntent was not exercised")
	}
}

func TestCompose_ComposeIntent_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := (Intent{}).Compose(nil)
		if got.Question != "" {
			t.Fatalf("got %q", got.Question)
		}
	})
	if !called {
		t.Fatal("ComposeIntent was not exercised")
	}
}

func TestCompose_ComposeIntent_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ComposeIntent(Intent{Question: "{{.Count}}"}, S("file", "x").Count(0))
		if got.Question != "0" {
			t.Fatalf("got %q", got.Question)
		}
	})
	if !called {
		t.Fatal("ComposeIntent was not exercised")
	}
}

func TestCompose_Intent_Compose_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := (Intent{Success: "Done {{.Subject}}"}).Compose(S("file", "config.yaml"))
		if got.Success != "Done config.yaml" {
			t.Fatalf("got %q", got.Success)
		}
	})
	if !called {
		t.Fatal("Intent_Compose was not exercised")
	}
}

func TestCompose_Intent_Compose_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := (Intent{}).Compose(nil)
		if got.Question != "" {
			t.Fatalf("got %q", got.Question)
		}
	})
	if !called {
		t.Fatal("Intent_Compose was not exercised")
	}
}

func TestCompose_Intent_Compose_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ComposeIntent(Intent{Question: "{{.Count}}"}, S("file", "x").Count(0))
		if got.Question != "0" {
			t.Fatalf("got %q", got.Question)
		}
	})
	if !called {
		t.Fatal("Intent_Compose was not exercised")
	}
}

func TestCompose_Subject_Count_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(3)
		if subj.CountInt() != 3 {
			t.Fatalf("want 3")
		}
	})
	if !called {
		t.Fatal("Subject_Count was not exercised")
	}
}

func TestCompose_Subject_Count_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.Count(3)
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_Count was not exercised")
	}
}

func TestCompose_Subject_Count_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(0)
		if subj.CountInt() != 0 {
			t.Fatalf("want 0")
		}
	})
	if !called {
		t.Fatal("Subject_Count was not exercised")
	}
}

func TestCompose_Subject_Gender_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Gender("f")
		if subj.GenderString() != "f" {
			t.Fatalf("want f")
		}
	})
	if !called {
		t.Fatal("Subject_Gender was not exercised")
	}
}

func TestCompose_Subject_Gender_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.Gender("f")
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_Gender was not exercised")
	}
}

func TestCompose_Subject_Gender_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Gender("")
		if subj.GenderString() != "" {
			t.Fatalf("want empty")
		}
	})
	if !called {
		t.Fatal("Subject_Gender was not exercised")
	}
}

func TestCompose_Subject_In_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").In("workspace")
		if subj.LocationString() != "workspace" {
			t.Fatalf("want workspace")
		}
	})
	if !called {
		t.Fatal("Subject_In was not exercised")
	}
}

func TestCompose_Subject_In_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.In("workspace")
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_In was not exercised")
	}
}

func TestCompose_Subject_In_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").In("")
		if subj.LocationString() != "" {
			t.Fatalf("want empty")
		}
	})
	if !called {
		t.Fatal("Subject_In was not exercised")
	}
}

func TestCompose_Subject_Formal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Formal()
		if !subj.IsFormal() {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("Subject_Formal was not exercised")
	}
}

func TestCompose_Subject_Formal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.Formal()
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_Formal was not exercised")
	}
}

func TestCompose_Subject_Formal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Formal().Informal()
		if subj.IsFormal() {
			t.Fatalf("want not formal")
		}
	})
	if !called {
		t.Fatal("Subject_Formal was not exercised")
	}
}

func TestCompose_Subject_Informal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Informal()
		if !subj.IsInformal() {
			t.Fatalf("want informal")
		}
	})
	if !called {
		t.Fatal("Subject_Informal was not exercised")
	}
}

func TestCompose_Subject_Informal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.Informal()
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_Informal was not exercised")
	}
}

func TestCompose_Subject_Informal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Informal().Formal()
		if subj.IsInformal() {
			t.Fatalf("want not informal")
		}
	})
	if !called {
		t.Fatal("Subject_Informal was not exercised")
	}
}

func TestCompose_Subject_SetFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").SetFormality(FormalityFormal)
		if subj.FormalityString() != "formal" {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("Subject_SetFormality was not exercised")
	}
}

func TestCompose_Subject_SetFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.SetFormality(FormalityFormal)
		if got != nil {
			t.Fatalf("nil receiver should return nil")
		}
	})
	if !called {
		t.Fatal("Subject_SetFormality was not exercised")
	}
}

func TestCompose_Subject_SetFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").SetFormality(FormalityNeutral)
		if subj.FormalityString() != "neutral" {
			t.Fatalf("want neutral")
		}
	})
	if !called {
		t.Fatal("Subject_SetFormality was not exercised")
	}
}

func TestCompose_Subject_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "config.yaml")
		got := subj.String()
		if got != "config.yaml" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_String was not exercised")
	}
}

func TestCompose_Subject_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.String()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_String was not exercised")
	}
}

func TestCompose_Subject_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("n", 42)
		got := subj.String()
		if got != "42" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_String was not exercised")
	}
}

func TestCompose_Subject_IsPlural_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(2)
		got := subj.IsPlural()
		if !got {
			t.Fatalf("want plural")
		}
	})
	if !called {
		t.Fatal("Subject_IsPlural was not exercised")
	}
}

func TestCompose_Subject_IsPlural_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.IsPlural()
		if got {
			t.Fatalf("nil is not plural")
		}
	})
	if !called {
		t.Fatal("Subject_IsPlural was not exercised")
	}
}

func TestCompose_Subject_IsPlural_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(1)
		got := subj.IsPlural()
		if got {
			t.Fatalf("one is not plural")
		}
	})
	if !called {
		t.Fatal("Subject_IsPlural was not exercised")
	}
}

func TestCompose_Subject_CountInt_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(3)
		got := subj.CountInt()
		if got != 3 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountInt was not exercised")
	}
}

func TestCompose_Subject_CountInt_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.CountInt()
		if got != 1 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountInt was not exercised")
	}
}

func TestCompose_Subject_CountInt_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(-1)
		got := subj.CountInt()
		if got != -1 {
			t.Fatalf("got %d", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountInt was not exercised")
	}
}

func TestCompose_Subject_CountString_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(1234)
		got := subj.CountString()
		if got != "1,234" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountString was not exercised")
	}
}

func TestCompose_Subject_CountString_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.CountString()
		if got != "1" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountString was not exercised")
	}
}

func TestCompose_Subject_CountString_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").Count(0)
		got := subj.CountString()
		if got != "0" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_CountString was not exercised")
	}
}

func TestCompose_Subject_GenderString_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Gender("f")
		got := subj.GenderString()
		if got != "f" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_GenderString was not exercised")
	}
}

func TestCompose_Subject_GenderString_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.GenderString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_GenderString was not exercised")
	}
}

func TestCompose_Subject_GenderString_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a")
		got := subj.GenderString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_GenderString was not exercised")
	}
}

func TestCompose_Subject_LocationString_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x").In("workspace")
		got := subj.LocationString()
		if got != "workspace" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_LocationString was not exercised")
	}
}

func TestCompose_Subject_LocationString_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.LocationString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_LocationString was not exercised")
	}
}

func TestCompose_Subject_LocationString_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x")
		got := subj.LocationString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_LocationString was not exercised")
	}
}

func TestCompose_Subject_NounString_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("file", "x")
		got := subj.NounString()
		if got != "file" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_NounString was not exercised")
	}
}

func TestCompose_Subject_NounString_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.NounString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_NounString was not exercised")
	}
}

func TestCompose_Subject_NounString_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("", "x")
		got := subj.NounString()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_NounString was not exercised")
	}
}

func TestCompose_Subject_FormalityString_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Formal()
		got := subj.FormalityString()
		if got != "formal" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_FormalityString was not exercised")
	}
}

func TestCompose_Subject_FormalityString_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.FormalityString()
		if got != "neutral" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_FormalityString was not exercised")
	}
}

func TestCompose_Subject_FormalityString_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a")
		got := subj.FormalityString()
		if got != "neutral" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Subject_FormalityString was not exercised")
	}
}

func TestCompose_Subject_IsFormal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Formal()
		got := subj.IsFormal()
		if !got {
			t.Fatalf("want formal")
		}
	})
	if !called {
		t.Fatal("Subject_IsFormal was not exercised")
	}
}

func TestCompose_Subject_IsFormal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.IsFormal()
		if got {
			t.Fatalf("nil is not formal")
		}
	})
	if !called {
		t.Fatal("Subject_IsFormal was not exercised")
	}
}

func TestCompose_Subject_IsFormal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a")
		got := subj.IsFormal()
		if got {
			t.Fatalf("neutral is not formal")
		}
	})
	if !called {
		t.Fatal("Subject_IsFormal was not exercised")
	}
}

func TestCompose_Subject_IsInformal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a").Informal()
		got := subj.IsInformal()
		if !got {
			t.Fatalf("want informal")
		}
	})
	if !called {
		t.Fatal("Subject_IsInformal was not exercised")
	}
}

func TestCompose_Subject_IsInformal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var subj *Subject
		got := subj.IsInformal()
		if got {
			t.Fatalf("nil is not informal")
		}
	})
	if !called {
		t.Fatal("Subject_IsInformal was not exercised")
	}
}

func TestCompose_Subject_IsInformal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		subj := S("user", "a")
		got := subj.IsInformal()
		if got {
			t.Fatalf("neutral is not informal")
		}
	})
	if !called {
		t.Fatal("Subject_IsInformal was not exercised")
	}
}
