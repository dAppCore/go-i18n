package i18n

import (
	"reflect"
	"testing"
	"testing/fstest"

	"dappco.re/go"
)

// --- Package-level T() ---

func TestT_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	got := T("prompt.yes")
	if "y" != got {
		t.Fatalf("want %v, got %v", "y", got)
	}
}

func TestT_Good_WithHandlers(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	got := T("i18n.label.status")
	if "Status:" != got {
		t.Fatalf("want %v, got %v", "Status:", got)
	}
}

func TestT_Good_MissingKey(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	got := T("nonexistent.key.test")
	if "nonexistent.key.test" != got {
		t.Fatalf("want %v, got %v", "nonexistent.key.test", got)
	}
}

func TestTranslate_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	result := Translate("prompt.yes")
	if !(result.OK) {
		t.Fatal("expected true")
	}
	if "y" != result.Value {
		t.Fatalf("want %v, got %v", "y", result.Value)
	}
}

func TestTranslate_Good_MissingKey(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	result := Translate("nonexistent.translation.key")
	if result.OK {
		t.Fatal("expected false")
	}
	if "nonexistent.translation.key" != result.Error() {
		t.Fatalf("want %v, got %v", "nonexistent.translation.key", result.Error())
	}
}

func TestTranslate_Good_SameTextAsKey(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	AddMessages("en", map[string]string{
		"exact.same.key": "exact.same.key",
	})

	result := Translate("exact.same.key")
	if !(result.OK) {
		t.Fatal("expected true")
	}
	if "exact.same.key" != result.Value {
		t.Fatalf("want %v, got %v", "exact.same.key", result.Value)
	}
}

func TestRaw_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	got := Raw("prompt.yes")
	if "y" != got {
		t.Fatalf("want %v, got %v", "y", got)
	}
}

func TestRaw_Good_BypassesHandlers(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	// i18n.label.status is not in messages, handlers don't apply in Raw
	got := Raw("i18n.label.status")
	if "i18n.label.status" != got {
		t.Fatalf("want %v, got %v", "i18n.label.status", got)
	}
}

func TestLoadFS_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	fsys := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{
			Data: []byte(`{"loadfs.key": "loaded via package helper"}`),
		},
	}

	LoadFS(fsys, "locales")

	got := T("loadfs.key")
	if "loaded via package helper" != got {
		t.Fatalf("want %v, got %v", "loaded via package helper", got)
	}
}

func TestAddMessages_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	AddMessages("en", map[string]string{
		"add.messages.key": "loaded via package helper",
	})

	got := T("add.messages.key")
	if "loaded via package helper" != got {
		t.Fatalf("want %v, got %v", "loaded via package helper", got)
	}
}

func TestSetLanguage_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	err = errorFromResult(SetLanguage("en"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !core.Contains(CurrentLanguage(), "en") {
		t.Fatalf("expected %q to contain %q", CurrentLanguage(), "en")
	}
}

func TestSetLanguage_Good_UnderscoreTag(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	err = errorFromResult(SetLanguage("fr_CA"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !(len(CurrentLanguage()) >= 2) {
		t.Fatal("expected true")
	}
	if "fr" != CurrentLanguage()[:2] {
		t.Fatalf("want %v, got %v", "fr", CurrentLanguage()[:2])
	}
}

func TestLanguage_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)
	if CurrentLanguage() != Language() {
		t.Fatalf("want %v, got %v", CurrentLanguage(), Language())
	}
}

func TestSetLanguage_Bad_Unsupported(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	_ = SetLanguage("xx")
}

func TestCurrentLanguage_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	lang := CurrentLanguage()
	if len(lang) == 0 {
		t.Fatalf("expected non-empty")
	}
	if lang != CurrentLang() {
		t.Fatalf("want %v, got %v", lang, CurrentLang())
	}
}

func TestAvailableLanguages_Good(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	langs := AvailableLanguages()
	if len(langs) == 0 {
		t.Fatalf("expected non-empty")
	}
	if !reflect.DeepEqual(svc.AvailableLanguages(), langs) {
		t.Fatalf("want %v, got %v", svc.AvailableLanguages(), langs)
	}

	langs[0] = "zz"
	if "zz" == svc.AvailableLanguages()[0] {
		t.Fatalf("did not expect %v", svc.AvailableLanguages()[0])
	}
}

func TestCurrentAvailableLanguages_Good(t *testing.T) {
	prev := Default()
	t.Cleanup(func() {
		SetDefault(prev)
	})

	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	langs := CurrentAvailableLanguages()
	if len(langs) == 0 {
		t.Fatalf("expected non-empty")
	}
	if !reflect.DeepEqual(svc.AvailableLanguages(), langs) {
		t.Fatalf("want %v, got %v", svc.AvailableLanguages(), langs)
	}
}

func TestFallback_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if "en" != Fallback() {
		t.Fatalf("want %v, got %v", "en", Fallback())
	}

	SetFallback("fr")
	if "fr" != Fallback() {
		t.Fatalf("want %v, got %v", "fr", Fallback())
	}
}

func TestDebug_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)
	if Debug() {
		t.Fatal("expected false")
	}

	SetDebug(true)
	if !(Debug()) {
		t.Fatal("expected true")
	}
}

func TestCurrentState_Good(t *testing.T) {
	svc, err := serviceFromResult(NewWithLoader(messageBaseFallbackLoader{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	state := CurrentState()
	if svc.Language() != state.Language {
		t.Fatalf("want %v, got %v", svc.Language(), state.Language)
	}
	if !reflect.DeepEqual(svc.AvailableLanguages(), state.AvailableLanguages) {
		t.Fatalf("want %v, got %v", svc.AvailableLanguages(), state.AvailableLanguages)
	}
	if svc.Mode() != state.Mode {
		t.Fatalf("want %v, got %v", svc.Mode(), state.Mode)
	}
	if svc.Fallback() != state.Fallback {
		t.Fatalf("want %v, got %v", svc.Fallback(), state.Fallback)
	}
	if svc.Formality() != state.Formality {
		t.Fatalf("want %v, got %v", svc.Formality(), state.Formality)
	}
	if svc.Location() != state.Location {
		t.Fatalf("want %v, got %v", svc.Location(), state.Location)
	}
	if svc.Direction() != state.Direction {
		t.Fatalf("want %v, got %v", svc.Direction(), state.Direction)
	}
	if svc.IsRTL() != state.IsRTL {
		t.Fatalf("want %v, got %v", svc.IsRTL(), state.IsRTL)
	}
	if svc.Debug() != state.Debug {
		t.Fatalf("want %v, got %v", svc.Debug(), state.Debug)
	}
	if len(state.Handlers) != len(svc.Handlers()) {
		t.Fatalf("expected length %v, got %v", len(svc.Handlers()), state.Handlers)
	}

	state.AvailableLanguages[0] = "zz"
	if "zz" == CurrentState().AvailableLanguages[0] {
		t.Fatalf("did not expect %v", CurrentState().AvailableLanguages[0])
	}
	state.Handlers[0] = nil
	if CurrentState().Handlers[0] == nil {
		t.Fatalf("expected non-nil")
	}
}

func TestState_Good_WithoutDefaultService(t *testing.T) {
	var svc *Service
	state := svc.State()
	if !reflect.DeepEqual(defaultServiceStateSnapshot(), state) {
		t.Fatalf("want %v, got %v", defaultServiceStateSnapshot(), state)
	}
}

func TestSetMode_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	SetMode(ModeCollect)
	if ModeCollect != CurrentMode() {
		t.Fatalf("want %v, got %v", ModeCollect, CurrentMode())
	}

	SetMode(ModeNormal)
	if ModeNormal != CurrentMode() {
		t.Fatalf("want %v, got %v", ModeNormal, CurrentMode())
	}
}

func TestCurrentMode_Good_Default(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)
	if ModeNormal != CurrentMode() {
		t.Fatalf("want %v, got %v", ModeNormal, CurrentMode())
	}
}

func TestN_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	tests := []struct {
		name   string
		format string
		value  any
		args   []any
		want   string
	}{
		{"number", "number", int64(1234567), nil, "1,234,567"},
		{"percent", "percent", 0.85, nil, "85%"},
		{byteUnitName(), byteUnitName(), int64(1536000), nil, "1.46 MB"},
		{"ordinal", "ordinal", 1, nil, "1st"},
		{"ago", "ago", 5, []any{"minutes"}, "5 minutes ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := N(tt.format, tt.value, tt.args...)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestN_Good_WithoutDefaultService(t *testing.T) {
	prev := Default()
	SetDefault(nil)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	tests := []struct {
		name   string
		format string
		value  any
		args   []any
		want   string
	}{
		{"number", "number", int64(1234567), nil, "1,234,567"},
		{"percent", "percent", 0.85, nil, "85%"},
		{byteUnitName(), byteUnitName(), int64(1536000), nil, "1.46 MB"},
		{"ordinal", "ordinal", 1, nil, "1st"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := N(tt.format, tt.value, tt.args...)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

// --- Prompt() prompt shorthand ---

func TestPrompt_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"yes", "yes", "y"},
		{"yes_trimmed", " yes ", "y"},
		{"yes_prefixed", "prompt.yes", "y"},
		{"confirm", "confirm", "Are you sure?"},
		{"confirm_prefixed", "prompt.confirm", "Are you sure?"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Prompt(tt.key)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCurrentPrompt_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)
	if Prompt("confirm") != CurrentPrompt("confirm") {
		t.Fatalf("want %v, got %v", Prompt("confirm"), CurrentPrompt("confirm"))
	}
}

func TestLang_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"de", "de", "German"},
		{"fr", "fr", "French"},
		{"fr_ca", "fr_CA", "French"},
		{"fr_prefixed", "lang.fr", "French"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Lang(tt.key)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestLang_MissingKeyHandler_FiresOnce(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetMissingKeyHandlers()
		SetMode(ModeNormal)
		SetDefault(prev)
	})

	SetMode(ModeCollect)
	calls := 0
	SetMissingKeyHandlers(func(MissingKey) {
		calls++
	})

	got := Lang("zz")
	if "[lang.zz]" != got {
		t.Fatalf("want %v, got %v", "[lang.zz]", got)
	}
	if 1 != calls {
		t.Fatalf("want %v, got %v", 1, calls)
	}
}

func TestAddHandler_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	initialCount := len(svc.Handlers())

	AddHandler(LabelHandler{})
	if initialCount+1 != len(svc.Handlers()) {
		t.Fatalf("want %v, got %v", initialCount+1, len(svc.Handlers()))
	}
}

func TestAddHandler_Good_Variadic(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	AddHandler(LabelHandler{}, ProgressHandler{})
	handlers := svc.Handlers()
	if 2 != len(handlers) {
		t.Fatalf("want %v, got %v", 2, len(handlers))
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
	if reflect.TypeOf(handlers[1]) != reflect.TypeOf(ProgressHandler{}) {
		t.Fatalf("expected type %T, got %T", ProgressHandler{}, handlers[1])
	}
}

func TestAddHandler_Good_SkipsNil(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	var nilHandler KeyHandler
	AddHandler(nilHandler, LabelHandler{})

	handlers := svc.Handlers()
	if len(handlers) != 1 {
		t.Fatalf("expected length %v, got %v", 1, handlers)
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
}

func TestAddHandler_DoesNotMutateInputSlice(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	handlers := []KeyHandler{nil, LabelHandler{}}
	AddHandler(handlers...)
	if handlers[0] != nil {
		t.Fatalf("expected nil, got %v", handlers[0])
	}
	if reflect.TypeOf(handlers[1]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[1])
	}
}

func TestPrependHandler_Good(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers())) // start with no handlers
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	PrependHandler(LabelHandler{})
	if 1 != len(svc.Handlers()) {
		t.Fatalf("want %v, got %v", 1, len(svc.Handlers()))
	}

	PrependHandler(ProgressHandler{})
	handlers := svc.Handlers()
	if 2 != len(handlers) {
		t.Fatalf("want %v, got %v", 2, len(handlers))
	}
}

func TestPrependHandler_Good_Variadic(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	PrependHandler(LabelHandler{}, ProgressHandler{})
	handlers := svc.Handlers()
	if 2 != len(handlers) {
		t.Fatalf("want %v, got %v", 2, len(handlers))
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
	if reflect.TypeOf(handlers[1]) != reflect.TypeOf(ProgressHandler{}) {
		t.Fatalf("expected type %T, got %T", ProgressHandler{}, handlers[1])
	}
}

func TestPrependHandler_Good_SkipsNil(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	var nilHandler KeyHandler
	PrependHandler(nilHandler, LabelHandler{})

	handlers := svc.Handlers()
	if len(handlers) != 1 {
		t.Fatalf("expected length %v, got %v", 1, handlers)
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
}

func TestPrependHandler_DoesNotMutateInputSlice(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	handlers := []KeyHandler{nil, ProgressHandler{}}
	PrependHandler(handlers...)
	if handlers[0] != nil {
		t.Fatalf("expected nil, got %v", handlers[0])
	}
	if reflect.TypeOf(handlers[1]) != reflect.TypeOf(ProgressHandler{}) {
		t.Fatalf("expected type %T, got %T", ProgressHandler{}, handlers[1])
	}
}

func TestClearHandlers_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	AddHandler(LabelHandler{})
	if len(svc.Handlers()) == 0 {
		t.Fatalf("expected non-empty")
	}

	ClearHandlers()
	if len(svc.Handlers()) != 0 {
		t.Fatalf("expected empty, got %v", svc.Handlers())
	}
}

func TestResetHandlers_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	ClearHandlers()
	if len(svc.Handlers()) != 0 {
		t.Fatalf("expected empty, got %v", svc.Handlers())
	}

	svc.ResetHandlers()
	if len(svc.Handlers()) != len(DefaultHandlers()) {
		t.Fatalf("expected length %v, got %v", len(DefaultHandlers()), svc.Handlers())
	}
	if reflect.TypeOf(svc.Handlers()[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, svc.Handlers()[0])
	}

	ClearHandlers()
	if len(svc.Handlers()) != 0 {
		t.Fatalf("expected empty, got %v", svc.Handlers())
	}

	ResetHandlers()
	handlers := svc.Handlers()
	if len(handlers) != len(DefaultHandlers()) {
		t.Fatalf("expected length %v, got %v", len(DefaultHandlers()), handlers)
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
	if "Status:" != T("i18n.label.status") {
		t.Fatalf("want %v, got %v", "Status:", T("i18n.label.status"))
	}
}

func TestSetHandlers_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	SetHandlers(serviceStubHandler{})

	handlers := CurrentHandlers()
	if len(handlers) != 1 {
		t.Fatalf("expected length %v, got %v", 1, handlers)
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(serviceStubHandler{}) {
		t.Fatalf("expected type %T, got %T", serviceStubHandler{}, handlers[0])
	}
	if "stub" != T("custom.stub") {
		t.Fatalf("want %v, got %v", "stub", T("custom.stub"))
	}
	if "i18n.label.status" != T("i18n.label.status") {
		t.Fatalf("want %v, got %v", "i18n.label.status", T("i18n.label.status"))
	}
}

func TestHandlers_Good(t *testing.T) {
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	handlers := Handlers()
	if len(handlers) != len(svc.Handlers()) {
		t.Fatalf("expected length %v, got %v", len(svc.Handlers()), handlers)
	}
	if !reflect.DeepEqual(svc.Handlers(), handlers) {
		t.Fatalf("want %v, got %v", svc.Handlers(), handlers)
	}
}

func TestNewWithHandlers_SkipsNil(t *testing.T) {
	svc, err := serviceFromResult(New(WithHandlers(nil, LabelHandler{})))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handlers := svc.Handlers()
	if len(handlers) != 1 {
		t.Fatalf("expected length %v, got %v", 1, handlers)
	}
	if reflect.TypeOf(handlers[0]) != reflect.TypeOf(LabelHandler{}) {
		t.Fatalf("expected type %T, got %T", LabelHandler{}, handlers[0])
	}
}

func TestExecuteIntentTemplate_Good(t *testing.T) {
	data := templateData{
		Subject: "config.yaml",
		Noun:    "file",
		Count:   1,
	}

	got := executeIntentTemplate("Delete {{.Subject}}?", data)
	if "Delete config.yaml?" != got {
		t.Fatalf("want %v, got %v", "Delete config.yaml?", got)
	}
}

func TestExecuteIntentTemplate_Good_Empty(t *testing.T) {
	got := executeIntentTemplate("", templateData{})
	if "" != got {
		t.Fatalf("want %v, got %v", "", got)
	}
}

func TestExecuteIntentTemplate_Bad_InvalidTemplate(t *testing.T) {
	got := executeIntentTemplate("{{.Invalid", templateData{})
	if "{{.Invalid" != got {
		t.Fatalf("want %v, got %v", "{{.Invalid", got)
	}
}

func TestExecuteIntentTemplate_Good_Cached(t *testing.T) {
	data := templateData{Subject: "test"}
	tmplStr := "Hello {{.Subject}} intent"

	// First call caches
	got1 := executeIntentTemplate(tmplStr, data)
	if "Hello test intent" != got1 {
		t.Fatalf("want %v, got %v", "Hello test intent", got1)
	}

	got2 := executeIntentTemplate(tmplStr, data)
	if "Hello test intent" != got2 {
		t.Fatalf("want %v, got %v", "Hello test intent", got2)
	}
}

func TestExecuteIntentTemplate_Good_WithFuncs(t *testing.T) {
	data := templateData{Subject: "build"}

	got := executeIntentTemplate("{{past .Subject}}!", data)
	if "built!" != got {
		t.Fatalf("want %v, got %v", "built!", got)
	}
}

func TestComposeIntent_Good(t *testing.T) {
	intent := Intent{
		Meta: IntentMeta{
			Type:      "action",
			Verb:      "delete",
			Dangerous: true,
			Default:   "no",
			Supports:  []string{"yes", "no"},
		},
		Question: "Delete {{.Subject}}?",
		Confirm:  "Really delete {{article .Subject}}?",
		Success:  "{{title .Subject}} deleted",
		Failure:  "Failed to delete {{lower .Subject}}",
	}

	got := ComposeIntent(intent, S("file", "config.yaml"))
	if "Delete config.yaml?" != got.Question {
		t.Fatalf("want %v, got %v", "Delete config.yaml?", got.Question)
	}
	if "Really delete a config.yaml?" != got.Confirm {
		t.Fatalf("want %v, got %v", "Really delete a config.yaml?", got.Confirm)
	}
	if "Config.yaml deleted" != got.Success {
		t.Fatalf("want %v, got %v", "Config.yaml deleted", got.Success)
	}
	if "Failed to delete config.yaml" != got.Failure {
		t.Fatalf("want %v, got %v", "Failed to delete config.yaml", got.Failure)
	}
	if !reflect.DeepEqual(intent.Meta, got.Meta) {
		t.Fatalf("want %v, got %v", intent.Meta, got.Meta)
	}
}

func TestIntentCompose_Good_NilSubject(t *testing.T) {
	intent := Intent{
		Question: "Proceed?",
	}

	got := intent.Compose(nil)
	if "Proceed?" != got.Question {
		t.Fatalf("want %v, got %v", "Proceed?", got.Question)
	}
	if len(got.Confirm) != 0 {
		t.Fatalf("expected empty, got %v", got.Confirm)
	}
	if len(got.Success) != 0 {
		t.Fatalf("expected empty, got %v", got.Success)
	}
	if len(got.Failure) != 0 {
		t.Fatalf("expected empty, got %v", got.Failure)
	}
}

// --- applyTemplate ---

func TestApplyTemplate_Good(t *testing.T) {
	got := applyTemplate("Hello {{.Name}}", map[string]any{"Name": "World"})
	if "Hello World" != got {
		t.Fatalf("want %v, got %v", "Hello World", got)
	}
}

func TestApplyTemplate_Good_NoTemplate(t *testing.T) {
	got := applyTemplate("plain text", nil)
	if "plain text" != got {
		t.Fatalf("want %v, got %v", "plain text", got)
	}
}

func TestApplyTemplate_Bad_InvalidTemplate(t *testing.T) {
	got := applyTemplate("{{.Invalid", nil)
	if "{{.Invalid" != got {
		t.Fatalf("want %v, got %v", "{{.Invalid", got)
	}
}

func TestApplyTemplate_Good_Cached(t *testing.T) {
	tmpl := "Cached {{.Val}} template test"
	data := map[string]any{"Val": "test"}

	got1 := applyTemplate(tmpl, data)
	if "Cached test template test" != got1 {
		t.Fatalf("want %v, got %v", "Cached test template test", got1)
	}

	got2 := applyTemplate(tmpl, data)
	if "Cached test template test" != got2 {
		t.Fatalf("want %v, got %v", "Cached test template test", got2)
	}
}

func TestApplyTemplate_Bad_ExecuteError(t *testing.T) {
	// A template that references a method on a non-struct type
	got := applyTemplate("{{.Missing}}", "not a map")
	if "{{.Missing}}" != got {
		t.Fatalf("want %v, got %v", "{{.Missing}}", got)
	}
}

func TestErrServiceNotInitialised_Good(t *testing.T) {
	if "i18n: service not initialised" != ErrServiceNotInitialised.Error() {
		t.Fatalf("want %v, got %v", "i18n: service not initialised", ErrServiceNotInitialised.Error())
	}
}

func TestErrServiceNotInitialized_DeprecatedAlias(t *testing.T) {
	if ErrServiceNotInitialised != ErrServiceNotInitialized {
		t.Fatalf("want %v, got %v", ErrServiceNotInitialised, ErrServiceNotInitialized)
	}
}

type ax7Handler struct {
	match bool
	value string
}

func (h ax7Handler) Match(string) bool { return h.match }

func (h ax7Handler) Handle(_ string, _ []any, next func() string) string {
	if h.value != "" {
		return h.value
	}
	if next != nil {
		return next()
	}
	return ""
}

func serviceForAudit(t *testing.T) *Service {
	t.Helper()
	svc, err := serviceFromResult(New())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	return svc
}

func setDefaultForAudit(t *testing.T) *Service {
	t.Helper()
	prev := defaultService.Load()
	svc := serviceForAudit(t)
	SetDefault(svc)
	t.Cleanup(func() { SetDefault(prev) })
	return svc
}

func coreServiceForAudit(t *testing.T) *CoreService {
	t.Helper()
	return &CoreService{svc: serviceForAudit(t), missingKeys: make([]MissingKey, 0)}
}

func testLocaleFS() fstest.MapFS {
	return fstest.MapFS{
		"locales/en.json": {Data: []byte(`{"prompt":{"yes":"y"},"lang":{"en":"English"}}`)},
		"locales/fr.json": {Data: []byte(`{"prompt":{"yes":"oui"},"lang":{"fr":"français"}}`)},
	}
}

func noPanicForAudit(t *testing.T, fn func()) {
	t.Helper()
	prevDefault := defaultService.Load()
	prevMissing := missingKeyHandlers()
	registeredLocalesMu.Lock()
	prevLocales := append([]localeRegistration(nil), registeredLocales...)
	prevProviders := append([]localeProviderRegistration(nil), registeredLocaleProviders...)
	prevLoaded := localesLoaded
	prevLocaleID := nextLocaleRegistrationID
	prevProviderID := nextLocaleProviderID
	registeredLocalesMu.Unlock()
	defer func() {
		SetDefault(prevDefault)
		missingKeyHandler.Store(prevMissing)
		registeredLocalesMu.Lock()
		registeredLocales = prevLocales
		registeredLocaleProviders = prevProviders
		localesLoaded = prevLoaded
		nextLocaleRegistrationID = prevLocaleID
		nextLocaleProviderID = prevProviderID
		registeredLocalesMu.Unlock()
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}

// --- AX-7 canonical triplets ---

func TestI18n_T_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := T("prompt.yes")
		if got == "" {
			t.Fatal("expected translation")
		}
	})
	if !called {
		t.Fatal("T was not exercised")
	}
}

func TestI18n_T_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := T("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("T was not exercised")
	}
}

func TestI18n_T_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := T("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("T was not exercised")
	}
}

func TestI18n_Translate_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		r := Translate("prompt.yes")
		if !r.OK {
			t.Fatalf("expected ok: %v", r)
		}
	})
	if !called {
		t.Fatal("Translate was not exercised")
	}
}

func TestI18n_Translate_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		r := Translate("missing")
		if r.OK {
			t.Fatalf("expected missing result")
		}
	})
	if !called {
		t.Fatal("Translate was not exercised")
	}
}

func TestI18n_Translate_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		r := Translate("missing")
		if r.OK {
			t.Fatalf("expected failed result")
		}
	})
	if !called {
		t.Fatal("Translate was not exercised")
	}
}

func TestI18n_Raw_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Raw("prompt.yes")
		if got == "" {
			t.Fatal("expected raw translation")
		}
	})
	if !called {
		t.Fatal("Raw was not exercised")
	}
}

func TestI18n_Raw_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Raw("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Raw was not exercised")
	}
}

func TestI18n_Raw_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Raw("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Raw was not exercised")
	}
}

func TestI18n_Compose_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Compose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("Compose was not exercised")
	}
}

func TestI18n_Compose_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Compose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("Compose was not exercised")
	}
}

func TestI18n_Compose_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Compose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("Compose was not exercised")
	}
}

func TestI18n_CurrentCompose_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentCompose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("CurrentCompose was not exercised")
	}
}

func TestI18n_CurrentCompose_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentCompose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("CurrentCompose was not exercised")
	}
}

func TestI18n_CurrentCompose_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentCompose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("CurrentCompose was not exercised")
	}
}

func TestI18n_SetLanguage_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		err := errorFromResult(SetLanguage("fr"))
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("SetLanguage was not exercised")
	}
}

func TestI18n_SetLanguage_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		err := errorFromResult(SetLanguage("zz"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("SetLanguage was not exercised")
	}
}

func TestI18n_SetLanguage_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		err := errorFromResult(SetLanguage("fr"))
		_ = err
	})
	if !called {
		t.Fatal("SetLanguage was not exercised")
	}
}

func TestI18n_CurrentLanguage_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentLanguage()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CurrentLanguage was not exercised")
	}
}

func TestI18n_CurrentLanguage_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentLanguage()
		if got == "" {
			t.Fatal("expected fallback language")
		}
	})
	if !called {
		t.Fatal("CurrentLanguage was not exercised")
	}
}

func TestI18n_CurrentLanguage_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		_ = SetLanguage("fr")
		got := CurrentLanguage()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CurrentLanguage was not exercised")
	}
}

func TestI18n_CurrentLang_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentLang()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CurrentLang was not exercised")
	}
}

func TestI18n_CurrentLang_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentLang()
		if got == "" {
			t.Fatal("expected fallback language")
		}
	})
	if !called {
		t.Fatal("CurrentLang was not exercised")
	}
}

func TestI18n_CurrentLang_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		_ = SetLanguage("fr")
		got := CurrentLang()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CurrentLang was not exercised")
	}
}

func TestI18n_Language_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Language()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Language was not exercised")
	}
}

func TestI18n_Language_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Language()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Language was not exercised")
	}
}

func TestI18n_Language_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		_ = SetLanguage("fr")
		got := Language()
		if got == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Language was not exercised")
	}
}

func TestI18n_AvailableLanguages_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		langs := AvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("AvailableLanguages was not exercised")
	}
}

func TestI18n_AvailableLanguages_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		langs := AvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("AvailableLanguages was not exercised")
	}
}

func TestI18n_AvailableLanguages_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		langs := AvailableLanguages()
		langs[0] = "mutated"
		if AvailableLanguages()[0] == "mutated" {
			t.Fatal("languages slice was not copied")
		}
	})
	if !called {
		t.Fatal("AvailableLanguages was not exercised")
	}
}

func TestI18n_CurrentAvailableLanguages_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		langs := CurrentAvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("CurrentAvailableLanguages was not exercised")
	}
}

func TestI18n_CurrentAvailableLanguages_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		langs := CurrentAvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("CurrentAvailableLanguages was not exercised")
	}
}

func TestI18n_CurrentAvailableLanguages_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		langs := CurrentAvailableLanguages()
		langs[0] = "mutated"
		if CurrentAvailableLanguages()[0] == "mutated" {
			t.Fatal("languages slice was not copied")
		}
	})
	if !called {
		t.Fatal("CurrentAvailableLanguages was not exercised")
	}
}

func TestI18n_SetMode_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetMode(ModeCollect)
		if svc.Mode() != ModeCollect {
			t.Fatal("mode not set")
		}
	})
	if !called {
		t.Fatal("SetMode was not exercised")
	}
}

func TestI18n_SetMode_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		SetMode(ModeCollect)
		_ = defaultService.Load()
	})
	if !called {
		t.Fatal("SetMode was not exercised")
	}
}

func TestI18n_SetMode_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetMode(ModeStrict)
		SetMode(ModeNormal)
		if svc.Mode() != ModeNormal {
			t.Fatal("mode not reset")
		}
	})
	if !called {
		t.Fatal("SetMode was not exercised")
	}
}

func TestI18n_SetFallback_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetFallback("fr")
		if svc.Fallback() != "fr" {
			t.Fatal("fallback not set")
		}
	})
	if !called {
		t.Fatal("SetFallback was not exercised")
	}
}

func TestI18n_SetFallback_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		SetFallback("fr")
		_ = defaultService.Load()
	})
	if !called {
		t.Fatal("SetFallback was not exercised")
	}
}

func TestI18n_SetFallback_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetFallback("")
		if svc.Fallback() != "" {
			t.Fatal("empty fallback not set")
		}
	})
	if !called {
		t.Fatal("SetFallback was not exercised")
	}
}

func TestI18n_Fallback_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetFallback("fr")
		got := Fallback()
		if got != "fr" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Fallback was not exercised")
	}
}

func TestI18n_Fallback_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Fallback()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Fallback was not exercised")
	}
}

func TestI18n_Fallback_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Fallback()
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("Fallback was not exercised")
	}
}

func TestI18n_CurrentMode_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetMode(ModeCollect)
		got := CurrentMode()
		if got != ModeCollect {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentMode was not exercised")
	}
}

func TestI18n_CurrentMode_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentMode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentMode was not exercised")
	}
}

func TestI18n_CurrentMode_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentMode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentMode was not exercised")
	}
}

func TestI18n_CurrentFallback_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetFallback("fr")
		got := CurrentFallback()
		if got != "fr" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CurrentFallback was not exercised")
	}
}

func TestI18n_CurrentFallback_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentFallback()
		if got != "en" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CurrentFallback was not exercised")
	}
}

func TestI18n_CurrentFallback_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentFallback()
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("CurrentFallback was not exercised")
	}
}

func TestI18n_CurrentFormality_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetFormality(FormalityFormal)
		got := CurrentFormality()
		if got != FormalityFormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentFormality was not exercised")
	}
}

func TestI18n_CurrentFormality_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentFormality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentFormality was not exercised")
	}
}

func TestI18n_CurrentFormality_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentFormality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CurrentFormality was not exercised")
	}
}

func TestI18n_CurrentDebug_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetDebug(true)
		got := CurrentDebug()
		if !got {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("CurrentDebug was not exercised")
	}
}

func TestI18n_CurrentDebug_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentDebug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("CurrentDebug was not exercised")
	}
}

func TestI18n_CurrentDebug_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentDebug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("CurrentDebug was not exercised")
	}
}

func TestI18n_State_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		state := State()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("State was not exercised")
	}
}

func TestI18n_State_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		state := State()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("State was not exercised")
	}
}

func TestI18n_State_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetDebug(true)
		state := State()
		if !state.Debug {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("State was not exercised")
	}
}

func TestI18n_CurrentState_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		state := CurrentState()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CurrentState was not exercised")
	}
}

func TestI18n_CurrentState_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		state := CurrentState()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("CurrentState was not exercised")
	}
}

func TestI18n_CurrentState_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetMode(ModeCollect)
		state := CurrentState()
		if state.Mode != ModeCollect {
			t.Fatal("expected collect")
		}
	})
	if !called {
		t.Fatal("CurrentState was not exercised")
	}
}

func TestI18n_Debug_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		svc.SetDebug(true)
		got := Debug()
		if !got {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("Debug was not exercised")
	}
}

func TestI18n_Debug_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Debug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("Debug was not exercised")
	}
}

func TestI18n_Debug_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Debug()
		if got {
			t.Fatal("unexpected debug")
		}
	})
	if !called {
		t.Fatal("Debug was not exercised")
	}
}

func TestI18n_N_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := N("number", 1234)
		if got == "" {
			t.Fatal("expected numeric text")
		}
	})
	if !called {
		t.Fatal("N was not exercised")
	}
}

func TestI18n_N_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := N("unknown", 1)
		if got == "" {
			t.Fatal("expected fallback text")
		}
	})
	if !called {
		t.Fatal("N was not exercised")
	}
}

func TestI18n_N_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := N("ago", 0, "second")
		if got == "" {
			t.Fatal("expected ago text")
		}
	})
	if !called {
		t.Fatal("N was not exercised")
	}
}

func TestI18n_Prompt_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Prompt("yes")
		if got == "" {
			t.Fatal("expected prompt")
		}
	})
	if !called {
		t.Fatal("Prompt was not exercised")
	}
}

func TestI18n_Prompt_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Prompt("yes")
		if got == "" {
			t.Fatal("expected fallback prompt")
		}
	})
	if !called {
		t.Fatal("Prompt was not exercised")
	}
}

func TestI18n_Prompt_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Prompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Prompt was not exercised")
	}
}

func TestI18n_CurrentPrompt_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected prompt")
		}
	})
	if !called {
		t.Fatal("CurrentPrompt was not exercised")
	}
}

func TestI18n_CurrentPrompt_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected fallback prompt")
		}
	})
	if !called {
		t.Fatal("CurrentPrompt was not exercised")
	}
}

func TestI18n_CurrentPrompt_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := CurrentPrompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CurrentPrompt was not exercised")
	}
}

func TestI18n_Lang_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Lang("en")
		if got == "" {
			t.Fatal("expected language label")
		}
	})
	if !called {
		t.Fatal("Lang was not exercised")
	}
}

func TestI18n_Lang_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		got := Lang("en")
		if got == "" {
			t.Fatal("expected fallback label")
		}
	})
	if !called {
		t.Fatal("Lang was not exercised")
	}
}

func TestI18n_Lang_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		got := Lang("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Lang was not exercised")
	}
}

func TestI18n_AddHandler_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		AddHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("AddHandler was not exercised")
	}
}

func TestI18n_AddHandler_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		AddHandler(nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected default handlers")
		}
	})
	if !called {
		t.Fatal("AddHandler was not exercised")
	}
}

func TestI18n_AddHandler_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		AddHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("AddHandler was not exercised")
	}
}

func TestI18n_SetHandlers_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetHandlers(ax7Handler{match: true})
		if len(svc.Handlers()) != 1 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("SetHandlers was not exercised")
	}
}

func TestI18n_SetHandlers_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetHandlers(nil)
		if len(svc.Handlers()) != 0 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("SetHandlers was not exercised")
	}
}

func TestI18n_SetHandlers_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		SetHandlers(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) != 1 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("SetHandlers was not exercised")
	}
}

func TestI18n_LoadFS_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		LoadFS(testLocaleFS(), "locales")
		if len(svc.AvailableLanguages()) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("LoadFS was not exercised")
	}
}

func TestI18n_LoadFS_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		LoadFS(testLocaleFS(), "missing")
		_ = AvailableLanguages()
	})
	if !called {
		t.Fatal("LoadFS was not exercised")
	}
}

func TestI18n_LoadFS_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		LoadFS(testLocaleFS(), "")
		_ = AvailableLanguages()
	})
	if !called {
		t.Fatal("LoadFS was not exercised")
	}
}

func TestI18n_AddMessages_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		AddMessages("en", map[string]string{"ax7.message": "ready"})
		if T("ax7.message") != "ready" {
			t.Fatal("message not added")
		}
	})
	if !called {
		t.Fatal("AddMessages was not exercised")
	}
}

func TestI18n_AddMessages_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		AddMessages("", nil)
		_ = AvailableLanguages()
	})
	if !called {
		t.Fatal("AddMessages was not exercised")
	}
}

func TestI18n_AddMessages_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		AddMessages("en", map[string]string{})
		_ = AvailableLanguages()
	})
	if !called {
		t.Fatal("AddMessages was not exercised")
	}
}

func TestI18n_PrependHandler_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		PrependHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("PrependHandler was not exercised")
	}
}

func TestI18n_PrependHandler_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		PrependHandler(nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected default handlers")
		}
	})
	if !called {
		t.Fatal("PrependHandler was not exercised")
	}
}

func TestI18n_PrependHandler_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		PrependHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("PrependHandler was not exercised")
	}
}

func TestI18n_CurrentHandlers_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		handlers := CurrentHandlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("CurrentHandlers was not exercised")
	}
}

func TestI18n_CurrentHandlers_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		handlers := CurrentHandlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("CurrentHandlers was not exercised")
	}
}

func TestI18n_CurrentHandlers_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		handlers := CurrentHandlers()
		handlers[0] = nil
		if CurrentHandlers()[0] == nil {
			t.Fatal("handlers were not copied")
		}
	})
	if !called {
		t.Fatal("CurrentHandlers was not exercised")
	}
}

func TestI18n_Handlers_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		handlers := Handlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Handlers was not exercised")
	}
}

func TestI18n_Handlers_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		handlers := Handlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("Handlers was not exercised")
	}
}

func TestI18n_Handlers_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		setDefaultForAudit(t)
		handlers := Handlers()
		handlers[0] = nil
		if Handlers()[0] == nil {
			t.Fatal("handlers were not copied")
		}
	})
	if !called {
		t.Fatal("Handlers was not exercised")
	}
}

func TestI18n_ClearHandlers_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		ClearHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("ClearHandlers was not exercised")
	}
}

func TestI18n_ClearHandlers_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		ClearHandlers()
		_ = defaultService.Load()
	})
	if !called {
		t.Fatal("ClearHandlers was not exercised")
	}
}

func TestI18n_ClearHandlers_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		ClearHandlers()
		ClearHandlers()
		if len(svc.Handlers()) != 0 {
			t.Fatalf("got %d", len(svc.Handlers()))
		}
	})
	if !called {
		t.Fatal("ClearHandlers was not exercised")
	}
}

func TestI18n_ResetHandlers_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		ClearHandlers()
		ResetHandlers()
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("ResetHandlers was not exercised")
	}
}

func TestI18n_ResetHandlers_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		SetDefault(nil)
		ResetHandlers()
		_ = defaultService.Load()
	})
	if !called {
		t.Fatal("ResetHandlers was not exercised")
	}
}

func TestI18n_ResetHandlers_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := setDefaultForAudit(t)
		ResetHandlers()
		ResetHandlers()
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("ResetHandlers was not exercised")
	}
}

func TestI18n_Buffer_Write_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		n, err := buf.Write([]byte("hello"))
		if err != nil || n != 5 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	if !called {
		t.Fatal("Buffer_Write was not exercised")
	}
}

func TestI18n_Buffer_Write_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		n, err := buf.Write(nil)
		if err != nil || n != 0 {
			t.Fatalf("n=%d err=%v", n, err)
		}
	})
	if !called {
		t.Fatal("Buffer_Write was not exercised")
	}
}

func TestI18n_Buffer_Write_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		_, _ = buf.Write([]byte("hello"))
		_, _ = buf.Write([]byte(" world"))
		if buf.String() != "hello world" {
			t.Fatalf("got %q", buf.String())
		}
	})
	if !called {
		t.Fatal("Buffer_Write was not exercised")
	}
}

func TestI18n_Buffer_String_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		_, _ = buf.Write([]byte("hello"))
		if buf.String() != "hello" {
			t.Fatalf("got %q", buf.String())
		}
	})
	if !called {
		t.Fatal("Buffer_String was not exercised")
	}
}

func TestI18n_Buffer_String_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		got := buf.String()
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("Buffer_String was not exercised")
	}
}

func TestI18n_Buffer_String_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		buf := core.NewBuffer()
		_, _ = buf.Write(nil)
		if buf.String() != "" {
			t.Fatalf("got %q", buf.String())
		}
	})
	if !called {
		t.Fatal("Buffer_String was not exercised")
	}
}
