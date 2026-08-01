package i18n

// End-to-end composition through the gram.word bridge for the West/South
// Slavic trio: Slovak, Slovenian, Croatian. All three are article-less
// (article.none) and compose Subject + passive participle with an elided
// copula — "Datoteka izbrisana", not "Datoteka je izbrisana". Slovak
// authors the masculine participle in the long -ný/-tý form and strips it
// for agreement (zmazaný → zmazaná/zmazané); Slovenian and Croatian author
// the masculine bare (izbrisan, pokrenut) and simply append (izbrisana/
// izbrisano). Composition is NOMINATIVE-only throughout — real Slovak/
// Slovenian/Croatian alternate case after numerals and in object position,
// which this engine's case-free noun table does not attempt (see each
// locale's _comment for the documented gap).

import "testing"

// TestSlovakComposition drives the i18n.* namespace under sk with English
// keys only — the sk word bridge and grammar tables do the rest. Agreement
// is a single uniform rule (strip "ý", add "á"/"é") verified across all
// three genders.
//
//	T("i18n.done.delete", "file") // "Súbor zmazaný"
func TestSlovakComposition(t *testing.T) {
	setCompositionLanguage(t, "sk")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file m", T("i18n.done.delete", "file"), "Súbor zmazaný"},
		{"done.fix error f", T("i18n.done.fix", "error"), "Chyba opravená"},
		{"done.find child n", T("i18n.done.find", "child"), "Dieťa nájdené"},
		{"progress.delete", T("i18n.progress.delete"), "Mazanie..."},
		{"progress.read", T("i18n.progress.read"), "Čítanie..."},
		{"count.file 3", T("i18n.count.file", 3), "3 súbory"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nepodarilo sa pushnúť vetva"},
		{"bare article phrase", ArticlePhrase("file"), "súbor"},   // Slovak has no articles
		{"bare definite phrase", DefinitePhrase("file"), "súbor"}, // none, definite or otherwise
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestSlovenianComposition drives the i18n.* namespace under sl with
// English keys only. Masculine participles are authored bare (no vowel
// ending), so agreement is a pure append: "a" (f) / "o" (n).
//
//	T("i18n.done.delete", "file") // "Datoteka izbrisana"
func TestSlovenianComposition(t *testing.T) {
	setCompositionLanguage(t, "sl")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file f", T("i18n.done.delete", "file"), "Datoteka izbrisana"},
		{"done.update message n", T("i18n.done.update", "message"), "Sporočilo posodobljeno"},
		{"done.stop server m", T("i18n.done.stop", "server"), "Strežnik ustavljen"},
		{"progress.delete — matches Croatian", T("i18n.progress.delete"), "Brisanje..."},
		{"progress.open", T("i18n.progress.open"), "Odpiranje..."},
		{"count.task 4", T("i18n.count.task", 4), "4 naloge"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Ni uspelo pushati veja"},
		{"bare article phrase", ArticlePhrase("file"), "datoteka"},   // Slovenian has no articles
		{"bare definite phrase", DefinitePhrase("file"), "datoteka"}, // none, definite or otherwise
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestCroatianComposition drives the i18n.* namespace under hr with
// English keys only. Same bare-masculine agreement shape as Slovenian
// (izbrisan → izbrisana/izbrisano), but its own distinct vocabulary
// throughout — deliberately NOT copied from sl.json (ovisnost not
// odvisnost, dijete not otrok).
//
//	T("i18n.done.delete", "file") // "Datoteka izbrisana"
func TestCroatianComposition(t *testing.T) {
	setCompositionLanguage(t, "hr")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file f", T("i18n.done.delete", "file"), "Datoteka izbrisana"},
		{"done.fix issue m", T("i18n.done.fix", "issue"), "Problem popravljen"},
		{"done.find child n", T("i18n.done.find", "child"), "Dijete pronađeno"},
		{"progress.delete — matches Slovenian", T("i18n.progress.delete"), "Brisanje..."},
		{"progress.check", T("i18n.progress.check"), "Provjera..."},
		{"count.version 2", T("i18n.count.version", 2), "2 verzije"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nije uspjelo pushati grana"},
		{"bare article phrase", ArticlePhrase("file"), "datoteka"},   // Croatian has no articles
		{"bare definite phrase", DefinitePhrase("file"), "datoteka"}, // none, definite or otherwise
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
