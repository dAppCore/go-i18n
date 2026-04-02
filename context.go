package i18n

import "fmt"

// TranslationContext provides disambiguation for translations.
//
//	T("direction.right", C("navigation")) // "rechts" (German)
//	T("status.right", C("correctness"))   // "richtig" (German)
type TranslationContext struct {
	Context   string
	Gender    string
	Location  string
	Formality Formality
	Extra     map[string]any
}

// C creates a TranslationContext.
func C(context string) *TranslationContext {
	return &TranslationContext{Context: context}
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
	return fmt.Sprint(c.Context)
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
