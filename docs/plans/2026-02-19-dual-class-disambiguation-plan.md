# Dual-Class Word Disambiguation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add multi-signal probabilistic disambiguation so dual-class words (commit, run, test, check, file, build) are correctly classified as verb or noun based on grammatical context, with confidence scores flowing into imprints.

**Architecture:** Two-pass tokenisation. Pass 1 classifies unambiguous tokens and marks dual-class base forms as ambiguous. Pass 2 evaluates 7 weighted signals against the resolved context to score ambiguous tokens. Confidence values weight imprint contributions.

**Tech Stack:** Pure Go, no new dependencies. Data-driven signal tables in `locales/en.json`.

---

### Task 1: Add SignalData to GrammarData and load from JSON

**Files:**
- Modify: `types.go:188-195` (GrammarData struct)
- Modify: `loader.go:99-211` (flattenWithGrammar function)
- Modify: `locales/en.json` (add gram.signal block)
- Test: `loader_test.go` (if exists, otherwise verify via Task 2 tests)

**Step 1: Write the failing test**

Create a test in the root package that verifies SignalData is loaded:

```go
// In a new file or existing test file
func TestGrammarData_Signals(t *testing.T) {
    svc, err := New()
    if err != nil {
        t.Fatalf("New() failed: %v", err)
    }
    SetDefault(svc)

    data := GetGrammarData("en")
    if data == nil {
        t.Fatal("GetGrammarData(\"en\") returned nil")
    }
    if len(data.Signals.NounDeterminers) == 0 {
        t.Error("Signals.NounDeterminers is empty")
    }
    if len(data.Signals.VerbAuxiliaries) == 0 {
        t.Error("Signals.VerbAuxiliaries is empty")
    }
    if len(data.Signals.VerbInfinitive) == 0 {
        t.Error("Signals.VerbInfinitive is empty")
    }

    // Spot-check known values
    found := false
    for _, d := range data.Signals.NounDeterminers {
        if d == "the" {
            found = true
        }
    }
    if !found {
        t.Error("NounDeterminers missing 'the'")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestGrammarData_Signals -v ./...`
Expected: FAIL — `SignalData` type doesn't exist yet

**Step 3: Add SignalData type to types.go**

Add after `PunctuationRules` struct (after line 222):

```go
// SignalData holds word lists used for disambiguation signals.
type SignalData struct {
    NounDeterminers []string // Words that precede nouns: "the", "a", "this", "my", ...
    VerbAuxiliaries []string // Auxiliaries/modals before verbs: "is", "was", "will", ...
    VerbInfinitive  []string // Infinitive markers: "to"
}
```

Add `Signals` field to `GrammarData` struct at line 195:

```go
type GrammarData struct {
    Verbs    map[string]VerbForms
    Nouns    map[string]NounForms
    Articles ArticleForms
    Words    map[string]string
    Punct    PunctuationRules
    Signals  SignalData
}
```

**Step 4: Add gram.signal handling to loader.go**

In `flattenWithGrammar`, add a new case before the existing `gram.article` handler (before line 158). The signal block is a map of string arrays:

```go
// Signal data for disambiguation
if grammar != nil && fullKey == "gram.signal" {
    if nd, ok := v["noun_determiner"]; ok {
        if arr, ok := nd.([]any); ok {
            for _, item := range arr {
                if s, ok := item.(string); ok {
                    grammar.Signals.NounDeterminers = append(grammar.Signals.NounDeterminers, s)
                }
            }
        }
    }
    if va, ok := v["verb_auxiliary"]; ok {
        if arr, ok := va.([]any); ok {
            for _, item := range arr {
                if s, ok := item.(string); ok {
                    grammar.Signals.VerbAuxiliaries = append(grammar.Signals.VerbAuxiliaries, s)
                }
            }
        }
    }
    if vi, ok := v["verb_infinitive"]; ok {
        if arr, ok := vi.([]any); ok {
            for _, item := range arr {
                if s, ok := item.(string); ok {
                    grammar.Signals.VerbInfinitive = append(grammar.Signals.VerbInfinitive, s)
                }
            }
        }
    }
    continue
}
```

**Step 5: Add gram.signal block to locales/en.json**

Add after the `"punct"` block (after line 112, before `"number"`):

```json
"signal": {
    "noun_determiner": [
        "the", "a", "an",
        "this", "that", "these", "those",
        "my", "your", "his", "her", "its", "our", "their",
        "every", "each", "some", "any", "no",
        "many", "few", "several", "all", "both"
    ],
    "verb_auxiliary": [
        "is", "are", "was", "were",
        "has", "had", "have",
        "do", "does", "did",
        "will", "would", "could", "should",
        "can", "may", "might", "shall", "must"
    ],
    "verb_infinitive": ["to"]
},
```

**Step 6: Run test to verify it passes**

Run: `go test -run TestGrammarData_Signals -v ./...`
Expected: PASS

**Step 7: Run full test suite**

Run: `go test ./...`
Expected: All existing tests still pass (no breaking changes)

**Step 8: Commit**

```bash
git add types.go loader.go locales/en.json
git commit -m "feat(grammar): add SignalData for disambiguation signals

Load noun_determiner, verb_auxiliary, and verb_infinitive word lists
from gram.signal in locale JSON. Data-driven signal tables enable
context-aware dual-class word disambiguation."
```

---

### Task 2: Add dual-class verb/noun entries to en.json

**Files:**
- Modify: `locales/en.json` (add verb entries for test/check/file; add noun entries for run/build)

**Step 1: Write the failing test**

```go
func TestGrammarData_DualClassEntries(t *testing.T) {
    svc, _ := New()
    SetDefault(svc)
    data := GetGrammarData("en")

    dualClass := []string{"commit", "run", "test", "check", "file", "build"}
    for _, word := range dualClass {
        if _, ok := data.Verbs[word]; !ok {
            t.Errorf("gram.verb missing dual-class word %q", word)
        }
        if _, ok := data.Nouns[word]; !ok {
            t.Errorf("gram.noun missing dual-class word %q", word)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestGrammarData_DualClassEntries -v .`
Expected: FAIL — missing "run" and "build" as nouns, missing "test", "check", "file" as verbs

**Step 3: Add missing entries to en.json**

Add to `gram.verb` (these are regular verbs, the grammar engine handles their morphology, but explicit entries ensure they're in the verb index):

```json
"test":  { "base": "test",  "past": "tested",  "gerund": "testing" },
"check": { "base": "check", "past": "checked", "gerund": "checking" },
"file":  { "base": "file",  "past": "filed",   "gerund": "filing" }
```

Add to `gram.noun`:

```json
"run":   { "one": "run",   "other": "runs" },
"build": { "one": "build", "other": "builds" }
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestGrammarData_DualClassEntries -v .`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./...`
Expected: All pass. Note: existing tokeniser tests may now show "commit", "run", "build" matching differently since they'll hit both verb and noun indexes. Current tests only check MatchVerb directly (not via Tokenise), so they should still pass.

**Step 6: Commit**

```bash
git add locales/en.json
git commit -m "feat(grammar): add dual-class verb/noun entries

Add test, check, file as verbs and run, build as nouns so the
tokeniser can detect them in both grammatical roles. The dual-class
set {commit, run, test, check, file, build} is now complete."
```

---

### Task 3: Add Token confidence fields and SignalBreakdown types

**Files:**
- Modify: `reversal/tokeniser.go:24-60` (Token struct, add new types)

**Step 1: Write the failing test**

```go
func TestToken_ConfidenceField(t *testing.T) {
    setup(t)
    tok := NewTokeniser()
    tokens := tok.Tokenise("Deleted the file")

    for _, token := range tokens {
        if token.Type != TokenUnknown && token.Confidence == 0 {
            t.Errorf("token %q (type %d) has zero Confidence", token.Raw, token.Type)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestToken_ConfidenceField -v ./reversal/`
Expected: FAIL — `Confidence` field doesn't exist on Token

**Step 3: Add fields to Token struct and new types**

In `reversal/tokeniser.go`, update the Token struct:

```go
// Token represents a single classified token from a text string.
type Token struct {
    Raw        string          // Original text as it appeared in input
    Lower      string          // Lowercased form
    Type       TokenType       // Classification
    Confidence float64         // 0.0-1.0 classification confidence
    AltType    TokenType       // Runner-up classification (dual-class only)
    AltConf    float64         // Runner-up confidence
    VerbInfo   VerbMatch       // Set when Type OR AltType == TokenVerb
    NounInfo   NounMatch       // Set when Type OR AltType == TokenNoun
    WordCat    string          // Set when Type == TokenWord
    ArtType    string          // Set when Type == TokenArticle
    PunctType  string          // Set when Type == TokenPunctuation
    Signals    *SignalBreakdown // Non-nil only when WithSignals() option is set
}

// SignalBreakdown provides detailed scoring for dual-class disambiguation.
type SignalBreakdown struct {
    VerbScore  float64
    NounScore  float64
    Components []SignalComponent
}

// SignalComponent describes a single signal's contribution to disambiguation.
type SignalComponent struct {
    Name    string  // "noun_determiner", "verb_auxiliary", etc.
    Weight  float64 // Signal weight (0.0-1.0)
    Value   float64 // Signal firing strength (0.0-1.0)
    Contrib float64 // weight × value
    Reason  string  // Human-readable: "preceded by 'the'"
}
```

**Step 4: Set Confidence=1.0 for all non-ambiguous classifications in Tokenise()**

In the existing `Tokenise()` method, every branch that classifies a token should set `Confidence: 1.0`. Update the article, verb, noun, word, and punctuation branches:

```go
if artType, ok := t.MatchArticle(word); ok {
    tok.Type = TokenArticle
    tok.ArtType = artType
    tok.Confidence = 1.0
} else if vm, ok := t.MatchVerb(word); ok {
    tok.Type = TokenVerb
    tok.VerbInfo = vm
    tok.Confidence = 1.0
} else if nm, ok := t.MatchNoun(word); ok {
    tok.Type = TokenNoun
    tok.NounInfo = nm
    tok.Confidence = 1.0
} else if cat, ok := t.MatchWord(word); ok {
    tok.Type = TokenWord
    tok.WordCat = cat
    tok.Confidence = 1.0
} else {
    tok.Type = TokenUnknown
}
```

And for punctuation tokens:

```go
tokens = append(tokens, Token{
    Raw:        punct,
    Lower:      punct,
    Type:       TokenPunctuation,
    PunctType:  punctType,
    Confidence: 1.0,
})
```

**Step 5: Run test to verify it passes**

Run: `go test -run TestToken_ConfidenceField -v ./reversal/`
Expected: PASS

**Step 6: Run full test suite**

Run: `go test ./...`
Expected: All pass

**Step 7: Commit**

```bash
git add reversal/tokeniser.go
git commit -m "feat(reversal): add Token confidence and SignalBreakdown types

Every classified token now carries a Confidence score (1.0 for
unambiguous tokens). SignalBreakdown and SignalComponent types
provide detailed scoring for the upcoming dual-class disambiguation."
```

---

### Task 4: Add TokeniserOption and signal/dual-class indexes

**Files:**
- Modify: `reversal/tokeniser.go` (Tokeniser struct, NewTokeniser signature, build methods)

**Step 1: Write the failing test**

```go
func TestTokeniser_WithSignals(t *testing.T) {
    setup(t)
    tok := NewTokeniser(WithSignals())
    _ = tok // just verify it compiles and accepts the option
}

func TestTokeniser_DualClassDetection(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    // "commit" should be detected as dual-class
    if !tok.IsDualClass("commit") {
        t.Error("commit should be dual-class")
    }
    if !tok.IsDualClass("run") {
        t.Error("run should be dual-class")
    }
    // "delete" should NOT be dual-class (only a verb)
    if tok.IsDualClass("delete") {
        t.Error("delete should not be dual-class")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run "TestTokeniser_WithSignals|TestTokeniser_DualClassDetection" -v ./reversal/`
Expected: FAIL — `WithSignals` and `IsDualClass` don't exist, `NewTokeniser` doesn't accept options

**Step 3: Add TokeniserOption type and update constructors**

```go
// TokeniserOption configures a Tokeniser.
type TokeniserOption func(*Tokeniser)

// WithSignals enables detailed SignalBreakdown on ambiguous tokens.
func WithSignals() TokeniserOption {
    return func(t *Tokeniser) { t.withSignals = true }
}
```

Add fields to Tokeniser struct:

```go
type Tokeniser struct {
    pastToBase   map[string]string
    gerundToBase map[string]string
    baseVerbs    map[string]bool
    pluralToBase map[string]string
    baseNouns    map[string]bool
    words        map[string]string
    lang         string

    dualClass   map[string]bool // words in both verb AND noun tables
    nounDet     map[string]bool // signal: noun determiners
    verbAux     map[string]bool // signal: verb auxiliaries
    verbInf     map[string]bool // signal: infinitive markers
    withSignals bool            // allocate SignalBreakdown on ambiguous tokens
}
```

Update `NewTokeniser` and `NewTokeniserForLang`:

```go
func NewTokeniser(opts ...TokeniserOption) *Tokeniser {
    return NewTokeniserForLang("en", opts...)
}

func NewTokeniserForLang(lang string, opts ...TokeniserOption) *Tokeniser {
    t := &Tokeniser{
        pastToBase:   make(map[string]string),
        gerundToBase: make(map[string]string),
        baseVerbs:    make(map[string]bool),
        pluralToBase: make(map[string]string),
        baseNouns:    make(map[string]bool),
        words:        make(map[string]string),
        lang:         lang,
    }
    for _, opt := range opts {
        opt(t)
    }
    t.buildVerbIndex()
    t.buildNounIndex()
    t.buildWordIndex()
    t.buildDualClassIndex()
    t.buildSignalIndex()
    return t
}
```

Add `IsDualClass` method:

```go
// IsDualClass returns true if the word exists in both verb and noun tables.
func (t *Tokeniser) IsDualClass(word string) bool {
    return t.dualClass[strings.ToLower(word)]
}
```

Add `buildDualClassIndex`:

```go
func (t *Tokeniser) buildDualClassIndex() {
    t.dualClass = make(map[string]bool)
    for base := range t.baseVerbs {
        if t.baseNouns[base] {
            t.dualClass[base] = true
        }
    }
}
```

Add `buildSignalIndex`:

```go
func (t *Tokeniser) buildSignalIndex() {
    t.nounDet = make(map[string]bool)
    t.verbAux = make(map[string]bool)
    t.verbInf = make(map[string]bool)

    data := i18n.GetGrammarData(t.lang)
    if data != nil && len(data.Signals.NounDeterminers) > 0 {
        for _, w := range data.Signals.NounDeterminers {
            t.nounDet[strings.ToLower(w)] = true
        }
        for _, w := range data.Signals.VerbAuxiliaries {
            t.verbAux[strings.ToLower(w)] = true
        }
        for _, w := range data.Signals.VerbInfinitive {
            t.verbInf[strings.ToLower(w)] = true
        }
        return
    }

    // Fallback: hardcoded English defaults
    for _, w := range []string{
        "the", "a", "an", "this", "that", "these", "those",
        "my", "your", "his", "her", "its", "our", "their",
        "every", "each", "some", "any", "no",
        "many", "few", "several", "all", "both",
    } {
        t.nounDet[w] = true
    }
    for _, w := range []string{
        "is", "are", "was", "were", "has", "had", "have",
        "do", "does", "did", "will", "would", "could", "should",
        "can", "may", "might", "shall", "must",
    } {
        t.verbAux[w] = true
    }
    t.verbInf["to"] = true
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run "TestTokeniser_WithSignals|TestTokeniser_DualClassDetection" -v ./reversal/`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./...`
Expected: All pass — `NewTokeniser()` with no args still works (variadic)

**Step 6: Commit**

```bash
git add reversal/tokeniser.go
git commit -m "feat(reversal): add TokeniserOption, dual-class and signal indexes

NewTokeniser now accepts variadic options (backwards compatible).
Builds dual-class index from verb∩noun overlap and signal word
lookup sets from gram.signal data."
```

---

### Task 5: Implement two-pass Tokenise with disambiguation scoring

**Files:**
- Modify: `reversal/tokeniser.go` (Tokenise method — major rewrite)

This is the core task. The existing `Tokenise()` becomes Pass 1 + Pass 2.

**Step 1: Write the failing tests**

```go
func TestTokeniser_Disambiguate_NounAfterDeterminer(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("the commit was approved")
    // "commit" after "the" should be noun
    if tokens[1].Type != TokenNoun {
        t.Errorf("'commit' after 'the': Type = %v, want TokenNoun", tokens[1].Type)
    }
    if tokens[1].Confidence < 0.8 {
        t.Errorf("'commit' after 'the': Confidence = %f, want >= 0.8", tokens[1].Confidence)
    }
    // Should also have verb as alt
    if tokens[1].AltType != TokenVerb {
        t.Errorf("'commit' AltType = %v, want TokenVerb", tokens[1].AltType)
    }
}

func TestTokeniser_Disambiguate_VerbImperative(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("Commit the changes")
    // "Commit" sentence-initial before article → verb
    if tokens[0].Type != TokenVerb {
        t.Errorf("'Commit' imperative: Type = %v, want TokenVerb", tokens[0].Type)
    }
    if tokens[0].Confidence < 0.8 {
        t.Errorf("'Commit' imperative: Confidence = %f, want >= 0.8", tokens[0].Confidence)
    }
}

func TestTokeniser_Disambiguate_NounWithVerbSaturation(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("The test failed")
    // "test" after "The" should be noun, "failed" is confident verb
    if tokens[1].Type != TokenNoun {
        t.Errorf("'test' in 'The test failed': Type = %v, want TokenNoun", tokens[1].Type)
    }
}

func TestTokeniser_Disambiguate_VerbBeforeNoun(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("Run tests")
    // "Run" sentence-initial before noun → verb
    if tokens[0].Type != TokenVerb {
        t.Errorf("'Run' in 'Run tests': Type = %v, want TokenVerb", tokens[0].Type)
    }
}

func TestTokeniser_Disambiguate_InflectedSelfResolve(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    // Inflected forms should NOT be ambiguous — they self-resolve in Pass 1
    tokens := tok.Tokenise("committed the files")
    if tokens[0].Type != TokenVerb {
        t.Errorf("'committed': Type = %v, want TokenVerb", tokens[0].Type)
    }
    if tokens[0].Confidence != 1.0 {
        t.Errorf("'committed': Confidence = %f, want 1.0 (inflected self-resolves)", tokens[0].Confidence)
    }

    tokens = tok.Tokenise("the commits were reviewed")
    if tokens[1].Type != TokenNoun {
        t.Errorf("'commits': Type = %v, want TokenNoun", tokens[1].Type)
    }
    if tokens[1].Confidence != 1.0 {
        t.Errorf("'commits': Confidence = %f, want 1.0 (inflected self-resolves)", tokens[1].Confidence)
    }
}

func TestTokeniser_Disambiguate_VerbAfterAuxiliary(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("will commit the changes")
    // "commit" after "will" → verb
    if tokens[1].Type != TokenVerb {
        t.Errorf("'commit' after 'will': Type = %v, want TokenVerb", tokens[1].Type)
    }
}

func TestTokeniser_Disambiguate_ProseMultiple(t *testing.T) {
    setup(t)
    tok := NewTokeniser()

    tokens := tok.Tokenise("The test failed because the commit introduced a regression")
    // Find "test" and "commit" — both should be nouns
    for _, token := range tokens {
        if token.Lower == "test" && token.Type != TokenNoun {
            t.Errorf("'test' in prose: Type = %v, want TokenNoun", token.Type)
        }
        if token.Lower == "commit" && token.Type != TokenNoun {
            t.Errorf("'commit' in prose: Type = %v, want TokenNoun", token.Type)
        }
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -run "TestTokeniser_Disambiguate" -v ./reversal/`
Expected: FAIL — current tokeniser always picks verb-first for dual-class base forms

**Step 3: Implement two-pass Tokenise**

Replace the body of `Tokenise()` with the two-pass algorithm. The existing method signature stays the same.

Internal constant for ambiguous state (not exported):

```go
const tokenAmbiguous TokenType = -1 // internal only, resolved in Pass 2
```

Pass 1 loop — same structure as current, but dual-class base forms get marked ambiguous:

```go
func (t *Tokeniser) Tokenise(text string) []Token {
    text = strings.TrimSpace(text)
    if text == "" {
        return nil
    }

    parts := strings.Fields(text)
    var tokens []Token

    // --- Pass 1: classify unambiguous tokens, mark dual-class base forms ---
    for _, raw := range parts {
        word, punct := splitTrailingPunct(raw)

        if word != "" {
            tok := Token{Raw: raw, Lower: strings.ToLower(word)}
            classified := false

            // Articles are never ambiguous
            if artType, ok := t.MatchArticle(word); ok {
                tok.Type = TokenArticle
                tok.ArtType = artType
                tok.Confidence = 1.0
                classified = true
            }

            if !classified {
                vm, verbOK := t.MatchVerb(word)
                nm, nounOK := t.MatchNoun(word)

                if verbOK && nounOK && t.dualClass[strings.ToLower(word)] {
                    // Both match AND it's a known dual-class word
                    if vm.Tense != "base" {
                        // Inflected verb form → self-resolves as verb
                        tok.Type = TokenVerb
                        tok.VerbInfo = vm
                        tok.Confidence = 1.0
                        classified = true
                    } else if nm.Plural {
                        // Plural noun form → self-resolves as noun
                        tok.Type = TokenNoun
                        tok.NounInfo = nm
                        tok.Confidence = 1.0
                        classified = true
                    } else {
                        // Base form, both match → ambiguous
                        tok.Type = tokenAmbiguous
                        tok.VerbInfo = vm
                        tok.NounInfo = nm
                        classified = true
                    }
                } else if verbOK {
                    tok.Type = TokenVerb
                    tok.VerbInfo = vm
                    tok.Confidence = 1.0
                    classified = true
                } else if nounOK {
                    tok.Type = TokenNoun
                    tok.NounInfo = nm
                    tok.Confidence = 1.0
                    classified = true
                }
            }

            if !classified {
                if cat, ok := t.MatchWord(word); ok {
                    tok.Type = TokenWord
                    tok.WordCat = cat
                    tok.Confidence = 1.0
                } else {
                    tok.Type = TokenUnknown
                }
            }

            tokens = append(tokens, tok)
        }

        if punct != "" {
            if punctType, ok := matchPunctuation(punct); ok {
                tokens = append(tokens, Token{
                    Raw:        punct,
                    Lower:      punct,
                    Type:       TokenPunctuation,
                    PunctType:  punctType,
                    Confidence: 1.0,
                })
            }
        }
    }

    // --- Pass 2: resolve ambiguous tokens ---
    t.resolveAmbiguous(tokens)

    return tokens
}
```

Pass 2 — the scoring engine:

```go
func (t *Tokeniser) resolveAmbiguous(tokens []Token) {
    for i := range tokens {
        if tokens[i].Type != tokenAmbiguous {
            continue
        }
        verbScore, nounScore, components := t.scoreAmbiguous(tokens, i)
        t.resolveToken(&tokens[i], verbScore, nounScore, components)
    }
}

func (t *Tokeniser) scoreAmbiguous(tokens []Token, idx int) (float64, float64, []SignalComponent) {
    var verbScore, nounScore float64
    var components []SignalComponent

    word := tokens[idx].Lower

    // Signal 1: Preceding noun determiner (weight 0.35)
    if idx > 0 {
        prev := tokens[idx-1].Lower
        if t.nounDet[prev] {
            nounScore += 0.35
            components = append(components, SignalComponent{
                Name: "noun_determiner", Weight: 0.35, Value: 1.0,
                Contrib: 0.35, Reason: "preceded by '" + prev + "'",
            })
        }
    }

    // Signal 2: Auxiliary/modal/infinitive preceding (weight 0.25)
    if idx > 0 {
        prev := tokens[idx-1].Lower
        if t.verbAux[prev] || t.verbInf[prev] {
            verbScore += 0.25
            components = append(components, SignalComponent{
                Name: "verb_auxiliary", Weight: 0.25, Value: 1.0,
                Contrib: 0.25, Reason: "preceded by '" + prev + "'",
            })
        }
    }

    // Signal 3: Following token class (weight 0.15)
    if idx+1 < len(tokens) {
        next := tokens[idx+1]
        switch {
        case next.Type == TokenArticle || t.nounDet[next.Lower]:
            // Followed by article/determiner → verb signal
            verbScore += 0.15
            components = append(components, SignalComponent{
                Name: "following_class", Weight: 0.15, Value: 1.0,
                Contrib: 0.15, Reason: "followed by determiner '" + next.Lower + "'",
            })
        case next.Type == TokenNoun:
            // Followed by noun → verb signal
            verbScore += 0.15
            components = append(components, SignalComponent{
                Name: "following_class", Weight: 0.15, Value: 1.0,
                Contrib: 0.15, Reason: "followed by noun '" + next.Lower + "'",
            })
        case next.Type == TokenVerb:
            // Followed by verb → noun signal
            nounScore += 0.15
            components = append(components, SignalComponent{
                Name: "following_class", Weight: 0.15, Value: 1.0,
                Contrib: 0.15, Reason: "followed by verb '" + next.Lower + "'",
            })
        }
    }

    // Signal 4: Sentence position (weight 0.10)
    if idx == 0 {
        verbScore += 0.10
        components = append(components, SignalComponent{
            Name: "sentence_position", Weight: 0.10, Value: 1.0,
            Contrib: 0.10, Reason: "sentence-initial position",
        })
    }

    // Signal 5: Verb saturation (weight 0.10)
    hasConfidentVerb := false
    for j, tok := range tokens {
        if j == idx {
            continue
        }
        if tok.Type == TokenVerb && tok.Confidence >= 1.0 {
            hasConfidentVerb = true
            break
        }
    }
    if hasConfidentVerb {
        nounScore += 0.10
        components = append(components, SignalComponent{
            Name: "verb_saturation", Weight: 0.10, Value: 1.0,
            Contrib: 0.10, Reason: "clause already has a confident verb",
        })
    }

    // Signal 6: Inflection echo (weight 0.03)
    for j, tok := range tokens {
        if j == idx {
            continue
        }
        // If an inflected form of the same base appears, the base form
        // is likely used in the OTHER role
        if tok.Type == TokenVerb && tok.VerbInfo.Base == word && tok.VerbInfo.Tense != "base" {
            nounScore += 0.03
            components = append(components, SignalComponent{
                Name: "inflection_echo", Weight: 0.03, Value: 1.0,
                Contrib: 0.03, Reason: "inflected verb form '" + tok.Lower + "' also present",
            })
            break
        }
        if tok.Type == TokenNoun && tok.NounInfo.Base == word && tok.NounInfo.Plural {
            verbScore += 0.03
            components = append(components, SignalComponent{
                Name: "inflection_echo", Weight: 0.03, Value: 1.0,
                Contrib: 0.03, Reason: "plural noun form '" + tok.Lower + "' also present",
            })
            break
        }
    }

    // Signal 7: Default prior (weight 0.02) — verb-first tiebreaker
    verbScore += 0.02
    components = append(components, SignalComponent{
        Name: "default_prior", Weight: 0.02, Value: 1.0,
        Contrib: 0.02, Reason: "verb-first default prior",
    })

    return verbScore, nounScore, components
}

func (t *Tokeniser) resolveToken(tok *Token, verbScore, nounScore float64, components []SignalComponent) {
    total := verbScore + nounScore
    if total == 0 {
        // Shouldn't happen (default prior ensures > 0), but guard
        tok.Type = TokenVerb
        tok.Confidence = 0.5
        tok.AltType = TokenNoun
        tok.AltConf = 0.5
        return
    }

    if verbScore >= nounScore {
        tok.Type = TokenVerb
        tok.Confidence = verbScore / total
        tok.AltType = TokenNoun
        tok.AltConf = nounScore / total
    } else {
        tok.Type = TokenNoun
        tok.Confidence = nounScore / total
        tok.AltType = TokenVerb
        tok.AltConf = verbScore / total
    }

    if t.withSignals {
        tok.Signals = &SignalBreakdown{
            VerbScore:  verbScore,
            NounScore:  nounScore,
            Components: components,
        }
    }
}
```

**Step 4: Run disambiguation tests to verify they pass**

Run: `go test -run "TestTokeniser_Disambiguate" -v ./reversal/`
Expected: All PASS

**Step 5: Run full test suite**

Run: `go test ./...`
Expected: All pass. Existing tests use inflected forms or non-dual-class words, so they're unaffected.

**Step 6: Commit**

```bash
git add reversal/tokeniser.go
git commit -m "feat(reversal): implement two-pass disambiguation with 7 signals

Pass 1 classifies unambiguous tokens and marks dual-class base forms.
Pass 2 evaluates noun_determiner, verb_auxiliary, following_class,
sentence_position, verb_saturation, inflection_echo, and default_prior
signals to resolve ambiguous tokens with confidence scores."
```

---

### Task 6: Add WithSignals test coverage

**Files:**
- Modify: `reversal/tokeniser_test.go` (add signal breakdown tests)

**Step 1: Write the test**

```go
func TestTokeniser_WithSignals_Breakdown(t *testing.T) {
    setup(t)
    tok := NewTokeniser(WithSignals())

    tokens := tok.Tokenise("the commit was approved")
    // "commit" should have a SignalBreakdown
    commitTok := tokens[1]
    if commitTok.Signals == nil {
        t.Fatal("WithSignals(): commit token has nil Signals")
    }
    if commitTok.Signals.NounScore <= commitTok.Signals.VerbScore {
        t.Errorf("NounScore (%f) should exceed VerbScore (%f) for 'the commit'",
            commitTok.Signals.NounScore, commitTok.Signals.VerbScore)
    }
    if len(commitTok.Signals.Components) == 0 {
        t.Error("Components should not be empty")
    }

    // Verify noun_determiner signal fired
    foundDet := false
    for _, c := range commitTok.Signals.Components {
        if c.Name == "noun_determiner" {
            foundDet = true
            if c.Contrib != 0.35 {
                t.Errorf("noun_determiner Contrib = %f, want 0.35", c.Contrib)
            }
        }
    }
    if !foundDet {
        t.Error("noun_determiner signal should have fired")
    }
}

func TestTokeniser_WithoutSignals_NilBreakdown(t *testing.T) {
    setup(t)
    tok := NewTokeniser() // no WithSignals

    tokens := tok.Tokenise("the commit was approved")
    if tokens[1].Signals != nil {
        t.Error("Without WithSignals(), Signals should be nil")
    }
}
```

**Step 2: Run tests**

Run: `go test -run "TestTokeniser_WithSignals|TestTokeniser_WithoutSignals" -v ./reversal/`
Expected: PASS

**Step 3: Commit**

```bash
git add reversal/tokeniser_test.go
git commit -m "test(reversal): add WithSignals breakdown coverage

Verify SignalBreakdown is populated when WithSignals() is set and
nil when not. Check individual signal components fire correctly."
```

---

### Task 7: Update imprint to use confidence weighting

**Files:**
- Modify: `reversal/imprint.go:41-71` (token processing loop)

**Step 1: Write the failing test**

```go
func TestImprint_ConfidenceWeighting(t *testing.T) {
    svc, _ := i18n.New()
    i18n.SetDefault(svc)
    tok := NewTokeniser()

    // "the commit was approved" — "commit" should be noun with ~0.96 confidence
    tokens := tok.Tokenise("the commit was approved")
    imp := NewImprint(tokens)

    // Commit should contribute primarily to noun distribution
    if imp.NounDistribution["commit"] == 0 {
        t.Error("NounDistribution should contain 'commit'")
    }

    // But also fractionally to verb distribution (via AltConf)
    if imp.VerbDistribution["commit"] == 0 {
        t.Error("VerbDistribution should contain fractional 'commit' from AltConf")
    }

    // Noun contribution should be much larger than verb contribution
    // (before normalisation, noun ~0.96, verb ~0.04)
    // After normalisation ratios depend on other tokens, but noun > verb
}

func TestImprint_ConfidenceWeighting_BackwardsCompat(t *testing.T) {
    svc, _ := i18n.New()
    i18n.SetDefault(svc)
    tok := NewTokeniser()

    // Non-ambiguous tokens should work identically (Confidence=1.0, AltConf=0)
    tokens := tok.Tokenise("Deleted the files")
    imp := NewImprint(tokens)

    if imp.VerbDistribution["delete"] == 0 {
        t.Error("VerbDistribution should contain 'delete'")
    }
    if imp.NounDistribution["file"] == 0 {
        t.Error("NounDistribution should contain 'file'")
    }
}
```

**Step 2: Run tests to verify the first fails**

Run: `go test -run "TestImprint_ConfidenceWeighting" -v ./reversal/`
Expected: First test FAIL (commit goes entirely to verb, nothing to noun)

**Step 3: Update NewImprint to use confidence weighting**

Replace the token processing loop in `NewImprint`:

```go
for _, tok := range tokens {
    switch tok.Type {
    case TokenVerb:
        conf := tok.Confidence
        if conf == 0 {
            conf = 1.0
        }
        verbCount++
        base := tok.VerbInfo.Base
        imp.VerbDistribution[base] += conf
        imp.TenseDistribution[tok.VerbInfo.Tense] += conf
        verbBases[base] = true

        // Dual-class: contribute alt confidence to noun distribution
        if tok.AltType == TokenNoun && tok.NounInfo.Base != "" {
            imp.NounDistribution[tok.NounInfo.Base] += tok.AltConf
            nounBases[tok.NounInfo.Base] = true
            totalNouns++
        }

    case TokenNoun:
        conf := tok.Confidence
        if conf == 0 {
            conf = 1.0
        }
        nounCount++
        base := tok.NounInfo.Base
        imp.NounDistribution[base] += conf
        nounBases[base] = true
        totalNouns++
        if tok.NounInfo.Plural {
            pluralNouns++
        }

        // Dual-class: contribute alt confidence to verb distribution
        if tok.AltType == TokenVerb && tok.VerbInfo.Base != "" {
            imp.VerbDistribution[tok.VerbInfo.Base] += tok.AltConf
            imp.TenseDistribution[tok.VerbInfo.Tense] += tok.AltConf
            verbBases[tok.VerbInfo.Base] = true
        }

    case TokenArticle:
        articleCount++
        imp.ArticleUsage[tok.ArtType]++

    case TokenWord:
        imp.DomainVocabulary[tok.WordCat]++

    case TokenPunctuation:
        punctCount++
        imp.PunctuationPattern[tok.PunctType]++
    }
}
```

**Step 4: Run tests**

Run: `go test -run "TestImprint_ConfidenceWeighting" -v ./reversal/`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./...`
Expected: All pass. Existing imprint tests use non-ambiguous tokens so behaviour is unchanged.

**Step 6: Commit**

```bash
git add reversal/imprint.go
git commit -m "feat(reversal): confidence-weighted imprint contributions

Dual-class tokens contribute to both verb and noun distributions
weighted by Confidence and AltConf. Non-ambiguous tokens (Confidence
1.0, AltConf 0.0) behave identically to before."
```

---

### Task 8: Update multiplier for confidence compatibility

**Files:**
- Modify: `reversal/multiplier.go` (minor — ensure Confidence is preserved on transformed tokens)

**Step 1: Write the test**

```go
func TestMultiplier_Expand_DualClass(t *testing.T) {
    svc, _ := i18n.New()
    i18n.SetDefault(svc)
    m := NewMultiplier()

    // "the commit" — commit is noun, should still produce variants
    variants := m.Expand("the commit")
    if len(variants) < 2 {
        t.Errorf("Expand('the commit') returned %d variants, want >= 2", len(variants))
    }

    // Should have at least original + plural toggle
    found := false
    for _, v := range variants {
        if v == "the commits" {
            found = true
        }
    }
    if !found {
        t.Errorf("Expected 'the commits' variant, got: %v", variants)
    }
}
```

**Step 2: Run test**

Run: `go test -run TestMultiplier_Expand_DualClass -v ./reversal/`
Expected: Likely already passes since multiplier works on token types. If it fails, update `applyVerbTransform` and `applyNounTransformOnTokens` to preserve Confidence.

**Step 3: If needed, update multiplier token construction**

In `applyVerbTransform`, set Confidence on the new token:

```go
result[vi] = Token{
    Raw:        newForm,
    Lower:      strings.ToLower(newForm),
    Type:       TokenVerb,
    Confidence: 1.0, // transformed tokens are unambiguous
    VerbInfo:   VerbMatch{Base: base, Tense: targetTense, Form: newForm},
}
```

Same for `applyNounTransformOnTokens`.

**Step 4: Run full test suite**

Run: `go test ./...`
Expected: All pass

**Step 5: Commit**

```bash
git add reversal/multiplier.go
git commit -m "fix(reversal): preserve Confidence on multiplier-transformed tokens

Transformed tokens get Confidence 1.0 since the transformation
is deterministic and unambiguous."
```

---

### Task 9: Add comprehensive round-trip tests for dual-class words

**Files:**
- Modify: `reversal/roundtrip_test.go`

**Step 1: Write the tests**

```go
func TestRoundTrip_DualClassDisambiguation(t *testing.T) {
    svc, _ := i18n.New()
    i18n.SetDefault(svc)
    tok := NewTokeniser()

    tests := []struct {
        name     string
        text     string
        word     string
        wantType TokenType
    }{
        {"commit as noun", "Delete the commit", "commit", TokenNoun},
        {"commit as verb", "Commit the changes", "commit", TokenVerb},
        {"run as verb", "Run the tests", "run", TokenVerb},
        {"test as noun", "The test passed", "test", TokenNoun},
        {"build as verb", "Build the project", "build", TokenVerb},
        {"build as noun", "The build failed", "build", TokenNoun},
        {"check as noun", "The check passed", "check", TokenNoun},
        {"check as verb", "Check the logs", "check", TokenVerb},
        {"file as noun", "Delete the file", "file", TokenNoun},
        {"file as verb", "File the report", "file", TokenVerb},
        {"test as verb after aux", "will test the system", "test", TokenVerb},
        {"run as noun with possessive", "his run was fast", "run", TokenNoun},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tokens := tok.Tokenise(tt.text)
            found := false
            for _, token := range tokens {
                if token.Lower == tt.word {
                    found = true
                    if token.Type != tt.wantType {
                        t.Errorf("%q in %q: got Type %v, want %v (Confidence: %.2f)",
                            tt.word, tt.text, token.Type, tt.wantType, token.Confidence)
                    }
                }
            }
            if !found {
                t.Errorf("did not find %q in tokens from %q", tt.word, tt.text)
            }
        })
    }
}

func TestRoundTrip_DualClassImprintConvergence(t *testing.T) {
    svc, _ := i18n.New()
    i18n.SetDefault(svc)
    tok := NewTokeniser()

    // Two texts using "commit" as noun should produce similar imprints
    imp1 := NewImprint(tok.Tokenise("the commit was approved"))
    imp2 := NewImprint(tok.Tokenise("the commit was merged"))

    sim := imp1.Similar(imp2)
    if sim < 0.7 {
        t.Errorf("Same-role imprint similarity = %f, want >= 0.7", sim)
    }

    // Text using "commit" as verb should diverge more
    imp3 := NewImprint(tok.Tokenise("Commit the changes now"))
    simDiff := imp1.Similar(imp3)

    if simDiff >= sim {
        t.Errorf("Different-role similarity (%f) should be less than same-role (%f)",
            simDiff, sim)
    }
}
```

**Step 2: Run tests**

Run: `go test -run "TestRoundTrip_DualClass" -v ./reversal/`
Expected: PASS

**Step 3: Commit**

```bash
git add reversal/roundtrip_test.go
git commit -m "test(reversal): dual-class round-trip and imprint convergence tests

Verify all 6 dual-class words disambiguate correctly in both verb
and noun contexts. Verify same-role imprints converge and different-
role imprints diverge."
```

---

### Task 10: Final validation and FINDINGS.md update

**Files:**
- Modify: `FINDINGS.md`
- Modify: `TODO.md` (mark task complete)

**Step 1: Run full test suite one final time**

Run: `go test -v ./...`
Expected: All pass, zero failures

**Step 2: Run tests with race detector**

Run: `go test -race ./...`
Expected: No data races

**Step 3: Update TODO.md**

Mark the dual-class ambiguity task as done with commit hash.

**Step 4: Update FINDINGS.md**

Add a new section documenting the disambiguation design, signal weights, and test results.

**Step 5: Commit**

```bash
git add FINDINGS.md TODO.md
git commit -m "docs: mark dual-class disambiguation complete

Update TODO.md and FINDINGS.md with implementation details,
signal weight table, and test coverage summary."
```
