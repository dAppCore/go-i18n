package i18n

import (
	"testing"
	"testing/fstest"
)

func TestFSLoaderLanguages(t *testing.T) {
	loader := NewFSLoader(localeFS, "locales")
	langs := loader.Languages()
	if len(langs) == 0 {
		t.Fatal("FSLoader.Languages() returned empty")
	}

	found := false
	for _, l := range langs {
		if l == "en" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Languages() = %v, expected 'en' in list", langs)
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
			},
			"noun": map[string]any{
				"widget": map[string]any{
					"one":   "widget",
					"other": "widgets",
				},
			},
			"word": map[string]any{
				"api": "API",
			},
			"punct": map[string]any{
				"label":    ":",
				"progress": "...",
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

	// Noun extracted
	if n, ok := grammar.Nouns["widget"]; !ok {
		t.Error("noun 'widget' not extracted")
	} else {
		if n.Other != "widgets" {
			t.Errorf("widget.other = %q, want 'widgets'", n.Other)
		}
	}

	// Word extracted
	if grammar.Words["api"] != "API" {
		t.Errorf("word 'api' = %q, want 'API'", grammar.Words["api"])
	}

	// Punct extracted
	if grammar.Punct.LabelSuffix != ":" {
		t.Errorf("punct.label = %q, want ':'", grammar.Punct.LabelSuffix)
	}

	// Articles extracted
	if grammar.Articles.IndefiniteDefault != "a" {
		t.Errorf("article.indefinite.default = %q, want 'a'", grammar.Articles.IndefiniteDefault)
	}

	// Regular keys flattened
	if msg, ok := messages["prompt.yes"]; !ok || msg.Text != "y" {
		t.Errorf("prompt.yes not flattened correctly, got %+v", messages["prompt.yes"])
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

func TestCustomFSLoader(t *testing.T) {
	fs := fstest.MapFS{
		"locales/test.json": &fstest.MapFile{
			Data: []byte(`{
				"gram": {
					"verb": {
						"zap": { "base": "zap", "past": "zapped", "gerund": "zapping" }
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
}
