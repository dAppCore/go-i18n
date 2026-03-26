package i18n

import (
	"embed"
	"io/fs"
	"maps"
	"path"
	"slices"
	"sync"
	"sync/atomic"

	"dappco.re/go/core"
	log "dappco.re/go/core/log"
	"golang.org/x/text/language"
)

// Service provides grammar-aware internationalisation.
type Service struct {
	loader         Loader
	messages       map[string]map[string]Message // lang -> key -> message
	currentLang    string
	fallbackLang   string
	availableLangs []language.Tag
	mode           Mode
	debug          bool
	formality      Formality
	handlers       []KeyHandler
	mu             sync.RWMutex
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

// WithHandlers sets custom handlers (replaces default handlers).
func WithHandlers(handlers ...KeyHandler) Option {
	return func(s *Service) { s.handlers = handlers }
}

// WithDefaultHandlers adds the default i18n.* namespace handlers.
func WithDefaultHandlers() Option {
	return func(s *Service) { s.handlers = append(s.handlers, DefaultHandlers()...) }
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
		if grammar != nil && (len(grammar.Verbs) > 0 || len(grammar.Nouns) > 0 || len(grammar.Words) > 0) {
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
			// CAS prevents overwriting a concurrent SetDefault call that
			// raced between the Load check above and this store.
			defaultService.CompareAndSwap(nil, svc)
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
	if len(grammarData.Verbs) > 0 || len(grammarData.Nouns) > 0 || len(grammarData.Words) > 0 {
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
	bestMatch, _, confidence := matcher.Match(requestedLang)
	if confidence == language.No {
		return log.E("Service.SetLanguage", "unsupported language: "+lang, nil)
	}
	s.currentLang = bestMatch.String()
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

func (s *Service) AddHandler(h KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, h)
}

func (s *Service) PrependHandler(h KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append([]KeyHandler{h}, s.handlers...)
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

func (s *Service) resolveWithFallback(messageID string, data any) string {
	if text := s.tryResolve(s.currentLang, messageID, data); text != "" {
		return text
	}
	if text := s.tryResolve(s.fallbackLang, messageID, data); text != "" {
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
	formality := s.getEffectiveFormality(data)
	if formality != FormalityNeutral {
		formalityKey := key + "._" + formality.String()
		if text := s.resolveMessage(lang, formalityKey, data); text != "" {
			return text
		}
	}
	return s.resolveMessage(lang, key, data)
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

func (s *Service) handleMissingKey(key string, args []any) string {
	switch s.mode {
	case ModeStrict:
		panic(core.Sprintf("i18n: missing translation key %q", key))
	case ModeCollect:
		var argsMap map[string]any
		if len(args) > 0 {
			if m, ok := args[0].(map[string]any); ok {
				argsMap = m
			}
		}
		dispatchMissingKey(key, argsMap)
		return "[" + key + "]"
	default:
		return key
	}
}

// Raw translates without i18n.* namespace magic.
func (s *Service) Raw(messageID string, args ...any) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var data any
	if len(args) > 0 {
		data = args[0]
	}
	text := s.resolveWithFallback(messageID, data)
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
		if grammar != nil && (len(grammar.Verbs) > 0 || len(grammar.Nouns) > 0 || len(grammar.Words) > 0) {
			MergeGrammarData(lang, grammar)
		}

		tag := language.Make(lang)
		if !slices.Contains(s.availableLangs, tag) {
			s.availableLangs = append(s.availableLangs, tag)
		}
		s.mu.Unlock()
	}
	return nil
}

// LoadFS loads additional locale files from a filesystem.
// Deprecated: Use AddLoader(NewFSLoader(fsys, dir)) instead for proper grammar handling.
func (s *Service) LoadFS(fsys fs.FS, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
