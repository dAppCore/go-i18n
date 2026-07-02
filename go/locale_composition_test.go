package i18n

// End-to-end composition through the gram.word bridge: the SAME English
// T() keys and arguments produce French and German output, because the
// bridge maps the key to the locale's own word and the locale's grammar
// tables conjugate, pluralise, gender and article it from there.

import "testing"

func setCompositionLanguage(t *testing.T, lang string) {
	t.Helper()
	prev := Default()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})
	if err := errorFromResult(SetLanguage(lang)); err != nil {
		t.Fatalf("SetLanguage(%s) failed: %v", lang, err)
	}
}

// TestFrenchComposition drives the i18n.* namespace under fr with English
// keys only — the fr word bridge and grammar tables do the rest.
//
//	T("i18n.done.delete", "file") // "Fichier supprimé"
func TestFrenchComposition(t *testing.T) {
	setCompositionLanguage(t, "fr")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Fichier supprimé"},
		// v1 limitation, pinned deliberately: the participle does not yet
		// agree with a feminine subject (proper French: "Branche créée").
		// VerbForms carries one past form; agreement is the v2 line.
		{"done.create branch", T("i18n.done.create", "branch"), "Branche créé"},
		{"progress.build", T("i18n.progress.build"), "Construisant..."},
		{"progress.delete", T("i18n.progress.delete"), "Supprimant..."},
		{"count.file 5", T("i18n.count.file", 5), "5 fichiers"},
		{"count.run 3", T("i18n.count.run", 3), "3 exécutions"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Impossible de pousser branche"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestFrenchArticleBridge proves an English key picks up French gender:
// the bridge resolves the word, the noun table supplies the gender, and
// the gendered indefinite/definite articles do the rest.
//
//	ArticlePhrase("file") // "un fichier" under fr
func TestFrenchArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "fr")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite m", ArticlePhrase("file"), "un fichier"},
		{"indefinite f", ArticlePhrase("branch"), "une branche"},
		{"definite m", DefinitePhrase("file"), "le fichier"},
		{"definite f", DefinitePhrase("task"), "la tâche"},
		{"definite elision", DefinitePhrase("error"), "l'erreur"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestGermanComposition drives the i18n.* namespace under de with English
// keys only. German word order matches the composed shape natively:
// subject + participle is exactly "Datei gelöscht".
//
//	T("i18n.done.delete", "file") // "Datei gelöscht"
func TestGermanComposition(t *testing.T) {
	setCompositionLanguage(t, "de")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Datei gelöscht"},
		{"done.send package", T("i18n.done.send", "package"), "Paket gesendet"},
		{"done.find error", T("i18n.done.find", "error"), "Fehler gefunden"},
		{"progress.build", T("i18n.progress.build"), "Bauen..."},
		{"progress.delete", T("i18n.progress.delete"), "Löschen..."},
		{"count.file 5", T("i18n.count.file", 5), "5 Dateien"},
		{"count.error 3", T("i18n.count.error", 3), "3 Fehler"}, // zero-plural noun
		{"count.run 2", T("i18n.count.run", 2), "2 Läufe"},      // umlaut plural
		{"fail.push commit", T("i18n.fail.push", "commit"), "Fehlgeschlagen: pushen Commit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestGermanArticleBridge pins der/die/das and ein/eine across genders,
// resolved from English keys through the bridge, plus the definite plural.
//
//	DefinitePhrase("directory") // "das Verzeichnis"
func TestGermanArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "de")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite f", ArticlePhrase("file"), "eine Datei"},
		{"indefinite m", ArticlePhrase("error"), "ein Fehler"},
		{"indefinite n", ArticlePhrase("package"), "ein Paket"},
		{"definite f", DefinitePhrase("file"), "die Datei"},
		{"definite m", DefinitePhrase("error"), "der Fehler"},
		{"definite n", DefinitePhrase("directory"), "das Verzeichnis"},
		{"definite plural", DefinitePhrase("Dateien"), "die Dateien"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
