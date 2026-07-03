package i18n

// TestRomanianComposition and TestBulgarianComposition exercise the SAME
// i18n.* namespace and English T() keys as the other composition tests
// (locale_composition_test.go) — the ro/bg word bridges and grammar tables
// do the rest. Both locales are the first to exercise SUFFIXED definite
// articles (definite_suffix / definite_suffix_plural): the article welds
// onto the noun instead of preceding it (fișierul, файлът), so alongside
// the usual done/progress/count/fail rows these tests specifically pin the
// suffixing mechanism, including a plural-definite row.

import "testing"

// TestRomanianComposition drives the i18n.* namespace under ro with
// English keys only. Participle agreement is exercised two ways: the
// regular locale-wide rule (creat -> creată) and a per-verb past_f
// override for a stem-mutating irregular (șters -> ștearsă).
//
//	T("i18n.done.delete", "file") // "Fișier șters"
func TestRomanianComposition(t *testing.T) {
	setCompositionLanguage(t, "ro")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file (m)", T("i18n.done.delete", "file"), "Fișier șters"},
		// Feminine agreement via the declared past_f override — șters is a
		// stem-mutating irregular (șters -> ștearsă), not the regular +ă.
		{"done.delete task (f, past_f override)", T("i18n.done.delete", "task"), "Sarcină ștearsă"},
		// Feminine agreement via the regular locale-wide rule (add ă).
		{"done.create task (f, regular)", T("i18n.done.create", "task"), "Sarcină creată"},
		{"done.send message (m)", T("i18n.done.send", "message"), "Mesaj trimis"},
		{"progress.delete", T("i18n.progress.delete"), "Ștergere..."},
		{"progress.create", T("i18n.progress.create"), "Creare..."},
		{"count.file 5", T("i18n.count.file", 5), "5 fișiere"},
		{"count.task 2", T("i18n.count.task", 2), "2 sarcini"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nu s-a putut împinge ramură"},
		{"prompt.yes", T("prompt.yes"), "da"},
		{"prompt.no", T("prompt.no"), "nu"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestRomanianArticleBridge proves the SUFFIXED definite mechanism: the
// article welds onto the noun (fișierul) rather than preceding it, by
// gender (m: -ul, f: strip ă + a), and definite_suffix_plural welds -le
// onto a known plural form (fișiere -> fișierele). The indefinite stays
// a preceding word (un/o), resolved by gender from the same noun table.
//
//	DefinitePhrase("file") // "fișierul"
func TestRomanianArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "ro")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite m", ArticlePhrase("file"), "un fișier"},
		{"indefinite f", ArticlePhrase("task"), "o sarcină"},
		{"suffixed definite m", DefinitePhrase("file"), "fișierul"},
		{"suffixed definite f (strip ă + a)", DefinitePhrase("task"), "sarcina"},
		{"suffixed definite plural (+le)", DefinitePhrase("fișiere"), "fișierele"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestBulgarianComposition drives the i18n.* namespace under bg with
// English keys only. Participle agreement covers all three genders from
// one masculine base (изтрит -> изтрита (f) / изтрито (n)) via the
// locale-wide +а / +о rule — Bulgarian's regular participle agreement
// needed no per-verb override anywhere in this verb table.
//
//	T("i18n.done.delete", "file") // "Файл изтрит"
func TestBulgarianComposition(t *testing.T) {
	setCompositionLanguage(t, "bg")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"done.delete file (m)", T("i18n.done.delete", "file"), "Файл изтрит"},
		{"done.delete task (f, +а)", T("i18n.done.delete", "task"), "Задача изтрита"},
		{"done.delete message (n, +о)", T("i18n.done.delete", "message"), "Съобщение изтрито"},
		{"progress.delete", T("i18n.progress.delete"), "Изтриване..."},
		{"progress.test", T("i18n.progress.test"), "Тестване..."},
		{"count.error 3", T("i18n.count.error", 3), "3 грешки"},
		{"count.task 2", T("i18n.count.task", 2), "2 задачи"},
		{"fail.push branch", T("i18n.fail.push", "branch"), "Неуспешно: бутане клон"},
		{"prompt.yes", T("prompt.yes"), "да"},
		{"prompt.no", T("prompt.no"), "не"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestBulgarianArticleBridge proves the two Bulgarian article mechanics
// together: indefinite_none means ArticlePhrase degrades to the bare noun
// (no indefinite article exists in Bulgarian at all), while the suffixed
// DEFINITE is alive and well — файлът (m), задачата (f) — including the
// plural-definite row (файлове -> файловете, +те on a known plural).
//
//	ArticlePhrase("file")  // "файл"    (bare — no indefinite article)
//	DefinitePhrase("file") // "файлът"  (suffixed definite)
func TestBulgarianArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "bg")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"no indefinite article — bare noun", ArticlePhrase("file"), "файл"},
		{"suffixed definite m (+ът)", DefinitePhrase("file"), "файлът"},
		{"suffixed definite f (+та)", DefinitePhrase("task"), "задачата"},
		{"suffixed definite plural (+те)", DefinitePhrase("файлове"), "файловете"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
