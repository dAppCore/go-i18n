package i18n

import (
	"strings"
	"testing"
	"text/template"
)

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
		{"delete", "config.yaml", "Config.Yaml deleted"},
		{"build", "project", "Project built"},
		{"", "file", ""},
		{"delete", "", ""},
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

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()
	expected := []string{
		"title",
		"lower",
		"upper",
		"past",
		"gerund",
		"plural",
		"pluralForm",
		"article",
		"quote",
		"label",
		"progress",
		"progressSubject",
		"actionResult",
		"actionFailed",
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
