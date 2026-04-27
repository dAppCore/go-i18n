package i18n

import (
	"dappco.re/go/core"
)

// stringer mirrors fmt.Stringer without pulling in the banned fmt package.
type stringer interface {
	String() string
}

// S creates a new Subject with the given noun and value.
//
//	S("file", "config.yaml")
//	S("file", path).Count(3).In("workspace")
func S(noun string, value any) *Subject {
	return &Subject{Noun: noun, Value: value, count: 1}
}

// ComposeIntent renders an intent's templates into concrete output.
func ComposeIntent(intent Intent, subject *Subject) Composed {
	return intent.Compose(subject)
}

// Compose renders an intent's templates into concrete output.
func (i Intent) Compose(subject *Subject) Composed {
	data := newTemplateData(subject)
	return Composed{
		Question: executeIntentTemplate(i.Question, data),
		Confirm:  executeIntentTemplate(i.Confirm, data),
		Success:  executeIntentTemplate(i.Success, data),
		Failure:  executeIntentTemplate(i.Failure, data),
		Meta:     i.Meta,
	}
}

// Count sets the subject's plural count and returns the receiver for
// chaining.
//
//	S("file", path).Count(3) // 3 files
func (s *Subject) Count(n int) *Subject {
	if s == nil {
		return nil
	}
	s.count = n
	return s
}

// Gender annotates the subject's grammatical gender for languages that
// inflect on it. Returns the receiver for chaining.
//
//	S("client", name).Gender("f")
func (s *Subject) Gender(g string) *Subject {
	if s == nil {
		return nil
	}
	s.gender = g
	return s
}

// In annotates a location/scope for the subject; chains.
//
//	S("file", path).In("workspace")
func (s *Subject) In(location string) *Subject {
	if s == nil {
		return nil
	}
	s.location = location
	return s
}

// Formal sets the subject's formality to formal and returns the receiver.
//
//	S("user", u).Formal()
func (s *Subject) Formal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityFormal
	return s
}

// Informal sets the subject's formality to informal and returns the receiver.
//
//	S("user", u).Informal()
func (s *Subject) Informal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityInformal
	return s
}

// SetFormality assigns the subject's formality (formal/informal/neutral)
// and returns the receiver.
//
//	S("user", u).SetFormality(FormalityFormal)
func (s *Subject) SetFormality(f Formality) *Subject {
	if s == nil {
		return nil
	}
	s.formality = f
	return s
}

// String renders the subject's Value via fmt.Stringer when available, or via
// core.Sprintf("%v", ...) otherwise. Nil subjects render as the empty string.
//
//	S("file", path).String()
func (s *Subject) String() string {
	if s == nil {
		return ""
	}
	if s, ok := s.Value.(stringer); ok {
		return s.String()
	}
	return core.Sprintf("%v", s.Value)
}

// IsPlural reports whether the subject's count is anything other than 1.
//
//	S("file", path).Count(3).IsPlural() // → true
func (s *Subject) IsPlural() bool { return s != nil && s.count != 1 }

// CountInt returns the subject's count as int (defaulting to 1 for nil).
//
//	S("file", path).Count(3).CountInt() // → 3
func (s *Subject) CountInt() int {
	if s == nil {
		return 1
	}
	return s.count
}

// CountString returns the subject's count formatted via FormatNumber.
//
//	S("file", path).Count(1234).CountString() // → "1,234"
func (s *Subject) CountString() string {
	if s == nil {
		return "1"
	}
	return FormatNumber(int64(s.count))
}

// GenderString returns the subject's gender annotation, empty when unset.
//
//	S("client", c).Gender("f").GenderString() // → "f"
func (s *Subject) GenderString() string {
	if s == nil {
		return ""
	}
	return s.gender
}

// LocationString returns the subject's location annotation, empty when unset.
//
//	S("file", path).In("workspace").LocationString() // → "workspace"
func (s *Subject) LocationString() string {
	if s == nil {
		return ""
	}
	return s.location
}

// NounString returns the subject's noun keyword.
//
//	S("file", path).NounString() // → "file"
func (s *Subject) NounString() string {
	if s == nil {
		return ""
	}
	return s.Noun
}

// FormalityString returns the subject's formality as a string label.
//
//	S("user", u).Formal().FormalityString() // → "formal"
func (s *Subject) FormalityString() string {
	if s == nil {
		return FormalityNeutral.String()
	}
	return s.formality.String()
}

// IsFormal reports whether the subject is marked formal.
//
//	S("user", u).Formal().IsFormal() // → true
func (s *Subject) IsFormal() bool { return s != nil && s.formality == FormalityFormal }

// IsInformal reports whether the subject is marked informal.
//
//	S("user", u).Informal().IsInformal() // → true
func (s *Subject) IsInformal() bool { return s != nil && s.formality == FormalityInformal }

func newTemplateData(s *Subject) templateData {
	if s == nil {
		return templateData{Count: 1}
	}
	return templateData{
		Subject:   s.String(),
		Noun:      s.Noun,
		Count:     s.count,
		Gender:    s.gender,
		Location:  s.location,
		Formality: s.formality,
		IsFormal:  s.formality == FormalityFormal,
		IsPlural:  s.count != 1,
		Value:     s.Value,
	}
}
