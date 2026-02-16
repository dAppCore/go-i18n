package reversal

// GrammarImprint is a low-dimensional grammar feature vector.
type GrammarImprint struct {
	VerbDistribution   map[string]float64 // verb base -> frequency
	TenseDistribution  map[string]float64 // "past"/"gerund"/"base" -> ratio
	NounDistribution   map[string]float64 // noun base -> frequency
	PluralRatio        float64            // proportion of plural nouns (0.0-1.0)
	DomainVocabulary   map[string]int     // gram.word category -> hit count
	ArticleUsage       map[string]float64 // "definite"/"indefinite" -> ratio
	PunctuationPattern map[string]float64 // "label"/"progress"/"question" -> ratio
	TokenCount         int
	UniqueVerbs        int
	UniqueNouns        int
}

// NewImprint calculates a GrammarImprint from classified tokens.
func NewImprint(tokens []Token) GrammarImprint {
	imp := GrammarImprint{
		VerbDistribution:   make(map[string]float64),
		TenseDistribution:  make(map[string]float64),
		NounDistribution:   make(map[string]float64),
		DomainVocabulary:   make(map[string]int),
		ArticleUsage:       make(map[string]float64),
		PunctuationPattern: make(map[string]float64),
	}

	if len(tokens) == 0 {
		return imp
	}

	imp.TokenCount = len(tokens)

	verbBases := make(map[string]bool)
	nounBases := make(map[string]bool)
	var verbCount, nounCount, articleCount, punctCount int
	var pluralNouns, totalNouns int

	for _, tok := range tokens {
		switch tok.Type {
		case TokenVerb:
			verbCount++
			base := tok.VerbInfo.Base
			imp.VerbDistribution[base]++
			imp.TenseDistribution[tok.VerbInfo.Tense]++
			verbBases[base] = true

		case TokenNoun:
			nounCount++
			base := tok.NounInfo.Base
			imp.NounDistribution[base]++
			nounBases[base] = true
			totalNouns++
			if tok.NounInfo.Plural {
				pluralNouns++
			}

		case TokenArticle:
			articleCount++
			imp.ArticleUsage[tok.ArtType]++

		case TokenWord:
			imp.DomainVocabulary[tok.WordCat]++

		case TokenPunctuation:
			punctCount++
			imp.PunctuationPattern[tok.PunctType]++
		}
	}

	imp.UniqueVerbs = len(verbBases)
	imp.UniqueNouns = len(nounBases)

	// Calculate plural ratio
	if totalNouns > 0 {
		imp.PluralRatio = float64(pluralNouns) / float64(totalNouns)
	}

	// Normalise frequency maps to sum to 1.0
	normaliseMap(imp.VerbDistribution)
	normaliseMap(imp.TenseDistribution)
	normaliseMap(imp.NounDistribution)
	normaliseMap(imp.ArticleUsage)
	normaliseMap(imp.PunctuationPattern)

	return imp
}

// normaliseMap scales all values in a map so they sum to 1.0.
// If the map is empty or sums to zero, it is left unchanged.
func normaliseMap(m map[string]float64) {
	var total float64
	for _, v := range m {
		total += v
	}
	if total == 0 {
		return
	}
	for k, v := range m {
		m[k] = v / total
	}
}
