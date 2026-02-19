package i18n

import "testing"

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
	if len(data.Signals.NounDeterminers) == 0 {
		t.Error("Signals.NounDeterminers is empty")
	}
	if len(data.Signals.VerbAuxiliaries) == 0 {
		t.Error("Signals.VerbAuxiliaries is empty")
	}
	if len(data.Signals.VerbInfinitive) == 0 {
		t.Error("Signals.VerbInfinitive is empty")
	}

	// Spot-check known values
	found := false
	for _, d := range data.Signals.NounDeterminers {
		if d == "the" {
			found = true
		}
	}
	if !found {
		t.Error("NounDeterminers missing 'the'")
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()
	expected := []string{"title", "lower", "upper", "past", "gerund", "plural", "pluralForm", "article", "quote"}
	for _, name := range expected {
		if _, ok := funcs[name]; !ok {
			t.Errorf("TemplateFuncs() missing %q", name)
		}
	}
}
