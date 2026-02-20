package reversal

import (
	"fmt"
	"math"
	"sort"
)

// ClassifiedText is a text sample with a domain label (from 1B model or ground truth).
type ClassifiedText struct {
	Text   string
	Domain string
}

// ReferenceDistribution holds the centroid imprint for a single domain.
type ReferenceDistribution struct {
	Domain      string
	Centroid    GrammarImprint
	SampleCount int
	// Per-key variance for Mahalanobis distance (flattened across all map fields).
	Variance map[string]float64
}

// ReferenceSet holds per-domain reference distributions for classification.
type ReferenceSet struct {
	Domains map[string]*ReferenceDistribution
}

// DistanceMetrics holds multiple distance measures between an imprint and a reference.
type DistanceMetrics struct {
	CosineSimilarity float64 // 0.0–1.0 (1.0 = identical)
	KLDivergence     float64 // 0.0+ (0.0 = identical)
	Mahalanobis      float64 // 0.0+ (0.0 = identical)
}

// ClassifyResult holds the domain classification from imprint comparison.
type ImprintClassification struct {
	Domain     string  // best-matching domain
	Confidence float64 // distance margin between best and second-best (0.0–1.0)
	Distances  map[string]DistanceMetrics
}

// BuildReferences computes per-domain reference distributions from classified samples.
// Each sample is tokenised and its imprint computed, then aggregated into a centroid
// per unique domain label.
func BuildReferences(tokeniser *Tokeniser, samples []ClassifiedText) (*ReferenceSet, error) {
	if len(samples) == 0 {
		return nil, fmt.Errorf("empty sample set")
	}

	// Group imprints by domain.
	grouped := make(map[string][]GrammarImprint)
	for _, s := range samples {
		if s.Domain == "" {
			continue
		}
		tokens := tokeniser.Tokenise(s.Text)
		imp := NewImprint(tokens)
		grouped[s.Domain] = append(grouped[s.Domain], imp)
	}

	if len(grouped) == 0 {
		return nil, fmt.Errorf("no samples with domain labels")
	}

	rs := &ReferenceSet{Domains: make(map[string]*ReferenceDistribution)}
	for domain, imprints := range grouped {
		centroid := computeCentroid(imprints)
		variance := computeVariance(imprints, centroid)
		rs.Domains[domain] = &ReferenceDistribution{
			Domain:      domain,
			Centroid:    centroid,
			SampleCount: len(imprints),
			Variance:    variance,
		}
	}

	return rs, nil
}

// Compare computes distance metrics between an imprint and all domain references.
func (rs *ReferenceSet) Compare(imprint GrammarImprint) map[string]DistanceMetrics {
	result := make(map[string]DistanceMetrics, len(rs.Domains))
	for domain, ref := range rs.Domains {
		result[domain] = DistanceMetrics{
			CosineSimilarity: imprint.Similar(ref.Centroid),
			KLDivergence:     klDivergence(imprint, ref.Centroid),
			Mahalanobis:      mahalanobis(imprint, ref.Centroid, ref.Variance),
		}
	}
	return result
}

// Classify returns the best-matching domain for an imprint based on cosine similarity.
// Confidence is the margin between the best and second-best similarity scores.
func (rs *ReferenceSet) Classify(imprint GrammarImprint) ImprintClassification {
	distances := rs.Compare(imprint)

	// Rank by cosine similarity (descending).
	type scored struct {
		domain string
		sim    float64
	}
	var ranked []scored
	for d, m := range distances {
		ranked = append(ranked, scored{d, m.CosineSimilarity})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].sim > ranked[j].sim })

	result := ImprintClassification{Distances: distances}
	if len(ranked) > 0 {
		result.Domain = ranked[0].domain
		if len(ranked) > 1 {
			result.Confidence = ranked[0].sim - ranked[1].sim
		} else {
			result.Confidence = ranked[0].sim
		}
	}
	return result
}

// DomainNames returns sorted domain names in the reference set.
func (rs *ReferenceSet) DomainNames() []string {
	names := make([]string, 0, len(rs.Domains))
	for d := range rs.Domains {
		names = append(names, d)
	}
	sort.Strings(names)
	return names
}

// computeCentroid averages imprints into a single centroid.
func computeCentroid(imprints []GrammarImprint) GrammarImprint {
	n := float64(len(imprints))
	if n == 0 {
		return GrammarImprint{}
	}

	centroid := GrammarImprint{
		VerbDistribution:   make(map[string]float64),
		TenseDistribution:  make(map[string]float64),
		NounDistribution:   make(map[string]float64),
		DomainVocabulary:   make(map[string]int),
		ArticleUsage:       make(map[string]float64),
		PunctuationPattern: make(map[string]float64),
	}

	for _, imp := range imprints {
		addMap(centroid.VerbDistribution, imp.VerbDistribution)
		addMap(centroid.TenseDistribution, imp.TenseDistribution)
		addMap(centroid.NounDistribution, imp.NounDistribution)
		addMap(centroid.ArticleUsage, imp.ArticleUsage)
		addMap(centroid.PunctuationPattern, imp.PunctuationPattern)
		for k, v := range imp.DomainVocabulary {
			centroid.DomainVocabulary[k] += v
		}
		centroid.PluralRatio += imp.PluralRatio
		centroid.TokenCount += imp.TokenCount
		centroid.UniqueVerbs += imp.UniqueVerbs
		centroid.UniqueNouns += imp.UniqueNouns
	}

	// Average scalar fields.
	centroid.PluralRatio /= n
	centroid.TokenCount = int(math.Round(float64(centroid.TokenCount) / n))
	centroid.UniqueVerbs = int(math.Round(float64(centroid.UniqueVerbs) / n))
	centroid.UniqueNouns = int(math.Round(float64(centroid.UniqueNouns) / n))

	// Normalise map fields (sums to 1.0 after accumulation).
	normaliseMap(centroid.VerbDistribution)
	normaliseMap(centroid.TenseDistribution)
	normaliseMap(centroid.NounDistribution)
	normaliseMap(centroid.ArticleUsage)
	normaliseMap(centroid.PunctuationPattern)

	return centroid
}

// computeVariance computes per-key variance across imprints relative to a centroid.
// Keys are prefixed: "verb:", "tense:", "noun:", "article:", "punct:".
func computeVariance(imprints []GrammarImprint, centroid GrammarImprint) map[string]float64 {
	n := float64(len(imprints))
	if n < 2 {
		return nil
	}

	variance := make(map[string]float64)

	for _, imp := range imprints {
		accumVariance(variance, "verb:", imp.VerbDistribution, centroid.VerbDistribution)
		accumVariance(variance, "tense:", imp.TenseDistribution, centroid.TenseDistribution)
		accumVariance(variance, "noun:", imp.NounDistribution, centroid.NounDistribution)
		accumVariance(variance, "article:", imp.ArticleUsage, centroid.ArticleUsage)
		accumVariance(variance, "punct:", imp.PunctuationPattern, centroid.PunctuationPattern)
	}

	for k := range variance {
		variance[k] /= (n - 1) // sample variance
	}
	return variance
}

// accumVariance adds squared deviation for each key.
func accumVariance(variance map[string]float64, prefix string, sample, centroid map[string]float64) {
	// All keys that appear in either sample or centroid.
	keys := make(map[string]bool)
	for k := range sample {
		keys[k] = true
	}
	for k := range centroid {
		keys[k] = true
	}
	for k := range keys {
		diff := sample[k] - centroid[k]
		variance[prefix+k] += diff * diff
	}
}

// addMap accumulates values from src into dst.
func addMap(dst, src map[string]float64) {
	for k, v := range src {
		dst[k] += v
	}
}

// klDivergence computes symmetric KL divergence between two imprints.
// Uses the averaged distributions (Jensen-Shannon style) for stability.
const klEpsilon = 1e-10

func klDivergence(a, b GrammarImprint) float64 {
	var total float64
	total += mapKL(a.VerbDistribution, b.VerbDistribution) * 0.30
	total += mapKL(a.TenseDistribution, b.TenseDistribution) * 0.20
	total += mapKL(a.NounDistribution, b.NounDistribution) * 0.25
	total += mapKL(a.ArticleUsage, b.ArticleUsage) * 0.15
	total += mapKL(a.PunctuationPattern, b.PunctuationPattern) * 0.10
	return total
}

// mapKL computes symmetric KL divergence between two frequency maps.
// Returns 0.0 if both are empty.
func mapKL(p, q map[string]float64) float64 {
	if len(p) == 0 && len(q) == 0 {
		return 0.0
	}

	// Collect union of keys.
	keys := make(map[string]bool)
	for k := range p {
		keys[k] = true
	}
	for k := range q {
		keys[k] = true
	}

	// Symmetric KL: (KL(P||Q) + KL(Q||P)) / 2
	var klPQ, klQP float64
	for k := range keys {
		pv := p[k] + klEpsilon
		qv := q[k] + klEpsilon
		klPQ += pv * math.Log(pv/qv)
		klQP += qv * math.Log(qv/pv)
	}
	return (klPQ + klQP) / 2.0
}

// mahalanobis computes a simplified Mahalanobis-like distance using per-key variance.
// Falls back to Euclidean distance when variance is unavailable.
func mahalanobis(a, b GrammarImprint, variance map[string]float64) float64 {
	var sumSq float64

	sumSq += mapMahalanobis("verb:", a.VerbDistribution, b.VerbDistribution, variance) * 0.30
	sumSq += mapMahalanobis("tense:", a.TenseDistribution, b.TenseDistribution, variance) * 0.20
	sumSq += mapMahalanobis("noun:", a.NounDistribution, b.NounDistribution, variance) * 0.25
	sumSq += mapMahalanobis("article:", a.ArticleUsage, b.ArticleUsage, variance) * 0.15
	sumSq += mapMahalanobis("punct:", a.PunctuationPattern, b.PunctuationPattern, variance) * 0.10

	return math.Sqrt(sumSq)
}

// mapMahalanobis computes variance-normalised squared distance between two maps.
func mapMahalanobis(prefix string, a, b map[string]float64, variance map[string]float64) float64 {
	keys := make(map[string]bool)
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}

	var sumSq float64
	for k := range keys {
		diff := a[k] - b[k]
		v := 1.0 // default: unit variance (Euclidean)
		if variance != nil {
			if vk, ok := variance[prefix+k]; ok && vk > klEpsilon {
				v = vk
			}
		}
		sumSq += (diff * diff) / v
	}
	return sumSq
}
