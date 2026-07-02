package i18n

// End-to-end composition through the gram.word bridge for Greek and
// Hungarian — the SAME English T() keys and arguments produce Greek and
// Hungarian output, because the bridge maps the key to the locale's own
// word and the locale's grammar tables conjugate, pluralise, gender/article
// it from there. See locales/el.json and locales/hu.json for the design
// notes (_comment) behind each documented composition quirk asserted here.

import "testing"

// TestGreekComposition drives the i18n.* namespace under el with English
// keys only. The done-shape is a bare 3rd-person passive aorist VERB (not a
// participle, and Greek declares no participle agreement), so subject +
// past composes article-less: "Αρχείο διαγράφηκε". Progress is the deverbal
// NOUN real Greek software shows ("Διαγραφή..."), not a verb form. The
// article proofs pin the by_gender indefinite/definite tables including the
// merged m/f plural definite "οι" (genuine Greek grammar, not a bug).
//
//	T("i18n.done.delete", "file") // "Αρχείο διαγράφηκε"
func TestGreekComposition(t *testing.T) {
	setCompositionLanguage(t, "el")

	tests := []struct {
		name string
		got  string
		want string
	}{
		// done.* — subject + 3sg passive aorist, no agreement (invariant).
		{"done.delete file", T("i18n.done.delete", "file"), "Αρχείο διαγράφηκε"},
		{"done.create branch", T("i18n.done.create", "branch"), "Κλάδος δημιουργήθηκε"},
		{"done.save message", T("i18n.done.save", "message"), "Μήνυμα αποθηκεύτηκε"},
		{"done.fix error", T("i18n.done.fix", "error"), "Σφάλμα διορθώθηκε"},

		// progress.* — the deverbal noun, not a verb gerund.
		{"progress.delete", T("i18n.progress.delete"), "Διαγραφή..."},
		{"progress.build", T("i18n.progress.build"), "Κατασκευή..."},
		{"progress.save", T("i18n.progress.save"), "Αποθήκευση..."},

		// count.* — plural proof, including the run/build/check noun
		// homograph (noun table outranks the verb bridge).
		{"count.file 5", T("i18n.count.file", 5), "5 αρχεία"},
		{"count.error 3", T("i18n.count.error", 3), "3 σφάλματα"},
		{"count.run 4 homograph noun", T("i18n.count.run", 4), "4 εκτελέσεις"},

		// fail.* — documented mechanical quirk: "Δεν ήταν δυνατή" wants a
		// nominalised action in natural Greek; the engine appends the bare
		// dictionary-form (1sg present) verb instead, since Greek has no
		// infinitive to fall back on. See locales/el.json _comment.
		{"fail.push branch", T("i18n.fail.push", "branch"), "Δεν ήταν δυνατή προωθώ κλάδος"},

		// prompt.* — short native forms, as real Greek CLIs write them.
		{"prompt.yes", T("prompt.yes"), "ναι"},
		{"prompt.no", T("prompt.no"), "όχι"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestGreekArticleBridge pins the by_gender indefinite (ένας/μία/ένα) and
// definite (ο/η/το) tables resolved from English keys through the bridge,
// plus the plural definite where masculine and feminine genuinely merge
// into "οι" (only neuter takes its own "τα") — real Greek grammar, not a
// simplification.
//
//	DefinitePhrase("file") // "το αρχείο"
func TestGreekArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "el")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"indefinite m", ArticlePhrase("server"), "ένας διακομιστής"},
		{"indefinite f", ArticlePhrase("task"), "μία εργασία"},
		{"indefinite n", ArticlePhrase("file"), "ένα αρχείο"},
		{"definite n", DefinitePhrase("file"), "το αρχείο"},
		{"definite f", DefinitePhrase("task"), "η εργασία"},
		{"definite m", DefinitePhrase("server"), "ο διακομιστής"},
		{"definite plural n, m/f merge to οι", DefinitePhrase("αρχεία"), "τα αρχεία"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestHungarianComposition drives the i18n.* namespace under hu with
// English keys only. The done-shape is the -va/-ve adverbial participle in
// the register every real Hungarian app status uses ("Fájl törölve"), and
// progress is the -ás/-és deverbal noun ("Törlés..."). No gender, no
// agreement. The headline article proof is the az/a phonetic definite
// split (definite_vowel), tested on a vowel-initial noun (adat) against a
// consonant-initial one (fájl).
//
//	T("i18n.done.delete", "file") // "Fájl törölve"
func TestHungarianComposition(t *testing.T) {
	setCompositionLanguage(t, "hu")

	tests := []struct {
		name string
		got  string
		want string
	}{
		// done.* — subject + -va/-ve participle, no gender/agreement.
		{"done.delete file", T("i18n.done.delete", "file"), "Fájl törölve"},
		{"done.save message", T("i18n.done.save", "message"), "Üzenet mentve"},
		{"done.create branch", T("i18n.done.create", "branch"), "Ág létrehozva"},
		{"done.install package", T("i18n.done.install", "package"), "Csomag telepítve"},

		// progress.* — the -ás/-és deverbal noun.
		{"progress.delete", T("i18n.progress.delete"), "Törlés..."},
		{"progress.save", T("i18n.progress.save"), "Mentés..."},
		{"progress.install", T("i18n.progress.install"), "Telepítés..."},

		// count.* — Hungarian numerals govern the SINGULAR (5 fájl, never
		// 5 fájlok): counts_use_singular. The plural forms still serve the
		// definite plural (a fájlok, az adatok).
		{"count.file 5", T("i18n.count.file", 5), "5 fájl"},
		{"count.error 3", T("i18n.count.error", 3), "3 hiba"},
		{"count.run 4 homograph noun", T("i18n.count.run", 4), "4 futtatás"},

		// fail.* — documented mechanical quirk: "Nem sikerült" wants a
		// following infinitive (törölni) in natural Hungarian; the engine
		// appends the bare dictionary-form (3sg present) verb instead. See
		// locales/hu.json _comment.
		{"fail.push branch", T("i18n.fail.push", "branch"), "Nem sikerült felküld ág"},

		// prompt.* — igen/nem, the short native forms real Hungarian CLIs
		// write.
		{"prompt.yes", T("prompt.yes"), "igen"},
		{"prompt.no", T("prompt.no"), "nem"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestHungarianArticleBridge pins the headline az/a split: "az" before a
// vowel-initial word (adat), "a" before a consonant-initial one (fájl),
// tested on the letter per docs/locale-schema.md. The indefinite "egy" is
// phonetically invariant (default and vowel forms are identical), proven
// across both nouns.
//
//	DefinitePhrase("data") // "az adat"
//	DefinitePhrase("file") // "a fájl"
func TestHungarianArticleBridge(t *testing.T) {
	setCompositionLanguage(t, "hu")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"definite vowel-initial: az adat", DefinitePhrase("data"), "az adat"},
		{"definite consonant-initial: a fájl", DefinitePhrase("file"), "a fájl"},
		{"definite PLURAL vowel-initial: az adatok", DefinitePhrase("adatok"), "az adatok"},
		{"definite plural consonant-initial: a fájlok", DefinitePhrase("fájlok"), "a fájlok"},
		{"indefinite invariant vowel-initial", ArticlePhrase("data"), "egy adat"},
		{"indefinite invariant consonant-initial", ArticlePhrase("file"), "egy fájl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
