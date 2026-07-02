package i18n

// End-to-end composition through the gram.word bridge for the Baltic pair:
// Lithuanian (lt) and Latvian (lv). Both are article-less (article.none)
// two-gender (m/f, no neuter) languages whose done-shape is Subject +
// passive past participle with an elided copula.

import "testing"

// TestLithuanianComposition drives the i18n.* namespace under lt with
// English keys only — the lt word bridge and grammar tables do the rest.
// Feminine agreement is a single mechanical rule from the masculine base
// (agreement.participle.f: strip "s") — ištrintas → ištrinta.
//
//	T("i18n.done.delete", "file") // "Failas ištrintas"
func TestLithuanianComposition(t *testing.T) {
	setCompositionLanguage(t, "lt")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file m", T("i18n.done.delete", "file"), "Failas ištrintas"},
		{"done.delete task f", T("i18n.done.delete", "task"), "Užduotis ištrinta"},
		{"progress.delete", T("i18n.progress.delete"), "Trynimas..."},
		{"count.error 3", T("i18n.count.error", 3), "3 klaidos"},
		{"fail.install package", T("i18n.fail.install", "package"), "Nepavyko įdiegti paketas"},
		{"bare article phrase", ArticlePhrase("file"), "failas"},
		{"bare definite phrase", DefinitePhrase("file"), "failas"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestLatvianComposition drives the i18n.* namespace under lv with English
// keys only. Feminine agreement needs both a strip AND an add from the
// masculine base (agreement.participle.f: strip "s", add "a") — izdzēsts →
// izdzēst → izdzēsta — since the masculine participle ends in a bare
// consonant (-ts), not a vowel, unlike Lithuanian's -as/-a pair.
//
//	T("i18n.done.delete", "file") // "Fails izdzēsts"
func TestLatvianComposition(t *testing.T) {
	setCompositionLanguage(t, "lv")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file m", T("i18n.done.delete", "file"), "Fails izdzēsts"},
		{"done.fix error f", T("i18n.done.fix", "error"), "Kļūda izlabota"},
		{"progress.delete", T("i18n.progress.delete"), "Dzēšana..."},
		{"count.error 3", T("i18n.count.error", 3), "3 kļūdas"},
		{"fail.install package", T("i18n.fail.install", "package"), "Neizdevās instalēt pakete"},
		{"bare article phrase", ArticlePhrase("file"), "fails"},
		{"bare definite phrase", DefinitePhrase("file"), "fails"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
