package i18n

func ExampleC() {
	_ = C
}

func ExampleTranslationContext_WithGender() {
	v := &TranslationContext{}
	_ = v.WithGender
}

func ExampleTranslationContext_In() {
	v := &TranslationContext{}
	_ = v.In
}

func ExampleTranslationContext_Formal() {
	v := &TranslationContext{}
	_ = v.Formal
}

func ExampleTranslationContext_Informal() {
	v := &TranslationContext{}
	_ = v.Informal
}

func ExampleTranslationContext_WithFormality() {
	v := &TranslationContext{}
	_ = v.WithFormality
}

func ExampleTranslationContext_Count() {
	v := &TranslationContext{}
	_ = v.Count
}

func ExampleTranslationContext_Set() {
	v := &TranslationContext{}
	_ = v.Set
}

func ExampleTranslationContext_Get() {
	v := &TranslationContext{}
	_ = v.Get
}

func ExampleTranslationContext_ContextString() {
	v := &TranslationContext{}
	_ = v.ContextString
}

func ExampleTranslationContext_String() {
	v := &TranslationContext{}
	_ = v.String
}

func ExampleTranslationContext_GenderString() {
	v := &TranslationContext{}
	_ = v.GenderString
}

func ExampleTranslationContext_LocationString() {
	v := &TranslationContext{}
	_ = v.LocationString
}

func ExampleTranslationContext_FormalityValue() {
	v := &TranslationContext{}
	_ = v.FormalityValue
}

func ExampleTranslationContext_CountInt() {
	v := &TranslationContext{}
	_ = v.CountInt
}

func ExampleTranslationContext_CountString() {
	v := &TranslationContext{}
	_ = v.CountString
}

func ExampleTranslationContext_IsPlural() {
	v := &TranslationContext{}
	_ = v.IsPlural
}
