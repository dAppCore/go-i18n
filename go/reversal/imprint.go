package reversal

import "math"

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
			conf := tok.Confidence
			if conf == 0 {
				conf = 1.0
			}
			verbCount++
			base := tok.VerbInfo.Base
			imp.VerbDistribution[base] += conf
			imp.TenseDistribution[tok.VerbInfo.Tense] += conf
			verbBases[base] = true

			// Dual-class: contribute alt confidence to noun distribution
			if tok.AltType == TokenNoun && tok.NounInfo.Base != "" {
				imp.NounDistribution[tok.NounInfo.Base] += tok.AltConf
				nounBases[tok.NounInfo.Base] = true
				totalNouns++
			}

		case TokenNoun:
			conf := tok.Confidence
			if conf == 0 {
				conf = 1.0
			}
			nounCount++
			base := tok.NounInfo.Base
			imp.NounDistribution[base] += conf
			nounBases[base] = true
			totalNouns++
			if tok.NounInfo.Plural {
				pluralNouns++
			}

			// Dual-class: contribute alt confidence to verb distribution
			if tok.AltType == TokenVerb && tok.VerbInfo.Base != "" {
				imp.VerbDistribution[tok.VerbInfo.Base] += tok.AltConf
				imp.TenseDistribution[tok.VerbInfo.Tense] += tok.AltConf
				verbBases[tok.VerbInfo.Base] = true
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

// Similar returns weighted cosine similarity between two imprints (0.0-1.0).
// Weights: verb(0.30), tense(0.20), noun(0.25), article(0.15), punct(0.10).
func (a GrammarImprint) Similar(b GrammarImprint) float64 {
	// Two empty imprints are identical.
	if a.TokenCount == 0 && b.TokenCount == 0 {
		return 1.0
	}

	type component struct {
		weight float64
		a, b   map[string]float64
	}

	components := []component{
		{0.30, a.VerbDistribution, b.VerbDistribution},
		{0.20, a.TenseDistribution, b.TenseDistribution},
		{0.25, a.NounDistribution, b.NounDistribution},
		{0.15, a.ArticleUsage, b.ArticleUsage},
		{0.10, a.PunctuationPattern, b.PunctuationPattern},
	}

	var totalWeight float64
	var weightedSum float64

	for _, c := range components {
		// Skip components where both maps are empty (no signal).
		if len(c.a) == 0 && len(c.b) == 0 {
			continue
		}
		totalWeight += c.weight
		weightedSum += c.weight * mapSimilarity(c.a, c.b)
	}

	if totalWeight == 0 {
		return 1.0
	}

	return weightedSum / totalWeight
}

// mapSimilarity computes cosine similarity between two frequency maps.
// Returns 1.0 for identical distributions, 0.0 for completely disjoint.
func mapSimilarity(a, b map[string]float64) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Collect the union of keys.
	keys := make(map[string]bool)
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}

	var dot, magA, magB float64
	for k := range keys {
		va := a[k]
		vb := b[k]
		dot += va * vb
		magA += va * va
		magB += vb * vb
	}

	denom := math.Sqrt(magA) * math.Sqrt(magB)
	if denom == 0 {
		return 0.0
	}

	return dot / denom
}
