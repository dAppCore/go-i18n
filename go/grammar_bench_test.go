package i18n

import "testing"

var (
	benchGrammarAnySink    any
	benchGrammarBoolSink   bool
	benchGrammarDataSink   *GrammarData
	benchGrammarStringSink string
	benchGrammarSliceSink  []string
)

func benchGrammarFixtureData() *GrammarData {
	return &GrammarData{
		Verbs: map[string]VerbForms{
			"delete": {Past: "removed", Gerund: "removing"},
			"build":  {Past: "built", Gerund: "building"},
		},
		Nouns: map[string]NounForms{
			"archive": {One: "archive", Other: "archives", Gender: "f"},
			"file":    {One: "file", Other: "files", Gender: "m"},
		},
		Articles: ArticleForms{
			IndefiniteDefault: "un",
			IndefiniteVowel:   "une",
			Definite:          "le",
			ByGender: map[string]string{
				"f": "la",
				"m": "le",
			},
		},
		Words: map[string]string{
			"failed_to": "Unable to",
			"file":      "fichier",
			"status":    "statut",
		},
		Punct: PunctuationRules{
			LabelSuffix:    " :",
			ProgressSuffix: "...",
		},
		Signals: SignalData{
			NounDeterminers: []string{"the", "this"},
			VerbAuxiliaries: []string{"will", "should"},
			VerbInfinitive:  []string{"to"},
			VerbNegation:    []string{"not", "never"},
			Priors: map[string]map[string]float64{
				"archive": {"noun": 0.7, "verb": 0.3},
			},
		},
		Intents: map[string]Intent{
			"delete": {
				Meta:    IntentMeta{Type: "action", Verb: "delete", Dangerous: true, Default: "no", Supports: []string{"force", "dry-run"}},
				Confirm: "Delete {{.Subject}}?",
				Success: "{{.Subject}} removed",
				Failure: "Unable to delete {{.Subject}}",
			},
		},
		Number: NumberFormat{
			ThousandsSep: " ",
			DecimalSep:   ",",
			PercentFmt:   "%s %%",
		},
	}
}

func benchInstallGrammarData(b *testing.B, lang string) *GrammarData {
	b.Helper()
	lang = normalizeLanguageTag(lang)
	data := benchGrammarFixtureData()

	grammarCacheMu.Lock()
	previous, existed := grammarCache[lang]
	grammarCacheMu.Unlock()

	SetGrammarData(lang, data)
	b.Cleanup(func() {
		grammarCacheMu.Lock()
		if existed {
			grammarCache[lang] = previous
		} else {
			delete(grammarCache, lang)
		}
		grammarCacheMu.Unlock()
	})
	return data
}

func benchSetDefaultLanguage(b *testing.B, lang string) {
	b.Helper()
	previous := Default()
	svc, _ := serviceFromResult(New())
	svc.mu.Lock()
	svc.currentLang = normalizeLanguageTag(lang)
	svc.languageExplicit = true
	svc.mu.Unlock()
	SetDefault(svc)
	b.Cleanup(func() {
		SetDefault(previous)
	})
}

func BenchmarkGetGrammarData(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarDataSink = GetGrammarData("fr-bench")
	}
}

func BenchmarkSetGrammarData(b *testing.B) {
	lang := normalizeLanguageTag("zz-bench-set")
	data := benchGrammarFixtureData()
	grammarCacheMu.Lock()
	previous, existed := grammarCache[lang]
	grammarCacheMu.Unlock()
	b.Cleanup(func() {
		grammarCacheMu.Lock()
		if existed {
			grammarCache[lang] = previous
		} else {
			delete(grammarCache, lang)
		}
		grammarCacheMu.Unlock()
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetGrammarData(lang, data)
	}
	benchGrammarDataSink = GetGrammarData(lang)
}

func BenchmarkMergeGrammarData(b *testing.B) {
	lang := "zz-bench-merge"
	benchInstallGrammarData(b, lang)
	data := benchGrammarFixtureData()
	data.Words["branch"] = "branche"
	data.Nouns["branch"] = NounForms{One: "branch", Other: "branches", Gender: "f"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeGrammarData(lang, data)
	}
	benchGrammarDataSink = GetGrammarData(lang)
}

func BenchmarkMergeArticleForms(b *testing.B) {
	dst := ArticleForms{IndefiniteDefault: "a", ByGender: map[string]string{"n": "an"}}
	src := benchGrammarFixtureData().Articles
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeArticleForms(&dst, src)
	}
	benchGrammarStringSink = dst.ByGender["f"]
}

func BenchmarkMergePunctuationRules(b *testing.B) {
	dst := PunctuationRules{LabelSuffix: ":"}
	src := PunctuationRules{LabelSuffix: " :", ProgressSuffix: "..."}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergePunctuationRules(&dst, src)
	}
	benchGrammarStringSink = dst.ProgressSuffix
}

func BenchmarkMergeSignalData(b *testing.B) {
	dst := SignalData{NounDeterminers: []string{"the"}, Priors: map[string]map[string]float64{"file": {"noun": 1}}}
	src := benchGrammarFixtureData().Signals
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeSignalData(&dst, src)
	}
	benchGrammarAnySink = dst.Priors
}

func BenchmarkAppendUniqueStrings(b *testing.B) {
	dst := []string{"the", "this"}
	values := []string{"the", "a", "each", ""}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = appendUniqueStrings(dst, values...)
	}
	benchGrammarSliceSink = dst
}

func BenchmarkGrammarDataHasContent(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = grammarDataHasContent(data)
	}
}

func BenchmarkCloneGrammarData(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarDataSink = cloneGrammarData(data)
	}
}

func BenchmarkMergeIntentData(b *testing.B) {
	dst := map[string]Intent{"build": {Success: "Built {{.Subject}}"}}
	src := benchGrammarFixtureData().Intents
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = mergeIntentData(dst, src)
	}
	benchGrammarAnySink = dst
}

func BenchmarkCloneIntentMap(b *testing.B) {
	intents := benchGrammarFixtureData().Intents
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = cloneIntentMap(intents)
	}
}

func BenchmarkCloneIntent(b *testing.B) {
	intent := benchGrammarFixtureData().Intents["delete"]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = cloneIntent(intent)
	}
}

func BenchmarkIrregularVerbs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = IrregularVerbs()
	}
}

func BenchmarkIrregularNouns(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = IrregularNouns()
	}
}

func BenchmarkDualClassVerbs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = DualClassVerbs()
	}
}

func BenchmarkDualClassNouns(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = DualClassNouns()
	}
}

func BenchmarkLower(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Lower("Repository")
	}
}

func BenchmarkUpper(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Upper("repository")
	}
}

func BenchmarkGetVerbForm(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = getVerbForm("fr-bench", "delete", "past")
	}
}

func BenchmarkGetWord(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = getWord("fr-bench", "file")
	}
}

func BenchmarkGetPunct(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = getPunct("fr-bench", "label", ":")
	}
}

func BenchmarkGetNounForm(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = getNounForm("fr-bench", "file", "other")
	}
}

func BenchmarkCurrentLangForGrammar(b *testing.B) {
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = currentLangForGrammar()
	}
}

func BenchmarkPastTenseForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = pastTenseForLanguages(langs, "delete")
	}
}

func BenchmarkApplyRegularPastTense(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = applyRegularPastTense("commit")
	}
}

func BenchmarkShouldDoubleConsonant(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = shouldDoubleConsonant("commit")
	}
}

func BenchmarkApplyRegularGerund(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = applyRegularGerund("commit")
	}
}

func BenchmarkPluralForm(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = PluralForm("file")
	}
}

func BenchmarkApplyRegularPlural(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = applyRegularPlural("repository")
	}
}

func BenchmarkArticleToken(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = ArticleToken("error")
	}
}

func BenchmarkArticleForCurrentLanguage(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := articleForCurrentLanguage("archive", "archive")
		benchGrammarStringSink = article
	}
}

func BenchmarkArticleByGender(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := articleByGender(data, "archive", "archive", "fr")
		benchGrammarStringSink = article
	}
}

func BenchmarkArticleForPluralForm(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := articleForPluralForm(data, "files", "fr")
		benchGrammarStringSink = article
	}
}

func BenchmarkArticleForFrenchPluralGuess(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := articleForFrenchPluralGuess(data, "journaux", "journaux", "fr")
		benchGrammarStringSink = article
	}
}

func BenchmarkIsKnownPluralNoun(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = isKnownPluralNoun(data, "files")
	}
}

func BenchmarkArticleFromGrammarForms(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := articleFromGrammarForms(data, "error")
		benchGrammarStringSink = article
	}
}

func BenchmarkMaybeElideArticle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = maybeElideArticle("la", "archive", "fr")
	}
}

func BenchmarkUsesVowelSoundArticle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = usesVowelSoundArticle(nil, "MRI")
	}
}

func BenchmarkLooksLikeFrenchPlural(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = looksLikeFrenchPlural("journaux")
	}
}

func BenchmarkStartsWithVowelSound(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = startsWithVowelSound("archive")
	}
}

func BenchmarkIsFrenchAspiratedHWord(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = isFrenchAspiratedHWord("haricot")
	}
}

func BenchmarkIsInitialism(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = isInitialism("API")
	}
}

func BenchmarkPreserveInitialCapitalization(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = preserveInitialCapitalization("File", "fichier")
	}
}

func BenchmarkInitialismUsesVowelSound(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = initialismUsesVowelSound("MRI")
	}
}

func BenchmarkIsVowel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarBoolSink = isVowel('e')
	}
}

func BenchmarkTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Title("build go.mod-cache")
	}
}

func BenchmarkRenderWord(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = renderWord("fr-bench", "file")
	}
}

func BenchmarkRenderWordOrTitle(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = renderWordOrTitle("fr-bench", "status")
	}
}

func BenchmarkQuote(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Quote(`go "mod"`)
	}
}

func BenchmarkArticlePhrase(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = ArticlePhrase("file")
	}
}

func BenchmarkDefiniteArticle(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = DefiniteArticle("archive")
	}
}

func BenchmarkDefiniteToken(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = DefiniteToken("archive")
	}
}

func BenchmarkDefinitePhrase(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = DefinitePhrase("archive")
	}
}

func BenchmarkDefiniteArticleForCurrentLanguage(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := definiteArticleForCurrentLanguage("archive", "archive")
		benchGrammarStringSink = article
	}
}

func BenchmarkGrammarDataForLang(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarDataSink = grammarDataForLang("fr-bench")
	}
}

func BenchmarkLanguageFallbackOrder(b *testing.B) {
	langs := []string{"fr-CA", "en-GB", "fr"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarSliceSink = languageFallbackOrder(langs)
	}
}

func BenchmarkBaseLanguageTag(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = baseLanguageTag("fr-CA")
	}
}

func BenchmarkDefiniteArticleFromGrammarForms(b *testing.B) {
	data := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		article, _ := definiteArticleFromGrammarForms(data, "files", "files", "fr")
		benchGrammarStringSink = article
	}
}

func BenchmarkTemplateFuncs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarAnySink = TemplateFuncs()
	}
}

func BenchmarkNumber(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Number(1234567)
	}
}

func BenchmarkDecimal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Decimal(1234.567)
	}
}

func BenchmarkPercent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Percent(0.42)
	}
}

func BenchmarkBytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Bytes(1048576)
	}
}

func BenchmarkOrdinal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Ordinal(42)
	}
}

func BenchmarkAgo(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Ago(3, "hour")
	}
}

func BenchmarkPrefixWithArticle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = prefixWithArticle("l'", "archive")
	}
}

func BenchmarkProgressSubject(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = ProgressSubject("build", "file")
	}
}

func BenchmarkActionResultForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = actionResultForLanguages(langs, "delete", "file")
	}
}

func BenchmarkActionFailed(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = ActionFailed("delete", "file")
	}
}

func BenchmarkActionFailedForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = actionFailedForLanguages(langs, "delete", "file")
	}
}

func BenchmarkFailedPrefix(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = failedPrefix("fr-bench")
	}
}

func BenchmarkFailedPrefixForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = failedPrefixForLanguages(langs)
	}
}

func BenchmarkRenderWordForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = renderWordForLanguages(langs, "file")
	}
}

func BenchmarkRenderWordOrTitleForLanguages(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	langs := []string{"fr-bench", "en"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = renderWordOrTitleForLanguages(langs, "status")
	}
}

func BenchmarkLabel(b *testing.B) {
	benchInstallGrammarData(b, "fr-bench")
	benchSetDefaultLanguage(b, "fr-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchGrammarStringSink = Label("status")
	}
}
