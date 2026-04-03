package i18n

import (
	"errors"
	"io/fs"
	"math"
	"path"
	"slices"
	"sync"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
)

// FSLoader loads translations from a filesystem (embedded or disk).
type FSLoader struct {
	fsys fs.FS
	dir  string

	languages []string
	langOnce  sync.Once
	langErr   error
}

// NewFSLoader creates a loader for the given filesystem and directory.
func NewFSLoader(fsys fs.FS, dir string) *FSLoader {
	return &FSLoader{fsys: fsys, dir: dir}
}

// Load implements Loader.Load.
func (l *FSLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	variants := localeFilenameCandidates(lang)
	var data []byte
	var err error
	var firstNonMissingErr error
	for _, filename := range variants {
		filePath := path.Join(l.dir, filename)
		data, err = fs.ReadFile(l.fsys, filePath)
		if err == nil {
			break
		}
		if firstNonMissingErr == nil && !errors.Is(err, fs.ErrNotExist) {
			firstNonMissingErr = err
		}
	}
	if err != nil {
		if firstNonMissingErr != nil {
			err = firstNonMissingErr
		}
		return nil, nil, log.E("FSLoader.Load", "locale not found: "+lang, err)
	}

	var raw map[string]any
	if r := core.JSONUnmarshal(data, &raw); !r.OK {
		return nil, nil, log.E("FSLoader.Load", "invalid JSON in locale: "+lang, r.Value.(error))
	}

	messages := make(map[string]Message)
	grammar := &GrammarData{
		Verbs: make(map[string]VerbForms),
		Nouns: make(map[string]NounForms),
		Words: make(map[string]string),
	}

	flattenWithGrammar("", raw, messages, grammar)

	return messages, grammar, nil
}

func localeFilenameCandidates(lang string) []string {
	// Preserve the documented lookup order: exact tag first, then underscore /
	// hyphen variants, then the base language tag.
	variants := make([]string, 0, 4)
	addVariant := func(candidate string) {
		for _, existing := range variants {
			if existing == candidate {
				return
			}
		}
		variants = append(variants, candidate)
	}
	canonical := normalizeLanguageTag(lang)
	addTag := func(tag string) {
		if tag == "" {
			return
		}
		addVariant(tag + ".json")
		addVariant(core.Replace(tag, "-", "_") + ".json")
		addVariant(core.Replace(tag, "_", "-") + ".json")
	}
	addTag(lang)
	if canonical != "" && canonical != lang {
		addTag(canonical)
	}
	if base := baseLanguageTag(canonical); base != "" && base != canonical {
		addTag(base)
	}
	return variants
}

// Languages implements Loader.Languages.
func (l *FSLoader) Languages() []string {
	l.langOnce.Do(func() {
		entries, err := fs.ReadDir(l.fsys, l.dir)
		if err != nil {
			l.langErr = log.E("FSLoader.Languages", "read locale directory: "+l.dir, err)
			return
		}
		seen := make(map[string]struct{}, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !core.HasSuffix(entry.Name(), ".json") {
				continue
			}
			lang := core.TrimSuffix(entry.Name(), ".json")
			lang = normalizeLanguageTag(core.Replace(lang, "_", "-"))
			if lang == "" {
				continue
			}
			if _, ok := seen[lang]; ok {
				continue
			}
			seen[lang] = struct{}{}
			l.languages = append(l.languages, lang)
		}
		slices.Sort(l.languages)
	})
	return append([]string(nil), l.languages...)
}

// LanguagesErr returns any error from the directory scan.
func (l *FSLoader) LanguagesErr() error {
	l.Languages()
	return l.langErr
}

var _ Loader = (*FSLoader)(nil)

// --- Flatten helpers ---

func flatten(prefix string, data map[string]any, out map[string]Message) {
	flattenWithGrammar(prefix, data, out, nil)
}

func flattenWithGrammar(prefix string, data map[string]any, out map[string]Message, grammar *GrammarData) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			if grammar != nil && loadGrammarWord(fullKey, v, grammar) {
				continue
			}
			out[fullKey] = Message{Text: v}

		case map[string]any:
			if grammar != nil && loadGrammarVerb(fullKey, key, v, grammar) {
				continue
			}

			if grammar != nil && loadGrammarNoun(fullKey, key, v, grammar) {
				continue
			}

			if grammar != nil && loadGrammarSignals(fullKey, v, grammar) {
				continue
			}

			if grammar != nil && loadGrammarArticle(fullKey, v, grammar) {
				continue
			}

			if grammar != nil && loadGrammarPunctuation(fullKey, v, grammar) {
				continue
			}

			if grammar != nil && loadGrammarNumber(fullKey, v, grammar) {
				continue
			}

			// CLDR plural object
			if isPluralObject(v) {
				msg := Message{}
				if zero, ok := v["zero"].(string); ok {
					msg.Zero = zero
				}
				if one, ok := v["one"].(string); ok {
					msg.One = one
				}
				if two, ok := v["two"].(string); ok {
					msg.Two = two
				}
				if few, ok := v["few"].(string); ok {
					msg.Few = few
				}
				if many, ok := v["many"].(string); ok {
					msg.Many = many
				}
				if other, ok := v["other"].(string); ok {
					msg.Other = other
				}
				out[fullKey] = msg
			} else {
				flattenWithGrammar(fullKey, v, out, grammar)
			}
		}
	}
}

func loadGrammarWord(fullKey, value string, grammar *GrammarData) bool {
	if grammar == nil || !core.HasPrefix(fullKey, "gram.word.") {
		return false
	}
	wordKey := core.TrimPrefix(fullKey, "gram.word.")
	if shouldSkipDeprecatedEnglishGrammarEntry(fullKey) {
		return true
	}
	grammar.Words[core.Lower(wordKey)] = value
	return true
}

func loadGrammarVerb(fullKey, key string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || !isVerbFormObject(v) {
		return false
	}
	verbName := key
	if base, ok := v["base"].(string); ok && base != "" {
		verbName = base
	}
	if core.HasPrefix(fullKey, "gram.verb.") {
		after := core.TrimPrefix(fullKey, "gram.verb.")
		if base, ok := v["base"].(string); !ok || base == "" {
			verbName = after
		}
	}
	forms := VerbForms{}
	if past, ok := v["past"].(string); ok {
		forms.Past = past
	}
	if gerund, ok := v["gerund"].(string); ok {
		forms.Gerund = gerund
	}
	grammar.Verbs[core.Lower(verbName)] = forms
	return true
}

func loadGrammarNoun(fullKey, key string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || !(core.HasPrefix(fullKey, "gram.noun.") || isNounFormObject(v)) {
		return false
	}
	nounName := key
	if core.HasPrefix(fullKey, "gram.noun.") {
		nounName = core.TrimPrefix(fullKey, "gram.noun.")
	}
	if shouldSkipDeprecatedEnglishGrammarEntry(fullKey) {
		return true
	}
	_, hasOne := v["one"]
	_, hasOther := v["other"]
	if !hasOne || !hasOther {
		return false
	}
	forms := NounForms{}
	if one, ok := v["one"].(string); ok {
		forms.One = one
	}
	if other, ok := v["other"].(string); ok {
		forms.Other = other
	}
	if gender, ok := v["gender"].(string); ok {
		forms.Gender = gender
	}
	grammar.Nouns[core.Lower(nounName)] = forms
	return true
}

func loadGrammarSignals(fullKey string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || fullKey != "gram.signal" {
		return false
	}
	loadSignalStringList := func(dst *[]string, raw any) {
		arr, ok := raw.([]any)
		if !ok {
			return
		}
		for _, item := range arr {
			if s, ok := item.(string); ok {
				*dst = append(*dst, core.Lower(s))
			}
		}
	}
	loadSignalStringList(&grammar.Signals.NounDeterminers, v["noun_determiner"])
	loadSignalStringList(&grammar.Signals.VerbAuxiliaries, v["verb_auxiliary"])
	loadSignalStringList(&grammar.Signals.VerbInfinitive, v["verb_infinitive"])
	loadSignalStringList(&grammar.Signals.VerbNegation, v["verb_negation"])
	if priors, ok := v["prior"].(map[string]any); ok {
		loadSignalPriors(grammar, priors)
	}
	if priors, ok := v["priors"].(map[string]any); ok {
		loadSignalPriors(grammar, priors)
	}
	return true
}

func loadGrammarArticle(fullKey string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || fullKey != "gram.article" {
		return false
	}
	if indef, ok := v["indefinite"].(map[string]any); ok {
		if def, ok := indef["default"].(string); ok {
			grammar.Articles.IndefiniteDefault = def
		}
		if vowel, ok := indef["vowel"].(string); ok {
			grammar.Articles.IndefiniteVowel = vowel
		}
	}
	if def, ok := v["definite"].(string); ok {
		grammar.Articles.Definite = def
	}
	if bg, ok := v["by_gender"].(map[string]any); ok {
		grammar.Articles.ByGender = make(map[string]string, len(bg))
		for g, art := range bg {
			if s, ok := art.(string); ok {
				grammar.Articles.ByGender[g] = s
			}
		}
	}
	return true
}

func loadGrammarPunctuation(fullKey string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || fullKey != "gram.punct" {
		return false
	}
	if label, ok := v["label"].(string); ok {
		grammar.Punct.LabelSuffix = label
	}
	if progress, ok := v["progress"].(string); ok {
		grammar.Punct.ProgressSuffix = progress
	}
	return true
}

func loadGrammarNumber(fullKey string, v map[string]any, grammar *GrammarData) bool {
	if grammar == nil || fullKey != "gram.number" {
		return false
	}
	if thousands, ok := v["thousands"].(string); ok {
		grammar.Number.ThousandsSep = thousands
	}
	if decimal, ok := v["decimal"].(string); ok {
		grammar.Number.DecimalSep = decimal
	}
	if percent, ok := v["percent"].(string); ok {
		grammar.Number.PercentFmt = percent
	}
	return true
}

func isVerbFormObject(m map[string]any) bool {
	_, hasPast := m["past"]
	_, hasGerund := m["gerund"]
	// Verb objects are identified by their inflected forms. A bare "base"
	// field is metadata, not enough to claim the object is a verb table.
	return (hasPast || hasGerund) && !isPluralObject(m)
}

func isNounFormObject(m map[string]any) bool {
	_, hasGender := m["gender"]
	return hasGender
}

func isPluralObject(m map[string]any) bool {
	_, hasZero := m["zero"]
	_, hasOne := m["one"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]
	_, hasOther := m["other"]
	if !hasZero && !hasOne && !hasTwo && !hasFew && !hasMany && !hasOther {
		return false
	}
	for _, v := range m {
		if _, isMap := v.(map[string]any); isMap {
			return false
		}
	}
	return true
}

func loadSignalPriors(grammar *GrammarData, priors map[string]any) {
	if grammar == nil || len(priors) == 0 {
		return
	}
	if grammar.Signals.Priors == nil {
		grammar.Signals.Priors = make(map[string]map[string]float64, len(priors))
	}
	for word, raw := range priors {
		bucket, ok := raw.(map[string]any)
		if !ok || len(bucket) == 0 {
			continue
		}
		key := core.Lower(word)
		if grammar.Signals.Priors[key] == nil {
			grammar.Signals.Priors[key] = make(map[string]float64, len(bucket))
		}
		for role, value := range bucket {
			score, ok := float64Value(value)
			if !ok || !validSignalPriorScore(score) {
				continue
			}
			grammar.Signals.Priors[key][core.Lower(role)] = score
		}
	}
}

func validSignalPriorScore(score float64) bool {
	return !math.IsNaN(score) && !math.IsInf(score, 0) && score >= 0
}

func float64Value(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case int16:
		return float64(n), true
	case int8:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint8:
		return float64(n), true
	default:
		return 0, false
	}
}

func shouldSkipDeprecatedEnglishGrammarEntry(fullKey string) bool {
	switch fullKey {
	case "gram.noun.passed", "gram.noun.failed", "gram.noun.skipped",
		"gram.word.passed", "gram.word.failed", "gram.word.skipped":
		return true
	default:
		return false
	}
}
