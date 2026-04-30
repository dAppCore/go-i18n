package i18n

func ExampleMode_String() {
	var v Mode
	_ = v.String
}

func ExampleMessage_ForCategory() {
	var v Message
	_ = v.ForCategory
}

func ExampleMessage_IsPlural() {
	var v Message
	_ = v.IsPlural
}
