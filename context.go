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

func (c *TranslationContext) WithGender(gender string) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Gender = gender
	return c
}

func (c *TranslationContext) In(location string) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Location = location
	return c
}

func (c *TranslationContext) Formal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityFormal
	return c
}

func (c *TranslationContext) Informal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityInformal
	return c
}

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

func (c *TranslationContext) Get(key string) any {
	if c == nil || c.Extra == nil {
		return nil
	}
	return c.Extra[key]
}

func (c *TranslationContext) ContextString() string {
	if c == nil {
		return ""
	}
	return c.Context
}

func (c *TranslationContext) String() string {
	if c == nil {
		return ""
	}
	return c.Context
}

func (c *TranslationContext) GenderString() string {
	if c == nil {
		return ""
	}
	return c.Gender
}

func (c *TranslationContext) LocationString() string {
	if c == nil {
		return ""
	}
	return c.Location
}

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
