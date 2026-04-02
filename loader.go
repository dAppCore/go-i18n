package i18n

import (
	"io/fs"
	"path"
	"strings"
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
	variants := []string{
		lang + ".json",
		core.Replace(lang, "-", "_") + ".json",
		core.Replace(lang, "_", "-") + ".json",
	}

	var data []byte
	var err error
	for _, filename := range variants {
		filePath := path.Join(l.dir, filename)
		data, err = fs.ReadFile(l.fsys, filePath)
		if err == nil {
			break
		}
	}
	if err != nil {
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

// Languages implements Loader.Languages.
func (l *FSLoader) Languages() []string {
	l.langOnce.Do(func() {
		entries, err := fs.ReadDir(l.fsys, l.dir)
		if err != nil {
			l.langErr = log.E("FSLoader.Languages", "read locale directory: "+l.dir, err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !core.HasSuffix(entry.Name(), ".json") {
				continue
			}
			lang := core.TrimSuffix(entry.Name(), ".json")
			lang = core.Replace(lang, "_", "-")
			l.languages = append(l.languages, lang)
		}
	})
	return l.languages
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
			if grammar != nil && core.HasPrefix(fullKey, "gram.word.") {
				wordKey := core.TrimPrefix(fullKey, "gram.word.")
				grammar.Words[core.Lower(wordKey)] = v
				continue
			}
			out[fullKey] = Message{Text: v}

		case map[string]any:
			// Verb form object (has base/past/gerund keys)
			if grammar != nil && isVerbFormObject(v) {
				verbName := key
				if base, ok := v["base"].(string); ok && base != "" {
					verbName = base
				}
				if after, ok := strings.CutPrefix(fullKey, "gram.verb."); ok {
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
				if forms.Past == "" || forms.Gerund == "" {
					continue
				}
				grammar.Verbs[core.Lower(verbName)] = forms
				continue
			}

			// Noun form object (under gram.noun.* or has gender field)
			if grammar != nil && (core.HasPrefix(fullKey, "gram.noun.") || isNounFormObject(v)) {
				nounName := key
				if after, ok := strings.CutPrefix(fullKey, "gram.noun."); ok {
					nounName = after
				}
				_, hasOne := v["one"]
				_, hasOther := v["other"]
				if hasOne && hasOther {
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
					continue
				}
			}

			// Signal data for disambiguation
			if grammar != nil && fullKey == "gram.signal" {
				if nd, ok := v["noun_determiner"]; ok {
					if arr, ok := nd.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.NounDeterminers = append(grammar.Signals.NounDeterminers, core.Lower(s))
							}
						}
					}
				}
				if va, ok := v["verb_auxiliary"]; ok {
					if arr, ok := va.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.VerbAuxiliaries = append(grammar.Signals.VerbAuxiliaries, core.Lower(s))
							}
						}
					}
				}
				if vi, ok := v["verb_infinitive"]; ok {
					if arr, ok := vi.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.VerbInfinitive = append(grammar.Signals.VerbInfinitive, core.Lower(s))
							}
						}
					}
				}
				if priors, ok := v["prior"].(map[string]any); ok {
					loadSignalPriors(grammar, priors)
				}
				if priors, ok := v["priors"].(map[string]any); ok {
					loadSignalPriors(grammar, priors)
				}
				continue
			}

			// Article configuration
			if grammar != nil && fullKey == "gram.article" {
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
				continue
			}

			// Punctuation rules
			if grammar != nil && fullKey == "gram.punct" {
				if label, ok := v["label"].(string); ok {
					grammar.Punct.LabelSuffix = label
				}
				if progress, ok := v["progress"].(string); ok {
					grammar.Punct.ProgressSuffix = progress
				}
				continue
			}

			// Number formatting rules
			if grammar != nil && fullKey == "gram.number" {
				if thousands, ok := v["thousands"].(string); ok {
					grammar.Number.ThousandsSep = thousands
				}
				if decimal, ok := v["decimal"].(string); ok {
					grammar.Number.DecimalSep = decimal
				}
				if percent, ok := v["percent"].(string); ok {
					grammar.Number.PercentFmt = percent
				}
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

func isVerbFormObject(m map[string]any) bool {
	_, hasBase := m["base"]
	_, hasPast := m["past"]
	_, hasGerund := m["gerund"]
	return (hasBase || hasPast || hasGerund) && !isPluralObject(m)
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
			if score := toFloat64(value); score != 0 {
				grammar.Signals.Priors[key][core.Lower(role)] = score
			}
		}
	}
}
