package i18n

import "testing"

// TestMalteseComposition pins the ASSIMILATING definite article — Arabic
// sun letters inside the EU: il- is the moon default, but the article
// agrees with the noun's first letter (ix-xogħol, iż-żball, it-test).
// Maltese has no indefinite article.
//
//	DefinitePhrase("task") // "ix-xogħol"
func TestMalteseComposition(t *testing.T) {
	setCompositionLanguage(t, "mt")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Fajl imħassar"},
		{"done.create change f agreement", T("i18n.done.create", "change"), "Bidla maħluqa"},
		{"done.delete branch f override", T("i18n.done.delete", "branch"), "Fergħa imħassra"},
		{"progress.delete", T("i18n.progress.delete"), "Tħassir..."},
		{"count.error 3 broken plural", T("i18n.count.error", 3), "3 żbalji"},
		{"fail.send message", T("i18n.fail.send", "message"), "Ma rnexxiex bagħat messaġġ"},
		{"moon letter", DefinitePhrase("file"), "il-fajl"},
		{"sun letter x", DefinitePhrase("task"), "ix-xogħol"},
		{"sun letter ż", DefinitePhrase("error"), "iż-żball"},
		{"sun letter t", DefinitePhrase("test"), "it-test"},
		{"no indefinite", ArticlePhrase("file"), "fajl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestIrishComposition pins the definite-article LENITION — the article
// mutates the noun: craobh → an chraobh, fadhb → an fhadhb — while d/t/s
// resist (an teachtaireacht, DNTLS) and masculine nouns stay unmutated
// (an comhad). The done-form is the invariant verbal adjective and Irish
// has no indefinite article.
//
//	DefinitePhrase("issue") // "an fhadhb"
func TestIrishComposition(t *testing.T) {
	setCompositionLanguage(t, "ga")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Comhad scriosta"},
		{"done.send message", T("i18n.done.send", "message"), "Teachtaireacht seolta"},
		{"done.find error", T("i18n.done.find", "error"), "Earráid aimsithe"},
		{"progress.delete", T("i18n.progress.delete"), "Ag scriosadh..."},
		{"progress.update", T("i18n.progress.update"), "Ag nuashonrú..."},
		{"count.task 3", T("i18n.count.task", 3), "3 tascanna"},
		{"fail.send message", T("i18n.fail.send", "message"), "Níorbh fhéidir seol teachtaireacht"},
		{"masculine unmutated", DefinitePhrase("file"), "an comhad"},
		{"feminine c lenites", DefinitePhrase("category"), "an chatagóir"},
		{"feminine f lenites", DefinitePhrase("issue"), "an fhadhb"},
		{"feminine t resists (DNTLS)", DefinitePhrase("message"), "an teachtaireacht"},
		{"feminine vowel untouched", DefinitePhrase("error"), "an earráid"},
		{"plural definite", DefinitePhrase("comhaid"), "na comhaid"},
		{"no indefinite", ArticlePhrase("file"), "comhad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
