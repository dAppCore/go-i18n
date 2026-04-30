package reversal

func ExampleWithSignals() {
	_ = WithSignals
}

func ExampleWithWeights() {
	_ = WithWeights
}

func ExampleNewTokeniser() {
	_ = NewTokeniser
}

func ExampleNewTokeniserForLang() {
	_ = NewTokeniserForLang
}

func ExampleTokeniser_MatchNoun() {
	v := &Tokeniser{}
	_ = v.MatchNoun
}

func ExampleTokeniser_MatchVerb() {
	v := &Tokeniser{}
	_ = v.MatchVerb
}

func ExampleTokeniser_IsDualClass() {
	v := &Tokeniser{}
	_ = v.IsDualClass
}

func ExampleDefaultWeights() {
	_ = DefaultWeights
}

func ExampleTokeniser_SignalWeights() {
	v := &Tokeniser{}
	_ = v.SignalWeights
}

func ExampleTokeniser_MatchWord() {
	v := &Tokeniser{}
	_ = v.MatchWord
}

func ExampleTokeniser_MatchArticle() {
	v := &Tokeniser{}
	_ = v.MatchArticle
}

func ExampleTokeniser_Tokenise() {
	v := &Tokeniser{}
	_ = v.Tokenise
}

func ExampleDisambiguationStatsFromTokens() {
	_ = DisambiguationStatsFromTokens
}

func ExampleTokeniser_DisambiguationStats() {
	v := &Tokeniser{}
	_ = v.DisambiguationStats
}
