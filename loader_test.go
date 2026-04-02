package i18n

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestFSLoaderLanguages(t *testing.T) {
	loader := NewFSLoader(localeFS, "locales")
	langs := loader.Languages()
	if len(langs) == 0 {
		t.Fatal("FSLoader.Languages() returned empty")
	}

	found := slices.Contains(langs, "en")
	if !found {
		t.Errorf("Languages() = %v, expected 'en' in list", langs)
	}
}

func TestFSLoaderLanguagesCanonicalAndUnique(t *testing.T) {
	fs := fstest.MapFS{
		"locales/en.json":    &fstest.MapFile{Data: []byte(`{}`)},
		"locales/en_US.json": &fstest.MapFile{Data: []byte(`{}`)},
		"locales/es-MX.json": &fstest.MapFile{Data: []byte(`{}`)},
		"locales/fr.json":    &fstest.MapFile{Data: []byte(`{}`)},
	}

	loader := NewFSLoader(fs, "locales")
	langs := loader.Languages()
	want := []string{"en", "en-US", "es-MX", "fr"}
	if !slices.Equal(langs, want) {
		t.Fatalf("Languages() = %v, want %v", langs, want)
	}
}

func TestFSLoaderLoad(t *testing.T) {
	loader := NewFSLoader(localeFS, "locales")
	messages, grammar, err := loader.Load("en")
	if err != nil {
		t.Fatalf("Load(en) error: %v", err)
	}

	// Should have messages from the JSON
	if len(messages) == 0 {
		t.Error("Load(en) returned 0 messages")
	}

	// Grammar data should be extracted from nested JSON
	if grammar == nil {
		t.Fatal("Load(en) returned nil grammar")
	}

	// Verbs from gram.verb.*
	if len(grammar.Verbs) == 0 {
		t.Error("grammar has 0 verbs")
	}
	if v, ok := grammar.Verbs["build"]; !ok {
		t.Error("grammar missing verb 'build'")
	} else {
		if v.Past != "built" {
			t.Errorf("build.past = %q, want 'built'", v.Past)
		}
		if v.Gerund != "building" {
			t.Errorf("build.gerund = %q, want 'building'", v.Gerund)
		}
	}

	// Nouns from gram.noun.*
	if len(grammar.Nouns) == 0 {
		t.Error("grammar has 0 nouns")
	}
	if n, ok := grammar.Nouns["file"]; !ok {
		t.Error("grammar missing noun 'file'")
	} else {
		if n.One != "file" {
			t.Errorf("file.one = %q, want 'file'", n.One)
		}
		if n.Other != "files" {
			t.Errorf("file.other = %q, want 'files'", n.Other)
		}
	}

	// Articles from gram.article
	if grammar.Articles.IndefiniteDefault != "a" {
		t.Errorf("article.indefinite.default = %q, want 'a'", grammar.Articles.IndefiniteDefault)
	}
	if grammar.Articles.IndefiniteVowel != "an" {
		t.Errorf("article.indefinite.vowel = %q, want 'an'", grammar.Articles.IndefiniteVowel)
	}
	if grammar.Articles.Definite != "the" {
		t.Errorf("article.definite = %q, want 'the'", grammar.Articles.Definite)
	}

	// Punctuation from gram.punct
	if grammar.Punct.LabelSuffix != ":" {
		t.Errorf("punct.label = %q, want ':'", grammar.Punct.LabelSuffix)
	}
	if grammar.Punct.ProgressSuffix != "..." {
		t.Errorf("punct.progress = %q, want '...'", grammar.Punct.ProgressSuffix)
	}

	// Number formatting from gram.number
	if grammar.Number.ThousandsSep != "," {
		t.Errorf("number.thousands = %q, want ','", grammar.Number.ThousandsSep)
	}
	if grammar.Number.DecimalSep != "." {
		t.Errorf("number.decimal = %q, want '.'", grammar.Number.DecimalSep)
	}
	if grammar.Number.PercentFmt != "%s%%" {
		t.Errorf("number.percent = %q, want '%%s%%%%'", grammar.Number.PercentFmt)
	}

	// Words from gram.word.*
	if len(grammar.Words) == 0 {
		t.Error("grammar has 0 words")
	}
	if grammar.Words["url"] != "URL" {
		t.Errorf("word.url = %q, want 'URL'", grammar.Words["url"])
	}
	if grammar.Words["api"] != "API" {
		t.Errorf("word.api = %q, want 'API'", grammar.Words["api"])
	}
}

func TestFSLoaderLoadMissing(t *testing.T) {
	loader := NewFSLoader(localeFS, "locales")
	_, _, err := loader.Load("xx")
	if err == nil {
		t.Error("Load(xx) should fail for non-existent locale")
	}
}

func TestFSLoaderLoadFallsBackToBaseLanguage(t *testing.T) {
	fs := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{
				"greeting": "hello",
				"gram": {
					"article": {
						"indefinite": { "default": "a", "vowel": "an" },
						"definite": "the"
					}
				}
			}`),
		},
	}

	loader := NewFSLoader(fs, "locales")
	messages, grammar, err := loader.Load("en-GB")
	if err != nil {
		t.Fatalf("Load(en-GB) error: %v", err)
	}
	if got := messages["greeting"].Text; got != "hello" {
		t.Fatalf("Load(en-GB) greeting = %q, want %q", got, "hello")
	}
	if grammar == nil {
		t.Fatal("Load(en-GB) returned nil grammar")
	}
	if grammar.Articles.Definite != "the" {
		t.Fatalf("Load(en-GB) grammar article = %q, want %q", grammar.Articles.Definite, "the")
	}
}

func TestFlattenWithGrammar(t *testing.T) {
	messages := make(map[string]Message)
	grammar := &GrammarData{
		Verbs: make(map[string]VerbForms),
		Nouns: make(map[string]NounForms),
		Words: make(map[string]string),
	}

	raw := map[string]any{
		"gram": map[string]any{
			"verb": map[string]any{
				"test": map[string]any{
					"base":   "test",
					"past":   "tested",
					"gerund": "testing",
				},
				"partial_past": map[string]any{
					"past": "partialed",
				},
				"partial_gerund": map[string]any{
					"gerund": "partialing",
				},
				"publish_draft": map[string]any{
					"base":   "publish",
					"past":   "published",
					"gerund": "publishing",
				},
			},
			"noun": map[string]any{
				"widget": map[string]any{
					"one":   "widget",
					"other": "widgets",
				},
				"passed": map[string]any{
					"one":   "passed",
					"other": "passed",
				},
			},
			"word": map[string]any{
				"api":     "API",
				"failed":  "failed",
				"skipped": "skipped",
			},
			"punct": map[string]any{
				"label":    ":",
				"progress": "...",
			},
			"number": map[string]any{
				"thousands": ",",
				"decimal":   ".",
				"percent":   "%s%%",
			},
			"signal": map[string]any{
				"prior": map[string]any{
					"commit": map[string]any{
						"verb": 0.25,
						"noun": 0.75,
					},
				},
				"verb_negation": []any{"not", "never"},
			},
			"article": map[string]any{
				"indefinite": map[string]any{
					"default": "a",
					"vowel":   "an",
				},
				"definite": "the",
			},
		},
		"prompt": map[string]any{
			"yes": "y",
			"no":  "n",
		},
	}

	flattenWithGrammar("", raw, messages, grammar)

	// Verb extracted
	if v, ok := grammar.Verbs["test"]; !ok {
		t.Error("verb 'test' not extracted")
	} else {
		if v.Past != "tested" {
			t.Errorf("test.past = %q, want 'tested'", v.Past)
		}
	}
	if v, ok := grammar.Verbs["publish"]; !ok {
		t.Error("verb base override 'publish' not extracted")
	} else {
		if v.Past != "published" {
			t.Errorf("publish.past = %q, want 'published'", v.Past)
		}
		if v.Gerund != "publishing" {
			t.Errorf("publish.gerund = %q, want 'publishing'", v.Gerund)
		}
	}
	if _, ok := grammar.Verbs["publish_draft"]; ok {
		t.Error("verb should be stored under explicit base, not JSON key")
	}
	if v, ok := grammar.Verbs["partial_past"]; !ok {
		t.Error("incomplete verb entry with only past should be extracted")
	} else if v.Past != "partialed" || v.Gerund != "" {
		t.Errorf("partial_past forms = %+v, want Past only", v)
	}
	if v, ok := grammar.Verbs["partial_gerund"]; !ok {
		t.Error("incomplete verb entry with only gerund should be extracted")
	} else if v.Past != "" || v.Gerund != "partialing" {
		t.Errorf("partial_gerund forms = %+v, want Gerund only", v)
	}
	if _, ok := messages["gram.verb.partial_past"]; ok {
		t.Error("gram.verb.partial_past should not be flattened into messages")
	}
	if _, ok := messages["gram.verb.partial_gerund"]; ok {
		t.Error("gram.verb.partial_gerund should not be flattened into messages")
	}

	// Noun extracted
	if n, ok := grammar.Nouns["widget"]; !ok {
		t.Error("noun 'widget' not extracted")
	} else {
		if n.Other != "widgets" {
			t.Errorf("widget.other = %q, want 'widgets'", n.Other)
		}
	}
	if _, ok := grammar.Nouns["passed"]; ok {
		t.Error("deprecated noun 'passed' should be ignored")
	}

	// Word extracted
	if grammar.Words["api"] != "API" {
		t.Errorf("word 'api' = %q, want 'API'", grammar.Words["api"])
	}
	if _, ok := grammar.Words["failed"]; ok {
		t.Error("deprecated word 'failed' should be ignored")
	}
	if _, ok := grammar.Words["skipped"]; ok {
		t.Error("deprecated word 'skipped' should be ignored")
	}

	// Punct extracted
	if grammar.Punct.LabelSuffix != ":" {
		t.Errorf("punct.label = %q, want ':'", grammar.Punct.LabelSuffix)
	}

	// Number formatting extracted
	if grammar.Number.ThousandsSep != "," {
		t.Errorf("number.thousands = %q, want ','", grammar.Number.ThousandsSep)
	}
	if len(grammar.Signals.VerbNegation) != 2 || grammar.Signals.VerbNegation[0] != "not" || grammar.Signals.VerbNegation[1] != "never" {
		t.Errorf("verb negation not extracted: %+v", grammar.Signals.VerbNegation)
	}

	// Articles extracted
	if grammar.Articles.IndefiniteDefault != "a" {
		t.Errorf("article.indefinite.default = %q, want 'a'", grammar.Articles.IndefiniteDefault)
	}

	// Regular keys flattened
	if msg, ok := messages["prompt.yes"]; !ok || msg.Text != "y" {
		t.Errorf("prompt.yes not flattened correctly, got %+v", messages["prompt.yes"])
	}
	if _, ok := messages["gram.number.thousands"]; ok {
		t.Error("gram.number.thousands should not be flattened into messages")
	}
}

func TestMergeGrammarData(t *testing.T) {
	const lang = "zz"
	original := GetGrammarData(lang)
	t.Cleanup(func() {
		SetGrammarData(lang, original)
	})

	SetGrammarData(lang, &GrammarData{
		Verbs: map[string]VerbForms{
			"keep": {Past: "kept", Gerund: "keeping"},
		},
		Nouns: map[string]NounForms{
			"file": {One: "file", Other: "files"},
		},
		Words: map[string]string{
			"url": "URL",
		},
		Articles: ArticleForms{
			IndefiniteDefault: "a",
			IndefiniteVowel:   "an",
			Definite:          "the",
			ByGender: map[string]string{
				"m": "le",
			},
		},
		Punct: PunctuationRules{
			LabelSuffix:    ":",
			ProgressSuffix: "...",
		},
		Signals: SignalData{
			NounDeterminers: []string{"the"},
			VerbAuxiliaries: []string{"will"},
			VerbInfinitive:  []string{"to"},
			VerbNegation:    []string{"not"},
			Priors: map[string]map[string]float64{
				"run": {
					"verb": 0.7,
				},
			},
		},
		Number: NumberFormat{
			ThousandsSep: ",",
			DecimalSep:   ".",
			PercentFmt:   "%s%%",
		},
	})

	MergeGrammarData(lang, &GrammarData{
		Verbs: map[string]VerbForms{
			"add": {Past: "added", Gerund: "adding"},
		},
		Nouns: map[string]NounForms{
			"repo": {One: "repo", Other: "repos"},
		},
		Words: map[string]string{
			"api": "API",
		},
		Articles: ArticleForms{
			ByGender: map[string]string{
				"f": "la",
			},
		},
		Punct: PunctuationRules{
			LabelSuffix: " !",
		},
		Signals: SignalData{
			NounDeterminers: []string{"a"},
			VerbAuxiliaries: []string{"can"},
			VerbInfinitive:  []string{"go"},
			VerbNegation:    []string{"never"},
			Priors: map[string]map[string]float64{
				"run": {
					"noun": 0.3,
				},
			},
		},
		Number: NumberFormat{
			ThousandsSep: ".",
		},
	})

	data := GetGrammarData(lang)
	if data == nil {
		t.Fatal("MergeGrammarData() cleared existing grammar data")
	}
	if _, ok := data.Verbs["keep"]; !ok {
		t.Error("existing verb entry was lost")
	}
	if _, ok := data.Verbs["add"]; !ok {
		t.Error("merged verb entry missing")
	}
	if _, ok := data.Nouns["file"]; !ok {
		t.Error("existing noun entry was lost")
	}
	if _, ok := data.Nouns["repo"]; !ok {
		t.Error("merged noun entry missing")
	}
	if data.Words["url"] != "URL" || data.Words["api"] != "API" {
		t.Errorf("words not merged correctly: %+v", data.Words)
	}
	if data.Articles.IndefiniteDefault != "a" || data.Articles.IndefiniteVowel != "an" || data.Articles.Definite != "the" {
		t.Errorf("article defaults changed unexpectedly: %+v", data.Articles)
	}
	if data.Articles.ByGender["m"] != "le" || data.Articles.ByGender["f"] != "la" {
		t.Errorf("article by_gender not merged correctly: %+v", data.Articles.ByGender)
	}
	if data.Punct.LabelSuffix != " !" || data.Punct.ProgressSuffix != "..." {
		t.Errorf("punctuation not merged correctly: %+v", data.Punct)
	}
	if len(data.Signals.NounDeterminers) != 2 || len(data.Signals.VerbAuxiliaries) != 2 || len(data.Signals.VerbInfinitive) != 2 || len(data.Signals.VerbNegation) != 2 {
		t.Errorf("signal slices not merged correctly: %+v", data.Signals)
	}
	if got := data.Signals.Priors["run"]["verb"]; got != 0.7 {
		t.Errorf("signal priors lost existing value: got %v", got)
	}
	if got := data.Signals.Priors["run"]["noun"]; got != 0.3 {
		t.Errorf("signal priors missing merged value: got %v", got)
	}
	if data.Signals.VerbNegation[0] != "not" || data.Signals.VerbNegation[1] != "never" {
		t.Errorf("signal negation not merged correctly: %+v", data.Signals.VerbNegation)
	}
	if data.Number.ThousandsSep != "." || data.Number.DecimalSep != "." || data.Number.PercentFmt != "%s%%" {
		t.Errorf("number format not merged correctly: %+v", data.Number)
	}
}

func TestNewWithLoader_LoadsGrammarOnlyLocale(t *testing.T) {
	loaderFS := fstest.MapFS{
		"fr.json": &fstest.MapFile{
			Data: []byte(`{
				"gram": {
					"article": {
						"indefinite": { "default": "el", "vowel": "l'" },
						"definite": "el",
						"by_gender": { "m": "el", "f": "la" }
					},
					"punct": { "label": " !", "progress": " ..." },
					"signal": {
						"noun_determiner": ["el"],
						"verb_auxiliary": ["va"],
						"verb_infinitive": ["a"],
						"verb_negation": ["no", "nunca"]
					},
					"number": { "thousands": ".", "decimal": ",", "percent": "%s %%"}
				}
			}`),
		},
	}

	svc, err := NewWithLoader(NewFSLoader(loaderFS, "."))
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}

	data := GetGrammarData("fr")
	if data == nil {
		t.Fatal("grammar-only locale was not loaded")
	}
	if data.Articles.ByGender["f"] != "la" {
		t.Errorf("article by_gender[f] = %q, want %q", data.Articles.ByGender["f"], "la")
	}
	if data.Punct.LabelSuffix != " !" || data.Punct.ProgressSuffix != " ..." {
		t.Errorf("punctuation not loaded: %+v", data.Punct)
	}
	if len(data.Signals.NounDeterminers) != 1 || data.Signals.NounDeterminers[0] != "el" {
		t.Errorf("signals not loaded: %+v", data.Signals)
	}
	if len(data.Signals.VerbNegation) != 2 || data.Signals.VerbNegation[0] != "no" || data.Signals.VerbNegation[1] != "nunca" {
		t.Errorf("negation signal not loaded: %+v", data.Signals.VerbNegation)
	}
	if data.Number.DecimalSep != "," || data.Number.ThousandsSep != "." {
		t.Errorf("number format not loaded: %+v", data.Number)
	}

	if err := svc.SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}
	SetDefault(svc)
	if got := Label("status"); got != "Status !" {
		t.Errorf("Label(status) = %q, want %q", got, "Status !")
	}
}

func TestFlattenPluralObject(t *testing.T) {
	messages := make(map[string]Message)
	raw := map[string]any{
		"time": map[string]any{
			"ago": map[string]any{
				"second": map[string]any{
					"one":   "{{.Count}} second ago",
					"other": "{{.Count}} seconds ago",
				},
			},
		},
	}

	flattenWithGrammar("", raw, messages, nil)

	msg, ok := messages["time.ago.second"]
	if !ok {
		t.Fatal("time.ago.second not found")
	}
	if !msg.IsPlural() {
		t.Error("time.ago.second should be plural")
	}
	if msg.One != "{{.Count}} second ago" {
		t.Errorf("time.ago.second.one = %q", msg.One)
	}
	if msg.Other != "{{.Count}} seconds ago" {
		t.Errorf("time.ago.second.other = %q", msg.Other)
	}
}

func TestFSLoaderLanguagesErr_Good(t *testing.T) {
	loader := NewFSLoader(localeFS, "locales")
	if err := loader.LanguagesErr(); err != nil {
		t.Errorf("LanguagesErr() = %v, want nil for valid dir", err)
	}
}

func TestFSLoaderLanguagesErr_Bad(t *testing.T) {
	loader := NewFSLoader(localeFS, "nonexistent")
	langs := loader.Languages()
	if len(langs) != 0 {
		t.Errorf("Languages() = %v, want empty for bad dir", langs)
	}
	if err := loader.LanguagesErr(); err == nil {
		t.Error("LanguagesErr() = nil, want error for bad dir")
	}
}

func TestFlatten_Good(t *testing.T) {
	messages := make(map[string]Message)
	raw := map[string]any{
		"hello": "world",
		"nested": map[string]any{
			"key": "value",
		},
	}
	flatten("", raw, messages)
	if msg, ok := messages["hello"]; !ok || msg.Text != "world" {
		t.Errorf("flatten: hello = %+v, want 'world'", messages["hello"])
	}
	if msg, ok := messages["nested.key"]; !ok || msg.Text != "value" {
		t.Errorf("flatten: nested.key = %+v, want 'value'", messages["nested.key"])
	}
}

func TestCustomFSLoader(t *testing.T) {
	fs := fstest.MapFS{
		"locales/test.json": &fstest.MapFile{
			Data: []byte(`{
				"gram": {
					"verb": {
						"draft": { "base": "draft", "past": "drafted", "gerund": "drafting" },
						"zap": { "base": "zap", "past": "zapped", "gerund": "zapping" }
					},
					"signal": {
						"priors": {
							"draft": { "verb": 0.6, "noun": 0.4 }
						}
					},
					"word": {
						"hello": "Hello"
					}
				},
				"greeting": "Hello, world!"
			}`),
		},
	}

	svc, err := NewWithFS(fs, "locales", WithFallback("test"))
	if err != nil {
		t.Fatalf("NewWithFS failed: %v", err)
	}

	got := svc.T("greeting")
	if got != "Hello, world!" {
		t.Errorf("T(greeting) = %q, want 'Hello, world!'", got)
	}

	// Grammar should be loaded
	gd := GetGrammarData("test")
	if gd == nil {
		t.Fatal("grammar data not loaded for 'test'")
	}
	if v, ok := gd.Verbs["zap"]; !ok || v.Past != "zapped" {
		t.Errorf("verb 'zap' not loaded correctly")
	}
	if v, ok := gd.Verbs["draft"]; !ok || v.Past != "drafted" {
		t.Errorf("verb base override 'draft' not loaded correctly")
	}
	if gd.Signals.Priors["draft"]["verb"] != 0.6 || gd.Signals.Priors["draft"]["noun"] != 0.4 {
		t.Errorf("signal priors not loaded correctly: %+v", gd.Signals.Priors["draft"])
	}
}

func TestCustomFSLoaderPreservesZeroSignalPriors(t *testing.T) {
	fs := fstest.MapFS{
		"locales/test.json": &fstest.MapFile{
			Data: []byte(`{
				"gram": {
					"signal": {
						"prior": {
							"commit": { "verb": 0, "noun": 1 }
						}
					}
				}
			}`),
		},
	}

	loader := NewFSLoader(fs, "locales")
	_, grammar, err := loader.Load("test")
	if err != nil {
		t.Fatalf("Load(test) failed: %v", err)
	}
	if grammar == nil {
		t.Fatal("expected grammar data")
	}
	bucket, ok := grammar.Signals.Priors["commit"]
	if !ok {
		t.Fatal("signal priors for commit were not loaded")
	}
	if got := bucket["verb"]; got != 0 {
		t.Fatalf("signal priors verb = %v, want 0", got)
	}
	if got := bucket["noun"]; got != 1 {
		t.Fatalf("signal priors noun = %v, want 1", got)
	}
}
