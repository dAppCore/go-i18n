package i18n

func ExampleSetDebug() {
	_ = SetDebug
}

func ExampleService_SetDebug() {
	v := &Service{}
	_ = v.SetDebug
}

func ExampleService_Debug() {
	v := &Service{}
	_ = v.Debug
}
