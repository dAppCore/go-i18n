// SPDX-License-Identifier: EUPL-1.2

package i18n

// Turkish composition: verb-final, article-less, genderless, numerals take
// the singular, and i upper-cases to İ. Every one of those is declared in
// locales/tr.json (gram.order, gram.article, gram.number, gram.case); the
// engine has no Turkish code. Fallback is en, the English of England.

import "testing"

// TestTurkishComposition drives the i18n.* namespace under tr with English
// keys only.
//
//	T("i18n.done.delete", "file") // "Dosya silindi"
//	T("i18n.fail.delete", "file") // "Dosya silinemedi"
func TestTurkishComposition(t *testing.T) {
	setCompositionLanguage(t, "tr")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file", T("i18n.done.delete", "file"), "Dosya silindi"},
		{"done.send message", T("i18n.done.send", "message"), "Mesaj gönderildi"},
		{"done.download, no subject", T("i18n.done.download"), "İndirildi"},
		{"progress.build", T("i18n.progress.build"), "Derleniyor..."},
		{"progress.download — dotted İ", T("i18n.progress.download"), "İndiriliyor..."},
		{"progress.import — dotted İ, sentence case", T("i18n.progress.import"), "İçe aktarılıyor..."},
		{"progress.build project — verb last", T("i18n.progress.build", "project"), "Proje derleniyor..."},
		{"fail.delete file — impotential, verb last", T("i18n.fail.delete", "file"), "Dosya silinemedi"},
		{"fail.find file", T("i18n.fail.find", "file"), "Dosya bulunamadı"},
		{"fail.download, no subject", T("i18n.fail.download"), "İndirilemedi"},
		{"count.file 3 — numeral takes the singular", T("i18n.count.file", 3), "3 dosya"},
		{"count.error 1", T("i18n.count.error", 1), "1 hata"},
		{"count.file 1200 — dot for thousands", T("i18n.count.file", 1200), "1.200 dosya"},
		{"bare article phrase", ArticlePhrase("file"), "dosya"}, // no articles in Turkish
		{"plural outside a count", Pluralize("file", 2), "dosyalar"},
		{"title case — dotted İ", Title("içe aktar"), "İçe Aktar"},
		{"title case — dotless ı stays I", Title("ılık"), "Ilık"},
		{"prompt", T("prompt.yes"), "Evet"},
		{"English fallback for an unauthored verb", T("i18n.progress.deploy"), "Deploying..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestTurkishCaseIsLocaleOnly proves the dotted i is Turkish's alone: the
// same Title under en keeps Unicode's I.
func TestTurkishCaseIsLocaleOnly(t *testing.T) {
	setCompositionLanguage(t, "en")
	if got := Title("install"); got != "Install" {
		t.Errorf("en Title(install) = %q, want Install", got)
	}
	if got := T("i18n.fail.delete", "file"); got != "Failed to delete file" {
		t.Errorf("en fail = %q: a failed form must not leak out of tr", got)
	}
	if got := T("i18n.progress.build", "project"); got != "Building project..." {
		t.Errorf("en progress = %q: verb-final must not leak out of tr", got)
	}
}
