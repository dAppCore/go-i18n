package i18n

// End-to-end composition through the gram.word bridge for the two Scandi
// locales: the SAME English T() keys and arguments produce Danish and
// Swedish output, because the bridge maps the key to the locale's own word
// and the locale's grammar tables conjugate, pluralise, gender and suffix
// it from there. Danish and Swedish sit beside nl.json/de.json as the other
// Germanic common/neuter (c/n) pair, but weld their definite article onto
// the noun instead of fronting it — the SUFFIXED definite mechanism.

import "testing"

// TestDanishComposition drives the i18n.* namespace under da with English
// keys only. Danish word order matches the composed shape natively (subject
// + supine is exactly "Fil slettet"), the progress form is the PRESENT
// TENSE by real UI convention ("Sletter..."), and participles never agree
// with the subject's gender — da.json declares no gram.agreement block, so
// a neuter subject (Problem) keeps the same bare supine as a common one.
//
//	T("i18n.done.delete", "file") // "Fil slettet"
func TestDanishComposition(t *testing.T) {
	setCompositionLanguage(t, "da")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Fil slettet"},
		{"done.create task", T("i18n.done.create", "task"), "Opgave oprettet"},
		{"done.send message", T("i18n.done.send", "message"), "Besked sendt"},
		{"done.find error", T("i18n.done.find", "error"), "Fejl fundet"},
		{"done.push branch dev-Danish loanword", T("i18n.done.push", "branch"), "Gren pushet"},
		// Neuter subject, no gram.agreement declared: the supine stays bare —
		// contrast with Swedish's "Paket raderat" below.
		{"done.fix issue neuter no agreement", T("i18n.done.fix", "issue"), "Problem rettet"},
		{"progress.build", T("i18n.progress.build"), "Bygger..."},
		{"progress.delete", T("i18n.progress.delete"), "Sletter..."},
		{"count.file 5", T("i18n.count.file", 5), "5 filer"},
		{"count.child 2 irregular", T("i18n.count.child", 2), "2 børn"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Kunne ikke pushe gren"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestDanishArticleBridge proves the SUFFIXED definite mechanism: fil →
// filen welds the gendered suffix onto the noun instead of fronting a
// standalone word, filer → filerne applies the plural rule to a known
// plural form, and the indefinite en/et split still fronts normally since
// only the DEFINITE article is suffixed in Danish.
//
//	DefinitePhrase("file") // "filen"
func TestDanishArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "da")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite common", ArticlePhrase("file"), "en fil"},
		{"indefinite neuter", ArticlePhrase("issue"), "et problem"},
		{"definite suffix common", DefinitePhrase("file"), "filen"},
		{"definite suffix neuter", DefinitePhrase("issue"), "problemet"},
		{"definite suffix plural", DefinitePhrase("filer"), "filerne"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestSwedishComposition drives the i18n.* namespace under sv with English
// keys only. The gerund slot holds the PRESENT TENSE by real UI convention
// ("Raderar..."), and — unlike Danish — participles DO agree, but only for
// a NEUTER subject: sv.json declares agreement.participle.n as strip "d" /
// add "t", so raderad (common, unmarked) becomes raderat (neuter) while
// every other gender stays on the base form.
//
//	T("i18n.done.delete", "file")    // "Fil raderad"    (common, unmarked)
//	T("i18n.done.delete", "package") // "Paket raderat"  (neuter, agreed)
func TestSwedishComposition(t *testing.T) {
	setCompositionLanguage(t, "sv")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file common unmarked", T("i18n.done.delete", "file"), "Fil raderad"},
		// The required neuter-agreement proof: strip "d" / add "t" turns the
		// common-gender base form raderad into the neuter-agreed raderat.
		{"done.delete package neuter agreement", T("i18n.done.delete", "package"), "Paket raderat"},
		{"done.create task common", T("i18n.done.create", "task"), "Uppgift skapad"},
		{"done.send message neuter agreement", T("i18n.done.send", "message"), "Meddelande skickat"},
		{"done.push branch dev-Swedish loanword", T("i18n.done.push", "branch"), "Gren pushad"},
		{"progress.delete", T("i18n.progress.delete"), "Raderar..."},
		{"progress.build", T("i18n.progress.build"), "Bygger..."},
		{"count.file 5", T("i18n.count.file", 5), "5 filer"},
		{"count.child 2 invariant plural", T("i18n.count.child", 2), "2 barn"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Kunde inte pusha gren"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestSwedishArticleBridge proves the SUFFIXED definite mechanism beside
// Danish's: fil → filen, filer → filerna (the "na" plural suffix, not
// Danish "ne"), and the en/ett indefinite split — Swedish's neuter
// INDEFINITE article is the double-t "ett", but the definite SUFFIX is the
// single-t "-et" (paket → paketet), a genuine asymmetry in the language,
// not a typo.
//
//	DefinitePhrase("file") // "filen"
func TestSwedishArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "sv")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite common", ArticlePhrase("file"), "en fil"},
		{"indefinite neuter double-t", ArticlePhrase("package"), "ett paket"},
		{"definite suffix common", DefinitePhrase("file"), "filen"},
		{"definite suffix neuter single-t", DefinitePhrase("package"), "paketet"},
		{"definite suffix plural na", DefinitePhrase("filer"), "filerna"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
