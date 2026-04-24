package reversal

import (
	"unicode"

	"dappco.re/go/core"
	i18n "dappco.re/go/i18n"
)

// Multiplier generates deterministic grammatical variants of text
// for training data augmentation. Zero API calls.
type Multiplier struct {
	tokeniser *Tokeniser
}

// NewMultiplier creates a Multiplier using the default English tokeniser.
func NewMultiplier() *Multiplier {
	return &Multiplier{tokeniser: NewTokeniser()}
}

// NewMultiplierForLang creates a Multiplier for the specified language.
func NewMultiplierForLang(lang string) *Multiplier {
	return &Multiplier{tokeniser: NewTokeniserForLang(lang)}
}

// Expand produces: original + tense flips (past, gerund) + number flips (plural toggle) + combinations.
// All output is deterministic and grammatically correct.
func (m *Multiplier) Expand(text string) []string {
	text = core.Trim(text)
	if text == "" {
		return nil
	}

	tokens := m.tokeniser.Tokenise(text)
	if len(tokens) == 0 {
		return nil
	}

	// Collect indices of verbs and nouns for targeted replacement.
	var verbIndices []int
	var nounIndices []int
	for i, tok := range tokens {
		switch tok.Type {
		case TokenVerb:
			verbIndices = append(verbIndices, i)
		case TokenNoun:
			nounIndices = append(nounIndices, i)
		}
	}

	// Build the list of variants in deterministic order:
	// 1. Original
	// 2. Single verb transforms (past, gerund) for each verb
	// 3. Single noun transforms (plural toggle) for each noun
	// 4. Combined transforms (verb transform + noun transform)
	seen := make(map[string]bool)
	var results []string

	addVariant := func(s string) {
		if !seen[s] {
			seen[s] = true
			results = append(results, s)
		}
	}

	// 1. Original text
	addVariant(text)

	// 2. Verb transforms: for each verb, produce past and gerund variants
	for _, vi := range verbIndices {
		addVariant(m.reconstructWithVerbTransform(tokens, vi, "past"))
		addVariant(m.reconstructWithVerbTransform(tokens, vi, "gerund"))
		addVariant(m.reconstructWithVerbTransform(tokens, vi, "base"))
	}

	// 3. Noun transforms: for each noun, toggle plural/singular
	for _, ni := range nounIndices {
		addVariant(m.reconstructWithNounTransform(tokens, ni))
	}

	// 4. Combinations: each verb transform + each noun transform
	for _, vi := range verbIndices {
		for _, ni := range nounIndices {
			// past + noun toggle
			addVariant(m.reconstructWithVerbAndNounTransform(tokens, vi, "past", ni))

			// gerund + noun toggle
			addVariant(m.reconstructWithVerbAndNounTransform(tokens, vi, "gerund", ni))

			// base + noun toggle
			addVariant(m.reconstructWithVerbAndNounTransform(tokens, vi, "base", ni))
		}
	}

	return results
}

// applyVerbTransform returns a copy of tokens with the verb at index vi
// transformed to the specified tense ("past", "gerund", or "base").
func (m *Multiplier) applyVerbTransform(tokens []Token, vi int, targetTense string) []Token {
	result := make([]Token, len(tokens))
	copy(result, tokens)

	tok := tokens[vi]
	base := tok.VerbInfo.Base
	currentTense := tok.VerbInfo.Tense

	if currentTense == targetTense {
		return result
	}

	var newForm string
	switch targetTense {
	case "past":
		newForm = i18n.PastTense(base)
	case "gerund":
		newForm = i18n.Gerund(base)
	case "base":
		newForm = base
	}

	if newForm == "" {
		return result
	}

	// Preserve capitalisation of the original token.
	newForm = preserveCase(tok.Raw, newForm)

	result[vi] = Token{
		Raw:        newForm,
		Lower:      core.Lower(newForm),
		Type:       TokenVerb,
		Confidence: 1.0,
		VerbInfo: VerbMatch{
			Base:  base,
			Tense: targetTense,
			Form:  newForm,
		},
	}

	return result
}

// applyNounTransform returns a copy of tokens with the noun at index ni
// toggled between singular and plural.
func (m *Multiplier) applyNounTransform(tokens []Token, ni int) []Token {
	return m.applyNounTransformOnTokens(tokens, ni)
}

// applyNounTransformOnTokens returns a copy of the given tokens with the
// noun at index ni toggled between singular and plural.
func (m *Multiplier) applyNounTransformOnTokens(tokens []Token, ni int) []Token {
	result := make([]Token, len(tokens))
	copy(result, tokens)

	tok := tokens[ni]
	base := tok.NounInfo.Base
	isPlural := tok.NounInfo.Plural

	var newForm string
	var newPlural bool

	if isPlural {
		// Already plural, revert to singular (base form).
		newForm = base
		newPlural = false
	} else {
		// Singular, generate plural.
		newForm = i18n.PluralForm(base)
		newPlural = true
	}

	if newForm == "" {
		return result
	}

	// Preserve capitalisation.
	newForm = preserveCase(tok.Raw, newForm)

	result[ni] = Token{
		Raw:        newForm,
		Lower:      core.Lower(newForm),
		Type:       TokenNoun,
		Confidence: 1.0,
		NounInfo: NounMatch{
			Base:   base,
			Plural: newPlural,
			Form:   newForm,
		},
	}

	return result
}

func (m *Multiplier) reconstructWithVerbTransform(tokens []Token, vi int, targetTense string) string {
	return m.reconstructWithTransforms(tokens, vi, targetTense, -1)
}

func (m *Multiplier) reconstructWithNounTransform(tokens []Token, ni int) string {
	return m.reconstructWithTransforms(tokens, -1, "", ni)
}

func (m *Multiplier) reconstructWithVerbAndNounTransform(tokens []Token, vi int, targetTense string, ni int) string {
	return m.reconstructWithTransforms(tokens, vi, targetTense, ni)
}

func (m *Multiplier) reconstructWithTransforms(tokens []Token, vi int, targetTense string, ni int) string {
	b := core.NewBuilder()
	for i, tok := range tokens {
		if i > 0 {
			// Punctuation tokens should stay attached to the preceding token.
			if tok.Type == TokenPunctuation {
				b.WriteString(tok.Raw)
				continue
			}
			b.WriteByte(' ')
		}
		switch {
		case i == vi:
			b.WriteString(transformedVerbRaw(tok, targetTense))
		case i == ni:
			b.WriteString(transformedNounRaw(tok))
		default:
			b.WriteString(tok.Raw)
		}
	}
	return b.String()
}

func transformedVerbRaw(tok Token, targetTense string) string {
	base := tok.VerbInfo.Base
	currentTense := tok.VerbInfo.Tense
	if currentTense == targetTense {
		return tok.Raw
	}

	var newForm string
	switch targetTense {
	case "past":
		newForm = i18n.PastTense(base)
	case "gerund":
		newForm = i18n.Gerund(base)
	case "base":
		newForm = base
	}
	if newForm == "" {
		return tok.Raw
	}
	return preserveCase(tok.Raw, newForm)
}

func transformedNounRaw(tok Token) string {
	base := tok.NounInfo.Base
	if base == "" {
		return tok.Raw
	}

	var newForm string
	if tok.NounInfo.Plural {
		newForm = base
	} else {
		newForm = i18n.PluralForm(base)
	}
	if newForm == "" {
		return tok.Raw
	}
	return preserveCase(tok.Raw, newForm)
}

// reconstruct joins tokens back into a string, preserving spacing.
func reconstruct(tokens []Token) string {
	b := core.NewBuilder()
	for i, tok := range tokens {
		if i > 0 {
			// Punctuation tokens that were split from the previous word
			// should not have a leading space.
			if tok.Type == TokenPunctuation {
				b.WriteString(tok.Raw)
				continue
			}
			b.WriteByte(' ')
		}
		b.WriteString(tok.Raw)
	}
	return b.String()
}

// preserveCase applies the capitalisation pattern of the original word
// to the replacement word. If the original started with an uppercase
// letter, the replacement will too.
func preserveCase(original, replacement string) string {
	if len(original) == 0 || len(replacement) == 0 {
		return replacement
	}

	origRunes := []rune(original)
	repRunes := []rune(replacement)

	// If the original is all uppercase (like "DELETE"), make replacement all uppercase.
	if isAllUpper(original) && len(original) > 1 {
		return core.Upper(replacement)
	}

	// If the first character of the original is uppercase, capitalise the replacement.
	if unicode.IsUpper(origRunes[0]) {
		repRunes[0] = unicode.ToUpper(repRunes[0])
		return string(repRunes)
	}

	// Otherwise, ensure the replacement starts lowercase.
	repRunes[0] = unicode.ToLower(repRunes[0])
	return string(repRunes)
}

// isAllUpper returns true if every letter in the string is uppercase.
func isAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
