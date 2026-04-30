package i18n

func ExampleFormality_String() {
	var v Formality
	_ = v.String
}

func ExampleTextDirection_String() {
	var v TextDirection
	_ = v.String
}

func ExamplePluralCategory_String() {
	var v PluralCategory
	_ = v.String
}

func ExampleGrammaticalGender_String() {
	var v GrammaticalGender
	_ = v.String
}

func ExampleIsRTLLanguage() {
	_ = IsRTLLanguage
}

func ExampleSetFormality() {
	_ = SetFormality
}

func ExampleSetLocation() {
	_ = SetLocation
}

func ExampleCurrentLocation() {
	_ = CurrentLocation
}

func ExampleLocation() {
	_ = Location
}

func ExampleDirection() {
	_ = Direction
}

func ExampleCurrentDirection() {
	_ = CurrentDirection
}

func ExampleCurrentTextDirection() {
	_ = CurrentTextDirection
}

func ExampleIsRTL() {
	_ = IsRTL
}

func ExampleRTL() {
	_ = RTL
}

func ExampleCurrentIsRTL() {
	_ = CurrentIsRTL
}

func ExampleCurrentRTL() {
	_ = CurrentRTL
}

func ExampleCurrentPluralCategory() {
	_ = CurrentPluralCategory
}

func ExamplePluralCategoryOf() {
	_ = PluralCategoryOf
}
