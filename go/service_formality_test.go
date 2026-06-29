package i18n

import "testing"

// TestServiceFormality_parseFormalityValue covers every branch of the formality
// coercion: typed Formality, the "formal"/"informal" string spellings, and the
// neutral/unknown rejections.
//
//	parseFormalityValue("formal")          // FormalityFormal, true
//	parseFormalityValue(FormalityInformal) // FormalityInformal, true
//	parseFormalityValue("nonsense")        // FormalityNeutral, false
func TestServiceFormality_parseFormalityValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  Formality
		ok    bool
	}{
		{"typed formal", FormalityFormal, FormalityFormal, true},
		{"typed informal", FormalityInformal, FormalityInformal, true},
		{"typed neutral", FormalityNeutral, FormalityNeutral, false},
		{"string formal", "formal", FormalityFormal, true},
		{"string informal", "INFORMAL", FormalityInformal, true},
		{"string unknown", "polite", FormalityNeutral, false},
		{"nil", nil, FormalityNeutral, false},
		{"int", 3, FormalityNeutral, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseFormalityValue(tt.value)
			if got != tt.want || ok != tt.ok {
				t.Errorf("parseFormalityValue(%v) = (%v, %v), want (%v, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// TestServiceFormality_getEffectiveFormality drives every data shape the
// resolver inspects: TranslationContext, Subject, map[string]any,
// map[string]string, and the service-default fallback.
func TestServiceFormality_getEffectiveFormality(t *testing.T) {
	svc := serviceForAudit(t)

	// TranslationContext with explicit formality wins.
	if got := svc.getEffectiveFormality(C("greeting").Formal()); got != FormalityFormal {
		t.Errorf("context formal = %v, want FormalityFormal", got)
	}

	// Subject with explicit formality wins.
	if got := svc.getEffectiveFormality(S("user", "u").Informal()); got != FormalityInformal {
		t.Errorf("subject informal = %v, want FormalityInformal", got)
	}

	// map[string]any carrying a Formality string.
	if got := svc.getEffectiveFormality(map[string]any{"Formality": "formal"}); got != FormalityFormal {
		t.Errorf("map[any] formal = %v, want FormalityFormal", got)
	}

	// map[string]string carrying a Formality string.
	if got := svc.getEffectiveFormality(map[string]string{"Formality": "informal"}); got != FormalityInformal {
		t.Errorf("map[string] informal = %v, want FormalityInformal", got)
	}

	// Neutral context falls through to the service default (neutral here).
	if got := svc.getEffectiveFormality(C("greeting")); got != svc.formality {
		t.Errorf("neutral context = %v, want service default %v", got, svc.formality)
	}

	// Nil data falls through to the service default.
	if got := svc.getEffectiveFormality(nil); got != svc.formality {
		t.Errorf("nil data = %v, want service default %v", got, svc.formality)
	}
}

// TestServiceFormality_getEffectiveFormality_ServiceDefault confirms the service
// default formality is honoured when no per-call shape overrides it.
func TestServiceFormality_getEffectiveFormality_ServiceDefault(t *testing.T) {
	svc := serviceForAudit(t)
	svc.SetFormality(FormalityFormal)
	if got := svc.getEffectiveFormality(map[string]any{"other": "x"}); got != FormalityFormal {
		t.Errorf("map without Formality = %v, want service default FormalityFormal", got)
	}
}

// TestServiceFormality_mergeContextExtra covers the three accepted value shapes
// plus the nil-destination and nil-value guards.
func TestServiceFormality_mergeContextExtra(t *testing.T) {
	// map[string]any source.
	dst := map[string]any{}
	mergeContextExtra(dst, map[string]any{"a": 1, "b": "two"})
	if dst["a"] != 1 || dst["b"] != "two" {
		t.Errorf("merge map[any] = %v", dst)
	}

	// map[string]string source.
	dst = map[string]any{}
	mergeContextExtra(dst, map[string]string{"c": "three"})
	if dst["c"] != "three" {
		t.Errorf("merge map[string] = %v", dst)
	}

	// *TranslationContext source merges its Extra map.
	dst = map[string]any{}
	mergeContextExtra(dst, C("greeting").Set("d", 4))
	if dst["d"] != 4 {
		t.Errorf("merge context Extra = %v", dst)
	}

	// nil destination and nil value are no-ops (must not panic).
	mergeContextExtra(nil, map[string]any{"x": 1})
	mergeContextExtra(map[string]any{}, nil)

	// *TranslationContext with empty Extra is a no-op.
	dst = map[string]any{}
	mergeContextExtra(dst, C("greeting"))
	if len(dst) != 0 {
		t.Errorf("merge empty-Extra context mutated dst = %v", dst)
	}
}
