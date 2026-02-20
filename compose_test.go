package i18n

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- S() constructor ---

func TestS_Good(t *testing.T) {
	subj := S("file", "config.yaml")
	require.NotNil(t, subj)
	assert.Equal(t, "file", subj.Noun)
	assert.Equal(t, "config.yaml", subj.Value)
	assert.Equal(t, 1, subj.CountInt(), "default count should be 1")
}

func TestS_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Equal(t, "", s.String())
	assert.False(t, s.IsPlural())
	assert.Equal(t, 1, s.CountInt())
	assert.Equal(t, "1", s.CountString())
	assert.Equal(t, "", s.GenderString())
	assert.Equal(t, "", s.LocationString())
	assert.Equal(t, "", s.NounString())
	assert.Equal(t, "neutral", s.FormalityString())
	assert.False(t, s.IsFormal())
	assert.False(t, s.IsInformal())
}

// --- Chaining methods ---

func TestSubject_Count_Good(t *testing.T) {
	subj := S("file", "test.txt").Count(5)
	assert.Equal(t, 5, subj.CountInt())
	assert.Equal(t, "5", subj.CountString())
	assert.True(t, subj.IsPlural())
}

func TestSubject_Count_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	result := s.Count(5)
	assert.Nil(t, result)
}

func TestSubject_Gender_Good(t *testing.T) {
	subj := S("user", "Alice").Gender("feminine")
	assert.Equal(t, "feminine", subj.GenderString())
}

func TestSubject_Gender_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Nil(t, s.Gender("masculine"))
}

func TestSubject_In_Good(t *testing.T) {
	subj := S("file", "config.yaml").In("workspace")
	assert.Equal(t, "workspace", subj.LocationString())
}

func TestSubject_In_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Nil(t, s.In("workspace"))
}

func TestSubject_Formal_Good(t *testing.T) {
	subj := S("document", "report").Formal()
	assert.True(t, subj.IsFormal())
	assert.False(t, subj.IsInformal())
	assert.Equal(t, "formal", subj.FormalityString())
}

func TestSubject_Formal_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Nil(t, s.Formal())
}

func TestSubject_Informal_Good(t *testing.T) {
	subj := S("message", "hello").Informal()
	assert.True(t, subj.IsInformal())
	assert.False(t, subj.IsFormal())
	assert.Equal(t, "informal", subj.FormalityString())
}

func TestSubject_Informal_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Nil(t, s.Informal())
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
			assert.Equal(t, tt.want, subj.FormalityString())
		})
	}
}

func TestSubject_SetFormality_Bad_NilReceiver(t *testing.T) {
	var s *Subject
	assert.Nil(t, s.SetFormality(FormalityFormal))
}

// --- String() ---

func TestSubject_String_Good(t *testing.T) {
	subj := S("file", "config.yaml")
	assert.Equal(t, "config.yaml", subj.String())
}

func TestSubject_String_Good_Stringer(t *testing.T) {
	// Use a type that implements fmt.Stringer
	subj := S("error", fmt.Errorf("something broke"))
	assert.Equal(t, "something broke", subj.String())
}

func TestSubject_String_Good_IntValue(t *testing.T) {
	subj := S("count", 42)
	assert.Equal(t, "42", subj.String())
}

// --- NounString ---

func TestSubject_NounString_Good(t *testing.T) {
	subj := S("repository", "go-i18n")
	assert.Equal(t, "repository", subj.NounString())
}

// --- IsPlural edge cases ---

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
			assert.Equal(t, tt.want, subj.IsPlural())
		})
	}
}

// --- newTemplateData ---

func TestNewTemplateData_Good(t *testing.T) {
	subj := S("file", "test.txt").Count(3).Gender("neuter").In("workspace").Formal()
	data := newTemplateData(subj)

	assert.Equal(t, "test.txt", data.Subject)
	assert.Equal(t, "file", data.Noun)
	assert.Equal(t, 3, data.Count)
	assert.Equal(t, "neuter", data.Gender)
	assert.Equal(t, "workspace", data.Location)
	assert.Equal(t, FormalityFormal, data.Formality)
	assert.True(t, data.IsFormal)
	assert.True(t, data.IsPlural)
	assert.Equal(t, "test.txt", data.Value)
}

func TestNewTemplateData_Bad_NilSubject(t *testing.T) {
	data := newTemplateData(nil)
	assert.Equal(t, 1, data.Count, "nil subject should default count to 1")
	assert.Equal(t, "", data.Subject)
	assert.Equal(t, "", data.Noun)
}

func TestNewTemplateData_Good_Singular(t *testing.T) {
	subj := S("item", "widget")
	data := newTemplateData(subj)
	assert.False(t, data.IsPlural)
	assert.False(t, data.IsFormal)
}

// --- Full chaining ---

func TestSubject_FullChain_Good(t *testing.T) {
	subj := S("file", "readme.md").
		Count(2).
		Gender("neuter").
		In("project").
		Formal()

	assert.Equal(t, "file", subj.NounString())
	assert.Equal(t, "readme.md", subj.String())
	assert.Equal(t, 2, subj.CountInt())
	assert.Equal(t, "2", subj.CountString())
	assert.Equal(t, "neuter", subj.GenderString())
	assert.Equal(t, "project", subj.LocationString())
	assert.True(t, subj.IsFormal())
	assert.True(t, subj.IsPlural())
}
