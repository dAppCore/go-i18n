package reversal

// AnomalyResult flags a potential domain mismatch between model classification
// and imprint-based classification.
type AnomalyResult struct {
	Text          string  `json:"text"`
	ModelDomain   string  `json:"model_domain"`   // domain from 1B model
	ImprintDomain string  `json:"imprint_domain"` // domain from imprint comparison
	Confidence    float64 `json:"confidence"`      // imprint classification margin
	IsAnomaly     bool    `json:"is_anomaly"`      // true when domains disagree
}

// AnomalyStats holds aggregate anomaly detection metrics.
type AnomalyStats struct {
	Total     int            `json:"total"`
	Anomalies int            `json:"anomalies"`
	Rate      float64        `json:"rate"` // anomalies / total
	ByPair    map[string]int `json:"by_pair"` // "model->imprint": count
}

// DetectAnomalies compares 1B model domain tags against imprint-based classification.
// Returns per-sample results and aggregate stats.
// Samples with empty Domain are skipped.
func (rs *ReferenceSet) DetectAnomalies(tokeniser *Tokeniser, samples []ClassifiedText) ([]AnomalyResult, *AnomalyStats) {
	stats := &AnomalyStats{ByPair: make(map[string]int)}
	var results []AnomalyResult

	for _, s := range samples {
		if s.Domain == "" {
			continue
		}

		tokens := tokeniser.Tokenise(s.Text)
		imp := NewImprint(tokens)
		cls := rs.Classify(imp)

		isAnomaly := s.Domain != cls.Domain
		result := AnomalyResult{
			Text:          s.Text,
			ModelDomain:   s.Domain,
			ImprintDomain: cls.Domain,
			Confidence:    cls.Confidence,
			IsAnomaly:     isAnomaly,
		}
		results = append(results, result)
		stats.Total++

		if isAnomaly {
			stats.Anomalies++
			key := s.Domain + "->" + cls.Domain
			stats.ByPair[key]++
		}
	}

	if stats.Total > 0 {
		stats.Rate = float64(stats.Anomalies) / float64(stats.Total)
	}

	return results, stats
}
