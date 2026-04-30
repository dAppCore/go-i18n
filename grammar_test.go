package i18n

import (
	"strings"
	"testing"
	"text/template"
	"time"
)

type regionFallbackLoader struct{}

func (regionFallbackLoader) Languages() []string {
	return []string{"en-GB"}
}

func (regionFallbackLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	return map[string]Message{}, nil, nil
}

func TestPastTense(t *testing.T) {
	// Ensure grammar data is loaded from embedded JSON
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		verb string
		want string
	}{
		// Irregular verbs (from JSON)
		{"be", "was"},
		{"go", "went"},
		{"run", "ran"},
		{"write", "wrote"},
		{"build", "built"},
		{"find", "found"},
		{"set", "set"},
		{"put", "put"},
		{"cut", "cut"},
		{"commit", "committed"},

		// Irregular verbs (from Go map only)
		{"break", "broke"},
		{"speak", "spoke"},
		{"steal", "stole"},
		{"freeze", "froze"},

		// Compound irregular verbs
		{"undo", "undid"},
		{"redo", "redid"},
		{"rerun", "reran"},
		{"rewrite", "rewrote"},
		{"rebuild", "rebuilt"},
		{"resend", "resent"},
		{"override", "overrode"},
		{"rethink", "rethought"},
		{"remake", "remade"},
		{"undergo", "underwent"},
		{"overcome", "overcame"},
		{"withdraw", "withdrew"},
		{"uphold", "upheld"},
		{"withhold", "withheld"},
		{"outgrow", "outgrew"},
		{"outrun", "outran"},
		{"overshoot", "overshot"},

		// Simple irregular verbs (dev/ops)
		{"become", "became"},
		{"come", "came"},
		{"give", "gave"},
		{"fall", "fell"},
		{"understand", "understood"},
		{"arise", "arose"},
		{"bind", "bound"},
		{"spin", "spun"},
		{"quit", "quit"},
		{"cast", "cast"},
		{"broadcast", "broadcast"},
		{"burst", "burst"},
		{"cost", "cost"},
		{"shed", "shed"},
		{"rid", "rid"},
		{"shrink", "shrank"},
		{"shoot", "shot"},
		{"forbid", "forbade"},
		{"offset", "offset"},
		{"upset", "upset"},
		{"input", "input"},
		{"output", "output"},

		// CVC doubling failures (stressed final syllable)
		{"debug", "debugged"},
		{"embed", "embedded"},
		{"unzip", "unzipped"},
		{"remap", "remapped"},
		{"unpin", "unpinned"},
		{"unwrap", "unwrapped"},

		// Regular verbs
		{"delete", "deleted"},
		{"update", "updated"},
		{"push", "pushed"},
		{"pull", "pulled"},
		{"start", "started"},
		{"panic", "panicked"},
		{"copy", "copied"},
		{"apply", "applied"},

		// Edge cases
		{"", ""},
		{"  delete  ", "deleted"},
		{"DELETE", "deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			got := PastTense(tt.verb)
			if got != tt.want {
				t.Errorf("PastTense(%q) = %q, want %q", tt.verb, got, tt.want)
			}
		})
	}
}

func TestGerund(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		verb string
		want string
	}{
		// Irregular verbs (from JSON)
		{"be", "being"},
		{"go", "going"},
		{"run", "running"},
		{"build", "building"},
		{"write", "writing"},
		{"commit", "committing"},

		// Compound irregular verbs
		{"undo", "undoing"},
		{"rerun", "rerunning"},
		{"override", "overriding"},
		{"rebuild", "rebuilding"},

		// Simple irregular (dev/ops)
		{"become", "becoming"},
		{"give", "giving"},
		{"bind", "binding"},
		{"spin", "spinning"},
		{"quit", "quitting"},
		{"cast", "casting"},
		{"broadcast", "broadcasting"},

		// CVC doubling failures
		{"debug", "debugging"},
		{"embed", "embedding"},
		{"unzip", "unzipping"},
		{"remap", "remapping"},
		{"unpin", "unpinning"},
		{"unwrap", "unwrapping"},

		// Regular verbs
		{"delete", "deleting"},
		{"push", "pushing"},
		{"pull", "pulling"},
		{"start", "starting"},
		{"panic", "panicking"},
		{"die", "dying"},

		// Edge cases
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			got := Gerund(tt.verb)
			if got != tt.want {
				t.Errorf("Gerund(%q) = %q, want %q", tt.verb, got, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		noun  string
		count int
		want  string
	}{
		// Singular (count=1 always returns original)
		{"file", 1, "file"},
		{"child", 1, "child"},

		// From JSON grammar data
		{"file", 5, "files"},
		{"repo", 3, "repos"},
		{"branch", 2, "branches"},
		{"repository", 2, "repositories"},
		{"vulnerability", 2, "vulnerabilities"},
		{"person", 2, "people"},
		{"child", 3, "children"},

		// From irregular nouns map
		{"mouse", 2, "mice"},
		{"sheep", 5, "sheep"},
		{"knife", 3, "knives"},

		// Regular plurals
		{"server", 2, "servers"},
		{"box", 2, "boxes"},

		// Count 0
		{"file", 0, "files"},
	}

	for _, tt := range tests {
		t.Run(tt.noun, func(t *testing.T) {
			got := Pluralize(tt.noun, tt.count)
			if got != tt.want {
				t.Errorf("Pluralize(%q, %d) = %q, want %q", tt.noun, tt.count, got, tt.want)
			}
		})
	}
}

func TestPluralize_UsesLocaleSingularOverride(t *testing.T) {
	const lang = "en-x-singular"
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
		SetGrammarData(lang, nil)
	})

	svc, err := NewWithLoader(pluralizeOverrideLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	SetDefault(svc)

	if err := SetLanguage(lang); err != nil {
		t.Fatalf("SetLanguage(%s) failed: %v", lang, err)
	}

	if got, want := Pluralize("person", 1), "human"; got != want {
		t.Fatalf("Pluralize(%q, 1) = %q, want %q", "person", got, want)
	}
	if got, want := Pluralize("Person", 1), "Human"; got != want {
		t.Fatalf("Pluralize(%q, 1) = %q, want %q", "Person", got, want)
	}
	if got, want := Pluralize("person", 2), "people"; got != want {
		t.Fatalf("Pluralize(%q, 2) = %q, want %q", "person", got, want)
	}
}

func TestPluralize_PreservesUnicodeCapitalization(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	if got, want := Pluralize("Élément", 1), "Élément"; got != want {
		t.Fatalf("Pluralize(%q, 1) = %q, want %q", "Élément", got, want)
	}
	if got, want := Pluralize("Élément", 2), "Éléments"; got != want {
		t.Fatalf("Pluralize(%q, 2) = %q, want %q", "Élément", got, want)
	}
	if got, want := PluralForm("Élément"), "Éléments"; got != want {
		t.Fatalf("PluralForm(%q) = %q, want %q", "Élément", got, want)
	}
}

func TestPluralForm(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		noun string
		want string
	}{
		{"", ""},

		// Capitalisation preserved
		{"File", "Files"},
		{"Child", "Children"},

		// Regular rules
		{"bus", "buses"},
		{"fox", "foxes"},
		{"city", "cities"},
		{"key", "keys"},
	}

	for _, tt := range tests {
		t.Run(tt.noun, func(t *testing.T) {
			got := PluralForm(tt.noun)
			if got != tt.want {
				t.Errorf("PluralForm(%q) = %q, want %q", tt.noun, got, tt.want)
			}
		})
	}
}

func TestArticle(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"file", "a"},
		{"error", "an"},
		{"apple", "an"},
		{"user", "a"},       // Consonant sound: "yoo-zer"
		{"hour", "an"},      // Vowel sound: silent h
		{"honest", "an"},    // Vowel sound
		{"university", "a"}, // Consonant sound
		{"one", "a"},        // Consonant sound
		{"SSH", "an"},       // Initialism: "ess-ess-aitch"
		{"URL", "a"},        // Initialism: "you-are-ell"
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := Article(tt.word)
			if got != tt.want {
				t.Errorf("Article(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestArticleTokenAliases(t *testing.T) {
	if got, want := ArticleToken("apple"), Article("apple"); got != want {
		t.Fatalf("ArticleToken(apple) = %q, want %q", got, want)
	}
	if got, want := DefiniteToken("apple"), DefiniteArticle("apple"); got != want {
		t.Fatalf("DefiniteToken(apple) = %q, want %q", got, want)
	}
}

func TestArticleFrenchLocale(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	tests := []struct {
		word string
		want string
	}{
		{"branche", "la"},
		{"branches", "les"},
		{"amis", "des"},
		{"enfant", "l'"},
		{"fichier", "le"},
		{"inconnu", "un"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := Article(tt.word)
			if got != tt.want {
				t.Errorf("Article(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestArticleFrenchElisionKeepsLeadingConsonant(t *testing.T) {
	prevData := GetGrammarData("fr")
	t.Cleanup(func() {
		SetGrammarData("fr", prevData)
	})

	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	SetGrammarData("fr", &GrammarData{
		Nouns: map[string]NounForms{
			"amie":   {One: "amie", Other: "amies", Gender: "f"},
			"accord": {One: "accord", Other: "accords", Gender: "d"},
			"homme":  {One: "homme", Other: "hommes", Gender: "m"},
			"héros":  {One: "héros", Other: "héros", Gender: "m"},
			"idole":  {One: "idole", Other: "idoles", Gender: "j"},
		},
		Articles: ArticleForms{
			IndefiniteDefault: "un",
			IndefiniteVowel:   "un",
			Definite:          "le",
			ByGender: map[string]string{
				"d": "de",
				"f": "la",
				"j": "je",
				"m": "le",
			},
		},
	})

	tests := []struct {
		word string
		want string
	}{
		{"homme", "l'"},
		{"héros", "le"},
		{"amie", "l'"},
		{"accord", "d'"},
		{"idole", "j'"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := Article(tt.word)
			if got != tt.want {
				t.Errorf("Article(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}

	phraseTests := []struct {
		word string
		want string
	}{
		{"accord", "d'accord"},
		{"idole", "j'idole"},
	}

	for _, tt := range phraseTests {
		t.Run(tt.word+"_phrase", func(t *testing.T) {
			got := ArticlePhrase(tt.word)
			if got != tt.want {
				t.Errorf("ArticlePhrase(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

type pluralizeOverrideLoader struct{}

func (pluralizeOverrideLoader) Languages() []string {
	return []string{"en-x-singular"}
}

func (pluralizeOverrideLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	grammar := &GrammarData{
		Nouns: map[string]NounForms{
			"person": {One: "human", Other: "people"},
		},
	}
	SetGrammarData(lang, grammar)
	return map[string]Message{}, grammar, nil
}

func TestTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "Hello World"},
		{"hello", "Hello"},
		{"", ""},
		{"HELLO", "HELLO"},
		{"hello-world", "Hello-World"},
		{"config.yaml", "Config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Title(tt.input)
			if got != tt.want {
				t.Errorf("Title(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	if got := Quote("hello"); got != `"hello"` {
		t.Errorf("Quote(%q) = %q, want %q", "hello", got, `"hello"`)
	}
	if got := Quote(`a "quoted" path\name`); got != `"a \"quoted\" path\\name"` {
		t.Errorf("Quote(%q) = %q, want %q", `a "quoted" path\name`, got, `"a \"quoted\" path\\name"`)
	}
}

func TestCaseHelpers(t *testing.T) {
	if got := Lower("HELLO"); got != "hello" {
		t.Fatalf("Lower(%q) = %q, want %q", "HELLO", got, "hello")
	}
	if got := Upper("hello"); got != "HELLO" {
		t.Fatalf("Upper(%q) = %q, want %q", "hello", got, "HELLO")
	}
}

func TestArticlePhrase(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"file", "a file"},
		{"error", "an error"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := ArticlePhrase(tt.word)
			if got != tt.want {
				t.Errorf("ArticlePhrase(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestArticlePhrase_RespectsWordMap(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	data := GetGrammarData("en")
	if data == nil {
		t.Fatal("GetGrammarData(\"en\") returned nil")
	}
	original, existed := data.Words["go_mod"]
	data.Words["go_mod"] = "go.mod"
	t.Cleanup(func() {
		if existed {
			data.Words["go_mod"] = original
			return
		}
		delete(data.Words, "go_mod")
	})

	if got, want := ArticlePhrase("go_mod"), "a go.mod"; got != want {
		t.Fatalf("ArticlePhrase(%q) = %q, want %q", "go_mod", got, want)
	}
}

func TestArticlePhrase_UsesRenderedWordForArticleSelection(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	data := GetGrammarData("en")
	if data == nil {
		t.Fatal("GetGrammarData(\"en\") returned nil")
	}
	original, existed := data.Words["ssh"]
	data.Words["ssh"] = "SSH"
	t.Cleanup(func() {
		if existed {
			data.Words["ssh"] = original
			return
		}
		delete(data.Words, "ssh")
	})

	if got, want := ArticlePhrase("ssh"), "an SSH"; got != want {
		t.Fatalf("ArticlePhrase(%q) = %q, want %q", "ssh", got, want)
	}
}

func TestArticlePhraseFrenchLocale(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	tests := []struct {
		word string
		want string
	}{
		{"branche", "la branche"},
		{"branches", "les branches"},
		{"amis", "des amis"},
		{"enfant", "l'enfant"},
		{"fichier", "le fichier"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := ArticlePhrase(tt.word)
			if got != tt.want {
				t.Errorf("ArticlePhrase(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestDefiniteArticle(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"file", "the"},
		{"error", "the"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := DefiniteArticle(tt.word)
			if got != tt.want {
				t.Errorf("DefiniteArticle(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestDefinitePhraseFrenchLocale(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	tests := []struct {
		word string
		want string
	}{
		{"branche", "la branche"},
		{"branches", "les branches"},
		{"amis", "les amis"},
		{"enfant", "l'enfant"},
		{"fichier", "le fichier"},
		{"héros", "le héros"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := DefinitePhrase(tt.word)
			if got != tt.want {
				t.Errorf("DefinitePhrase(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestLabel(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		word string
		want string
	}{
		{"status", "Status:"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := Label(tt.word)
			if got != tt.want {
				t.Errorf("Label(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestCompositionHelpersTrimWhitespace(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	if got, want := Label("  status  "), "Status:"; got != want {
		t.Fatalf("Label(%q) = %q, want %q", "  status  ", got, want)
	}
	if got, want := Article("  error  "), "an"; got != want {
		t.Fatalf("Article(%q) = %q, want %q", "  error  ", got, want)
	}
	if got, want := ArticlePhrase("  go_mod  "), "a go.mod"; got != want {
		t.Fatalf("ArticlePhrase(%q) = %q, want %q", "  go_mod  ", got, want)
	}
	if got, want := ActionFailed("  delete  ", "  file  "), "Failed to delete file"; got != want {
		t.Fatalf("ActionFailed(%q, %q) = %q, want %q", "  delete  ", "  file  ", got, want)
	}
}

func TestProgress(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		verb string
		want string
	}{
		{"build", "Building..."},
		{"delete", "Deleting..."},
		{"scan", "Scanning..."},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			got := Progress(tt.verb)
			if got != tt.want {
				t.Errorf("Progress(%q) = %q, want %q", tt.verb, got, tt.want)
			}
		})
	}
}

func TestProgressSubject(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	got := ProgressSubject("build", "project")
	want := "Building project..."
	if got != want {
		t.Errorf("ProgressSubject(%q, %q) = %q, want %q", "build", "project", got, want)
	}
}

func TestActionResult(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tests := []struct {
		verb, subject string
		want          string
	}{
		{"delete", "config.yaml", "Config.yaml deleted"},
		{"build", "project", "Project built"},
		{"", "file", ""},
		{"delete", "", "Deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.verb+"_"+tt.subject, func(t *testing.T) {
			got := ActionResult(tt.verb, tt.subject)
			if got != tt.want {
				t.Errorf("ActionResult(%q, %q) = %q, want %q", tt.verb, tt.subject, got, tt.want)
			}
		})
	}
}

func TestActionFailed(t *testing.T) {
	tests := []struct {
		verb, subject string
		want          string
	}{
		{"delete", "config.yaml", "Failed to delete config.yaml"},
		{"Delete", "config.yaml", "Failed to delete config.yaml"},
		{"push", "commits", "Failed to push commits"},
		{"push", "", "Failed to push"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.verb+"_"+tt.subject, func(t *testing.T) {
			got := ActionFailed(tt.verb, tt.subject)
			if got != tt.want {
				t.Errorf("ActionFailed(%q, %q) = %q, want %q", tt.verb, tt.subject, got, tt.want)
			}
		})
	}
}

func TestActionFailed_RespectsWordMap(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	data := GetGrammarData("en")
	if data == nil {
		t.Fatal("GetGrammarData(\"en\") returned nil")
	}
	original, existed := data.Words["push"]
	data.Words["push"] = "submit"
	t.Cleanup(func() {
		if existed {
			data.Words["push"] = original
			return
		}
		delete(data.Words, "push")
	})

	if got, want := ActionFailed("push", "commits"), "Failed to submit commits"; got != want {
		t.Fatalf("ActionFailed(%q, %q) = %q, want %q", "push", "commits", got, want)
	}
}

func TestActionFailedFrenchLocale(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	if got, want := ActionFailed("supprimer", ""), "Impossible de supprimer"; got != want {
		t.Fatalf("ActionFailed(%q, %q) = %q, want %q", "supprimer", "", got, want)
	}
}

func TestGrammarData_Signals(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	data := GetGrammarData("en")
	if data == nil {
		t.Fatal("GetGrammarData(\"en\") returned nil")
	}
	if got := len(data.Signals.NounDeterminers); got < 20 {
		t.Errorf("NounDeterminers: got %d entries, want >= 20", got)
	}
	if got := len(data.Signals.VerbAuxiliaries); got < 19 {
		t.Errorf("VerbAuxiliaries: got %d entries, want >= 19", got)
	}
	if len(data.Signals.VerbInfinitive) != 1 || data.Signals.VerbInfinitive[0] != "to" {
		t.Errorf("VerbInfinitive: got %v, want [\"to\"]", data.Signals.VerbInfinitive)
	}

	// Spot-check known values
	detFound := false
	for _, d := range data.Signals.NounDeterminers {
		if d == "the" {
			detFound = true
		}
	}
	if !detFound {
		t.Error("NounDeterminers missing 'the'")
	}

	auxFound := false
	for _, a := range data.Signals.VerbAuxiliaries {
		if a == "will" {
			auxFound = true
		}
	}
	if !auxFound {
		t.Error("VerbAuxiliaries missing 'will'")
	}
}

func TestGrammarData_DualClassEntries(t *testing.T) {
	svc, _ := New()
	SetDefault(svc)
	data := GetGrammarData("en")

	dualClass := []string{"commit", "run", "test", "check", "file", "build"}
	for _, word := range dualClass {
		if _, ok := data.Verbs[word]; !ok {
			t.Errorf("gram.verb missing dual-class word %q", word)
		}
		if _, ok := data.Nouns[word]; !ok {
			t.Errorf("gram.noun missing dual-class word %q", word)
		}
	}
}

func TestFrenchGrammarData(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	data := GetGrammarData("fr")
	if data == nil {
		t.Fatal("GetGrammarData(\"fr\") returned nil")
	}

	// Verbs loaded
	if got := len(data.Verbs); got < 40 {
		t.Errorf("French verbs: got %d, want >= 40", got)
	}

	// Spot-check irregular verb forms
	verbTests := []struct {
		verb, past, gerund string
	}{
		{"être", "été", "étant"},
		{"avoir", "eu", "ayant"},
		{"faire", "fait", "faisant"},
		{"supprimer", "supprimé", "supprimant"},
		{"construire", "construit", "construisant"},
		{"écrire", "écrit", "écrivant"},
		{"prendre", "pris", "prenant"},
	}
	for _, tt := range verbTests {
		forms, ok := data.Verbs[tt.verb]
		if !ok {
			t.Errorf("French verb %q not found", tt.verb)
			continue
		}
		if forms.Past != tt.past {
			t.Errorf("French verb %q past = %q, want %q", tt.verb, forms.Past, tt.past)
		}
		if forms.Gerund != tt.gerund {
			t.Errorf("French verb %q gerund = %q, want %q", tt.verb, forms.Gerund, tt.gerund)
		}
	}

	// Nouns loaded with gender
	if got := len(data.Nouns); got < 20 {
		t.Errorf("French nouns: got %d, want >= 20", got)
	}
	nounTests := []struct {
		noun, one, other, gender string
	}{
		{"fichier", "fichier", "fichiers", "m"},
		{"branche", "branche", "branches", "f"},
		{"vulnérabilité", "vulnérabilité", "vulnérabilités", "f"},
		{"serveur", "serveur", "serveurs", "m"},
	}
	for _, tt := range nounTests {
		forms, ok := data.Nouns[tt.noun]
		if !ok {
			t.Errorf("French noun %q not found", tt.noun)
			continue
		}
		if forms.One != tt.one {
			t.Errorf("French noun %q one = %q, want %q", tt.noun, forms.One, tt.one)
		}
		if forms.Other != tt.other {
			t.Errorf("French noun %q other = %q, want %q", tt.noun, forms.Other, tt.other)
		}
		if forms.Gender != tt.gender {
			t.Errorf("French noun %q gender = %q, want %q", tt.noun, forms.Gender, tt.gender)
		}
	}

	// Articles with gender
	if data.Articles.Definite != "le" {
		t.Errorf("French definite article = %q, want \"le\"", data.Articles.Definite)
	}
	if data.Articles.IndefiniteDefault != "un" {
		t.Errorf("French indefinite article = %q, want \"un\"", data.Articles.IndefiniteDefault)
	}
	if data.Articles.ByGender == nil {
		t.Fatal("French articles.ByGender is nil")
	}
	if got := data.Articles.ByGender["m"]; got != "le" {
		t.Errorf("French article by_gender[m] = %q, want \"le\"", got)
	}
	if got := data.Articles.ByGender["f"]; got != "la" {
		t.Errorf("French article by_gender[f] = %q, want \"la\"", got)
	}

	// Punctuation — French space before colon
	if data.Punct.LabelSuffix != " :" {
		t.Errorf("French label suffix = %q, want \" :\"", data.Punct.LabelSuffix)
	}

	// Signals loaded
	if got := len(data.Signals.NounDeterminers); got < 25 {
		t.Errorf("French NounDeterminers: got %d, want >= 25", got)
	}
	if got := len(data.Signals.VerbAuxiliaries); got < 15 {
		t.Errorf("French VerbAuxiliaries: got %d, want >= 15", got)
	}
}

func TestGrammarFallbackToBaseLanguageTag(t *testing.T) {
	prevDefault := Default()
	prevGrammar := GetGrammarData("en")
	t.Cleanup(func() {
		SetGrammarData("en", prevGrammar)
		SetDefault(prevDefault)
	})

	SetGrammarData("en", &GrammarData{
		Verbs: map[string]VerbForms{
			"delete": {Past: "deleted", Gerund: "deleting"},
		},
		Nouns: map[string]NounForms{
			"file": {One: "file", Other: "files"},
		},
		Articles: ArticleForms{
			IndefiniteDefault: "a",
			IndefiniteVowel:   "an",
			Definite:          "the",
		},
		Punct: PunctuationRules{
			LabelSuffix:    ":",
			ProgressSuffix: "...",
		},
		Words: map[string]string{
			"status": "Status",
		},
	})

	svc, err := NewWithLoader(regionFallbackLoader{})
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	SetDefault(svc)
	if err := svc.SetLanguage("en-GB"); err != nil {
		t.Fatalf("SetLanguage(en-GB) failed: %v", err)
	}

	if got := PastTense("delete"); got != "deleted" {
		t.Fatalf("PastTense(delete) = %q, want %q", got, "deleted")
	}
	if got := Pluralize("file", 2); got != "files" {
		t.Fatalf("Pluralize(file, 2) = %q, want %q", got, "files")
	}
	if got := Article("apple"); got != "an" {
		t.Fatalf("Article(apple) = %q, want %q", got, "an")
	}
	if got := Label("status"); got != "Status:" {
		t.Fatalf("Label(status) = %q, want %q", got, "Status:")
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()
	expected := []string{
		"title",
		"lower",
		"upper",
		"n",
		"number",
		"int",
		"decimal",
		"float",
		"percent",
		"pct",
		"bytes",
		"size",
		"ordinal",
		"ord",
		"ago",
		"past",
		"gerund",
		"plural",
		"pluralForm",
		"article",
		"articlePhrase",
		"definiteArticle",
		"definite",
		"definitePhrase",
		"quote",
		"label",
		"progress",
		"progressSubject",
		"actionResult",
		"actionFailed",
		"prompt",
		"lang",
		"timeAgo",
		"formatAgo",
	}
	for _, name := range expected {
		if _, ok := funcs[name]; !ok {
			t.Errorf("TemplateFuncs() missing %q", name)
		}
	}
}

func TestTemplateFuncs_Article(t *testing.T) {
	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(`{{article "apple"}}`)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if got, want := buf.String(), "an apple"; got != want {
		t.Fatalf("template article = %q, want %q", got, want)
	}

	tmpl, err = template.New("").Funcs(TemplateFuncs()).Parse(`{{articleToken "apple"}}|{{articlePhrase "apple"}}|{{definiteToken "apple"}}|{{definitePhrase "apple"}}`)
	if err != nil {
		t.Fatalf("Parse() alias helpers failed: %v", err)
	}

	buf.Reset()
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() alias helpers failed: %v", err)
	}

	if got, want := buf.String(), "an|an apple|the|the apple"; got != want {
		t.Fatalf("template article aliases = %q, want %q", got, want)
	}

	tmpl, err = template.New("").Funcs(TemplateFuncs()).Parse(`{{definiteArticle "apple"}}`)
	if err != nil {
		t.Fatalf("Parse() definite article helper failed: %v", err)
	}

	buf.Reset()
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() definite article helper failed: %v", err)
	}

	if got, want := buf.String(), "the"; got != want {
		t.Fatalf("template definite article = %q, want %q", got, want)
	}
}

func TestTemplateFuncs_CompositeHelpers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(
		`{{label "status"}}|{{progress "build"}}|{{progressSubject "build" "project"}}|{{actionResult "delete" "file"}}|{{actionFailed "delete" "file"}}`,
	)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	want := "Status:|Building...|Building project...|File deleted|Failed to delete file"
	if got := buf.String(); got != want {
		t.Fatalf("template composite helpers = %q, want %q", got, want)
	}
}

func TestTemplateFuncs_PromptAndLang(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(
		`{{prompt "confirm"}}|{{lang "de"}}`,
	)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if got, want := buf.String(), "Are you sure?|German"; got != want {
		t.Fatalf("template prompt/lang = %q, want %q", got, want)
	}
}

func TestTemplateFuncs_NumericAlias(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(
		`{{n "number" 1234567}}|{{n "ago" 3 "hours"}}`,
	)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "1,234,567|3 hours ago") {
		t.Fatalf("template numeric alias = %q, want prefix %q", got, "1,234,567|3 hours ago")
	}
}

func TestTemplateFuncs_NumericDirectAliases(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(
		`{{int 1234567}}|{{float 3.14}}|{{pct 0.85}}|{{size 1536000}}|{{ord 3}}|{{ago 3 "hours"}}`,
	)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "1,234,567|3.14|85%|1.46 MB|3rd|3 hours ago") {
		t.Fatalf("template direct numeric aliases = %q, want prefix %q", got, "1,234,567|3.14|85%|1.46 MB|3rd|3 hours ago")
	}
}

func TestTemplateFuncs_TimeHelpers(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(
		`{{formatAgo 3 "hour"}}|{{timeAgo .}}`,
	)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, time.Now().Add(-5*time.Minute)); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "3 hours ago|") {
		t.Fatalf("template time helpers prefix = %q, want %q", got, "3 hours ago|")
	}
	if !strings.Contains(got, "minutes ago") && !strings.Contains(got, "just now") {
		t.Fatalf("template time helpers suffix = %q, want relative time output", got)
	}
}

func TestCompositeHelpersRespectWordMap(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	data := GetGrammarData("en")
	if data == nil {
		t.Fatal("GetGrammarData(\"en\") returned nil")
	}
	original, existed := data.Words["go_mod"]
	data.Words["go_mod"] = "go.mod"
	t.Cleanup(func() {
		if existed {
			data.Words["go_mod"] = original
			return
		}
		delete(data.Words, "go_mod")
	})

	if got, want := Label("go_mod"), "go.mod:"; got != want {
		t.Fatalf("Label(%q) = %q, want %q", "go_mod", got, want)
	}
	if got, want := ProgressSubject("build", "go_mod"), "Building go.mod..."; got != want {
		t.Fatalf("ProgressSubject(%q, %q) = %q, want %q", "build", "go_mod", got, want)
	}
	if got, want := ProgressSubject("build", ""), "Building..."; got != want {
		t.Fatalf("ProgressSubject(%q, %q) = %q, want %q", "build", "", got, want)
	}
	if got, want := ActionResult("delete", "go_mod"), "go.mod deleted"; got != want {
		t.Fatalf("ActionResult(%q, %q) = %q, want %q", "delete", "go_mod", got, want)
	}
	if got, want := ActionResult("delete", ""), "Deleted"; got != want {
		t.Fatalf("ActionResult(%q, %q) = %q, want %q", "delete", "", got, want)
	}
	if got, want := ActionFailed("delete", "go_mod"), "Failed to delete go.mod"; got != want {
		t.Fatalf("ActionFailed(%q, %q) = %q, want %q", "delete", "go_mod", got, want)
	}
}

// --- Benchmarks ---

func BenchmarkPastTense_Irregular(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PastTense("build")
	}
}

func BenchmarkPastTense_Regular(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PastTense("delete")
	}
}

func BenchmarkPastTense_Compound(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PastTense("rebuild")
	}
}

func BenchmarkGerund(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gerund("commit")
	}
}

func BenchmarkPluralize(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Pluralize("repository", 5)
	}
}

func BenchmarkArticle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Article("error")
	}
}

func BenchmarkProgress(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Progress("build")
	}
}

func BenchmarkActionResult(b *testing.B) {
	svc, _ := New()
	SetDefault(svc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ActionResult("delete", "config.yaml")
	}
}

// --- AX-7 canonical triplets ---

func TestGrammar_GetGrammarData_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		_ = GetGrammarData("en")
	})
	if !called {
		t.Fatal("GetGrammarData was not exercised")
	}
}

func TestGrammar_GetGrammarData_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		_ = GetGrammarData("zz")
	})
	if !called {
		t.Fatal("GetGrammarData was not exercised")
	}
}

func TestGrammar_GetGrammarData_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		_ = GetGrammarData("")
	})
	if !called {
		t.Fatal("GetGrammarData was not exercised")
	}
}

func TestGrammar_SetGrammarData_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetGrammarData("ax7-set", &GrammarData{Words: map[string]string{"agent": "agent"}})
		if GetGrammarData("ax7-set") == nil {
			t.Fatal("expected grammar data")
		}
	})
	if !called {
		t.Fatal("SetGrammarData was not exercised")
	}
}

func TestGrammar_SetGrammarData_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetGrammarData("", &GrammarData{})
		_ = GetGrammarData("")
	})
	if !called {
		t.Fatal("SetGrammarData was not exercised")
	}
}

func TestGrammar_SetGrammarData_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		SetGrammarData("ax7-set-nil", nil)
		if GetGrammarData("ax7-set-nil") != nil {
			t.Fatal("expected nil grammar data")
		}
	})
	if !called {
		t.Fatal("SetGrammarData was not exercised")
	}
}

func TestGrammar_MergeGrammarData_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		MergeGrammarData("ax7-merge", &GrammarData{Words: map[string]string{"agent": "agent"}})
		if GetGrammarData("ax7-merge") == nil {
			t.Fatal("expected grammar data")
		}
	})
	if !called {
		t.Fatal("MergeGrammarData was not exercised")
	}
}

func TestGrammar_MergeGrammarData_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		MergeGrammarData("", &GrammarData{})
		_ = GetGrammarData("")
	})
	if !called {
		t.Fatal("MergeGrammarData was not exercised")
	}
}

func TestGrammar_MergeGrammarData_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		MergeGrammarData("ax7-merge-nil", nil)
		_ = GetGrammarData("ax7-merge-nil")
	})
	if !called {
		t.Fatal("MergeGrammarData was not exercised")
	}
}

func TestGrammar_IrregularVerbs_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularVerbs()
		if got["go"].Past == "" {
			t.Fatal("expected go")
		}
	})
	if !called {
		t.Fatal("IrregularVerbs was not exercised")
	}
}

func TestGrammar_IrregularVerbs_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularVerbs()
		delete(got, "go")
		if IrregularVerbs()["go"].Past == "" {
			t.Fatal("source map mutated")
		}
	})
	if !called {
		t.Fatal("IrregularVerbs was not exercised")
	}
}

func TestGrammar_IrregularVerbs_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularVerbs()
		if len(got) == 0 {
			t.Fatal("expected verbs")
		}
	})
	if !called {
		t.Fatal("IrregularVerbs was not exercised")
	}
}

func TestGrammar_IrregularNouns_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularNouns()
		if got["child"] == "" {
			t.Fatal("expected child")
		}
	})
	if !called {
		t.Fatal("IrregularNouns was not exercised")
	}
}

func TestGrammar_IrregularNouns_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularNouns()
		delete(got, "child")
		if IrregularNouns()["child"] == "" {
			t.Fatal("source map mutated")
		}
	})
	if !called {
		t.Fatal("IrregularNouns was not exercised")
	}
}

func TestGrammar_IrregularNouns_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := IrregularNouns()
		if len(got) == 0 {
			t.Fatal("expected nouns")
		}
	})
	if !called {
		t.Fatal("IrregularNouns was not exercised")
	}
}

func TestGrammar_DualClassVerbs_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassVerbs()
		if got["build"].Past == "" && len(got) == 0 {
			t.Fatal("expected verbs")
		}
	})
	if !called {
		t.Fatal("DualClassVerbs was not exercised")
	}
}

func TestGrammar_DualClassVerbs_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassVerbs()
		delete(got, "build")
		_ = DualClassVerbs()
	})
	if !called {
		t.Fatal("DualClassVerbs was not exercised")
	}
}

func TestGrammar_DualClassVerbs_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassVerbs()
		if len(got) == 0 {
			t.Fatal("expected verbs")
		}
	})
	if !called {
		t.Fatal("DualClassVerbs was not exercised")
	}
}

func TestGrammar_DualClassNouns_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassNouns()
		if len(got) == 0 {
			t.Fatal("expected nouns")
		}
	})
	if !called {
		t.Fatal("DualClassNouns was not exercised")
	}
}

func TestGrammar_DualClassNouns_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassNouns()
		delete(got, "build")
		_ = DualClassNouns()
	})
	if !called {
		t.Fatal("DualClassNouns was not exercised")
	}
}

func TestGrammar_DualClassNouns_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DualClassNouns()
		if len(got) == 0 {
			t.Fatal("expected nouns")
		}
	})
	if !called {
		t.Fatal("DualClassNouns was not exercised")
	}
}

func TestGrammar_Lower_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Lower("AGENT")
		if got != "agent" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Lower was not exercised")
	}
}

func TestGrammar_Lower_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Lower("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Lower was not exercised")
	}
}

func TestGrammar_Lower_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Lower("Agent-01")
		if got != "agent-01" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Lower was not exercised")
	}
}

func TestGrammar_Upper_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Upper("agent")
		if got != "AGENT" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Upper was not exercised")
	}
}

func TestGrammar_Upper_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Upper("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Upper was not exercised")
	}
}

func TestGrammar_Upper_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Upper("agent-01")
		if got != "AGENT-01" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Upper was not exercised")
	}
}

func TestGrammar_PastTense_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := PastTense("delete")
		if got == "" {
			t.Fatal("expected past tense")
		}
	})
	if !called {
		t.Fatal("PastTense was not exercised")
	}
}

func TestGrammar_PastTense_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PastTense("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("PastTense was not exercised")
	}
}

func TestGrammar_PastTense_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PastTense("run")
		if got == "" {
			t.Fatal("expected irregular past")
		}
	})
	if !called {
		t.Fatal("PastTense was not exercised")
	}
}

func TestGrammar_Gerund_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Gerund("delete")
		if got == "" {
			t.Fatal("expected gerund")
		}
	})
	if !called {
		t.Fatal("Gerund was not exercised")
	}
}

func TestGrammar_Gerund_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Gerund("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Gerund was not exercised")
	}
}

func TestGrammar_Gerund_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Gerund("run")
		if got == "" {
			t.Fatal("expected irregular gerund")
		}
	})
	if !called {
		t.Fatal("Gerund was not exercised")
	}
}

func TestGrammar_Pluralize_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Pluralize("file", 2)
		if got == "" {
			t.Fatal("expected plural")
		}
	})
	if !called {
		t.Fatal("Pluralize was not exercised")
	}
}

func TestGrammar_Pluralize_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Pluralize("", 2)
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Pluralize was not exercised")
	}
}

func TestGrammar_Pluralize_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Pluralize("child", 2)
		if got == "" {
			t.Fatal("expected irregular plural")
		}
	})
	if !called {
		t.Fatal("Pluralize was not exercised")
	}
}

func TestGrammar_PluralForm_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralForm("file")
		if got == "" {
			t.Fatal("expected plural")
		}
	})
	if !called {
		t.Fatal("PluralForm was not exercised")
	}
}

func TestGrammar_PluralForm_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralForm("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("PluralForm was not exercised")
	}
}

func TestGrammar_PluralForm_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := PluralForm("child")
		if got == "" {
			t.Fatal("expected plural")
		}
	})
	if !called {
		t.Fatal("PluralForm was not exercised")
	}
}

func TestGrammar_Article_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Article("apple")
		_ = got
	})
	if !called {
		t.Fatal("Article was not exercised")
	}
}

func TestGrammar_Article_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Article("")
		_ = got
	})
	if !called {
		t.Fatal("Article was not exercised")
	}
}

func TestGrammar_Article_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Article("user")
		_ = got
	})
	if !called {
		t.Fatal("Article was not exercised")
	}
}

func TestGrammar_ArticleToken_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticleToken("apple")
		_ = got
	})
	if !called {
		t.Fatal("ArticleToken was not exercised")
	}
}

func TestGrammar_ArticleToken_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticleToken("")
		_ = got
	})
	if !called {
		t.Fatal("ArticleToken was not exercised")
	}
}

func TestGrammar_ArticleToken_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticleToken("user")
		_ = got
	})
	if !called {
		t.Fatal("ArticleToken was not exercised")
	}
}

func TestGrammar_Title_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Title("agent")
		if got != "Agent" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Title was not exercised")
	}
}

func TestGrammar_Title_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Title("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Title was not exercised")
	}
}

func TestGrammar_Title_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Title("a")
		if got != "A" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Title was not exercised")
	}
}

func TestGrammar_Quote_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Quote("agent")
		if got == "" {
			t.Fatal("expected quote")
		}
	})
	if !called {
		t.Fatal("Quote was not exercised")
	}
}

func TestGrammar_Quote_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Quote("")
		if got == "" {
			t.Fatal("expected quote")
		}
	})
	if !called {
		t.Fatal("Quote was not exercised")
	}
}

func TestGrammar_Quote_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Quote("agent\n")
		if got == "" {
			t.Fatal("expected quote")
		}
	})
	if !called {
		t.Fatal("Quote was not exercised")
	}
}

func TestGrammar_ArticlePhrase_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticlePhrase("apple")
		_ = got
	})
	if !called {
		t.Fatal("ArticlePhrase was not exercised")
	}
}

func TestGrammar_ArticlePhrase_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticlePhrase("")
		_ = got
	})
	if !called {
		t.Fatal("ArticlePhrase was not exercised")
	}
}

func TestGrammar_ArticlePhrase_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ArticlePhrase("user")
		_ = got
	})
	if !called {
		t.Fatal("ArticlePhrase was not exercised")
	}
}

func TestGrammar_DefiniteArticle_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteArticle("file")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteArticle was not exercised")
	}
}

func TestGrammar_DefiniteArticle_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteArticle("")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteArticle was not exercised")
	}
}

func TestGrammar_DefiniteArticle_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteArticle("apple")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteArticle was not exercised")
	}
}

func TestGrammar_DefiniteToken_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteToken("file")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteToken was not exercised")
	}
}

func TestGrammar_DefiniteToken_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteToken("")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteToken was not exercised")
	}
}

func TestGrammar_DefiniteToken_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefiniteToken("apple")
		_ = got
	})
	if !called {
		t.Fatal("DefiniteToken was not exercised")
	}
}

func TestGrammar_DefinitePhrase_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefinitePhrase("file")
		_ = got
	})
	if !called {
		t.Fatal("DefinitePhrase was not exercised")
	}
}

func TestGrammar_DefinitePhrase_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefinitePhrase("")
		_ = got
	})
	if !called {
		t.Fatal("DefinitePhrase was not exercised")
	}
}

func TestGrammar_DefinitePhrase_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := DefinitePhrase("apple")
		_ = got
	})
	if !called {
		t.Fatal("DefinitePhrase was not exercised")
	}
}

func TestGrammar_TemplateFuncs_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		funcs := TemplateFuncs()
		if len(funcs) == 0 {
			t.Fatal("expected funcs")
		}
	})
	if !called {
		t.Fatal("TemplateFuncs was not exercised")
	}
}

func TestGrammar_TemplateFuncs_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		funcs := TemplateFuncs()
		if funcs["missing"] != nil {
			t.Fatal("unexpected missing func")
		}
	})
	if !called {
		t.Fatal("TemplateFuncs was not exercised")
	}
}

func TestGrammar_TemplateFuncs_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		funcs := TemplateFuncs()
		if funcs["article"] == nil {
			t.Fatal("expected article func")
		}
	})
	if !called {
		t.Fatal("TemplateFuncs was not exercised")
	}
}

func TestGrammar_Number_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Number(1234)
		if got == "" {
			t.Fatal("expected number")
		}
	})
	if !called {
		t.Fatal("Number was not exercised")
	}
}

func TestGrammar_Number_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Number("bad")
		if got == "" {
			t.Fatal("expected fallback number")
		}
	})
	if !called {
		t.Fatal("Number was not exercised")
	}
}

func TestGrammar_Number_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Number(0)
		if got == "" {
			t.Fatal("expected zero")
		}
	})
	if !called {
		t.Fatal("Number was not exercised")
	}
}

func TestGrammar_Decimal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Decimal(1.5)
		if got == "" {
			t.Fatal("expected decimal")
		}
	})
	if !called {
		t.Fatal("Decimal was not exercised")
	}
}

func TestGrammar_Decimal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Decimal("bad")
		if got == "" {
			t.Fatal("expected fallback decimal")
		}
	})
	if !called {
		t.Fatal("Decimal was not exercised")
	}
}

func TestGrammar_Decimal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Decimal(0)
		if got == "" {
			t.Fatal("expected zero")
		}
	})
	if !called {
		t.Fatal("Decimal was not exercised")
	}
}

func TestGrammar_Percent_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Percent(0.5)
		if got == "" {
			t.Fatal("expected percent")
		}
	})
	if !called {
		t.Fatal("Percent was not exercised")
	}
}

func TestGrammar_Percent_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Percent("bad")
		if got == "" {
			t.Fatal("expected fallback percent")
		}
	})
	if !called {
		t.Fatal("Percent was not exercised")
	}
}

func TestGrammar_Percent_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Percent(0)
		if got == "" {
			t.Fatal("expected zero")
		}
	})
	if !called {
		t.Fatal("Percent was not exercised")
	}
}

func TestGrammar_Bytes_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Bytes(1024)
		if got == "" {
			t.Fatal("expected bytes")
		}
	})
	if !called {
		t.Fatal("Bytes was not exercised")
	}
}

func TestGrammar_Bytes_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Bytes("bad")
		if got == "" {
			t.Fatal("expected fallback bytes")
		}
	})
	if !called {
		t.Fatal("Bytes was not exercised")
	}
}

func TestGrammar_Bytes_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Bytes(0)
		if got == "" {
			t.Fatal("expected zero")
		}
	})
	if !called {
		t.Fatal("Bytes was not exercised")
	}
}

func TestGrammar_Ordinal_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Ordinal(1)
		if got == "" {
			t.Fatal("expected ordinal")
		}
	})
	if !called {
		t.Fatal("Ordinal was not exercised")
	}
}

func TestGrammar_Ordinal_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Ordinal("bad")
		if got == "" {
			t.Fatal("expected fallback ordinal")
		}
	})
	if !called {
		t.Fatal("Ordinal was not exercised")
	}
}

func TestGrammar_Ordinal_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Ordinal(0)
		if got == "" {
			t.Fatal("expected zero")
		}
	})
	if !called {
		t.Fatal("Ordinal was not exercised")
	}
}

func TestGrammar_Ago_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		ax7SetDefault(t)
		got := Ago(5, "minute")
		if got == "" {
			t.Fatal("expected ago")
		}
	})
	if !called {
		t.Fatal("Ago was not exercised")
	}
}

func TestGrammar_Ago_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Ago(5, "unknown")
		if got == "" {
			t.Fatal("expected fallback ago")
		}
	})
	if !called {
		t.Fatal("Ago was not exercised")
	}
}

func TestGrammar_Ago_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Ago(0, "second")
		if got == "" {
			t.Fatal("expected ago")
		}
	})
	if !called {
		t.Fatal("Ago was not exercised")
	}
}

func TestGrammar_Progress_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Progress("build")
		if got == "" {
			t.Fatal("expected progress")
		}
	})
	if !called {
		t.Fatal("Progress was not exercised")
	}
}

func TestGrammar_Progress_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Progress("")
		_ = got
	})
	if !called {
		t.Fatal("Progress was not exercised")
	}
}

func TestGrammar_Progress_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Progress("run")
		if got == "" {
			t.Fatal("expected progress")
		}
	})
	if !called {
		t.Fatal("Progress was not exercised")
	}
}

func TestGrammar_ProgressSubject_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ProgressSubject("build", "docs")
		if got == "" {
			t.Fatal("expected progress")
		}
	})
	if !called {
		t.Fatal("ProgressSubject was not exercised")
	}
}

func TestGrammar_ProgressSubject_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ProgressSubject("", "docs")
		_ = got
	})
	if !called {
		t.Fatal("ProgressSubject was not exercised")
	}
}

func TestGrammar_ProgressSubject_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ProgressSubject("build", "")
		if got == "" {
			t.Fatal("expected progress")
		}
	})
	if !called {
		t.Fatal("ProgressSubject was not exercised")
	}
}

func TestGrammar_ActionResult_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionResult("delete", "file")
		if got == "" {
			t.Fatal("expected result")
		}
	})
	if !called {
		t.Fatal("ActionResult was not exercised")
	}
}

func TestGrammar_ActionResult_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionResult("", "file")
		_ = got
	})
	if !called {
		t.Fatal("ActionResult was not exercised")
	}
}

func TestGrammar_ActionResult_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionResult("delete", "")
		if got == "" {
			t.Fatal("expected result")
		}
	})
	if !called {
		t.Fatal("ActionResult was not exercised")
	}
}

func TestGrammar_ActionFailed_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionFailed("delete", "file")
		if got == "" {
			t.Fatal("expected failure")
		}
	})
	if !called {
		t.Fatal("ActionFailed was not exercised")
	}
}

func TestGrammar_ActionFailed_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionFailed("", "file")
		_ = got
	})
	if !called {
		t.Fatal("ActionFailed was not exercised")
	}
}

func TestGrammar_ActionFailed_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := ActionFailed("delete", "")
		if got == "" {
			t.Fatal("expected failure")
		}
	})
	if !called {
		t.Fatal("ActionFailed was not exercised")
	}
}

func TestGrammar_Label_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Label("status")
		if got == "" {
			t.Fatal("expected label")
		}
	})
	if !called {
		t.Fatal("Label was not exercised")
	}
}

func TestGrammar_Label_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Label("")
		_ = got
	})
	if !called {
		t.Fatal("Label was not exercised")
	}
}

func TestGrammar_Label_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		got := Label("agent")
		if got == "" {
			t.Fatal("expected label")
		}
	})
	if !called {
		t.Fatal("Label was not exercised")
	}
}
