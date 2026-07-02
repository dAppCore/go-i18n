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
		// Participle agreement: feminine subjects agree the participle via
		// the locale's declared rule (agreement.participle add "e").
		{"done.create branch", T("i18n.done.create", "branch"), "Branche créée"},
		{"done.finish task", T("i18n.done.finish", "task"), "Tâche finie"},
		{"done.lose branch", T("i18n.done.lose", "branch"), "Branche perdue"},
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
		// Feminine agreement via the declared -o → -a swap — which covers
		// the irregular participles free: resuelto → resuelta.
		{"done.delete task", T("i18n.done.delete", "task"), "Tarea eliminada"},
		{"done.push branch", T("i18n.done.push", "branch"), "Rama subida"},
		{"done.resolve query", T("i18n.done.resolve", "query"), "Consulta resuelta"},
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

// TestItalianComposition — Spanish's sibling rides identical machinery:
// gerundio progress, -o → -a feminine agreement, gendered plural definites
// i/le, and the Greek -ma masculine trap for the second Romance language.
// Known v1 gap, deliberate: elision/allomorphs (l'errore, lo, gli) are not
// attempted.
//
//	T("i18n.done.delete", "file") // "File eliminato" (file IS Italian)
func TestItalianComposition(t *testing.T) {
	setCompositionLanguage(t, "it")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "File eliminato"},
		{"done.create branch m", T("i18n.done.create", "branch"), "Ramo creato"},
		{"done.update task f", T("i18n.done.update", "change"), "Modifica aggiornata"},
		{"done.resolve issue -ma trap", T("i18n.done.resolve", "issue"), "Problema risolto"},
		{"done.split query f", T("i18n.done.split", "query"), "Richiesta divisa"},
		{"progress.delete", T("i18n.progress.delete"), "Eliminando..."},
		{"progress.build", T("i18n.progress.build"), "Costruendo..."},
		{"count.file invariable loan", T("i18n.count.file", 3), "3 file"},
		{"count.test 2", T("i18n.count.test", 2), "2 prove"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Impossibile spingere ramo"},
		{"indefinite f", ArticlePhrase("query"), "una richiesta"},
		{"definite m", DefinitePhrase("branch"), "il ramo"},
		{"definite plural m", DefinitePhrase("rami"), "i rami"},
		{"definite plural f", DefinitePhrase("prove"), "le prove"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestPortugueseComposition — European Portuguese (ficheiro, pasta,
// artefacto), third Romance language on the same rails: gerúndio progress,
// -o → -a agreement covering the irregular pago → paga, os/as plural
// definites, nasal plurals authored per noun (versões), and o problema.
//
//	T("i18n.done.delete", "file") // "Ficheiro eliminado"
func TestPortugueseComposition(t *testing.T) {
	setCompositionLanguage(t, "pt")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Ficheiro eliminado"},
		{"done.delete task f", T("i18n.done.delete", "task"), "Tarefa eliminada"},
		{"done.make change f irregular", T("i18n.done.make", "change"), "Alteração feita"},
		{"done.resolve issue -ma trap", T("i18n.done.resolve", "issue"), "Problema resolvido"},
		{"progress.delete", T("i18n.progress.delete"), "Eliminando..."},
		{"count.version nasal plural", T("i18n.count.version", 3), "3 versões"},
		{"count.run 2", T("i18n.count.run", 2), "2 execuções"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Não foi possível empurrar ramo"},
		{"indefinite f", ArticlePhrase("task"), "uma tarefa"},
		{"definite m", DefinitePhrase("file"), "o ficheiro"},
		{"definite plural f", DefinitePhrase("tarefas"), "as tarefas"},
		{"definite plural m", DefinitePhrase("ficheiros"), "os ficheiros"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestDutchComposition — the Germanic shape beside de.json: Bestand
// verwijderd word order, infinitive progress forms, invariant participles,
// de/het by common/neuter gender with de as the universal plural definite,
// and dev-Dutch anglicisms (gepusht).
//
//	T("i18n.done.delete", "file") // "Bestand verwijderd"
func TestDutchComposition(t *testing.T) {
	setCompositionLanguage(t, "nl")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Bestand verwijderd"},
		{"done.send message", T("i18n.done.send", "message"), "Bericht verzonden"},
		{"done.find error", T("i18n.done.find", "error"), "Fout gevonden"},
		{"done.push branch dev-Dutch", T("i18n.done.push", "branch"), "Tak gepusht"},
		{"progress.delete", T("i18n.progress.delete"), "Verwijderen..."},
		{"progress.build", T("i18n.progress.build"), "Bouwen..."},
		{"count.file 5", T("i18n.count.file", 5), "5 bestanden"},
		{"count.child 2 irregular", T("i18n.count.child", 2), "2 kinderen"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Mislukt: pushen tak"},
		{"definite common", DefinitePhrase("task"), "de taak"},
		{"definite neuter", DefinitePhrase("file"), "het bestand"},
		{"indefinite", ArticlePhrase("file"), "een bestand"},
		{"definite plural", DefinitePhrase("bestanden"), "de bestanden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestLatinComposition — the vocabulary comes home. Subject + perfect
// passive participle with elided 'est' is lapidary Latin (TABVLA DELETA),
// the Progress form is the GERUNDIVE in Cato's register (Delenda...), and
// the per-gender agreement schema earns its keep: deletus → deleta (f) →
// deletum (n) from one masculine base.
//
//	T("i18n.done.delete", "file") // "Tabula deleta"
func TestLatinComposition(t *testing.T) {
	setCompositionLanguage(t, "la")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file f", T("i18n.done.delete", "file"), "Tabula deleta"},
		{"done.find error n", T("i18n.done.find", "error"), "Erratum inventum"},
		{"done.send message m", T("i18n.done.send", "message"), "Nuntius missus"},
		{"done.do task n", T("i18n.done.do", "task"), "Opus factum"},
		{"done.resolve... fix vulnerability n", T("i18n.done.fix", "vulnerability"), "Vulnus reparatum"},
		{"progress.delete — Cato", T("i18n.progress.delete"), "Delenda..."},
		{"progress.run — agenda", T("i18n.progress.run"), "Agenda..."},
		{"progress.read — legenda", T("i18n.progress.read"), "Legenda..."},
		{"count.error 3", T("i18n.count.error", 3), "3 errata"},
		{"count.task 2", T("i18n.count.task", 2), "2 opera"},
		{"count.directory 4", T("i18n.count.directory", 4), "4 indices"},
		{"count.category 2", T("i18n.count.category", 2), "2 genera"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Non potuit pellere ramus"},
		{"prompt.yes", T("prompt.yes"), "sic"},
		{"prompt.no", T("prompt.no"), "minime"},
		{"bare article phrase", ArticlePhrase("file"), "tabula"}, // Latin has no articles
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
