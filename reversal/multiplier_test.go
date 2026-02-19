package reversal

import (
	"testing"

	i18n "forge.lthn.ai/core/go-i18n"
)

func TestMultiplier_Expand(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	m := NewMultiplier()
	// Use "branch" (noun-only) to avoid dual-class ambiguity with "file" (now both verb and noun).
	variants := m.Expand("Delete the configuration branch")

	if len(variants) < 4 {
		t.Errorf("Expand() returned %d variants, want >= 4", len(variants))
	}

	expected := map[string]bool{
		"Delete the configuration branch":   true, // original
		"Deleted the configuration branch":  true, // past
		"Deleting the configuration branch": true, // gerund
		"Delete the configuration branches": true, // plural
	}
	for _, v := range variants {
		delete(expected, v)
	}
	for missing := range expected {
		t.Errorf("Expand() missing expected variant: %q", missing)
	}
}

func TestMultiplier_Expand_NoVerbs(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	m := NewMultiplier()
	variants := m.Expand("the configuration file")
	if len(variants) < 2 {
		t.Errorf("Expand() returned %d variants, want >= 2", len(variants))
	}
}

func TestMultiplier_Expand_Empty(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	m := NewMultiplier()
	variants := m.Expand("")
	if len(variants) != 0 {
		t.Errorf("Expand(\"\") returned %d variants, want 0", len(variants))
	}
}

func TestMultiplier_Expand_Deterministic(t *testing.T) {
	svc, _ := i18n.New()
	i18n.SetDefault(svc)
	m := NewMultiplier()
	v1 := m.Expand("Delete the file")
	v2 := m.Expand("Delete the file")
	if len(v1) != len(v2) {
		t.Fatalf("Non-deterministic: %d vs %d variants", len(v1), len(v2))
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Errorf("Non-deterministic at [%d]: %q vs %q", i, v1[i], v2[i])
		}
	}
}
