// SPDX-License-Identifier: EUPL-1.2

package i18n

import (
	"runtime"
	"testing"
)

// Good: on Windows the display language comes back as a locale name; elsewhere
// it's "" (the environment is read instead).
func TestLocalise_systemLanguage_Good(t *testing.T) {
	got := systemLanguage()
	if (runtime.GOOS == "windows") && got == "" {
		t.Fatal("Windows gave no display language")
	}
	if !(runtime.GOOS == "windows") && got != "" {
		t.Fatalf("systemLanguage() = %q off Windows", got)
	}
}

// Bad: an unsupported system language detects nothing.
func TestLocalise_systemLanguage_Bad(t *testing.T) {
	if got := detectLanguageFromEnv("xx-XX", nil); got != "" {
		t.Fatalf("detected %q from nothing supported", got)
	}
}

// Ugly: whatever Windows says, it parses as a language tag detectLanguage can use.
func TestLocalise_systemLanguage_Ugly(t *testing.T) {
	if got := systemLanguage(); got != "" && normalizeLanguageTag(got) == "" {
		t.Fatalf("systemLanguage() = %q doesn't parse", got)
	}
}
