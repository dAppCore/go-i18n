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

// TestSpanishComposition proves the Spanish theory: es.json is PURE DATA —
// no Spanish-specific code exists anywhere in the engine. The gerundio is
// exactly the Spanish UI progress convention (Eliminando...), and subject +
// participle is native word order (Archivo eliminado).
//
//	T("i18n.done.delete", "file") // "Archivo eliminado"
func TestSpanishComposition(t *testing.T) {
	setCompositionLanguage(t, "es")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Archivo eliminado"},
		{"done.send package", T("i18n.done.send", "package"), "Paquete enviado"},
		{"done.resolve issue", T("i18n.done.resolve", "issue"), "Problema resuelto"}, // irregular participle
		{"progress.delete", T("i18n.progress.delete"), "Eliminando..."},
		{"progress.build", T("i18n.progress.build"), "Construyendo..."},
		{"count.file 5", T("i18n.count.file", 5), "5 archivos"},
		{"count.run 3", T("i18n.count.run", 3), "3 ejecuciones"},
		{"count.error 2", T("i18n.count.error", 2), "2 errores"}, // -or → -ores
		{"fail.push branch", T("i18n.fail.push", "branch"), "No se pudo subir rama"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestSpanishArticleBridge pins el/la, un/una AND the gendered plural
// definites los/las — the case that forced definite_plural_by_gender into
// the schema. Includes the Greek -ma masculine trap: EL problema.
//
//	DefinitePhrase("issue") // "el problema"
func TestSpanishArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "es")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite m", ArticlePhrase("file"), "un archivo"},
		{"indefinite f", ArticlePhrase("branch"), "una rama"},
		{"definite m", DefinitePhrase("file"), "el archivo"},
		{"definite f", DefinitePhrase("task"), "la tarea"},
		{"greek -ma masculine", DefinitePhrase("issue"), "el problema"},
		{"plural definite m", DefinitePhrase("archivos"), "los archivos"},
		{"plural definite f", DefinitePhrase("tareas"), "las tareas"},
		{"plural indefinite→definite m", ArticlePhrase("archivos"), "los archivos"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestKlingonComposition proves two things at once: the article-less
// language mechanism (article.none — phrases degrade to the bare noun,
// which is how Japanese and Russian will work too), and that Klingon's
// aspect suffixes map onto the schema mechanically: -pu' perfective is
// the past slot, -taH continuous the gerund slot. Vocabulary is canonical
// Okrand. Qapla'.
//
//	T("i18n.done.delete", "file") // "De' teqpu'"
func TestKlingonComposition(t *testing.T) {
	setCompositionLanguage(t, "tlh")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "De' teqpu'"},
		{"done.send message", T("i18n.done.send", "message"), "QIn labpu'"},
		{"done.find error", T("i18n.done.find", "error"), "Qagh Sampu'"},
		{"progress.scan", T("i18n.progress.scan"), "HotlhtaH..."},
		{"count.error 2", T("i18n.count.error", 2), "2 Qaghmey"},  // things take -mey
		{"count.person 3", T("i18n.count.person", 3), "3 nuvpu'"}, // speakers take -pu'
		{"fail.push task", T("i18n.fail.push", "task"), "lujpu': yuv Qu'"},
		{"prompt.yes", T("prompt.yes"), "HIja'"},
		{"prompt.no", T("prompt.no"), "ghobe'"},
		{"bare article phrase", ArticlePhrase("file"), "De'"},   // no articles in Klingon
		{"bare definite phrase", DefinitePhrase("file"), "De'"}, // none, definite or otherwise
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestPirateComposition pins the en-x-pirate dialect file: pure vocabulary
// over the en fallback chain — English morphology conjugates pirate words
// for free (scuttled, rigged, krakens), exactly as en-US rides the same
// chain for its spellings.
//
//	T("i18n.fail.push", "branch") // "Blimey! Couldn't heave plank"
func TestPirateComposition(t *testing.T) {
	setCompositionLanguage(t, "en-x-pirate")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Scroll scuttled"},
		{"done.build branch", T("i18n.done.build", "branch"), "Plank rigged"},
		{"progress.build", T("i18n.progress.build"), "Rigging... arrr"},
		{"progress.test", T("i18n.progress.test"), "Keelhauling... arrr"},
		{"count.error 3", T("i18n.count.error", 3), "3 krakens"},
		{"count.package 2", T("i18n.count.package", 2), "2 cargoes"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Blimey! Couldn't heave plank"},
		{"prompt.confirm", T("prompt.confirm"), "Arrr, ye be certain?"},
		{"unbridged verb falls back", T("i18n.done.merge", "file"), "Scroll merged"},
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
