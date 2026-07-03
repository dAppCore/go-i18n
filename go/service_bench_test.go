package i18n

import (
	"testing"
	"testing/fstest"

	core "dappco.re/go"
	"golang.org/x/text/language"
)

var (
	benchServiceAnySink       any
	benchServiceBoolSink      bool
	benchServiceComposedSink  Composed
	benchServiceDirectionSink TextDirection
	benchServiceFormalitySink Formality
	benchServiceHandlersSink  []KeyHandler
	benchServiceIntSink       int
	benchServiceIntentSink    Intent
	benchServiceMessageSink   Message
	benchServiceModeSink      Mode
	benchServicePluralSink    PluralCategory
	benchServiceResultSink    core.Result
	benchServiceRuneSink      rune
	benchServiceServiceSink   *Service
	benchServiceStringSink    string
	benchServiceStringsSink   []string
)

type benchServiceLoader struct {
	langs    []string
	messages map[string]map[string]Message
	grammar  map[string]*GrammarData
}

func (l benchServiceLoader) Languages() []string {
	return l.langs
}

func (l benchServiceLoader) Load(lang string) core.Result {
	lang = normalizeLanguageTag(lang)
	return localeLoadResult(l.messages[lang], l.grammar[lang])
}

func benchServiceMessages() map[string]Message {
	return map[string]Message{
		"app.hello": {
			Text: "Hello {{.Name}}",
		},
		"app.files": {
			One:   "{{.Count}} file",
			Other: "{{.Count}} files",
		},
		"app.context._dashboard._f._eu._formal._tier._gold": {
			Text: "Gold dashboard",
		},
		"common.action.delete": {
			Text: "Delete common",
		},
		"common.delete": {
			Text: "Delete fallback",
		},
		"prompt.delete": {
			Text: "Delete?",
		},
		"common.prompt.delete": {
			Text: "Common delete?",
		},
		"lang.en": {
			Text: "English",
		},
		"core.delete.question": {
			Text: "Delete {{.Subject}}?",
		},
		"core.delete.confirm": {
			Text: "Really delete {{.Subject}}?",
		},
		"core.delete.success": {
			Text: "{{.Subject}} deleted",
		},
		"core.delete.failure": {
			Text: "Failed to delete {{.Subject}}",
		},
	}
}

func benchServiceCatalog() map[string]map[string]Message {
	return map[string]map[string]Message{
		"zz-bench": benchServiceMessages(),
		"en": {
			"app.hello": {Text: "Hello {{.Name}}"},
		},
	}
}

func benchServiceFixture() *Service {
	return &Service{
		loader:           benchServiceLoaderFixture(nil),
		messages:         benchServiceCatalog(),
		currentLang:      "zz-bench",
		fallbackLang:     "en",
		languageExplicit: true,
		availableLangs: []language.Tag{
			language.Make("ar"),
			language.Make("en"),
			language.Make("fr"),
			language.Make("zz-bench"),
		},
		mode:            ModeNormal,
		formality:       FormalityFormal,
		location:        "eu",
		handlers:        DefaultHandlers(),
		loadedLocales:   make(map[int]struct{}),
		loadedProviders: make(map[int]struct{}),
	}
}

func benchServiceWithGrammar(b *testing.B) *Service {
	b.Helper()
	benchInstallGrammarData(b, "zz-bench")
	return benchServiceFixture()
}

func benchServiceLoaderFixture(grammar *GrammarData) benchServiceLoader {
	grammars := map[string]*GrammarData{}
	if grammar != nil {
		grammars["zz-bench"] = grammar
	}
	return benchServiceLoader{
		langs:    []string{"zz-bench"},
		messages: map[string]map[string]Message{"zz-bench": benchServiceMessages()},
		grammar:  grammars,
	}
}

func benchServiceFS() fstest.MapFS {
	return fstest.MapFS{
		"locales/en.json": {
			Data: []byte(`{"app":{"hello":"Hello {{.Name}}","files":{"one":"{{.Count}} file","other":"{{.Count}} files"}}}`),
		},
	}
}

func benchServiceSubject() *Subject {
	return &Subject{
		Noun:      "file",
		Value:     "config.yaml",
		count:     2,
		gender:    "f",
		location:  "eu",
		formality: FormalityFormal,
	}
}

func benchServiceContext() *TranslationContext {
	return &TranslationContext{
		Context:   "dashboard",
		Gender:    "f",
		Location:  "eu",
		Formality: FormalityFormal,
		count:     2,
		countSet:  true,
		Extra:     map[string]any{"tier": "gold"},
	}
}

func benchSetDefaultService(b *testing.B, svc *Service) {
	b.Helper()
	previous := defaultService.Load()
	SetDefault(svc)
	b.Cleanup(func() {
		SetDefault(previous)
	})
}

func BenchmarkWithFallback(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithFallback("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceStringSink = svc.fallbackLang
}

func BenchmarkWithLanguage(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithLanguage("fr")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceStringSink = svc.requestedLang
}

func BenchmarkWithFormality(b *testing.B) {
	svc := benchServiceFixture()
	ctx := C("greeting")
	opt := WithFormality(FormalityInformal)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
		benchContextSink = ctx.WithFormality(FormalityFormal)
	}
	benchServiceFormalitySink = svc.formality
}

func BenchmarkWithLocation(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithLocation("dashboard")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceStringSink = svc.location
}

func BenchmarkWithHandlers(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithHandlers(LabelHandler{}, ProgressHandler{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkWithDefaultHandlers(b *testing.B) {
	svc := benchServiceFixture()
	svc.handlers = nil
	opt := WithDefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkWithMode(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithMode(ModeCollect)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceModeSink = svc.mode
}

func BenchmarkWithDebug(b *testing.B) {
	svc := benchServiceFixture()
	opt := WithDebug(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(svc)
	}
	benchServiceBoolSink = svc.debug
}

func BenchmarkLocaleLoadResult(b *testing.B) {
	messages := benchServiceMessages()
	grammar := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = localeLoadResult(messages, grammar)
	}
}

func BenchmarkLocaleFromResult(b *testing.B) {
	result := localeLoadResult(benchServiceMessages(), benchGrammarFixtureData())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		messages, grammar, r := localeFromResult(result)
		benchServiceAnySink = messages
		benchGrammarDataSink = grammar
		benchServiceResultSink = r
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = New()
	}
}

func BenchmarkNewService(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = NewService()
	}
}

func BenchmarkNewWithFS(b *testing.B) {
	fsys := benchServiceFS()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = NewWithFS(fsys, "locales")
	}
}

func BenchmarkNewServiceWithFS(b *testing.B) {
	fsys := benchServiceFS()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = NewServiceWithFS(fsys, "locales")
	}
}

func BenchmarkNewWithLoader(b *testing.B) {
	loader := benchServiceLoaderFixture(nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = NewWithLoader(loader)
	}
}

func BenchmarkNewServiceWithLoader(b *testing.B) {
	loader := benchServiceLoaderFixture(nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = NewServiceWithLoader(loader)
	}
}

func BenchmarkInit(b *testing.B) {
	benchSetDefaultService(b, benchServiceFixture())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = Init()
	}
}

func BenchmarkDefault(b *testing.B) {
	benchSetDefaultService(b, benchServiceFixture())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceServiceSink = Default()
	}
}

func BenchmarkSetDefault(b *testing.B) {
	svc := benchServiceFixture()
	previous := defaultService.Load()
	b.Cleanup(func() {
		SetDefault(previous)
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetDefault(svc)
	}
}

func BenchmarkAddLoader(b *testing.B) {
	svc := benchServiceFixture()
	benchSetDefaultService(b, svc)
	loader := benchServiceLoaderFixture(nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddLoader(loader)
	}
	benchServiceServiceSink = svc
}

func BenchmarkLoadJSON(b *testing.B) {
	svc := benchServiceFixture()
	data := []byte(`{"bench":{"json":"JSON loaded"}}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = svc.loadJSON("zz-json", data)
	}
}

func BenchmarkSetLanguage(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = svc.SetLanguage("fr")
	}
}

func BenchmarkLanguage(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Language()
	}
}

func BenchmarkCurrentLanguage(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.CurrentLanguage()
	}
}

func BenchmarkCurrentLang(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.CurrentLang()
	}
}

func BenchmarkPrompt(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Prompt("delete")
	}
}

func BenchmarkCurrentPrompt(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.CurrentPrompt("delete")
	}
}

func BenchmarkPromptLookupKeys(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = promptLookupKeys("delete")
	}
}

func BenchmarkLang(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Lang("en")
	}
}

func BenchmarkAvailableLanguages(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = svc.AvailableLanguages()
	}
}

func BenchmarkCurrentAvailableLanguages(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = svc.CurrentAvailableLanguages()
	}
}

func BenchmarkSetMode(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.SetMode(ModeCollect)
	}
	benchServiceModeSink = svc.mode
}

func BenchmarkMode(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceModeSink = svc.Mode()
	}
}

func BenchmarkCurrentMode(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceModeSink = svc.CurrentMode()
	}
}

func BenchmarkSetFormality(b *testing.B) {
	svc := benchServiceFixture()
	subject := S("user", "Alex")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.SetFormality(FormalityInformal)
		benchComposeSubjectSink = subject.SetFormality(FormalityFormal)
	}
	benchServiceFormalitySink = svc.formality
}

func BenchmarkFormality(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceFormalitySink = svc.Formality()
	}
}

func BenchmarkCurrentFormality(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceFormalitySink = svc.CurrentFormality()
	}
}

func BenchmarkSetFallback(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.SetFallback("fr")
	}
	benchServiceStringSink = svc.fallbackLang
}

func BenchmarkFallback(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Fallback()
	}
}

func BenchmarkCurrentFallback(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.CurrentFallback()
	}
}

func BenchmarkSetLocation(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.SetLocation("dashboard")
	}
	benchServiceStringSink = svc.location
}

func BenchmarkLocation(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Location()
	}
}

func BenchmarkCurrentLocation(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.CurrentLocation()
	}
}

func BenchmarkDirection(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceDirectionSink = svc.Direction()
	}
}

func BenchmarkCurrentDirection(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceDirectionSink = svc.CurrentDirection()
	}
}

func BenchmarkCurrentTextDirection(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceDirectionSink = svc.CurrentTextDirection()
	}
}

func BenchmarkIsRTL(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.IsRTL()
	}
}

func BenchmarkCurrentIsRTL(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.CurrentIsRTL()
	}
}

func BenchmarkRTL(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.RTL()
	}
}

func BenchmarkCurrentRTL(b *testing.B) {
	svc := benchServiceFixture()
	svc.currentLang = "ar"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.CurrentRTL()
	}
}

func BenchmarkCurrentDebug(b *testing.B) {
	svc := benchServiceFixture()
	svc.debug = true
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.CurrentDebug()
	}
}

func BenchmarkPluralCategory(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServicePluralSink = svc.PluralCategory(2)
	}
}

func BenchmarkCurrentPluralCategory(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServicePluralSink = svc.CurrentPluralCategory(2)
	}
}

func BenchmarkPluralCategoryOf(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServicePluralSink = svc.PluralCategoryOf(2)
	}
}

func BenchmarkJoinAvailableLanguagesLocked(b *testing.B) {
	tags := []language.Tag{language.Make("fr"), language.Make("en"), language.Make("ar")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = joinAvailableLanguagesLocked(tags)
	}
}

func BenchmarkAddHandler(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.handlers = svc.handlers[:0]
		svc.AddHandler(LabelHandler{})
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkSetHandlers(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.SetHandlers(LabelHandler{}, ProgressHandler{})
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkPrependHandler(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.handlers = svc.handlers[:0]
		svc.PrependHandler(LabelHandler{})
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkClearHandlers(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ClearHandlers()
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkResetHandlers(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ResetHandlers()
	}
	benchServiceHandlersSink = svc.handlers
}

func BenchmarkHandlers(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceHandlersSink = svc.Handlers()
	}
}

func BenchmarkCurrentHandlers(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceHandlersSink = svc.CurrentHandlers()
	}
}

func BenchmarkT(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.T("app.hello", data)
	}
}

func BenchmarkCompose(b *testing.B) {
	svc := benchServiceWithGrammar(b)
	subject := benchServiceSubject()
	intent := benchComposeIntent()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceComposedSink = svc.Compose("delete", subject)
		benchServiceComposedSink = intent.Compose(subject)
	}
}

func BenchmarkCurrentCompose(b *testing.B) {
	svc := benchServiceWithGrammar(b)
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceComposedSink = svc.CurrentCompose("delete", subject)
	}
}

func BenchmarkTranslate(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = svc.Translate("app.hello", data)
	}
}

func BenchmarkTranslateWithStatus(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text, ok := svc.translateWithStatus("app.hello", data)
		benchServiceStringSink = text
		benchServiceBoolSink = ok
	}
}

func BenchmarkResolveDirectLocked(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.resolveDirectLocked("app.hello", data)
	}
}

func BenchmarkResolveWithFallbackLocked(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.resolveWithFallbackLocked("app.delete", nil)
	}
}

func BenchmarkLookupIntentDefinitionLocked(b *testing.B) {
	svc := benchServiceWithGrammar(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intent, ok := svc.lookupIntentDefinitionLocked("delete")
		benchServiceIntentSink = intent
		benchServiceBoolSink = ok
	}
}

func BenchmarkLookupIntentInLanguage(b *testing.B) {
	benchInstallGrammarData(b, "zz-bench")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intent, ok := lookupIntentInLanguage("zz-bench", "delete")
		benchServiceIntentSink = intent
		benchServiceBoolSink = ok
	}
}

func BenchmarkResolveIntentQuestionLocked(b *testing.B) {
	svc := benchServiceWithGrammar(b)
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.resolveIntentQuestionLocked("delete", subject)
	}
}

func BenchmarkResolveIntentFieldLocked(b *testing.B) {
	svc := benchServiceWithGrammar(b)
	intent, _ := lookupIntentInLanguage("zz-bench", "delete")
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.resolveIntentFieldLocked("delete", "success", subject, intent)
	}
}

func BenchmarkGrammarLanguagesLocked(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = svc.grammarLanguagesLocked()
	}
}

func BenchmarkIntentTemplateForField(b *testing.B) {
	intent := benchGrammarFixtureData().Intents["delete"]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = intentTemplateForField(intent, "success")
	}
}

func BenchmarkRenderIntentTemplate(b *testing.B) {
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = renderIntentTemplate("Delete {{.Subject}}?", subject)
	}
}

func BenchmarkIntentVerbForKey(b *testing.B) {
	intent := Intent{Meta: IntentMeta{Verb: "core.action.delete"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = intentVerbForKey("fallback.remove", intent)
	}
}

func BenchmarkIntentVerbBase(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = intentVerbBase("core.action.delete")
	}
}

func BenchmarkIntentBaseFromKey(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = intentBaseFromKey("core.action.delete")
	}
}

func BenchmarkIntentSubjectText(b *testing.B) {
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = intentSubjectText(subject)
	}
}

func BenchmarkTryResolveLocked(b *testing.B) {
	svc := benchServiceFixture()
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.tryResolveLocked("zz-bench", "app.context", ctx)
	}
}

func BenchmarkResolveMessageLocked(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Count": 3}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.resolveMessageLocked("zz-bench", "app.files", data)
	}
}

func BenchmarkGetEffectiveContextGenderLocationAndFormality(b *testing.B) {
	svc := benchServiceFixture()
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		context, gender, location, formality := svc.getEffectiveContextGenderLocationAndFormality(ctx)
		benchServiceStringSink = location
		benchServiceBoolSink = context != "" && gender != ""
		benchServiceFormalitySink = formality
	}
}

func BenchmarkGetEffectiveContextExtra(b *testing.B) {
	svc := benchServiceFixture()
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = svc.getEffectiveContextExtra(ctx)
	}
}

func BenchmarkMergeContextExtra(b *testing.B) {
	dst := map[string]any{"existing": "value"}
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeContextExtra(dst, ctx)
	}
	benchServiceAnySink = dst
}

func BenchmarkGetEffectiveFormality(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Formality": "formal"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceFormalitySink = svc.getEffectiveFormality(data)
	}
}

func BenchmarkParseFormalityValue(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formality, ok := parseFormalityValue("formal")
		benchServiceFormalitySink = formality
		benchServiceBoolSink = ok
	}
}

func BenchmarkLookupVariants(b *testing.B) {
	extra := map[string]any{"tier": "gold"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringsSink = lookupVariants("app.context", "dashboard", "f", "eu", FormalityFormal, extra)
	}
}

func BenchmarkLookupVariantBaseCount(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceIntSink = lookupVariantBaseCount("dashboard", "f", "eu", "formal")
	}
}

func BenchmarkAppendLookupBaseVariants(b *testing.B) {
	dst := make([]string, 0, 31)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = dst[:0]
		benchServiceStringsSink = appendLookupBaseVariants(dst, "app.context", "dashboard", "f", "eu", "formal", "._tier._gold")
	}
}

func BenchmarkLookupExtraSuffix(b *testing.B) {
	extra := map[string]any{"tier": "gold", "zone": "eu-west"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = lookupExtraSuffix(extra)
	}
}

func BenchmarkLookupSegment(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = lookupSegment("Tier Gold")
	}
}

func BenchmarkTrimRuneRun(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = trimRuneRun("__dashboard__", '_')
	}
}

func BenchmarkCompareStrings(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceIntSink = compareStrings("en", "fr")
	}
}

func BenchmarkFirstRuneOf(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, size := firstRuneOf("archive")
		benchServiceRuneSink = r
		benchServiceIntSink = size
	}
}

func BenchmarkLastRuneOf(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, size := lastRuneOf("archive")
		benchServiceRuneSink = r
		benchServiceIntSink = size
	}
}

func BenchmarkHandleMissingKey(b *testing.B) {
	svc := benchServiceFixture()
	args := []any{map[string]any{"Context": "dashboard"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.handleMissingKey("missing.key", args)
	}
}

func BenchmarkMissingKeyArgs(b *testing.B) {
	args := []any{benchServiceContext(), benchServiceSubject(), map[string]string{"Extra": "value"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = missingKeyArgs(args)
	}
}

func BenchmarkMergeMissingKeyArgs(b *testing.B) {
	dst := map[string]any{}
	value := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeMissingKeyArgs(dst, value)
	}
	benchServiceAnySink = dst
}

func BenchmarkMissingKeyContextArgs(b *testing.B) {
	ctx := benchServiceContext()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = missingKeyContextArgs(ctx)
	}
}

func BenchmarkMissingKeySubjectArgs(b *testing.B) {
	subject := benchServiceSubject()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceAnySink = missingKeySubjectArgs(subject)
	}
}

func BenchmarkRaw(b *testing.B) {
	svc := benchServiceFixture()
	data := map[string]any{"Name": "Alex"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = svc.Raw("app.hello", data)
	}
}

func BenchmarkGetMessage(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, ok := svc.getMessage("zz-bench", "app.hello")
		benchServiceMessageSink = msg
		benchServiceBoolSink = ok
	}
}

func BenchmarkAddMessages(b *testing.B) {
	svc := benchServiceFixture()
	messages := map[string]string{"bench.added": "Added"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.AddMessages("zz-bench", messages)
	}
	benchServiceMessageSink, benchServiceBoolSink = svc.getMessage("zz-bench", "bench.added")
}

func BenchmarkIngestLocaleData(b *testing.B) {
	svc := benchServiceFixture()
	messages := benchServiceMessages()
	grammar := benchGrammarFixtureData()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ingestLocaleData("zz-ingest", messages, grammar)
	}
	benchServiceServiceSink = svc
}

func BenchmarkHasLocaleRegistrationLoaded(b *testing.B) {
	svc := benchServiceFixture()
	svc.markLocaleRegistrationLoaded(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.hasLocaleRegistrationLoaded(42)
	}
}

func BenchmarkMarkLocaleRegistrationLoaded(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.markLocaleRegistrationLoaded(42)
	}
	benchServiceBoolSink = svc.hasLocaleRegistrationLoaded(42)
}

func BenchmarkHasLocaleProviderLoaded(b *testing.B) {
	svc := benchServiceFixture()
	svc.markLocaleProviderLoaded(43)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = svc.hasLocaleProviderLoaded(43)
	}
}

func BenchmarkMarkLocaleProviderLoaded(b *testing.B) {
	svc := benchServiceFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.markLocaleProviderLoaded(43)
	}
	benchServiceBoolSink = svc.hasLocaleProviderLoaded(43)
}

func BenchmarkAddAvailableLanguageLocked(b *testing.B) {
	svc := benchServiceFixture()
	tag := language.Make("de")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.mu.Lock()
		svc.addAvailableLanguageLocked(tag)
		svc.mu.Unlock()
	}
	benchServiceStringsSink = svc.AvailableLanguages()
}

func BenchmarkLoadFS(b *testing.B) {
	svc := benchServiceFixture()
	fsys := benchServiceFS()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceResultSink = svc.LoadFS(fsys, "locales")
	}
}

func BenchmarkAutoDetectLanguage(b *testing.B) {
	svc := benchServiceFixture()
	svc.languageExplicit = false
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.autoDetectLanguage()
	}
	benchServiceStringSink = svc.currentLang
}

func BenchmarkHasHandlerType(b *testing.B) {
	handlers := DefaultHandlers()
	candidate := LabelHandler{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasHandlerType(handlers, candidate)
	}
}

func BenchmarkHasLabelHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasLabelHandler(handlers)
	}
}

func BenchmarkHasProgressHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasProgressHandler(handlers)
	}
}

func BenchmarkHasCountHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasCountHandler(handlers)
	}
}

func BenchmarkHasDoneHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasDoneHandler(handlers)
	}
}

func BenchmarkHasFailHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasFailHandler(handlers)
	}
}

func BenchmarkHasNumericHandler(b *testing.B) {
	handlers := DefaultHandlers()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceBoolSink = hasNumericHandler(handlers)
	}
}
