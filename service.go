package i18n

import (
	"embed"
	"io/fs"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

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
	requestedLang    string
	languageExplicit bool
	availableLangs   []language.Tag
	mode             Mode
	debug            bool
	formality        Formality
	location         string
	handlers         []KeyHandler
	loadedLocales    map[int]struct{}
	loadedProviders  map[int]struct{}
	mu               sync.RWMutex
}

// Option configures a Service during construction.
type Option func(*Service)

// WithFallback sets the fallback language for missing translations.
func WithFallback(lang string) Option {
	return func(s *Service) { s.fallbackLang = normalizeLanguageTag(lang) }
}

// WithLanguage sets an explicit initial language for the service.
//
// The language is applied after the loader has populated the available
// languages, so it can resolve to the best supported tag instead of failing
// during option construction.
func WithLanguage(lang string) Option {
	return func(s *Service) { s.requestedLang = normalizeLanguageTag(lang) }
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
	return func(s *Service) {
		s.handlers = filterNilHandlers(append([]KeyHandler(nil), handlers...))
	}
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
	defaultInitMu  sync.Mutex
)

//go:embed locales/*.json
var localeFS embed.FS

var _ Translator = (*Service)(nil)
var _ core.Translator = (*Service)(nil)

// New creates a new i18n service with embedded locales.
func New(opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(localeFS, "locales"), opts...)
}

// NewService creates a new i18n service with embedded locales.
//
// This is a named alias for New that keeps the constructor intent explicit
// for callers that prefer service-oriented naming.
func NewService(opts ...Option) (*Service, error) {
	return New(opts...)
}

// NewWithFS creates a new i18n service loading locales from the given filesystem.
func NewWithFS(fsys fs.FS, dir string, opts ...Option) (*Service, error) {
	return NewWithLoader(NewFSLoader(fsys, dir), opts...)
}

// NewServiceWithFS creates a new i18n service loading locales from the given filesystem.
func NewServiceWithFS(fsys fs.FS, dir string, opts ...Option) (*Service, error) {
	return NewWithFS(fsys, dir, opts...)
}

// NewWithLoader creates a new i18n service with a custom loader.
func NewWithLoader(loader Loader, opts ...Option) (*Service, error) {
	if loader == nil {
		return nil, log.E("NewWithLoader", "nil loader", nil)
	}
	s := &Service{
		loader:          loader,
		messages:        make(map[string]map[string]Message),
		fallbackLang:    "en",
		handlers:        DefaultHandlers(),
		loadedLocales:   make(map[int]struct{}),
		loadedProviders: make(map[int]struct{}),
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
		lang = normalizeLanguageTag(lang)
		_, seen := s.messages[lang]

		if existing, ok := s.messages[lang]; ok {
			if existing == nil {
				existing = make(map[string]Message)
				s.messages[lang] = existing
			}
			maps.Copy(existing, messages)
		} else {
			s.messages[lang] = messages
		}
		if grammarDataHasContent(grammar) {
			if seen {
				MergeGrammarData(lang, grammar)
			} else {
				SetGrammarData(lang, grammar)
			}
		}
		tag := language.Make(lang)
		if !slices.Contains(s.availableLangs, tag) {
			s.availableLangs = append(s.availableLangs, tag)
		}
	}

	if detected := detectLanguage(s.availableLangs); detected != "" {
		s.currentLang = detected
	} else {
		s.currentLang = s.fallbackLang
	}

	if s.requestedLang != "" {
		if err := s.SetLanguage(s.requestedLang); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// NewServiceWithLoader creates a new i18n service with a custom loader.
func NewServiceWithLoader(loader Loader, opts ...Option) (*Service, error) {
	return NewWithLoader(loader, opts...)
}

// Init initialises the default global service if none has been set via SetDefault.
func Init() error {
	if defaultService.Load() != nil {
		return nil
	}
	defaultInitMu.Lock()
	defer defaultInitMu.Unlock()
	// Re-check after taking the lock so concurrent callers do not create
	// duplicate services.
	if defaultService.Load() != nil {
		return nil
	}
	svc, err := New()
	if err != nil {
		return err
	}
	// Register and load any locales queued before initialisation.
	loadRegisteredLocales(svc)
	// CAS prevents overwriting a concurrent SetDefault call that raced between
	// the Load check above and this store.
	if !defaultService.CompareAndSwap(nil, svc) {
		// If a concurrent caller already installed a service, load registered
		// locales into that active default service instead.
		loadRegisteredLocales(defaultService.Load())
	}
	return nil
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
	if s == nil {
		return
	}
	registeredLocalesMu.Lock()
	hasRegistrations := len(registeredLocales) > 0 || len(registeredLocaleProviders) > 0
	registeredLocalesMu.Unlock()
	if hasRegistrations {
		loadRegisteredLocales(s)
	}
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
	if err := svc.AddLoader(loader); err != nil {
		log.Error("i18n: AddLoader failed", "err", err)
	}
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
	lang = normalizeLanguageTag(lang)
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
		return log.E("Service.SetLanguage", "unsupported language: "+lang+" (available: "+joinAvailableLanguagesLocked(s.availableLangs)+")", nil)
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
	slices.Sort(langs)
	return langs
}

func (s *Service) SetMode(m Mode)           { s.mu.Lock(); s.mode = m; s.mu.Unlock() }
func (s *Service) Mode() Mode               { s.mu.RLock(); defer s.mu.RUnlock(); return s.mode }
func (s *Service) SetFormality(f Formality) { s.mu.Lock(); s.formality = f; s.mu.Unlock() }
func (s *Service) Formality() Formality     { s.mu.RLock(); defer s.mu.RUnlock(); return s.formality }
func (s *Service) SetFallback(lang string) {
	s.mu.Lock()
	s.fallbackLang = normalizeLanguageTag(lang)
	s.mu.Unlock()
}
func (s *Service) Fallback() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fallbackLang
}

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

func joinAvailableLanguagesLocked(tags []language.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	langs := make([]string, len(tags))
	for i, tag := range tags {
		langs[i] = tag.String()
	}
	slices.Sort(langs)
	return strings.Join(langs, ", ")
}

func (s *Service) AddHandler(handlers ...KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, filterNilHandlers(handlers)...)
}

func (s *Service) PrependHandler(handlers ...KeyHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	handlers = filterNilHandlers(handlers)
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
	handlers := append([]KeyHandler(nil), s.handlers...)
	debug := s.debug
	s.mu.RUnlock()

	result := RunHandlerChain(handlers, messageID, args, func() string {
		var data any
		if len(args) > 0 {
			data = args[0]
		}

		s.mu.RLock()
		text := s.resolveWithFallbackLocked(messageID, data)
		s.mu.RUnlock()
		if text == "" {
			return s.handleMissingKey(messageID, args)
		}
		return text
	})
	if debug {
		return debugFormat(messageID, result)
	}
	return result
}

// Translate translates a message by its ID and returns a Core result.
func (s *Service) Translate(messageID string, args ...any) core.Result {
	value := s.T(messageID, args...)
	return core.Result{Value: value, OK: translateOK(messageID, value)}
}

// resolveDirect performs exact-key lookup in the current language, its base
// language tag, and then the configured fallback language.
func (s *Service) resolveDirectLocked(messageID string, data any) string {
	if text := s.tryResolveLocked(s.currentLang, messageID, data); text != "" {
		return text
	}
	if base := baseLanguageTag(s.currentLang); base != "" && base != s.currentLang {
		if text := s.tryResolveLocked(base, messageID, data); text != "" {
			return text
		}
	}
	if text := s.tryResolveLocked(s.fallbackLang, messageID, data); text != "" {
		return text
	}
	if base := baseLanguageTag(s.fallbackLang); base != "" && base != s.fallbackLang {
		return s.tryResolveLocked(base, messageID, data)
	}
	return ""
}

func (s *Service) resolveWithFallbackLocked(messageID string, data any) string {
	if text := s.resolveDirectLocked(messageID, data); text != "" {
		return text
	}
	if core.Contains(messageID, ".") {
		parts := core.Split(messageID, ".")
		verb := parts[len(parts)-1]
		commonKey := "common.action." + verb
		if text := s.resolveDirectLocked(commonKey, data); text != "" {
			return text
		}
		commonKey = "common." + verb
		if text := s.resolveDirectLocked(commonKey, data); text != "" {
			return text
		}
	}
	return ""
}

func (s *Service) tryResolveLocked(lang, key string, data any) string {
	context, gender, location, formality := s.getEffectiveContextGenderLocationAndFormality(data)
	extra := s.getEffectiveContextExtra(data)
	for _, lookupKey := range lookupVariants(key, context, gender, location, formality, extra) {
		if text := s.resolveMessageLocked(lang, lookupKey, data); text != "" {
			return text
		}
	}
	return ""
}

func (s *Service) resolveMessageLocked(lang, key string, data any) string {
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
		location := ctx.LocationString()
		if location == "" {
			location = s.location
		}
		return ctx.ContextString(), ctx.GenderString(), location, formality
	}
	if subj, ok := data.(*Subject); ok && subj != nil {
		formality := subj.formality
		if formality == FormalityNeutral {
			formality = s.formality
		}
		location := subj.location
		if location == "" {
			location = s.location
		}
		return "", subj.gender, location, formality
	}
	if m, ok := data.(map[string]any); ok {
		var context string
		var gender string
		location := s.location
		formality := s.formality
		if v, ok := mapValueString(m, "Context"); ok {
			context = v
		}
		if v, ok := mapValueString(m, "Gender"); ok {
			gender = v
		}
		if v, ok := mapValueString(m, "Location"); ok {
			location = v
		}
		if f, ok := parseFormalityValue(m["Formality"]); ok {
			formality = f
		}
		return context, gender, location, formality
	}
	if m, ok := data.(map[string]string); ok {
		var context string
		var gender string
		location := s.location
		formality := s.formality
		if v, ok := mapValueString(m, "Context"); ok {
			context = v
		}
		if v, ok := mapValueString(m, "Gender"); ok {
			gender = v
		}
		if v, ok := mapValueString(m, "Location"); ok {
			location = v
		}
		if f, ok := parseFormalityValue(m["Formality"]); ok {
			formality = f
		}
		return context, gender, location, formality
	}
	return "", "", s.location, s.getEffectiveFormality(data)
}

func (s *Service) getEffectiveContextExtra(data any) map[string]any {
	switch v := data.(type) {
	case *TranslationContext:
		if v == nil || len(v.Extra) == 0 {
			return nil
		}
		return v.Extra
	case map[string]any:
		return contextMapValues(v)
	case map[string]string:
		return contextMapValues(v)
	default:
		return nil
	}
}

func mergeContextExtra(dst map[string]any, value any) {
	if dst == nil || value == nil {
		return
	}
	switch extra := value.(type) {
	case map[string]any:
		for key, item := range extra {
			dst[key] = item
		}
	case map[string]string:
		for key, item := range extra {
			dst[key] = item
		}
	case *TranslationContext:
		if extra == nil || len(extra.Extra) == 0 {
			return
		}
		for key, item := range extra.Extra {
			dst[key] = item
		}
	}
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
		if f, ok := parseFormalityValue(m["Formality"]); ok {
			return f
		}
	}
	if m, ok := data.(map[string]string); ok {
		if f, ok := parseFormalityValue(m["Formality"]); ok {
			return f
		}
	}
	return s.formality
}

func parseFormalityValue(value any) (Formality, bool) {
	switch f := value.(type) {
	case Formality:
		if f != FormalityNeutral {
			return f, true
		}
	case string:
		switch core.Lower(f) {
		case "formal":
			return FormalityFormal, true
		case "informal":
			return FormalityInformal, true
		}
	}
	return FormalityNeutral, false
}

func lookupVariants(key, context, gender, location string, formality Formality, extra map[string]any) []string {
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
	if extraSuffix := lookupExtraSuffix(extra); extraSuffix != "" {
		base := slices.Clone(variants)
		var extraVariants []string
		for _, variant := range base {
			extraVariants = append(extraVariants, variant+extraSuffix)
		}
		variants = append(extraVariants, variants...)
	}
	variants = append(variants, key)
	return variants
}

func lookupExtraSuffix(extra map[string]any) string {
	if len(extra) == 0 {
		return ""
	}
	keys := slices.Sorted(maps.Keys(extra))
	var b strings.Builder
	for _, key := range keys {
		name := lookupSegment(key)
		if name == "" {
			continue
		}
		value := lookupSegment(core.Sprintf("%v", extra[key]))
		if value == "" {
			continue
		}
		b.WriteString("._")
		b.WriteString(name)
		b.WriteString("._")
		b.WriteString(value)
	}
	return b.String()
}

func lookupSegment(s string) string {
	s = core.Trim(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range core.Lower(s) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == '.' || unicode.IsSpace(r):
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
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
	case map[string]string:
		return contextMapValues(v)
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
	var data any
	if len(args) > 0 {
		data = args[0]
	}
	text := s.resolveDirectLocked(messageID, data)
	debug := s.debug
	s.mu.RUnlock()
	if text == "" {
		text = s.handleMissingKey(messageID, args)
	}
	if debug {
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
	lang = normalizeLanguageTag(lang)
	if lang == "" {
		return
	}

	s.mu.Lock()
	if s.messages[lang] == nil {
		s.messages[lang] = make(map[string]Message)
	}
	for key, text := range messages {
		s.messages[lang][key] = Message{Text: text}
	}
	tag := language.Make(lang)
	if !slices.Contains(s.availableLangs, tag) {
		s.availableLangs = append(s.availableLangs, tag)
	}
	s.mu.Unlock()

	s.autoDetectLanguage()
}

// AddLoader loads translations from an additional Loader, merging messages
// and grammar data into the existing service. This is the correct way to
// add package-specific translations at runtime.
func (s *Service) AddLoader(loader Loader) error {
	if loader == nil {
		return log.E("Service.AddLoader", "nil loader", nil)
	}
	langs := loader.Languages()
	if len(langs) == 0 {
		if el, ok := loader.(interface{ LanguagesErr() error }); ok {
			if langErr := el.LanguagesErr(); langErr != nil {
				return log.E("Service.AddLoader", "read locales directory", langErr)
			}
		}
		return log.E("Service.AddLoader", "no languages available from loader", nil)
	}
	for _, lang := range langs {
		messages, grammar, err := loader.Load(lang)
		if err != nil {
			return log.E("Service.AddLoader", "load locale: "+lang, err)
		}
		lang = normalizeLanguageTag(lang)

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

func (s *Service) hasLocaleRegistrationLoaded(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.loadedLocales) == 0 {
		return false
	}
	_, ok := s.loadedLocales[id]
	return ok
}

func (s *Service) markLocaleRegistrationLoaded(id int) {
	if id == 0 || s == nil {
		return
	}
	s.mu.Lock()
	if s.loadedLocales == nil {
		s.loadedLocales = make(map[int]struct{})
	}
	s.loadedLocales[id] = struct{}{}
	s.mu.Unlock()
}

func (s *Service) hasLocaleProviderLoaded(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.loadedProviders) == 0 {
		return false
	}
	_, ok := s.loadedProviders[id]
	return ok
}

func (s *Service) markLocaleProviderLoaded(id int) {
	if id == 0 || s == nil {
		return
	}
	s.mu.Lock()
	if s.loadedProviders == nil {
		s.loadedProviders = make(map[int]struct{})
	}
	s.loadedProviders[id] = struct{}{}
	s.mu.Unlock()
}

func translateOK(messageID, value string) bool {
	if value == "" {
		return false
	}
	prefix := "[" + messageID + "] "
	if core.HasPrefix(value, prefix) {
		value = core.TrimPrefix(value, prefix)
	} else if value == "["+messageID+"]" {
		value = ""
	}
	if value == "" {
		return false
	}
	if value == messageID {
		return false
	}
	if value == "["+messageID+"]" {
		return false
	}
	return true
}

// LoadFS loads additional locale files from a filesystem.
// Deprecated: Use AddLoader(NewFSLoader(fsys, dir)) instead for proper grammar handling.
func (s *Service) LoadFS(fsys fs.FS, dir string) error {
	loader := NewFSLoader(fsys, dir)
	langs := loader.Languages()
	if len(langs) == 0 {
		if langErr := loader.LanguagesErr(); langErr != nil {
			return log.E("Service.LoadFS", "read locales directory", langErr)
		}
		return log.E("Service.LoadFS", "no languages available", nil)
	}
	return s.AddLoader(loader)
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
