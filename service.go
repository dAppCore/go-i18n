package i18n

import (
	"embed"
	"io/fs"
	"maps"
	"path"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
	"golang.org/x/text/language"
)

// Service provides grammar-aware internationalisation.
type Service struct {
	loader           Loader
	messages         map[string]map[string]Message // lang -> key -> message
	currentLang      string
	fallbackLang     string
	languageExplicit bool
	availableLangs   []language.Tag
	mode             Mode
	debug            bool
	formality        Formality
	location         string
	handlers         []KeyHandler
	mu               sync.RWMutex
}

// Option configures a Service during construction.
type Option func(*Service)

// WithFallback sets the fallback language for missing translations.
func WithFallback(lang string) Option {
	return func(s *Service) { s.fallbackLang = lang }
}

// WithFormality sets the default formality level.
func WithFormality(f Formality) Option {
	return func(s *Service) { s.formality = f }
}

// WithLocation sets the default location context.
func WithLocation(location string) Option {
	return func(s *Service) { s.location = location }
}

// WithHandlers sets custom handlers (replaces default handlers).
func WithHandlers(handlers ...KeyHandler) Option {
	return func(s *Service) { s.handlers = handlers }
}

// WithDefaultHandlers adds the default i18n.* namespace handlers.
func WithDefaultHandlers() Option {
	return func(s *Service) {
		for _, handler := range DefaultHandlers() {
			if hasHandlerType(s.handlers, handler) {
				continue
			}
			s.handlers = append(s.handlers, handler)
		}
	}
}

// WithMode sets the translation mode.
func WithMode(m Mode) Option {
	return func(s *Service) { s.mode = m }
}

// WithDebug enables or disables debug mode.
func WithDebug(enabled bool) Option {
	return func(s *Service) { s.debug = enabled }
}

var (
	defaultService atomic.Pointer[Service]
	defaultOnce    sync.Once
	defaultErr     error
)

//go:embed locales/*.json
var localeFS embed.FS

var _ Translator = (*Service)(nil)

// New creates a new i18n service with embedded locales.
func New(opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(localeFS, "locales"), opts...)
}

// NewWithFS creates a new i18n service loading locales from the given filesystem.
func NewWithFS(fsys fs.FS, dir string, opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(fsys, dir), opts...)
}

// NewWithLoader creates a new i18n service with a custom loader.
func NewWithLoader(loader Loader, opts ...Option) (*Service, error) {
	s := &Service{
		loader:       loader,
		messages:     make(map[string]map[string]Message),
		fallbackLang: "en",
		handlers:     DefaultHandlers(),
	}
	for _, opt := range opts {
		opt(s)
	}

	langs := loader.Languages()
	if len(langs) == 0 {
		// Check if the loader exposes a scan error (e.g. FSLoader).
		if el, ok := loader.(interface{ LanguagesErr() error }); ok {
			if langErr := el.LanguagesErr(); langErr != nil {
				return nil, log.E("NewWithLoader", "no languages available", langErr)
			}
		}
		return nil, log.E("NewWithLoader", "no languages available from loader", nil)
	}

	for _, lang := range langs {
		messages, grammar, err := loader.Load(lang)
		if err != nil {
			return nil, log.E("NewWithLoader", "load locale: "+lang, err)
		}
		s.messages[lang] = messages
		if grammarDataHasContent(grammar) {
			SetGrammarData(lang, grammar)
		}
		tag := language.Make(lang)
		s.availableLangs = append(s.availableLangs, tag)
	}

	if detected := detectLanguage(s.availableLangs); detected != "" {
		s.currentLang = detected
	} else {
		s.currentLang = s.fallbackLang
	}

	return s, nil
}

// Init initialises the default global service if none has been set via SetDefault.
func Init() error {
	defaultOnce.Do(func() {
		// If SetDefault was already called, don't overwrite
		if defaultService.Load() != nil {
			return
		}
		svc, err := New()
		if err == nil {
			// Register and load any locales queued before initialisation.
			loadRegisteredLocales(svc)
			// CAS prevents overwriting a concurrent SetDefault call that
			// raced between the Load check above and this store.
			if !defaultService.CompareAndSwap(nil, svc) {
				// If a concurrent caller already installed a service, load
				// registered locales into that active default service instead.
				loadRegisteredLocales(defaultService.Load())
			}
		}
		defaultErr = err
	})
	return defaultErr
}

// Default returns the global i18n service, initialising if needed.
// Returns nil if initialisation fails (error is logged).
func Default() *Service {
	if svc := defaultService.Load(); svc != nil {
		return svc
	}
	if err := Init(); err != nil {
		log.Error("failed to initialise default service", "err", err)
	}
	return defaultService.Load()
}

// SetDefault sets the global i18n service.
// Passing nil clears the default service.
func SetDefault(s *Service) {
	defaultService.Store(s)
}

// AddLoader loads translations from a Loader into the default service.
// Call this from init() in packages that ship their own locale files:
//
//	//go:embed *.json
//	var localeFS embed.FS
//	func init() { i18n.AddLoader(i18n.NewFSLoader(localeFS, ".")) }
//
// Note: When using the Core framework, NewCoreService creates a fresh Service
// and calls SetDefault, so init-time AddLoader calls are superseded. In that
// context, packages should implement LocaleProvider instead.
func AddLoader(loader Loader) {
	svc := Default()
	if svc == nil {
		return
	}
	_ = svc.AddLoader(loader)
}

func (s *Service) loadJSON(lang string, data []byte) error {
	var raw map[string]any
	if r := core.JSONUnmarshal(data, &raw); !r.OK {
		return r.Value.(error)
	}
	messages := make(map[string]Message)
	grammarData := &GrammarData{
		Verbs: make(map[string]VerbForms),
		Nouns: make(map[string]NounForms),
		Words: make(map[string]string),
	}
	flattenWithGrammar("", raw, messages, grammarData)
	if existing, ok := s.messages[lang]; ok {
		maps.Copy(existing, messages)
	} else {
		s.messages[lang] = messages
	}
	if grammarDataHasContent(grammarData) {
		MergeGrammarData(lang, grammarData)
	}
	return nil
}

// SetLanguage sets the language for translations.
func (s *Service) SetLanguage(lang string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	requestedLang, err := language.Parse(lang)
	if err != nil {
		return log.E("Service.SetLanguage", "invalid language tag: "+lang, err)
	}
	if len(s.availableLangs) == 0 {
		return log.E("Service.SetLanguage", "no languages available", nil)
	}
	matcher := language.NewMatcher(s.availableLangs)
	bestMatch, bestIndex, confidence := matcher.Match(requestedLang)
	if confidence == language.No {
		return log.E("Service.SetLanguage", "unsupported language: "+lang, nil)
	}
	if bestIndex >= 0 && bestIndex < len(s.availableLangs) {
		s.currentLang = s.availableLangs[bestIndex].String()
	} else {
		s.currentLang = bestMatch.String()
	}
	s.languageExplicit = true
	return nil
}

func (s *Service) Language() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLang
}

func (s *Service) AvailableLanguages() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	langs := make([]string, len(s.availableLangs))
	for i, tag := range s.availableLangs {
		langs[i] = tag.String()
	}
	return langs
}

func (s *Service) SetMode(m Mode)           { s.mu.Lock(); s.mode = m; s.mu.Unlock() }
func (s *Service) Mode() Mode               { s.mu.RLock(); defer s.mu.RUnlock(); return s.mode }
func (s *Service) SetFormality(f Formality) { s.mu.Lock(); s.formality = f; s.mu.Unlock() }
func (s *Service) Formality() Formality     { s.mu.RLock(); defer s.mu.RUnlock(); return s.formality }

func (s *Service) SetLocation(location string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.location = location
}

func (s *Service) Location() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.location
}

func (s *Service) Direction() TextDirection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if IsRTLLanguage(s.currentLang) {
		return DirRTL
	}
	return DirLTR
}

func (s *Service) IsRTL() bool { return s.Direction() == DirRTL }

func (s *Service) PluralCategory(n int) PluralCategory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return GetPluralCategory(s.currentLang, n)
}

func (s *Service) AddHandler(handlers ...KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handlers...)
}

func (s *Service) PrependHandler(handlers ...KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(handlers) == 0 {
		return
	}
	s.handlers = append(append([]KeyHandler(nil), handlers...), s.handlers...)
}

func (s *Service) ClearHandlers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = nil
}

func (s *Service) Handlers() []KeyHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]KeyHandler, len(s.handlers))
	copy(result, s.handlers)
	return result
}

// T translates a message by its ID with handler chain support.
//
//	T("i18n.label.status")              // "Status:"
//	T("i18n.progress.build")            // "Building..."
//	T("i18n.count.file", 5)             // "5 files"
//	T("i18n.done.delete", "file")       // "File deleted"
//	T("i18n.fail.delete", "file")       // "Failed to delete file"
func (s *Service) T(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := RunHandlerChain(s.handlers, messageID, args, func() string {
		var data any
		if len(args) > 0 {
			data = args[0]
		}
		text := s.resolveWithFallback(messageID, data)
		if text == "" {
			return s.handleMissingKey(messageID, args)
		}
		return text
	})
	if s.debug {
		return debugFormat(messageID, result)
	}
	return result
}

// resolveDirect performs exact-key lookup in the current language and fallback language.
func (s *Service) resolveDirect(messageID string, data any) string {
	if text := s.tryResolve(s.currentLang, messageID, data); text != "" {
		return text
	}
	return s.tryResolve(s.fallbackLang, messageID, data)
}

func (s *Service) resolveWithFallback(messageID string, data any) string {
	if text := s.resolveDirect(messageID, data); text != "" {
		return text
	}
	if core.Contains(messageID, ".") {
		parts := core.Split(messageID, ".")
		verb := parts[len(parts)-1]
		commonKey := "common.action." + verb
		if text := s.tryResolve(s.currentLang, commonKey, data); text != "" {
			return text
		}
		if text := s.tryResolve(s.fallbackLang, commonKey, data); text != "" {
			return text
		}
		commonKey = "common." + verb
		if text := s.tryResolve(s.currentLang, commonKey, data); text != "" {
			return text
		}
		if text := s.tryResolve(s.fallbackLang, commonKey, data); text != "" {
			return text
		}
	}
	return ""
}

func (s *Service) tryResolve(lang, key string, data any) string {
	context, gender, location, formality := s.getEffectiveContextGenderLocationAndFormality(data)
	for _, lookupKey := range lookupVariants(key, context, gender, location, formality) {
		if text := s.resolveMessage(lang, lookupKey, data); text != "" {
			return text
		}
	}
	return ""
}

func (s *Service) resolveMessage(lang, key string, data any) string {
	msg, ok := s.getMessage(lang, key)
	if !ok {
		return ""
	}
	text := msg.Text
	if msg.IsPlural() {
		count := getCount(data)
		category := GetPluralCategory(lang, count)
		text = msg.ForCategory(category)
	}
	if text == "" {
		return ""
	}
	if data != nil {
		text = applyTemplate(text, data)
	}
	return text
}

func (s *Service) getEffectiveContextGenderLocationAndFormality(data any) (string, string, string, Formality) {
	if ctx, ok := data.(*TranslationContext); ok && ctx != nil {
		formality := ctx.FormalityValue()
		if formality == FormalityNeutral {
			formality = s.formality
		}
		return ctx.ContextString(), ctx.GenderString(), ctx.LocationString(), formality
	}
	if subj, ok := data.(*Subject); ok && subj != nil {
		formality := subj.formality
		if formality == FormalityNeutral {
			formality = s.formality
		}
		return "", subj.gender, subj.location, formality
	}
	if m, ok := data.(map[string]any); ok {
		var context string
		var gender string
		var location string
		formality := s.formality
		if v, ok := m["Context"].(string); ok {
			context = core.Trim(v)
		}
		if v, ok := m["Gender"].(string); ok {
			gender = core.Trim(v)
		}
		if v, ok := m["Location"].(string); ok {
			location = core.Trim(v)
		}
		if v, ok := m["Formality"]; ok {
			switch f := v.(type) {
			case Formality:
				if f != FormalityNeutral {
					formality = f
				}
			case string:
				switch core.Lower(f) {
				case "formal":
					formality = FormalityFormal
				case "informal":
					formality = FormalityInformal
				}
			}
		}
		return context, gender, location, formality
	}
	return "", "", s.location, s.getEffectiveFormality(data)
}

func (s *Service) getEffectiveFormality(data any) Formality {
	if ctx, ok := data.(*TranslationContext); ok && ctx != nil {
		if ctx.Formality != FormalityNeutral {
			return ctx.Formality
		}
	}
	if subj, ok := data.(*Subject); ok && subj != nil {
		if subj.formality != FormalityNeutral {
			return subj.formality
		}
	}
	if m, ok := data.(map[string]any); ok {
		switch f := m["Formality"].(type) {
		case Formality:
			if f != FormalityNeutral {
				return f
			}
		case string:
			switch core.Lower(f) {
			case "formal":
				return FormalityFormal
			case "informal":
				return FormalityInformal
			}
		}
	}
	return s.formality
}

func lookupVariants(key, context, gender, location string, formality Formality) []string {
	var variants []string
	if context != "" {
		if gender != "" && location != "" && formality != FormalityNeutral {
			variants = append(variants, key+"._"+context+"._"+gender+"._"+location+"._"+formality.String())
		}
		if gender != "" && location != "" {
			variants = append(variants, key+"._"+context+"._"+gender+"._"+location)
		}
		if gender != "" && formality != FormalityNeutral {
			variants = append(variants, key+"._"+context+"._"+gender+"._"+formality.String())
		}
		if gender != "" {
			variants = append(variants, key+"._"+context+"._"+gender)
		}
		if location != "" && formality != FormalityNeutral {
			variants = append(variants, key+"._"+context+"._"+location+"._"+formality.String())
		}
		if location != "" {
			variants = append(variants, key+"._"+context+"._"+location)
		}
		if formality != FormalityNeutral {
			variants = append(variants, key+"._"+context+"._"+formality.String())
		}
		variants = append(variants, key+"._"+context)
	}
	if gender != "" && location != "" && formality != FormalityNeutral {
		variants = append(variants, key+"._"+gender+"._"+location+"._"+formality.String())
	}
	if gender != "" && location != "" {
		variants = append(variants, key+"._"+gender+"._"+location)
	}
	if gender != "" && formality != FormalityNeutral {
		variants = append(variants, key+"._"+gender+"._"+formality.String())
	}
	if gender != "" {
		variants = append(variants, key+"._"+gender)
	}
	if location != "" && formality != FormalityNeutral {
		variants = append(variants, key+"._"+location+"._"+formality.String())
	}
	if location != "" {
		variants = append(variants, key+"._"+location)
	}
	if formality != FormalityNeutral {
		variants = append(variants, key+"._"+formality.String())
	}
	variants = append(variants, key)
	return variants
}

func (s *Service) handleMissingKey(key string, args []any) string {
	switch s.mode {
	case ModeStrict:
		panic(core.Sprintf("i18n: missing translation key %q", key))
	case ModeCollect:
		argsMap := missingKeyArgs(args)
		dispatchMissingKey(key, argsMap)
		return "[" + key + "]"
	default:
		return key
	}
}

func missingKeyArgs(args []any) map[string]any {
	if len(args) == 0 {
		return nil
	}
	switch v := args[0].(type) {
	case map[string]any:
		return v
	case *TranslationContext:
		return missingKeyContextArgs(v)
	case *Subject:
		return missingKeySubjectArgs(v)
	default:
		return nil
	}
}

func missingKeyContextArgs(ctx *TranslationContext) map[string]any {
	if ctx == nil {
		return nil
	}
	data := templateDataForRendering(ctx)
	result, _ := data.(map[string]any)
	return result
}

func missingKeySubjectArgs(subj *Subject) map[string]any {
	if subj == nil {
		return nil
	}
	data := templateDataForRendering(subj)
	result, _ := data.(map[string]any)
	return result
}

// Raw translates without i18n.* namespace magic.
func (s *Service) Raw(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var data any
	if len(args) > 0 {
		data = args[0]
	}
	text := s.resolveDirect(messageID, data)
	if text == "" {
		return s.handleMissingKey(messageID, args)
	}
	if s.debug {
		return debugFormat(messageID, text)
	}
	return text
}

func (s *Service) getMessage(lang, key string) (Message, bool) {
	msgs, ok := s.messages[lang]
	if !ok {
		return Message{}, false
	}
	msg, ok := msgs[key]
	return msg, ok
}

// AddMessages adds messages for a language at runtime.
func (s *Service) AddMessages(lang string, messages map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.messages[lang] == nil {
		s.messages[lang] = make(map[string]Message)
	}
	for key, text := range messages {
		s.messages[lang][key] = Message{Text: text}
	}
}

// AddLoader loads translations from an additional Loader, merging messages
// and grammar data into the existing service. This is the correct way to
// add package-specific translations at runtime.
func (s *Service) AddLoader(loader Loader) error {
	langs := loader.Languages()
	for _, lang := range langs {
		messages, grammar, err := loader.Load(lang)
		if err != nil {
			return log.E("Service.AddLoader", "load locale: "+lang, err)
		}

		s.mu.Lock()
		if s.messages[lang] == nil {
			s.messages[lang] = make(map[string]Message)
		}
		for k, v := range messages {
			s.messages[lang][k] = v
		}

		// Merge grammar data into the global grammar store (merge, not replace,
		// so that multiple loaders contribute entries for the same language).
		if grammarDataHasContent(grammar) {
			MergeGrammarData(lang, grammar)
		}

		tag := language.Make(lang)
		if !slices.Contains(s.availableLangs, tag) {
			s.availableLangs = append(s.availableLangs, tag)
		}
		s.mu.Unlock()
	}
	s.autoDetectLanguage()
	return nil
}

// LoadFS loads additional locale files from a filesystem.
// Deprecated: Use AddLoader(NewFSLoader(fsys, dir)) instead for proper grammar handling.
func (s *Service) LoadFS(fsys fs.FS, dir string) error {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		if s.languageExplicit {
			return
		}
		if detected := detectLanguage(s.availableLangs); detected != "" {
			s.mu.Lock()
			s.currentLang = detected
			s.mu.Unlock()
		}
	}()
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return log.E("Service.LoadFS", "read locales directory", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !core.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := path.Join(dir, entry.Name())
		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return log.E("Service.LoadFS", "read locale: "+entry.Name(), err)
		}
		lang := core.TrimSuffix(entry.Name(), ".json")
		lang = core.Replace(lang, "_", "-")
		if err := s.loadJSON(lang, data); err != nil {
			return log.E("Service.LoadFS", "parse locale: "+entry.Name(), err)
		}
		tag := language.Make(lang)
		found := slices.Contains(s.availableLangs, tag)
		if !found {
			s.availableLangs = append(s.availableLangs, tag)
		}
	}
	return nil
}

func (s *Service) autoDetectLanguage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.languageExplicit {
		return
	}
	if detected := detectLanguage(s.availableLangs); detected != "" {
		s.currentLang = detected
	}
}

func hasHandlerType(handlers []KeyHandler, candidate KeyHandler) bool {
	want := reflect.TypeOf(candidate)
	for _, handler := range handlers {
		if reflect.TypeOf(handler) == want {
			return true
		}
	}
	return false
}
