package i18n

// Finnic-family composition: Finnish and Estonian both decline articles and
// grammatical gender entirely (article.none, no gender on any noun table
// entry), and both render done/progress in their native impersonal mood
// rather than a personal voice — this is the register real Finnish and
// Estonian software uses (Tiedosto poistettu, Fail kustutatud), not a
// simplification forced by the schema.
//
// The count rows prove numeral government, not plural formation: Finnic
// numerals govern the PARTITIVE SINGULAR (3 virhettä, 3 viga — never the
// nominative plural *3 virheet, *3 vead). Both locales therefore author the
// noun "other" slot as the partitive singular — in an article-less locale
// the count path is the slot's only consumer, so it carries the form the
// numeral actually demands.

import "testing"

// TestFinnishComposition drives the i18n.* namespace under fi with English
// keys only — the fi word bridge and grammar tables do the rest. The past
// slot is the passive past participle, the progress slot is the passive
// present: exactly what Finnish software shows.
//
//	T("i18n.done.delete", "file") // "Tiedosto poistettu"
func TestFinnishComposition(t *testing.T) {
	setCompositionLanguage(t, "fi")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Tiedosto poistettu"},
		{"done.send message", T("i18n.done.send", "message"), "Viesti lähetetty"},
		{"progress.build", T("i18n.progress.build"), "Rakennetaan..."},
		{"count.error 3 — numeral partitive", T("i18n.count.error", 3), "3 virhettä"},
		{"count.file 5 — numeral partitive", T("i18n.count.file", 5), "5 tiedostoa"},
		{"count.file 1 — nominative", T("i18n.count.file", 1), "1 tiedosto"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Ei voitu työntää haara"},
		{"bare article phrase", ArticlePhrase("file"), "tiedosto"}, // no articles in Finnish
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestEstonianComposition drives the i18n.* namespace under et with English
// keys only. Like Finnish, Estonian software renders done/progress in the
// impersonal mood: the impersonal past participle (-tud) for a completed
// action, the impersonal present (-takse) for one in progress.
//
//	T("i18n.done.delete", "file") // "Fail kustutatud"
func TestEstonianComposition(t *testing.T) {
	setCompositionLanguage(t, "et")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Fail kustutatud"},
		{"done.send message", T("i18n.done.send", "message"), "Sõnum saadetud"},
		{"progress.build", T("i18n.progress.build"), "Ehitatakse..."},
		// Estonian "viga" is partitive-homophonous with its nominative —
		// 3 viga is the partitive, not an unpluralised slip.
		{"count.error 3 — numeral partitive", T("i18n.count.error", 3), "3 viga"},
		{"count.file 5 — numeral partitive", T("i18n.count.file", 5), "5 faili"},
		{"count.file 1 — nominative", T("i18n.count.file", 1), "1 fail"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Ei õnnestunud lükata haru"},
		{"bare article phrase", ArticlePhrase("file"), "fail"}, // no articles in Estonian
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
