package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
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
		strings.ReplaceAll(lang, "-", "_") + ".json",
		strings.ReplaceAll(lang, "_", "-") + ".json",
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
		return nil, nil, fmt.Errorf("locale %q not found: %w", lang, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON in locale %q: %w", lang, err)
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
			l.langErr = fmt.Errorf("failed to read locale directory %q: %w", l.dir, err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			lang := strings.TrimSuffix(entry.Name(), ".json")
			lang = strings.ReplaceAll(lang, "_", "-")
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
			if grammar != nil && strings.HasPrefix(fullKey, "gram.word.") {
				wordKey := strings.TrimPrefix(fullKey, "gram.word.")
				grammar.Words[strings.ToLower(wordKey)] = v
				continue
			}
			out[fullKey] = Message{Text: v}

		case map[string]any:
			// Verb form object (has base/past/gerund keys)
			if grammar != nil && isVerbFormObject(v) {
				verbName := key
				if strings.HasPrefix(fullKey, "gram.verb.") {
					verbName = strings.TrimPrefix(fullKey, "gram.verb.")
				}
				forms := VerbForms{}
				if past, ok := v["past"].(string); ok {
					forms.Past = past
				}
				if gerund, ok := v["gerund"].(string); ok {
					forms.Gerund = gerund
				}
				grammar.Verbs[strings.ToLower(verbName)] = forms
				continue
			}

			// Noun form object (under gram.noun.* or has gender field)
			if grammar != nil && (strings.HasPrefix(fullKey, "gram.noun.") || isNounFormObject(v)) {
				nounName := key
				if strings.HasPrefix(fullKey, "gram.noun.") {
					nounName = strings.TrimPrefix(fullKey, "gram.noun.")
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
					grammar.Nouns[strings.ToLower(nounName)] = forms
					continue
				}
			}

			// Signal data for disambiguation
			if grammar != nil && fullKey == "gram.signal" {
				if nd, ok := v["noun_determiner"]; ok {
					if arr, ok := nd.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.NounDeterminers = append(grammar.Signals.NounDeterminers, strings.ToLower(s))
							}
						}
					}
				}
				if va, ok := v["verb_auxiliary"]; ok {
					if arr, ok := va.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.VerbAuxiliaries = append(grammar.Signals.VerbAuxiliaries, strings.ToLower(s))
							}
						}
					}
				}
				if vi, ok := v["verb_infinitive"]; ok {
					if arr, ok := vi.([]any); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok {
								grammar.Signals.VerbInfinitive = append(grammar.Signals.VerbInfinitive, strings.ToLower(s))
							}
						}
					}
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
