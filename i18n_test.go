package i18n

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// --- Package-level T() ---

func TestT_Good(t *testing.T) {
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	result := Translate("nonexistent.translation.key")
	if result.OK {
		t.Fatal("expected false")
	}
	if "nonexistent.translation.key" != result.Value {
		t.Fatalf("want %v, got %v", "nonexistent.translation.key", result.Value)
	}
}

func TestTranslate_Good_SameTextAsKey(t *testing.T) {
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	err = SetLanguage("en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(CurrentLanguage(), "en") {
		t.Fatalf("expected %q to contain %q", CurrentLanguage(), "en")
	}
}

func TestSetLanguage_Good_UnderscoreTag(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	err = SetLanguage("fr_CA")
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
	svc, err := New()
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
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	_ = SetLanguage("xx")
}

func TestCurrentLanguage_Good(t *testing.T) {
	svc, err := New()
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

	svc, err := New()
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

	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := NewWithLoader(messageBaseFallbackLoader{})
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
		{"bytes", "bytes", int64(1536000), nil, "1.46 MB"},
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
		{"bytes", "bytes", int64(1536000), nil, "1.46 MB"},
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New(WithHandlers())
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
	svc, err := New(WithHandlers())
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
	svc, err := New(WithHandlers())
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
	svc, err := New(WithHandlers()) // start with no handlers
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
	svc, err := New(WithHandlers())
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
	svc, err := New(WithHandlers())
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
	svc, err := New(WithHandlers())
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New()
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
	svc, err := New(WithHandlers(nil, LabelHandler{}))
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
