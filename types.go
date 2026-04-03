// Package i18n provides grammar-aware internationalisation.
//
// Unlike flat key-value translation systems, this package composes
// grammatically correct output from verbs, nouns, and articles.
// The i18n.* namespace provides magic key handlers that auto-compose
// labels, progress messages, counts, and action results:
//
//	T("i18n.label.status")              // "Status:"
//	T("i18n.progress.build")            // "Building..."
//	T("i18n.count.file", 5)             // "5 files"
//	T("i18n.done.delete", "config.yaml") // "Config.yaml deleted"
//	T("i18n.fail.push", "commits")      // "Failed to push commits"
package i18n

import "sync"

// --- Core Types ---

// Mode determines how the service handles missing translation keys.
//
//	i18n.SetMode(i18n.ModeStrict)
type Mode int

const (
	ModeNormal  Mode = iota // Returns key as-is (production)
	ModeStrict              // Panics on missing key (dev/CI)
	ModeCollect             // Dispatches MissingKey events, returns [key] (QA)
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeStrict:
		return "strict"
	case ModeCollect:
		return "collect"
	default:
		return "unknown"
	}
}

// Formality represents the level of formality in translations.
//
//	i18n.S("user", "Alex").Formal()
type Formality int

const (
	FormalityNeutral  Formality = iota // Context-appropriate (default)
	FormalityInformal                  // du, tu, you
	FormalityFormal                    // Sie, vous, usted
)

// TextDirection represents text directionality.
//
//	if i18n.Direction() == i18n.DirRTL { /* ... */ }
type TextDirection int

const (
	DirLTR TextDirection = iota // Left-to-right
	DirRTL                      // Right-to-left
)

// PluralCategory represents CLDR plural categories.
//
//	cat := i18n.CurrentPluralCategory(2)
type PluralCategory int

const (
	PluralOther PluralCategory = iota // Default/fallback
	PluralZero                        // n=0 (Arabic, Latvian)
	PluralOne                         // n=1 (most languages)
	PluralTwo                         // n=2 (Arabic, Welsh)
	PluralFew                         // Small numbers (Slavic: 2-4)
	PluralMany                        // Larger numbers (Slavic: 5+)
)

// GrammaticalGender represents grammatical gender for nouns.
//
//	i18n.S("user", "Alex").Gender("feminine")
type GrammaticalGender int

const (
	GenderNeuter    GrammaticalGender = iota // das, it
	GenderMasculine                          // der, le
	GenderFeminine                           // die, la
	GenderCommon                             // Swedish, Dutch
)

// --- Message Types ---

// Message represents a translation — either a simple string or plural forms.
//
//	msg := i18n.Message{One: "{{.Count}} file", Other: "{{.Count}} files"}
type Message struct {
	Text  string // Simple string value (non-plural)
	Zero  string // count == 0 (Arabic, Latvian, Welsh)
	One   string // count == 1 (most languages)
	Two   string // count == 2 (Arabic, Welsh)
	Few   string // Small numbers (Slavic: 2-4, Arabic: 3-10)
	Many  string // Larger numbers (Slavic: 5+, Arabic: 11-99)
	Other string // Default/fallback form
}

// ForCategory returns the appropriate text for a plural category.
func (m Message) ForCategory(cat PluralCategory) string {
	switch cat {
	case PluralZero:
		if m.Zero != "" {
			return m.Zero
		}
	case PluralOne:
		if m.One != "" {
			return m.One
		}
	case PluralTwo:
		if m.Two != "" {
			return m.Two
		}
	case PluralFew:
		if m.Few != "" {
			return m.Few
		}
	case PluralMany:
		if m.Many != "" {
			return m.Many
		}
	}
	if m.Other != "" {
		return m.Other
	}
	if m.One != "" {
		return m.One
	}
	return m.Text
}

// IsPlural returns true if this message has any plural forms.
func (m Message) IsPlural() bool {
	return m.Zero != "" || m.One != "" || m.Two != "" ||
		m.Few != "" || m.Many != "" || m.Other != ""
}

// --- Subject Types ---

// Subject represents a typed subject with metadata for semantic translations.
//
//	subj := i18n.S("file", "config.yaml").Count(3).In("workspace")
type Subject struct {
	Noun      string    // The noun type (e.g., "file", "repo")
	Value     any       // The actual value (e.g., filename)
	count     int       // Count for pluralisation (default 1)
	gender    string    // Grammatical gender
	location  string    // Location context (e.g., "in workspace")
	formality Formality // Formality level override
}

// --- Intent Types ---

// IntentMeta defines the behaviour of an intent.
//
//	intent := i18n.Intent{Meta: i18n.IntentMeta{Type: "action", Verb: "delete"}}
type IntentMeta struct {
	Type      string   // "action", "question", "info"
	Verb      string   // Reference to verb key
	Dangerous bool     // If true, requires confirmation
	Default   string   // Default response: "yes" or "no"
	Supports  []string // Extra options
}

// Composed holds all output forms for an intent after template resolution.
//
//	composed := i18n.ComposeIntent(i18n.Intent{Question: "Delete {{.Subject}}?"}, i18n.S("file", "config.yaml"))
type Composed struct {
	Question string     // "Delete config.yaml?"
	Confirm  string     // "Really delete config.yaml?"
	Success  string     // "config.yaml deleted"
	Failure  string     // "Failed to delete config.yaml"
	Meta     IntentMeta // Intent metadata
}

// Intent defines a semantic intent with templates for all output forms.
//
//	intent := i18n.Intent{Question: "Delete {{.Subject}}?"}
type Intent struct {
	Meta     IntentMeta
	Question string // Template for question form
	Confirm  string // Template for confirmation form
	Success  string // Template for success message
	Failure  string // Template for failure message
}

// templateData is passed to intent templates during execution.
type templateData struct {
	Subject   string
	Noun      string
	Count     int
	Gender    string
	Location  string
	Formality Formality
	IsFormal  bool
	IsPlural  bool
	Value     any
}

// --- Grammar Types ---

// GrammarData holds language-specific grammar forms loaded from JSON.
//
//	i18n.SetGrammarData("en", &i18n.GrammarData{Articles: i18n.ArticleForms{IndefiniteDefault: "a"}})
type GrammarData struct {
	Verbs    map[string]VerbForms // verb -> forms
	Nouns    map[string]NounForms // noun -> forms
	Articles ArticleForms         // article configuration
	Words    map[string]string    // base word translations
	Punct    PunctuationRules     // language-specific punctuation
	Signals  SignalData           // disambiguation signal word lists
	Number   NumberFormat         // locale-specific number formatting
}

// VerbForms holds verb conjugations.
//
//	forms := i18n.VerbForms{Past: "deleted", Gerund: "deleting"}
type VerbForms struct {
	Past   string // "deleted"
	Gerund string // "deleting"
}

// NounForms holds plural and gender information for a noun.
//
//	forms := i18n.NounForms{One: "file", Other: "files"}
type NounForms struct {
	One    string // Singular form
	Other  string // Plural form
	Gender string // Grammatical gender
}

// ArticleForms holds article configuration for a language.
//
//	articles := i18n.ArticleForms{IndefiniteDefault: "a", IndefiniteVowel: "an"}
type ArticleForms struct {
	IndefiniteDefault string            // "a"
	IndefiniteVowel   string            // "an"
	Definite          string            // "the"
	ByGender          map[string]string // Gender-specific articles
}

// PunctuationRules holds language-specific punctuation patterns.
//
//	rules := i18n.PunctuationRules{LabelSuffix: ":", ProgressSuffix: "..."}
type PunctuationRules struct {
	LabelSuffix    string // ":" (French uses " :")
	ProgressSuffix string // "..."
}

// SignalData holds word lists used for disambiguation signals.
//
//	signals := i18n.SignalData{VerbAuxiliaries: []string{"is", "was"}}
type SignalData struct {
	NounDeterminers []string                      // Words that precede nouns: "the", "a", "this", "my", ...
	VerbAuxiliaries []string                      // Auxiliaries/modals before verbs: "is", "was", "will", ...
	VerbInfinitive  []string                      // Infinitive markers: "to"
	VerbNegation    []string                      // Negation cues that weakly signal a verb: "not", "never", ...
	Priors          map[string]map[string]float64 // Corpus-derived verb/noun priors for ambiguous words, consumed by the reversal tokeniser.
}

// --- Number Formatting ---

// NumberFormat defines locale-specific number formatting rules.
//
//	fmt := i18n.NumberFormat{ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"}
type NumberFormat struct {
	ThousandsSep string // "," for en, "." for de
	DecimalSep   string // "." for en, "," for de
	PercentFmt   string // "%s%%" for en, "%s %%" for de
}

// --- Function Types ---

// PluralRule determines the plural category for a count.
//
//	rule := i18n.GetPluralRule("en")
type PluralRule func(n int) PluralCategory

// MissingKeyHandler receives missing key events.
//
//	i18n.OnMissingKey(func(m i18n.MissingKey) {})
type MissingKeyHandler func(missing MissingKey)

// MissingKey is dispatched when a translation key is not found in ModeCollect.
//
//	func handle(m i18n.MissingKey) { _ = m.Key }
type MissingKey struct {
	Key        string
	Args       map[string]any
	CallerFile string
	CallerLine int
}

// --- Interfaces ---

// KeyHandler processes translation keys before standard lookup.
// Handlers form a chain; each can handle a key or delegate to the next.
//
//	i18n.AddHandler(i18n.LabelHandler{})
type KeyHandler interface {
	Match(key string) bool
	Handle(key string, args []any, next func() string) string
}

// Loader provides translation data to the Service.
//
//	svc, err := i18n.NewWithLoader(loader)
type Loader interface {
	Load(lang string) (map[string]Message, *GrammarData, error)
	Languages() []string
}

// Translator defines the interface for translation services.
//
//	var t i18n.Translator = i18n.Default()
type Translator interface {
	T(messageID string, args ...any) string
	SetLanguage(lang string) error
	Language() string
	SetMode(m Mode)
	Mode() Mode
	SetDebug(enabled bool)
	Debug() bool
	SetFormality(f Formality)
	Formality() Formality
	Direction() TextDirection
	IsRTL() bool
	PluralCategory(n int) PluralCategory
	AvailableLanguages() []string
}

// --- Package Variables ---

var (
	grammarCache   = make(map[string]*GrammarData)
	grammarCacheMu sync.RWMutex
)

var templateCache sync.Map

var numberFormats = map[string]NumberFormat{
	"en": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
	"de": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s %%"},
	"fr": {ThousandsSep: " ", DecimalSep: ",", PercentFmt: "%s %%"},
	"es": {ThousandsSep: ".", DecimalSep: ",", PercentFmt: "%s%%"},
	"zh": {ThousandsSep: ",", DecimalSep: ".", PercentFmt: "%s%%"},
}

var rtlLanguages = map[string]bool{
	"ar": true, "ar-SA": true, "ar-EG": true,
	"he": true, "he-IL": true,
	"fa": true, "fa-IR": true,
	"ur": true, "ur-PK": true,
	"yi": true, "ps": true, "sd": true, "ug": true,
}

var pluralRules = map[string]PluralRule{
	"en": pluralRuleEnglish, "en-GB": pluralRuleEnglish, "en-US": pluralRuleEnglish,
	"de": pluralRuleGerman, "de-DE": pluralRuleGerman, "de-AT": pluralRuleGerman, "de-CH": pluralRuleGerman,
	"fr": pluralRuleFrench, "fr-FR": pluralRuleFrench, "fr-CA": pluralRuleFrench,
	"es": pluralRuleSpanish, "es-ES": pluralRuleSpanish, "es-MX": pluralRuleSpanish,
	"ru": pluralRuleRussian, "ru-RU": pluralRuleRussian,
	"pl": pluralRulePolish, "pl-PL": pluralRulePolish,
	"ar": pluralRuleArabic, "ar-SA": pluralRuleArabic,
	"cy": pluralRuleWelsh, "cy-GB": pluralRuleWelsh,
	"zh": pluralRuleChinese, "zh-CN": pluralRuleChinese, "zh-TW": pluralRuleChinese,
	"ja": pluralRuleJapanese, "ja-JP": pluralRuleJapanese,
	"ko": pluralRuleKorean, "ko-KR": pluralRuleKorean,
}

// --- Irregular Forms ---

var irregularVerbs = map[string]VerbForms{
	"be": {Past: "was", Gerund: "being"}, "have": {Past: "had", Gerund: "having"},
	"do": {Past: "did", Gerund: "doing"}, "go": {Past: "went", Gerund: "going"},
	"make": {Past: "made", Gerund: "making"}, "get": {Past: "got", Gerund: "getting"},
	"run": {Past: "ran", Gerund: "running"}, "set": {Past: "set", Gerund: "setting"},
	"put": {Past: "put", Gerund: "putting"}, "cut": {Past: "cut", Gerund: "cutting"},
	"let": {Past: "let", Gerund: "letting"}, "hit": {Past: "hit", Gerund: "hitting"},
	"shut": {Past: "shut", Gerund: "shutting"}, "split": {Past: "split", Gerund: "splitting"},
	"spread": {Past: "spread", Gerund: "spreading"}, "read": {Past: "read", Gerund: "reading"},
	"write": {Past: "wrote", Gerund: "writing"}, "send": {Past: "sent", Gerund: "sending"},
	"build": {Past: "built", Gerund: "building"}, "begin": {Past: "began", Gerund: "beginning"},
	"find": {Past: "found", Gerund: "finding"}, "take": {Past: "took", Gerund: "taking"},
	"see": {Past: "saw", Gerund: "seeing"}, "keep": {Past: "kept", Gerund: "keeping"},
	"hold": {Past: "held", Gerund: "holding"}, "tell": {Past: "told", Gerund: "telling"},
	"bring": {Past: "brought", Gerund: "bringing"}, "think": {Past: "thought", Gerund: "thinking"},
	"buy": {Past: "bought", Gerund: "buying"}, "catch": {Past: "caught", Gerund: "catching"},
	"teach": {Past: "taught", Gerund: "teaching"}, "throw": {Past: "threw", Gerund: "throwing"},
	"grow": {Past: "grew", Gerund: "growing"}, "know": {Past: "knew", Gerund: "knowing"},
	"show": {Past: "showed", Gerund: "showing"}, "draw": {Past: "drew", Gerund: "drawing"},
	"break": {Past: "broke", Gerund: "breaking"}, "speak": {Past: "spoke", Gerund: "speaking"},
	"choose": {Past: "chose", Gerund: "choosing"}, "forget": {Past: "forgot", Gerund: "forgetting"},
	"lose": {Past: "lost", Gerund: "losing"}, "win": {Past: "won", Gerund: "winning"},
	"swim": {Past: "swam", Gerund: "swimming"}, "drive": {Past: "drove", Gerund: "driving"},
	"rise": {Past: "rose", Gerund: "rising"}, "shine": {Past: "shone", Gerund: "shining"},
	"sing": {Past: "sang", Gerund: "singing"}, "ring": {Past: "rang", Gerund: "ringing"},
	"drink": {Past: "drank", Gerund: "drinking"}, "sink": {Past: "sank", Gerund: "sinking"},
	"sit": {Past: "sat", Gerund: "sitting"}, "stand": {Past: "stood", Gerund: "standing"},
	"hang": {Past: "hung", Gerund: "hanging"}, "dig": {Past: "dug", Gerund: "digging"},
	"stick": {Past: "stuck", Gerund: "sticking"}, "bite": {Past: "bit", Gerund: "biting"},
	"hide": {Past: "hid", Gerund: "hiding"}, "feed": {Past: "fed", Gerund: "feeding"},
	"meet": {Past: "met", Gerund: "meeting"}, "lead": {Past: "led", Gerund: "leading"},
	"sleep": {Past: "slept", Gerund: "sleeping"}, "feel": {Past: "felt", Gerund: "feeling"},
	"leave": {Past: "left", Gerund: "leaving"}, "mean": {Past: "meant", Gerund: "meaning"},
	"lend": {Past: "lent", Gerund: "lending"}, "spend": {Past: "spent", Gerund: "spending"},
	"bend": {Past: "bent", Gerund: "bending"}, "deal": {Past: "dealt", Gerund: "dealing"},
	"lay": {Past: "laid", Gerund: "laying"}, "pay": {Past: "paid", Gerund: "paying"},
	"say": {Past: "said", Gerund: "saying"}, "sell": {Past: "sold", Gerund: "selling"},
	"seek": {Past: "sought", Gerund: "seeking"}, "fight": {Past: "fought", Gerund: "fighting"},
	"fly": {Past: "flew", Gerund: "flying"}, "wear": {Past: "wore", Gerund: "wearing"},
	"tear": {Past: "tore", Gerund: "tearing"}, "bear": {Past: "bore", Gerund: "bearing"},
	"swear": {Past: "swore", Gerund: "swearing"}, "wake": {Past: "woke", Gerund: "waking"},
	"freeze": {Past: "froze", Gerund: "freezing"}, "steal": {Past: "stole", Gerund: "stealing"},
	"overwrite": {Past: "overwritten", Gerund: "overwriting"}, "reset": {Past: "reset", Gerund: "resetting"},
	"reboot": {Past: "rebooted", Gerund: "rebooting"},
	"submit": {Past: "submitted", Gerund: "submitting"}, "permit": {Past: "permitted", Gerund: "permitting"},
	"admit": {Past: "admitted", Gerund: "admitting"}, "omit": {Past: "omitted", Gerund: "omitting"},
	"commit": {Past: "committed", Gerund: "committing"}, "transmit": {Past: "transmitted", Gerund: "transmitting"},
	"prefer": {Past: "preferred", Gerund: "preferring"}, "refer": {Past: "referred", Gerund: "referring"},
	"transfer": {Past: "transferred", Gerund: "transferring"}, "defer": {Past: "deferred", Gerund: "deferring"},
	"confer": {Past: "conferred", Gerund: "conferring"}, "infer": {Past: "inferred", Gerund: "inferring"},
	"occur": {Past: "occurred", Gerund: "occurring"}, "recur": {Past: "recurred", Gerund: "recurring"},
	"incur": {Past: "incurred", Gerund: "incurring"}, "deter": {Past: "deterred", Gerund: "deterring"},
	"control": {Past: "controlled", Gerund: "controlling"}, "patrol": {Past: "patrolled", Gerund: "patrolling"},
	"compel": {Past: "compelled", Gerund: "compelling"}, "expel": {Past: "expelled", Gerund: "expelling"},
	"propel": {Past: "propelled", Gerund: "propelling"}, "repel": {Past: "repelled", Gerund: "repelling"},
	"rebel": {Past: "rebelled", Gerund: "rebelling"}, "excel": {Past: "excelled", Gerund: "excelling"},
	"cancel": {Past: "cancelled", Gerund: "cancelling"}, "travel": {Past: "travelled", Gerund: "travelling"},
	"label": {Past: "labelled", Gerund: "labelling"}, "model": {Past: "modelled", Gerund: "modelling"},
	"level":       {Past: "levelled", Gerund: "levelling"},
	"format":      {Past: "formatted", Gerund: "formatting"},
	"analyse":     {Past: "analysed", Gerund: "analysing"},
	"organise":    {Past: "organised", Gerund: "organising"},
	"recognise":   {Past: "recognised", Gerund: "recognising"},
	"realise":     {Past: "realised", Gerund: "realising"},
	"customise":   {Past: "customised", Gerund: "customising"},
	"optimise":    {Past: "optimised", Gerund: "optimising"},
	"initialise":  {Past: "initialised", Gerund: "initialising"},
	"synchronise": {Past: "synchronised", Gerund: "synchronising"},
	// Compound irregular verbs (prefix + base)
	"undo": {Past: "undid", Gerund: "undoing"}, "redo": {Past: "redid", Gerund: "redoing"},
	"rerun": {Past: "reran", Gerund: "rerunning"}, "rewrite": {Past: "rewrote", Gerund: "rewriting"},
	"rebuild": {Past: "rebuilt", Gerund: "rebuilding"}, "resend": {Past: "resent", Gerund: "resending"},
	"override": {Past: "overrode", Gerund: "overriding"}, "rethink": {Past: "rethought", Gerund: "rethinking"},
	"remake": {Past: "remade", Gerund: "remaking"}, "undergo": {Past: "underwent", Gerund: "undergoing"},
	"overcome": {Past: "overcame", Gerund: "overcoming"}, "withdraw": {Past: "withdrew", Gerund: "withdrawing"},
	"uphold": {Past: "upheld", Gerund: "upholding"}, "withhold": {Past: "withheld", Gerund: "withholding"},
	"outgrow": {Past: "outgrew", Gerund: "outgrowing"}, "outrun": {Past: "outran", Gerund: "outrunning"},
	"overshoot": {Past: "overshot", Gerund: "overshooting"},
	// Simple irregular verbs (dev/ops)
	"become": {Past: "became", Gerund: "becoming"}, "come": {Past: "came", Gerund: "coming"},
	"give": {Past: "gave", Gerund: "giving"}, "fall": {Past: "fell", Gerund: "falling"},
	"understand": {Past: "understood", Gerund: "understanding"}, "arise": {Past: "arose", Gerund: "arising"},
	"bind": {Past: "bound", Gerund: "binding"}, "spin": {Past: "spun", Gerund: "spinning"},
	"quit": {Past: "quit", Gerund: "quitting"}, "cast": {Past: "cast", Gerund: "casting"},
	"broadcast": {Past: "broadcast", Gerund: "broadcasting"}, "burst": {Past: "burst", Gerund: "bursting"},
	"cost": {Past: "cost", Gerund: "costing"}, "shed": {Past: "shed", Gerund: "shedding"},
	"rid": {Past: "rid", Gerund: "ridding"}, "shrink": {Past: "shrank", Gerund: "shrinking"},
	"shoot": {Past: "shot", Gerund: "shooting"}, "forbid": {Past: "forbade", Gerund: "forbidding"},
	"offset": {Past: "offset", Gerund: "offsetting"}, "upset": {Past: "upset", Gerund: "upsetting"},
	"input": {Past: "input", Gerund: "inputting"}, "output": {Past: "output", Gerund: "outputting"},
	// CVC doubling failures (stressed final syllable, >4 chars)
	"debug": {Past: "debugged", Gerund: "debugging"}, "embed": {Past: "embedded", Gerund: "embedding"},
	"unzip": {Past: "unzipped", Gerund: "unzipping"}, "remap": {Past: "remapped", Gerund: "remapping"},
	"unpin": {Past: "unpinned", Gerund: "unpinning"}, "unwrap": {Past: "unwrapped", Gerund: "unwrapping"},
}

var noDoubleConsonant = map[string]bool{
	"open": true, "listen": true, "happen": true, "enter": true, "offer": true,
	"suffer": true, "differ": true, "cover": true, "deliver": true, "develop": true,
	"visit": true, "limit": true, "edit": true, "credit": true, "orbit": true,
	"total": true, "target": true, "budget": true, "market": true, "benefit": true, "focus": true,
}

var irregularNouns = map[string]string{
	"child": "children", "person": "people", "man": "men", "woman": "women",
	"foot": "feet", "tooth": "teeth", "mouse": "mice", "goose": "geese",
	"ox": "oxen", "index": "indices", "appendix": "appendices", "matrix": "matrices",
	"vertex": "vertices", "crisis": "crises", "analysis": "analyses", "diagnosis": "diagnoses",
	"thesis": "theses", "hypothesis": "hypotheses", "parenthesis": "parentheses",
	"datum": "data", "medium": "media", "bacterium": "bacteria", "criterion": "criteria",
	"phenomenon": "phenomena", "curriculum": "curricula", "alumnus": "alumni",
	"cactus": "cacti", "focus": "foci", "fungus": "fungi", "nucleus": "nuclei",
	"radius": "radii", "stimulus": "stimuli", "syllabus": "syllabi",
	"fish": "fish", "sheep": "sheep", "deer": "deer", "species": "species",
	"series": "series", "aircraft": "aircraft",
	"life": "lives", "wife": "wives", "knife": "knives", "leaf": "leaves",
	"half": "halves", "self": "selves", "shelf": "shelves", "wolf": "wolves",
	"calf": "calves", "loaf": "loaves", "thief": "thieves",
}

// dualClassVerbs seeds additional regular verbs that are also common nouns in
// dev/ops text. The forms are regular, but listing them here makes the
// reversal tokeniser treat them as known bases for dual-class disambiguation.
var dualClassVerbs = map[string]VerbForms{
	"change":   {Past: "changed", Gerund: "changing"},
	"export":   {Past: "exported", Gerund: "exporting"},
	"function": {Past: "functioned", Gerund: "functioning"},
	"handle":   {Past: "handled", Gerund: "handling"},
	"host":     {Past: "hosted", Gerund: "hosting"},
	"import":   {Past: "imported", Gerund: "importing"},
	"link":     {Past: "linked", Gerund: "linking"},
	"log":      {Past: "logged", Gerund: "logging"},
	"merge":    {Past: "merged", Gerund: "merging"},
	"patch":    {Past: "patched", Gerund: "patching"},
	"process":  {Past: "processed", Gerund: "processing"},
	"queue":    {Past: "queued", Gerund: "queuing"},
	"release":  {Past: "released", Gerund: "releasing"},
	"pull":     {Past: "pulled", Gerund: "pulling"},
	"push":     {Past: "pushed", Gerund: "pushing"},
	"stream":   {Past: "streamed", Gerund: "streaming"},
	"tag":      {Past: "tagged", Gerund: "tagging"},
	"trigger":  {Past: "triggered", Gerund: "triggering"},
	"watch":    {Past: "watched", Gerund: "watching"},
	"update":   {Past: "updated", Gerund: "updating"},
}

// dualClassNouns mirrors the same vocabulary as nouns so the tokeniser can
// classify the base forms as ambiguous when they appear without inflection.
var dualClassNouns = map[string]string{
	"change":   "changes",
	"export":   "exports",
	"function": "functions",
	"handle":   "handles",
	"host":     "hosts",
	"import":   "imports",
	"link":     "links",
	"log":      "logs",
	"merge":    "merges",
	"patch":    "patches",
	"process":  "processes",
	"queue":    "queues",
	"release":  "releases",
	"pull":     "pulls",
	"push":     "pushes",
	"stream":   "streams",
	"tag":      "tags",
	"trigger":  "triggers",
	"watch":    "watches",
	"update":   "updates",
}

var vowelSounds = map[string]bool{
	"hour": true, "honest": true, "honour": true, "honor": true, "heir": true, "herb": true,
}

var consonantSounds = map[string]bool{
	"user": true, "union": true, "unique": true, "unit": true, "universe": true,
	"university": true, "uniform": true, "usage": true, "usual": true, "utility": true,
	"utensil": true, "one": true, "once": true, "euro": true, "eulogy": true, "euphemism": true,
}
