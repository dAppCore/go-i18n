package i18n

import (
	"math"
	"testing"
	"testing/fstest"
)

var (
	benchLoaderFloatSink   float64
	benchLoaderGrammarSink *GrammarData
	benchLoaderMetaSink    IntentMeta
	benchLoaderMessageSink Message
	benchLoaderSink        *FSLoader
)

const benchLoaderLocaleJSON = `{
  "app": {
    "hello": "Hello {{.Name}}",
    "files": {"one": "{{.Count}} file", "other": "{{.Count}} files"},
    "nested": {"label": "Nested"}
  },
  "gram": {
    "word": {"file": "fichier", "passed": "passed"},
    "verb": {"delete": {"past": "removed", "gerund": "removing"}},
    "noun": {"file": {"one": "file", "other": "files", "gender": "m"}},
    "signal": {
      "noun_determiner": ["the", "this"],
      "verb_auxiliary": ["will"],
      "verb_infinitive": ["to"],
      "verb_negation": ["not"],
      "prior": {"archive": {"noun": 0.7, "verb": 0.3}}
    },
    "article": {
      "the": "le",
      "a": {"default": "un", "vowel": "une"},
      "by_gender": {"m": "le", "f": "la"}
    },
    "punct": {"label": " :", "progress": "..."},
    "number": {"thousands": " ", "decimal": ",", "percent": "%s %%"}
  },
  "core": {
    "delete": {
      "_meta": {"type": "action", "verb": "delete", "dangerous": "yes", "default": "no", "supports": ["force", "dry-run"]},
      "question": "Delete {{.Subject}}?",
      "confirm": "Really delete {{.Subject}}?",
      "success": "{{.Subject}} removed",
      "failure": {"other": "Unable to delete {{.Subject}}"}
    }
  }
}`

func benchLoaderLocaleFS() fstest.MapFS {
	return fstest.MapFS{
		"locales/en_GB.json": &fstest.MapFile{Data: []byte(benchLoaderLocaleJSON)},
		"locales/en.json":    &fstest.MapFile{Data: []byte(`{"fallback":"Fallback"}`)},
		"locales/fr.json":    &fstest.MapFile{Data: []byte(`{"bonjour":"Bonjour"}`)},
		"locales/readme.txt": &fstest.MapFile{Data: []byte("not a locale")},
	}
}

func benchLoaderRaw() map[string]any {
	return map[string]any{
		"app": map[string]any{
			"hello": "Hello {{.Name}}",
			"files": map[string]any{
				"one":   "{{.Count}} file",
				"other": "{{.Count}} files",
			},
			"nested": map[string]any{"label": "Nested"},
		},
		"gram": map[string]any{
			"word": map[string]any{
				"file":   "fichier",
				"passed": "passed",
			},
			"verb": map[string]any{
				"delete": map[string]any{"past": "removed", "gerund": "removing"},
			},
			"noun": map[string]any{
				"file": map[string]any{"one": "file", "other": "files", "gender": "m"},
			},
			"signal": map[string]any{
				"noun_determiner": []any{"the", "this"},
				"verb_auxiliary":  []any{"will"},
				"verb_infinitive": []any{"to"},
				"verb_negation":   []any{"not"},
				"prior": map[string]any{
					"archive": map[string]any{"noun": 0.7, "verb": 0.3},
				},
			},
			"article": map[string]any{
				"the":       "le",
				"a":         map[string]any{"default": "un", "vowel": "une"},
				"by_gender": map[string]any{"m": "le", "f": "la"},
			},
			"punct":  map[string]any{"label": " :", "progress": "..."},
			"number": map[string]any{"thousands": " ", "decimal": ",", "percent": "%s %%"},
		},
		"core": map[string]any{
			"delete": map[string]any{
				"_meta": map[string]any{
					"type":      "action",
					"verb":      "delete",
					"dangerous": "yes",
					"default":   "no",
					"supports":  []any{"force", "dry-run"},
				},
				"question": "Delete {{.Subject}}?",
				"confirm":  "Really delete {{.Subject}}?",
				"success":  "{{.Subject}} removed",
				"failure":  map[string]any{"other": "Unable to delete {{.Subject}}"},
			},
		},
	}
}

func benchLoaderGrammar() *GrammarData {
	return &GrammarData{
		Verbs:   make(map[string]VerbForms),
		Nouns:   make(map[string]NounForms),
		Words:   make(map[string]string),
		Intents: make(map[string]Intent),
	}
}

func BenchmarkNewFSLoader(b *testing.B) {
	fsys := benchLoaderLocaleFS()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchLoaderSink = NewFSLoader(fsys, "locales")
	}
}

func BenchmarkLoad(b *testing.B) {
	loader := NewFSLoader(benchLoaderLocaleFS(), "locales")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = loader.Load("en-GB")
	}
}

func BenchmarkFSLocalePath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = fsLocalePath("locales/", "en.json")
		benchServiceStringSink = fsLocalePath(".", "en.json")
	}
}

func BenchmarkLocaleFilenameCandidates(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = localeFilenameCandidates("en-GB")
		benchServiceStringsSink = localeFilenameCandidates("pt_BR")
	}
}

func BenchmarkLanguages(b *testing.B) {
	loader := NewFSLoader(benchLoaderLocaleFS(), "locales")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = loader.Languages()
	}
}

func BenchmarkLanguagesErr(b *testing.B) {
	loader := NewFSLoader(benchLoaderLocaleFS(), "missing")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = loader.LanguagesErr()
	}
}

func BenchmarkFlatten(b *testing.B) {
	raw := benchLoaderRaw()
	out := make(map[string]Message)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flatten("", raw, out)
	}
	benchLoaderMessageSink = out["app.hello"]
}

func BenchmarkFlattenWithGrammar(b *testing.B) {
	raw := benchLoaderRaw()
	out := make(map[string]Message)
	grammar := benchLoaderGrammar()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenWithGrammar("", raw, out, grammar)
	}
	benchLoaderGrammarSink = grammar
}

func BenchmarkFlattenWithGrammarAndIntents(b *testing.B) {
	raw := benchLoaderRaw()
	out := make(map[string]Message)
	grammar := benchLoaderGrammar()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenWithGrammarAndIntents("", raw, out, grammar)
	}
	benchLoaderGrammarSink = grammar
}

func BenchmarkLoadIntentBlock(b *testing.B) {
	block := benchLoaderRaw()["core"].(map[string]any)["delete"].(map[string]any)
	out := make(map[string]Message)
	grammar := benchLoaderGrammar()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadIntentBlock("core.delete", block, out, grammar)
	}
}

func BenchmarkIsIntentBlock(b *testing.B) {
	block := benchLoaderRaw()["core"].(map[string]any)["delete"].(map[string]any)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isIntentBlock("core.delete", block)
		benchServiceBoolSink = isIntentBlock("app.hello", block)
	}
}

func BenchmarkMessageFromValue(b *testing.B) {
	plural := map[string]any{"one": "one file", "other": "files"}
	invalid := map[string]any{"nested": map[string]any{"text": "bad"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchLoaderMessageSink, benchServiceBoolSink = messageFromValue("Hello")
		benchLoaderMessageSink, benchServiceBoolSink = messageFromValue(plural)
		benchLoaderMessageSink, benchServiceBoolSink = messageFromValue(invalid)
	}
}

func BenchmarkParseIntentMeta(b *testing.B) {
	meta := map[string]any{
		"type":      "action",
		"verb":      "delete",
		"dangerous": "yes",
		"default":   "no",
		"supports":  []any{"force", "dry-run"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchLoaderMetaSink = parseIntentMeta(meta)
	}
}

func BenchmarkStringFromMap(b *testing.B) {
	values := map[string]any{"Type": " Action "}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink, benchServiceBoolSink = stringFromMap(values, "type", "Type")
	}
}

func BenchmarkBoolFromMap(b *testing.B) {
	values := map[string]any{"Dangerous": "yes", "Enabled": false}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink, benchGrammarBoolSink = boolFromMap(values, "dangerous", "Dangerous")
		benchServiceBoolSink, benchGrammarBoolSink = boolFromMap(values, "Enabled")
	}
}

func BenchmarkStringSliceFromMap(b *testing.B) {
	values := map[string]any{
		"supports": []any{"Force", "dry-run", ""},
		"aliases":  []string{"CLI", "api"},
		"single":   "One",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = stringSliceFromMap(values, "supports")
		benchServiceStringsSink = stringSliceFromMap(values, "aliases")
		benchServiceStringsSink = stringSliceFromMap(values, "single")
	}
}

func BenchmarkLoadGrammarWord(b *testing.B) {
	grammar := benchLoaderGrammar()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarWord("gram.word.file", "fichier", grammar)
		benchServiceBoolSink = loadGrammarWord("gram.word.passed", "passed", grammar)
	}
}

func BenchmarkLoadGrammarVerb(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{"past": "removed", "gerund": "removing"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarVerb("gram.verb.delete", "delete", value, grammar)
		benchServiceBoolSink = loadGrammarVerb("common.verb.build", "build", value, grammar)
	}
}

func BenchmarkLoadGrammarNoun(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{"one": "file", "other": "files", "gender": "m"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarNoun("gram.noun.file", "file", value, grammar)
		benchServiceBoolSink = loadGrammarNoun("common.noun.folder", "folder", value, grammar)
	}
}

func BenchmarkLoadGrammarSignals(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{
		"noun_determiner": []any{"the", "this"},
		"verb_auxiliary":  []any{"will"},
		"verb_infinitive": []any{"to"},
		"verb_negation":   []any{"not"},
		"prior": map[string]any{
			"archive": map[string]any{"noun": 0.7, "verb": 0.3},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarSignals("gram.signal", value, grammar)
	}
}

func BenchmarkLoadGrammarArticle(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{
		"the":      "le",
		"a":        map[string]string{"default": "un", "vowel": "une"},
		"byGender": map[string]any{"m": "le", "f": "la"},
		"definite": "le",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarArticle("gram.article", value, grammar)
	}
}

func BenchmarkFirstNonEmptyString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = firstNonEmptyString("", "  ", "value")
	}
}

func BenchmarkLoadGrammarPunctuation(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{"label": " :", "progress": "..."}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarPunctuation("gram.punct", value, grammar)
	}
}

func BenchmarkLoadGrammarNumber(b *testing.B) {
	grammar := benchLoaderGrammar()
	value := map[string]any{"thousands": " ", "decimal": ",", "percent": "%s %%"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = loadGrammarNumber("gram.number", value, grammar)
	}
}

func BenchmarkIsVerbFormObject(b *testing.B) {
	verb := map[string]any{"past": "removed", "gerund": "removing"}
	plural := map[string]any{"one": "file", "other": "files"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isVerbFormObject(verb)
		benchServiceBoolSink = isVerbFormObject(plural)
	}
}

func BenchmarkIsNounFormObject(b *testing.B) {
	noun := map[string]any{"gender": "m"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isNounFormObject(noun)
	}
}

func BenchmarkIsPluralObject(b *testing.B) {
	plural := map[string]any{"one": "file", "other": "files"}
	nested := map[string]any{"one": map[string]any{"text": "bad"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = isPluralObject(plural)
		benchServiceBoolSink = isPluralObject(nested)
	}
}

func BenchmarkLoadSignalPriors(b *testing.B) {
	grammar := benchLoaderGrammar()
	priors := map[string]any{
		"archive": map[string]any{"noun": 0.7, "verb": 0.3, "bad": -1},
		"skip":    map[string]any{},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadSignalPriors(grammar, priors)
	}
}

func BenchmarkFloat64Value(b *testing.B) {
	values := []any{float64(1), float32(2), int(3), int64(4), int32(5), int16(6), int8(7), uint(8), uint64(9), uint32(10), uint16(11), uint8(12), "bad"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchLoaderFloatSink, benchServiceBoolSink = float64Value(value)
		}
	}
}

func BenchmarkValidSignalPriorScore(b *testing.B) {
	values := []float64{0, 0.7, -1, math.NaN(), math.Inf(1)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			benchServiceBoolSink = validSignalPriorScore(value)
		}
	}
}

func BenchmarkShouldSkipDeprecatedEnglishGrammarEntry(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = shouldSkipDeprecatedEnglishGrammarEntry("gram.word.passed")
		benchServiceBoolSink = shouldSkipDeprecatedEnglishGrammarEntry("gram.word.file")
	}
}
