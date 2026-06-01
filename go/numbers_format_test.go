package i18n

import (
	"testing"

	"dappco.re/go"
)

// deRegionNumberLoader exposes a region-suffixed language whose base ("de") has
// a built-in number format. The language matcher preserves the region tag, so
// Language() returns "de-AT" and getNumberFormat must strip the region to find
// the base format.
type deRegionNumberLoader struct{}

func (deRegionNumberLoader) Languages() []string { return []string{"de-AT"} }

func (deRegionNumberLoader) Load(string) core.Result {
	return localeLoadResult(map[string]Message{}, &GrammarData{})
}

// TestNumbersFormat_getLocaleNumberFormat_MapHit confirms a registered language
// resolves directly from the built-in numberFormats table.
//
//	getLocaleNumberFormat("de") // {ThousandsSep: ".", DecimalSep: ",", ...}, true
func TestNumbersFormat_getLocaleNumberFormat_MapHit(t *testing.T) {
	nf, ok := getLocaleNumberFormat("de")
	if !ok {
		t.Fatal("getLocaleNumberFormat(de) reported not-found")
	}
	if nf.ThousandsSep != "." || nf.DecimalSep != "," {
		t.Errorf("getLocaleNumberFormat(de) = %+v, want dotted thousands / comma decimal", nf)
	}
}

// TestNumbersFormat_getLocaleNumberFormat_NotFound confirms an unregistered
// language reports not-found with the zero NumberFormat.
func TestNumbersFormat_getLocaleNumberFormat_NotFound(t *testing.T) {
	nf, ok := getLocaleNumberFormat("zz")
	if ok {
		t.Errorf("getLocaleNumberFormat(zz) = %+v, want not-found", nf)
	}
	if nf != (NumberFormat{}) {
		t.Errorf("getLocaleNumberFormat(zz) returned non-zero %+v", nf)
	}
}

// TestNumbersFormat_getLocaleNumberFormat_GrammarOverride confirms a non-empty
// Number block on the loaded grammar data takes precedence over the built-in
// table.
func TestNumbersFormat_getLocaleNumberFormat_GrammarOverride(t *testing.T) {
	const lang = "qaa-x-num"
	prev := GetGrammarData(lang)
	t.Cleanup(func() { SetGrammarData(lang, prev) })

	SetGrammarData(lang, &GrammarData{
		Number: NumberFormat{ThousandsSep: "_", DecimalSep: "·", PercentFmt: "%s pct"},
	})

	nf, ok := getLocaleNumberFormat(lang)
	if !ok {
		t.Fatal("getLocaleNumberFormat with grammar override reported not-found")
	}
	if nf.ThousandsSep != "_" || nf.DecimalSep != "·" {
		t.Errorf("grammar override = %+v, want underscore/middot separators", nf)
	}
}

// TestNumbersFormat_getNumberFormat_RegionFallback drives getNumberFormat's
// region-stripping branch: a region tag with no exact entry falls back to its
// base language's format.
func TestNumbersFormat_getNumberFormat_RegionFallback(t *testing.T) {
	prev := Default()
	svc, err := serviceFromResult(NewWithLoader(deRegionNumberLoader{}))
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() { SetDefault(prev) })

	if err := errorFromResult(SetLanguage("de-AT")); err != nil {
		t.Fatalf("SetLanguage(de-AT) failed: %v", err)
	}

	// Language() returns "de-AT" which has no exact numberFormats entry, so
	// getNumberFormat strips the region and falls back to the "de" format.
	if got := FormatNumber(1234567); got != "1.234.567" {
		t.Errorf("FormatNumber(de-AT) = %q, want %q (de fallback)", got, "1.234.567")
	}
}

// TestNumbersFormat_getNumberFormat_DefaultEnglish confirms an entirely unknown
// language falls all the way through to the English default.
func TestNumbersFormat_getNumberFormat_DefaultEnglish(t *testing.T) {
	prev := Default()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() { SetDefault(prev) })

	nf := getNumberFormat()
	if nf.ThousandsSep != "," || nf.DecimalSep != "." {
		t.Errorf("default getNumberFormat() = %+v, want English comma/dot", nf)
	}
}
