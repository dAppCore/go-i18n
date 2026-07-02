package i18n

// Base English in this package is the English of England: bare `en` IS
// en-GB, and American English is the override locale (locales/en-US.json).
// These tests pin the dialect boundary — the en-GB defaults, the en-US
// re-hearings — and the phonology-driven morphology the tables and rules
// now encode (qu as /kw/, soft-g e-retention, the closed f→ves class,
// the /juː/ glide words).

import "testing"

// TestArticleEnglishGBPhonetics pins Article() to en-GB pronunciation under
// the default locale: /juː/ and /w/ onsets take "a" however they are
// spelled, silent-h words take "an", and herb keeps its British /h/.
//
//	Article("unicorn") // "a"  (/juː/)
//	Article("herb")    // "a"  (en-GB sounds the h)
func TestArticleEnglishGBPhonetics(t *testing.T) {
	prev := Default()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	tests := []struct {
		word string
		want string
	}{
		{"unicorn", "a"},   // /juː/ glide, not a vowel onset
		{"ewe", "a"},       // /juː/ — sounds like "you"
		{"U-turn", "a"},    // letter-name hyphenation, /juː/
		{"unicode", "a"},   // /juː/
		{"European", "a"},  // /jʊər/
		{"one-off", "a"},   // /w/ onset
		{"ouija", "a"},     // /w/ onset
		{"user", "a"},      // /juː/
		{"hour", "an"},     // silent h
		{"honest", "an"},   // silent h
		{"heir", "an"},     // silent h
		{"honour", "an"},   // silent h, UK spelling
		{"X-ray", "an"},    // letter-name /ɛks/
		{"Xbox", "an"},     // /ɛks/
		{"herb", "a"},      // en-GB sounds the h; en-US re-hears it via locale data
		{"umbrella", "an"}, // plain vowel onset
		{"error", "an"},    // plain vowel onset
		{"file", "a"},      // plain consonant onset
	}
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := Article(tt.word); got != tt.want {
				t.Errorf("Article(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

// TestArticleEnUSLocaleOverride proves the dialect seam: switching to en-US
// re-hears "herb" as vowel-onset (silent h) through the locale's
// vowel_sound_words list, while the shared phonetic truths stay put.
//
//	Article("herb") // "an" under en-US
func TestArticleEnUSLocaleOverride(t *testing.T) {
	prev := Default()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := errorFromResult(SetLanguage("en-US")); err != nil {
		t.Fatalf("SetLanguage(en-US) failed: %v", err)
	}

	if got, want := Article("herb"), "an"; got != want {
		t.Errorf("Article(\"herb\") under en-US = %q, want %q", got, want)
	}
	// Dialect-neutral phonetics are unchanged by the override file.
	if got, want := Article("unicorn"), "a"; got != want {
		t.Errorf("Article(\"unicorn\") under en-US = %q, want %q", got, want)
	}
	if got, want := Article("hour"), "an"; got != want {
		t.Errorf("Article(\"hour\") under en-US = %q, want %q", got, want)
	}
}

// TestVerbFormsEnUSLocaleOverride pins the American single-l and e-drop
// forms carried by locales/en-US.json, and that verbs absent from the
// override fall through to the shared tables.
//
//	PastTense("travel") // "traveled" under en-US
func TestVerbFormsEnUSLocaleOverride(t *testing.T) {
	prev := Default()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := errorFromResult(SetLanguage("en-US")); err != nil {
		t.Fatalf("SetLanguage(en-US) failed: %v", err)
	}

	pastTests := []struct{ verb, want string }{
		{"travel", "traveled"},
		{"cancel", "canceled"},
		{"total", "totaled"},
		{"marshal", "marshaled"},
		{"dial", "dialed"},
		{"commit", "committed"}, // not in en-US.json → shared tables
	}
	for _, tt := range pastTests {
		if got := PastTense(tt.verb); got != tt.want {
			t.Errorf("PastTense(%q) under en-US = %q, want %q", tt.verb, got, tt.want)
		}
	}
	if got, want := Gerund("age"), "aging"; got != want {
		t.Errorf("Gerund(\"age\") under en-US = %q, want %q", got, want)
	}
}

// TestGerundSoftGRetention pins the soft-g gerunds that keep their "e" so
// /dʒ/ survives, against the regular e-drop they would otherwise get —
// singeing is scorching; singing is what choirs do.
//
//	Gerund("singe") // "singeing"
func TestGerundSoftGRetention(t *testing.T) {
	tests := []struct{ verb, want string }{
		{"singe", "singeing"},
		{"whinge", "whingeing"},
		{"binge", "bingeing"},
		{"tinge", "tingeing"},
		{"cringe", "cringing"}, // regular e-drop: no collision, no retention
		{"change", "changing"}, // regular e-drop
		{"age", "ageing"},      // en-GB e-retention (en-US overrides to "aging")
		{"dye", "dyeing"},      // y-exclusion keeps the e: dyeing ≠ dying
		{"canoe", "canoeing"},  // o-exclusion keeps the e
	}
	for _, tt := range tests {
		if got := Gerund(tt.verb); got != tt.want {
			t.Errorf("Gerund(%q) = %q, want %q", tt.verb, got, tt.want)
		}
	}
}

// TestPastTenseUKDoublingAndQuOnset pins the en-GB final-l doubling and the
// qu-as-/kw/ correction: "qu" is an onset cluster, so quiz and equal are CVC
// shapes and double their final consonant.
//
//	PastTense("marshal") // "marshalled"
//	PastTense("quiz")    // "quizzed"
func TestPastTenseUKDoublingAndQuOnset(t *testing.T) {
	tests := []struct{ verb, want string }{
		{"marshal", "marshalled"},
		{"unmarshal", "unmarshalled"},
		{"signal", "signalled"},
		{"total", "totalled"},
		{"equal", "equalled"},
		{"spiral", "spiralled"},
		{"pencil", "pencilled"},
		{"enrol", "enrolled"},
		{"fulfil", "fulfilled"},
		{"dial", "dialled"},
		{"fuel", "fuelled"},
		{"quiz", "quizzed"},
		{"equip", "equipped"},
		{"parallel", "paralleled"}, // both dialects prefer single l here
		{"reveal", "revealed"},     // vowel digraph — no doubling
		{"email", "emailed"},       // vowel digraph — no doubling
		{"curl", "curled"},         // consonant before l — no doubling
	}
	for _, tt := range tests {
		if got := PastTense(tt.verb); got != tt.want {
			t.Errorf("PastTense(%q) = %q, want %q", tt.verb, got, tt.want)
		}
	}
}

// TestPluralFormClosedVesClass pins the f→ves plural as a closed Old English
// class: the enumerated survivors voice the fricative, everything else —
// modern words, loans, ff-finals — takes plain -s.
//
//	PluralForm("wolf") // "wolves"
//	PluralForm("roof") // "roofs"
func TestPluralFormClosedVesClass(t *testing.T) {
	tests := []struct{ noun, want string }{
		{"wolf", "wolves"},
		{"knife", "knives"},
		{"leaf", "leaves"},
		{"elf", "elves"},
		{"hoof", "hooves"},
		{"scarf", "scarves"},
		{"sheaf", "sheaves"},
		{"wharf", "wharves"},
		{"roof", "roofs"},
		{"chief", "chiefs"},
		{"belief", "beliefs"},
		{"safe", "safes"},
		{"chef", "chefs"},
		{"cliff", "cliffs"},
		{"staff", "staffs"},
		{"dwarf", "dwarfs"}, // Tolkien pluralised his own way; standard is -s
		{"quiz", "quizzes"},
		{"whiz", "whizzes"},
		{"waltz", "waltzes"},
		{"buzz", "buzzes"},
	}
	for _, tt := range tests {
		if got := PluralForm(tt.noun); got != tt.want {
			t.Errorf("PluralForm(%q) = %q, want %q", tt.noun, got, tt.want)
		}
	}
}
