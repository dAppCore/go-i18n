package i18n

import "fmt"

// S creates a new Subject with the given noun and value.
//
//	S("file", "config.yaml")
//	S("file", path).Count(3).In("workspace")
func S(noun string, value any) *Subject {
	return &Subject{Noun: noun, Value: value, count: 1}
}

func (s *Subject) Count(n int) *Subject {
	if s == nil {
		return nil
	}
	s.count = n
	return s
}

func (s *Subject) Gender(g string) *Subject {
	if s == nil {
		return nil
	}
	s.gender = g
	return s
}

func (s *Subject) In(location string) *Subject {
	if s == nil {
		return nil
	}
	s.location = location
	return s
}

func (s *Subject) Formal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityFormal
	return s
}

func (s *Subject) Informal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityInformal
	return s
}

func (s *Subject) SetFormality(f Formality) *Subject {
	if s == nil {
		return nil
	}
	s.formality = f
	return s
}

func (s *Subject) String() string {
	if s == nil {
		return ""
	}
	if stringer, ok := s.Value.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprint(s.Value)
}

func (s *Subject) IsPlural() bool  { return s != nil && s.count != 1 }
func (s *Subject) CountInt() int   { if s == nil { return 1 }; return s.count }
func (s *Subject) CountString() string { if s == nil { return "1" }; return fmt.Sprint(s.count) }
func (s *Subject) GenderString() string { if s == nil { return "" }; return s.gender }
func (s *Subject) LocationString() string { if s == nil { return "" }; return s.location }
func (s *Subject) NounString() string { if s == nil { return "" }; return s.Noun }
func (s *Subject) FormalityString() string {
	if s == nil {
		return FormalityNeutral.String()
	}
	return s.formality.String()
}
func (s *Subject) IsFormal() bool   { return s != nil && s.formality == FormalityFormal }
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
