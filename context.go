package i18n

// TranslationContext provides disambiguation for translations.
//
//	T("direction.right", C("navigation")) // "rechts" (German)
//	T("status.right", C("correctness"))   // "richtig" (German)
type TranslationContext struct {
	Context   string
	Gender    string
	Location  string
	Formality Formality
	count     int
	countSet  bool
	Extra     map[string]any
}

// C creates a TranslationContext.
func C(context string) *TranslationContext {
	return &TranslationContext{Context: context, count: 1}
}

// WithGender annotates the context with a grammatical gender for languages
// that inflect on it. Returns the receiver for chaining.
//
//	C("greeting").WithGender("f")
func (c *TranslationContext) WithGender(gender string) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Gender = gender
	return c
}

// In annotates a location/scope on the context; chains.
//
//	C("status").In("dashboard")
func (c *TranslationContext) In(location string) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Location = location
	return c
}

// Formal sets the context's formality to formal and returns the receiver.
//
//	C("greeting").Formal()
func (c *TranslationContext) Formal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityFormal
	return c
}

// Informal sets the context's formality to informal and returns the receiver.
//
//	C("greeting").Informal()
func (c *TranslationContext) Informal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityInformal
	return c
}

// WithFormality assigns the context's formality (formal/informal/neutral)
// and returns the receiver.
//
//	C("greeting").WithFormality(FormalityFormal)
func (c *TranslationContext) WithFormality(f Formality) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = f
	return c
}

// Count sets the count used for plural-sensitive translations.
func (c *TranslationContext) Count(n int) *TranslationContext {
	if c == nil {
		return nil
	}
	c.count = n
	c.countSet = true
	return c
}

// Set stores a key-value pair in the context's Extra map for use by
// templates that read named fields. Chains.
//
//	C("notify").Set("user_name", "Alice")
func (c *TranslationContext) Set(key string, value any) *TranslationContext {
	if c == nil {
		return nil
	}
	if c.Extra == nil {
		c.Extra = make(map[string]any)
	}
	c.Extra[key] = value
	return c
}

// Get retrieves a previously-Set value by key. Returns nil when unset or
// when called on a nil context.
//
//	C("notify").Set("k", "v").Get("k") // → "v"
func (c *TranslationContext) Get(key string) any {
	if c == nil || c.Extra == nil {
		return nil
	}
	return c.Extra[key]
}

// ContextString returns the disambiguation context label.
//
//	C("navigation").ContextString() // → "navigation"
func (c *TranslationContext) ContextString() string {
	if c == nil {
		return ""
	}
	return c.Context
}

// String returns the context label (alias for ContextString) so the type
// satisfies fmt.Stringer.
//
//	core.Sprintf("%s", C("navigation"))
func (c *TranslationContext) String() string {
	if c == nil {
		return ""
	}
	return c.Context
}

// GenderString returns the context's gender annotation, empty when unset.
//
//	C("greeting").WithGender("f").GenderString() // → "f"
func (c *TranslationContext) GenderString() string {
	if c == nil {
		return ""
	}
	return c.Gender
}

// LocationString returns the context's location annotation, empty when
// unset.
//
//	C("status").In("dashboard").LocationString() // → "dashboard"
func (c *TranslationContext) LocationString() string {
	if c == nil {
		return ""
	}
	return c.Location
}

// FormalityValue returns the context's formality (defaulting to
// FormalityNeutral on a nil context).
//
//	C("greeting").Formal().FormalityValue() // → FormalityFormal
func (c *TranslationContext) FormalityValue() Formality {
	if c == nil {
		return FormalityNeutral
	}
	return c.Formality
}

// CountInt returns the current count value.
func (c *TranslationContext) CountInt() int {
	if c == nil {
		return 1
	}
	return c.count
}

// CountString returns the current count value formatted as text.
func (c *TranslationContext) CountString() string {
	if c == nil {
		return "1"
	}
	return FormatNumber(int64(c.count))
}

// IsPlural reports whether the count is plural.
func (c *TranslationContext) IsPlural() bool {
	return c != nil && c.count != 1
}

func (c *TranslationContext) countValue() (int, bool) {
	if c == nil {
		return 1, false
	}
	return c.count, c.countSet
}
