package i18n

func ExampleWithFallback() {
	_ = WithFallback
}

func ExampleWithLanguage() {
	_ = WithLanguage
}

func ExampleWithFormality() {
	_ = WithFormality
}

func ExampleWithLocation() {
	_ = WithLocation
}

func ExampleWithHandlers() {
	_ = WithHandlers
}

func ExampleWithDefaultHandlers() {
	_ = WithDefaultHandlers
}

func ExampleWithMode() {
	_ = WithMode
}

func ExampleWithDebug() {
	_ = WithDebug
}

func ExampleNew() {
	_ = New
}

func ExampleNewService() {
	_ = NewService
}

func ExampleNewWithFS() {
	_ = NewWithFS
}

func ExampleNewServiceWithFS() {
	_ = NewServiceWithFS
}

func ExampleNewWithLoader() {
	_ = NewWithLoader
}

func ExampleNewServiceWithLoader() {
	_ = NewServiceWithLoader
}

func ExampleInit() {
	_ = Init
}

func ExampleDefault() {
	_ = Default
}

func ExampleSetDefault() {
	_ = SetDefault
}

func ExampleAddLoader() {
	_ = AddLoader
}

func ExampleService_SetLanguage() {
	v := &Service{}
	_ = v.SetLanguage
}

func ExampleService_Language() {
	v := &Service{}
	_ = v.Language
}

func ExampleService_CurrentLanguage() {
	v := &Service{}
	_ = v.CurrentLanguage
}

func ExampleService_CurrentLang() {
	v := &Service{}
	_ = v.CurrentLang
}

func ExampleService_Prompt() {
	v := &Service{}
	_ = v.Prompt
}

func ExampleService_CurrentPrompt() {
	v := &Service{}
	_ = v.CurrentPrompt
}

func ExampleService_Lang() {
	v := &Service{}
	_ = v.Lang
}

func ExampleService_AvailableLanguages() {
	v := &Service{}
	_ = v.AvailableLanguages
}

func ExampleService_CurrentAvailableLanguages() {
	v := &Service{}
	_ = v.CurrentAvailableLanguages
}

func ExampleService_SetMode() {
	v := &Service{}
	_ = v.SetMode
}

func ExampleService_Mode() {
	v := &Service{}
	_ = v.Mode
}

func ExampleService_CurrentMode() {
	v := &Service{}
	_ = v.CurrentMode
}

func ExampleService_SetFormality() {
	v := &Service{}
	_ = v.SetFormality
}

func ExampleService_Formality() {
	v := &Service{}
	_ = v.Formality
}

func ExampleService_CurrentFormality() {
	v := &Service{}
	_ = v.CurrentFormality
}

func ExampleService_SetFallback() {
	v := &Service{}
	_ = v.SetFallback
}

func ExampleService_Fallback() {
	v := &Service{}
	_ = v.Fallback
}

func ExampleService_CurrentFallback() {
	v := &Service{}
	_ = v.CurrentFallback
}

func ExampleService_SetLocation() {
	v := &Service{}
	_ = v.SetLocation
}

func ExampleService_Location() {
	v := &Service{}
	_ = v.Location
}

func ExampleService_CurrentLocation() {
	v := &Service{}
	_ = v.CurrentLocation
}

func ExampleService_Direction() {
	v := &Service{}
	_ = v.Direction
}

func ExampleService_CurrentDirection() {
	v := &Service{}
	_ = v.CurrentDirection
}

func ExampleService_CurrentTextDirection() {
	v := &Service{}
	_ = v.CurrentTextDirection
}

func ExampleService_IsRTL() {
	v := &Service{}
	_ = v.IsRTL
}

func ExampleService_CurrentIsRTL() {
	v := &Service{}
	_ = v.CurrentIsRTL
}

func ExampleService_RTL() {
	v := &Service{}
	_ = v.RTL
}

func ExampleService_CurrentRTL() {
	v := &Service{}
	_ = v.CurrentRTL
}

func ExampleService_CurrentDebug() {
	v := &Service{}
	_ = v.CurrentDebug
}

func ExampleService_PluralCategory() {
	v := &Service{}
	_ = v.PluralCategory
}

func ExampleService_CurrentPluralCategory() {
	v := &Service{}
	_ = v.CurrentPluralCategory
}

func ExampleService_PluralCategoryOf() {
	v := &Service{}
	_ = v.PluralCategoryOf
}

func ExampleService_AddHandler() {
	v := &Service{}
	_ = v.AddHandler
}

func ExampleService_SetHandlers() {
	v := &Service{}
	_ = v.SetHandlers
}

func ExampleService_PrependHandler() {
	v := &Service{}
	_ = v.PrependHandler
}

func ExampleService_ClearHandlers() {
	v := &Service{}
	_ = v.ClearHandlers
}

func ExampleService_ResetHandlers() {
	v := &Service{}
	_ = v.ResetHandlers
}

func ExampleService_Handlers() {
	v := &Service{}
	_ = v.Handlers
}

func ExampleService_CurrentHandlers() {
	v := &Service{}
	_ = v.CurrentHandlers
}

func ExampleService_T() {
	v := &Service{}
	_ = v.T
}

func ExampleService_Compose() {
	v := &Service{}
	_ = v.Compose
}

func ExampleService_CurrentCompose() {
	v := &Service{}
	_ = v.CurrentCompose
}

func ExampleService_Translate() {
	v := &Service{}
	_ = v.Translate
}

func ExampleService_Raw() {
	v := &Service{}
	_ = v.Raw
}

func ExampleService_AddMessages() {
	v := &Service{}
	_ = v.AddMessages
}

func ExampleService_AddLoader() {
	v := &Service{}
	_ = v.AddLoader
}

func ExampleService_LoadFS() {
	v := &Service{}
	_ = v.LoadFS
}
