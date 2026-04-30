package i18n

func ExampleServiceState_HandlerTypeNames() {
	var v ServiceState
	_ = v.HandlerTypeNames
}

func ExampleServiceState_String() {
	var v ServiceState
	_ = v.String
}

func ExampleService_State() {
	v := &Service{}
	_ = v.State
}

func ExampleService_String() {
	v := &Service{}
	_ = v.String
}

func ExampleService_CurrentState() {
	v := &Service{}
	_ = v.CurrentState
}
