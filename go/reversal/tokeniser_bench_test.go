package reversal

import "testing"

func BenchmarkGetTokeniseScratch(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := getTokeniseScratch()
		putTokeniseScratch(s)
	}
}

func BenchmarkPutTokeniseScratch(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := getTokeniseScratch()
		s.fields = append(s.fields, "delete", "files")
		s.lowerWords = append(s.lowerWords, "delete", "files")
		s.phraseBuf = append(s.phraseBuf, "delete"...)
		s.rawBuf = append(s.rawBuf, "Delete"...)
		putTokeniseScratch(s)
	}
}

func BenchmarkWithWeights(b *testing.B) {
	benchSetup(b)
	weights := map[string]float64{"noun_determiner": 0.50, "default_prior": 0.05}
	opt := WithWeights(weights)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{}
		opt(tok)
	}
}

func BenchmarkNewTokeniser(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTokeniser()
	}
}

func BenchmarkNewTokeniserForLang(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTokeniserForLang("fr")
	}
}

func BenchmarkBuildVerbIndex(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{
			pastToBase:   make(map[string]string),
			gerundToBase: make(map[string]string),
			baseVerbs:    make(map[string]bool),
			lang:         "en",
		}
		tok.buildVerbIndex()
	}
}

func BenchmarkBuildNounIndex(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{
			pluralToBase: make(map[string]string),
			baseNouns:    make(map[string]bool),
			lang:         "en",
		}
		tok.buildNounIndex()
	}
}

func BenchmarkMatchNounLowered(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.matchNounLowered("repositories")
	}
}

func BenchmarkReverseRegularPlural(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.reverseRegularPlural("repositories")
	}
}

func BenchmarkMatchVerbLowered(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.matchVerbLowered("committed")
	}
}

func BenchmarkBestRoundTrip(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	candidates := []string{"commit", "commite"}
	forward := func(s string) string {
		if s == "commit" || s == "commite" {
			return "committed"
		}
		return ""
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.bestRoundTrip("committed", candidates, forward)
	}
}

func BenchmarkHasVCeEnding(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasVCeEnding("delete")
	}
}

func BenchmarkIsVowelByte(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isVowelByte('e')
	}
}

func BenchmarkReverseRegularPast(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.reverseRegularPast("committed")
	}
}

func BenchmarkReverseRegularGerund(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.reverseRegularGerund("committing")
	}
}

func BenchmarkBuildWordIndex(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{words: make(map[string]string), lang: "en"}
		tok.buildWordIndex()
	}
}

func BenchmarkIsDualClass(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.IsDualClass("commit")
	}
}

func BenchmarkBuildDualClassIndex(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{
			baseVerbs: map[string]bool{"commit": true, "delete": true},
			baseNouns: map[string]bool{"commit": true, "file": true},
		}
		tok.buildDualClassIndex()
	}
}

func BenchmarkBuildSignalIndex(b *testing.B) {
	benchSetup(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := &Tokeniser{lang: "en"}
		tok.buildSignalIndex()
	}
}

func BenchmarkDefaultNounDeterminers(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defaultNounDeterminers()
	}
}

func BenchmarkDefaultVerbAuxiliaries(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defaultVerbAuxiliaries()
	}
}

func BenchmarkDefaultVerbNegations(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defaultVerbNegations()
	}
}

func BenchmarkDefaultWeights(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultWeights()
	}
}

func BenchmarkSignalWeights(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.SignalWeights()
	}
}

func BenchmarkSkipDeprecatedEnglishGrammarEntry(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = skipDeprecatedEnglishGrammarEntry("passed")
	}
}

func BenchmarkMatchWord(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.MatchWord("up to date")
	}
}

func BenchmarkMatchWordLowered(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.matchWordLowered("up to date")
	}
}

func BenchmarkMatchArticleLowered(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.matchArticleLowered("de l'enfant")
	}
}

func BenchmarkMatchConfiguredArticleText(b *testing.B) {
	benchSetup(b)
	data := NewTokeniser().grammarData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchConfiguredArticleText("the file", data)
	}
}

func BenchmarkMatchConfiguredArticleCandidate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchConfiguredArticleCandidate("l'enfant", "l'", "definite")
	}
}

func BenchmarkMatchFrenchLeadingArticlePhrase(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchFrenchLeadingArticlePhrase("de l'enfant")
	}
}

func BenchmarkMatchFrenchArticleText(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchFrenchArticleText("de l'enfant")
	}
}

func BenchmarkMatchFrenchAttachedArticle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchFrenchAttachedArticle("l'enfant")
	}
}

func BenchmarkMatchWordPhrase(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	parts := splitFields("up to date.")
	scratch := &tokeniseScratch{lowerWords: []string{"up", "to", "date"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tok.matchWordPhrase(parts, scratch, 0)
	}
}

func BenchmarkMatchFrenchArticlePhrase(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr")
	parts := splitFields("de l'enfant.")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = tok.matchFrenchArticlePhrase(parts, 0)
	}
}

func BenchmarkClassifyElidedFrenchWord(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.classifyElidedFrenchWord("enfant")
	}
}

func BenchmarkResolveAmbiguous(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := []Token{{
		Raw: "commit", Lower: "commit", Type: tokenAmbiguous,
		VerbInfo: VerbMatch{Base: "commit", Tense: "base", Form: "commit"},
		NounInfo: NounMatch{Base: "commit", Form: "commit"},
	}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokens[0].Type = tokenAmbiguous
		tok.resolveAmbiguous(tokens)
	}
}

func BenchmarkScoreAmbiguous(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser(WithSignals())
	tokens := []Token{
		{Raw: "the", Lower: "the", Type: TokenArticle, Confidence: 1.0},
		{
			Raw: "commit", Lower: "commit", Type: tokenAmbiguous,
			VerbInfo: VerbMatch{Base: "commit", Tense: "base", Form: "commit"},
			NounInfo: NounMatch{Base: "commit", Form: "commit"},
		},
		{Raw: "build", Lower: "build", Type: TokenVerb, Confidence: 1.0},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tok.scoreAmbiguous(tokens, 1)
	}
}

func BenchmarkHasNoLongerBefore(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := []Token{{Lower: "no"}, {Lower: "longer"}, {Lower: "commit"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.hasNoLongerBefore(tokens, 2)
	}
}

func BenchmarkCorpusPrior(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tok.corpusPrior("commit")
	}
}

func BenchmarkValidSignalPriorScore(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validSignalPriorScore(0.5)
	}
}

func BenchmarkHasConfidentVerbInClause(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := []Token{
		{Lower: "delete", Type: TokenVerb, Confidence: 1.0},
		{Lower: "commit", Type: tokenAmbiguous},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.hasConfidentVerbInClause(tokens, 1)
	}
}

func BenchmarkCheckInflectionEcho(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := []Token{
		{Lower: "committed", Type: TokenVerb, VerbInfo: VerbMatch{Base: "commit", Tense: "past"}},
		{Lower: "commit", Type: tokenAmbiguous, VerbInfo: VerbMatch{Base: "commit"}, NounInfo: NounMatch{Base: "commit"}},
		{Lower: "commits", Type: TokenNoun, NounInfo: NounMatch{Base: "commit", Plural: true}},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tok.checkInflectionEcho(tokens, 1)
	}
}

func BenchmarkResolveToken(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser(WithSignals())
	components := []SignalComponent{{Name: "default_prior", Weight: 1, Value: 1, Contrib: 1}}
	var token Token
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token = Token{}
		tok.resolveToken(&token, 0.7, 0.3, components)
	}
}

func BenchmarkClassifyAmbiguousToken(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = classifyAmbiguousToken(0.7, 0.3)
	}
}

func BenchmarkSplitTrailingPunct(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = splitTrailingPunct("configuration.")
	}
}

func BenchmarkSplitFrenchElision(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tok.splitFrenchElision("l'enfant")
	}
}

func BenchmarkSplitConfiguredElision(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = tok.splitConfiguredElision("l'enfant")
	}
}

func BenchmarkIsFrenchLanguage(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr-CA")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.isFrenchLanguage()
	}
}

func BenchmarkGrammarData(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniserForLang("fr-CA")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.grammarData()
	}
}

func BenchmarkTokeniserLanguageBase(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokeniserLanguageBase("fr_CA")
	}
}

func BenchmarkNormaliseFrenchApostrophes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normaliseFrenchApostrophes("l’enfant")
	}
}

func BenchmarkIsFrenchApostrophe(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isFrenchApostrophe('’')
	}
}

func BenchmarkMatchPunctuation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchPunctuation("...")
	}
}

func BenchmarkDisambiguationStatsFromTokens(b *testing.B) {
	tokens := []Token{
		{Type: TokenVerb, Confidence: 0.55, AltType: TokenNoun, AltConf: 0.45},
		{Type: TokenNoun, Confidence: 1.0},
		{Type: TokenUnknown},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DisambiguationStatsFromTokens(tokens)
	}
}

func BenchmarkDisambiguationStats(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	tokens := tok.Tokenise("maybe commit")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tok.DisambiguationStats(tokens)
	}
}

func BenchmarkSplitFields(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitFields("  delete\tconfiguration files  ")
	}
}

func BenchmarkSplitFieldsInto(b *testing.B) {
	dst := make([]string, 0, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = splitFieldsInto("  delete\tconfiguration files  ", dst)
	}
}

func BenchmarkIndexAny(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = indexAny("fr-CA", "-_")
	}
}
