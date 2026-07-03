package i18n

// West Slavic composition: pl.json and cs.json ride the SAME article-less,
// three-gender rails as la.json (see locale_composition_test.go) — no code
// changes were needed to add these two languages, only data.
//
// Both are short-passive-participle registers, exactly how Polish and Czech
// software phrases outcomes: "Plik usunięty", "Soubor smazán". Polish
// agreement strips the masculine -y and adds -a/-e (usunięty → usunięta /
// usunięte); Czech agreement is add-only because the masculine already ends
// on the bare consonant (smazán → smazána / smazáno). The gerund slot is
// each language's progress-form verbal noun: Polish "Usuwanie...", Czech
// "Mazání...".

import "testing"

// TestPolishComposition drives the i18n.* namespace under pl with English
// keys only — the pl word bridge and grammar tables do the rest. One row
// per gender proves the strip-y/add-a-or-e participle rule generalises
// across four different verbs, not just the worked delete/file example.
//
//	T("i18n.done.delete", "file") // "Plik usunięty"
func TestPolishComposition(t *testing.T) {
	setCompositionLanguage(t, "pl")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file m", T("i18n.done.delete", "file"), "Plik usunięty"},
		{"done.create branch f", T("i18n.done.create", "branch"), "Gałąź utworzona"},
		{"done.create task n", T("i18n.done.create", "task"), "Zadanie utworzone"},
		{"done.fix vulnerability f", T("i18n.done.fix", "vulnerability"), "Podatność naprawiona"},
		{"done.install package m", T("i18n.done.install", "package"), "Pakiet zainstalowany"},
		{"progress.delete", T("i18n.progress.delete"), "Usuwanie..."},
		{"progress.check", T("i18n.progress.check"), "Sprawdzanie..."},
		{"count.error 3", T("i18n.count.error", 3), "3 błędy"},
		{"count.child 2 irregular", T("i18n.count.child", 2), "2 dzieci"},
		{"count.category 2", T("i18n.count.category", 2), "2 kategorie"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nie udało się wypchnąć gałąź"},
		{"prompt.yes", T("prompt.yes"), "tak"},
		{"prompt.no", T("prompt.no"), "nie"},
		{"bare article phrase", ArticlePhrase("file"), "plik"}, // Polish has no articles
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestCzechComposition drives the i18n.* namespace under cs with English
// keys only. The neuter row (find/child) stands in for create/task above —
// dítě is the only neuter noun in this table, so it is paired with a verb
// that makes sense with it ("child found") rather than forcing an artificial
// pairing.
//
//	T("i18n.done.delete", "file") // "Soubor smazán"
func TestCzechComposition(t *testing.T) {
	setCompositionLanguage(t, "cs")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file m", T("i18n.done.delete", "file"), "Soubor smazán"},
		{"done.create branch f", T("i18n.done.create", "branch"), "Větev vytvořena"},
		{"done.find child n", T("i18n.done.find", "child"), "Dítě nalezeno"},
		{"done.fix vulnerability f", T("i18n.done.fix", "vulnerability"), "Zranitelnost opravena"},
		{"done.install package m", T("i18n.done.install", "package"), "Balíček nainstalován"},
		{"progress.delete", T("i18n.progress.delete"), "Mazání..."},
		{"progress.run", T("i18n.progress.run"), "Spouštění..."},
		{"count.error 3", T("i18n.count.error", 3), "3 chyby"},
		{"count.child 2 irregular", T("i18n.count.child", 2), "2 děti"},
		{"count.task 4", T("i18n.count.task", 4), "4 úkoly"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nepodařilo se pushnout větev"},
		{"prompt.yes", T("prompt.yes"), "ano"},
		{"prompt.no", T("prompt.no"), "ne"},
		{"bare article phrase", ArticlePhrase("file"), "soubor"}, // Czech has no articles
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
