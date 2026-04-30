package i18n

func ExampleLabelHandler_Match() {
	var v LabelHandler
	_ = v.Match
}

func ExampleLabelHandler_Handle() {
	var v LabelHandler
	_ = v.Handle
}

func ExampleProgressHandler_Match() {
	var v ProgressHandler
	_ = v.Match
}

func ExampleProgressHandler_Handle() {
	var v ProgressHandler
	_ = v.Handle
}

func ExampleCountHandler_Match() {
	var v CountHandler
	_ = v.Match
}

func ExampleCountHandler_Handle() {
	var v CountHandler
	_ = v.Handle
}

func ExampleDoneHandler_Match() {
	var v DoneHandler
	_ = v.Match
}

func ExampleDoneHandler_Handle() {
	var v DoneHandler
	_ = v.Handle
}

func ExampleFailHandler_Match() {
	var v FailHandler
	_ = v.Match
}

func ExampleFailHandler_Handle() {
	var v FailHandler
	_ = v.Handle
}

func ExampleNumericHandler_Match() {
	var v NumericHandler
	_ = v.Match
}

func ExampleNumericHandler_Handle() {
	var v NumericHandler
	_ = v.Handle
}

func ExampleDefaultHandlers() {
	_ = DefaultHandlers
}

func ExampleRunHandlerChain() {
	_ = RunHandlerChain
}
