package i18n

func ExampleNewFSLoader() {
	_ = NewFSLoader
}

func ExampleFSLoader_Load() {
	v := &FSLoader{}
	_ = v.Load
}

func ExampleFSLoader_Languages() {
	v := &FSLoader{}
	_ = v.Languages
}

func ExampleFSLoader_LanguagesErr() {
	v := &FSLoader{}
	_ = v.LanguagesErr
}
