package i18n

func ExampleS() {
	_ = S
}

func ExampleComposeIntent() {
	_ = ComposeIntent
}

func ExampleIntent_Compose() {
	var v Intent
	_ = v.Compose
}

func ExampleSubject_Count() {
	v := &Subject{}
	_ = v.Count
}

func ExampleSubject_Gender() {
	v := &Subject{}
	_ = v.Gender
}

func ExampleSubject_In() {
	v := &Subject{}
	_ = v.In
}

func ExampleSubject_Formal() {
	v := &Subject{}
	_ = v.Formal
}

func ExampleSubject_Informal() {
	v := &Subject{}
	_ = v.Informal
}

func ExampleSubject_SetFormality() {
	v := &Subject{}
	_ = v.SetFormality
}

func ExampleSubject_String() {
	v := &Subject{}
	_ = v.String
}

func ExampleSubject_IsPlural() {
	v := &Subject{}
	_ = v.IsPlural
}

func ExampleSubject_CountInt() {
	v := &Subject{}
	_ = v.CountInt
}

func ExampleSubject_CountString() {
	v := &Subject{}
	_ = v.CountString
}

func ExampleSubject_GenderString() {
	v := &Subject{}
	_ = v.GenderString
}

func ExampleSubject_LocationString() {
	v := &Subject{}
	_ = v.LocationString
}

func ExampleSubject_NounString() {
	v := &Subject{}
	_ = v.NounString
}

func ExampleSubject_FormalityString() {
	v := &Subject{}
	_ = v.FormalityString
}

func ExampleSubject_IsFormal() {
	v := &Subject{}
	_ = v.IsFormal
}

func ExampleSubject_IsInformal() {
	v := &Subject{}
	_ = v.IsInformal
}
