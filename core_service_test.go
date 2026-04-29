package i18n

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCoreServiceNilSafe(t *testing.T) {
	var svc *CoreService
	savedDefault := defaultService.Load()
	t.Cleanup(func() {
		defaultService.Store(savedDefault)
	})
	defaultService.Store(nil)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()

		func() {
			if (ModeNormal) != (svc.Mode()) {
				t.Fatalf("want %v, got %v", ModeNormal, svc.Mode())
			}
			if ("en") != (svc.Language()) {
				t.Fatalf("want %v, got %v", "en", svc.Language())
			}
			if ("en") != (svc.Fallback()) {
				t.Fatalf("want %v, got %v", "en", svc.Fallback())
			}
			if (FormalityNeutral) != (svc.Formality()) {
				t.Fatalf("want %v, got %v", FormalityNeutral, svc.Formality())
			}
			if ("") != (svc.Location()) {
				t.Fatalf("want %v, got %v", "", svc.Location())
			}
			if svc.Debug() {
				t.Fatal("expected false")
			}
			if (DirLTR) != (svc.Direction()) {
				t.Fatalf("want %v, got %v", DirLTR, svc.Direction())
			}
			if svc.IsRTL() {
				t.Fatal("expected false")
			}
			if (PluralOther) != (svc.PluralCategory(3)) {
				t.Fatalf("want %v, got %v", PluralOther, svc.PluralCategory(3))
			}
			if len(svc.AvailableLanguages()) != 0 {
				t.Fatalf("expected empty, got %v", svc.AvailableLanguages())
			}
			if len(svc.Handlers()) != 0 {
				t.Fatalf("expected empty, got %v", svc.Handlers())
			}
			if ("prompt.confirm") != (svc.Prompt("confirm")) {
				t.Fatalf("want %v, got %v", "prompt.confirm", svc.Prompt("confirm"))
			}
			if ("lang.fr") != (svc.Lang("fr")) {
				t.Fatalf("want %v, got %v", "lang.fr", svc.Lang("fr"))
			}
			if ("hello") != (svc.T("hello")) {
				t.Fatalf("want %v, got %v", "hello", svc.T("hello"))
			}
			if ("hello") != (svc.Raw("hello")) {
				t.Fatalf("want %v, got %v", "hello", svc.Raw("hello"))
			}
			result := svc.Translate("hello")
			if result.OK {
				t.Fatalf("expected failed translation result, got %v", result)
			}
			if result.Error() != "hello" {
				t.Fatalf("want %v, got %v", "hello", result.Error())
			}
			if !reflect.DeepEqual(defaultServiceStateSnapshot(), svc.State()) {
				t.Fatalf("want %v, got %v", defaultServiceStateSnapshot(), svc.State())
			}
			if !reflect.DeepEqual(defaultServiceStateSnapshot(), svc.CurrentState()) {
				t.Fatalf("want %v, got %v", defaultServiceStateSnapshot(), svc.CurrentState())
			}
			if (defaultServiceStateSnapshot().String()) != (svc.String()) {
				t.Fatalf("want %v, got %v", defaultServiceStateSnapshot().String(), svc.String())
			}
		}()
	}()
	if (defaultService.Load()) != (nil) {
		t.Fatalf("expected nil, got %v", defaultService.Load())
	}
	if result := svc.OnStartup(nil); !result.OK {
		t.Fatalf("expected startup OK, got %v", result)
	}
	if result := svc.OnShutdown(nil); !result.OK {
		t.Fatalf("expected shutdown OK, got %v", result)
	}
	svc.SetMode(ModeCollect)
	svc.SetFallback("fr")
	svc.SetFormality(FormalityFormal)
	svc.SetLocation("workspace")
	svc.SetDebug(true)
	svc.SetHandlers(nil)
	svc.AddHandler(nil)
	svc.PrependHandler(nil)
	svc.ClearHandlers()
	svc.ResetHandlers()
	svc.AddMessages("en", nil)
	if err := svc.SetLanguage("fr"); !errors.Is(err, ErrServiceNotInitialised) {
		t.Fatalf("expected error %v, got %v", ErrServiceNotInitialised, err)
	}
	if err := svc.AddLoader(nil); !errors.Is(err, ErrServiceNotInitialised) {
		t.Fatalf("expected error %v, got %v", ErrServiceNotInitialised, err)
	}
	if err := svc.LoadFS(nil, "locales"); !errors.Is(err, ErrServiceNotInitialised) {
		t.Fatalf("expected error %v, got %v", ErrServiceNotInitialised, err)
	}
}

func TestCoreServiceMissingKeysReturnsCopies(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	coreSvc := &CoreService{svc: svc}

	coreSvc.SetMode(ModeCollect)
	_ = svc.T("missing.copy.key", map[string]any{"foo": "bar"})

	missing := coreSvc.MissingKeys()
	if len(missing) != 1 {
		t.Fatalf("expected length %v, got %v", 1, missing)
	}
	if ("bar") != (missing[0].Args["foo"]) {
		t.Fatalf("want %v, got %v", "bar", missing[0].Args["foo"])
	}

	missing[0].Args["foo"] = "mutated"

	again := coreSvc.MissingKeys()
	if len(again) != 1 {
		t.Fatalf("expected length %v, got %v", 1, again)
	}
	if ("bar") != (again[0].Args["foo"]) {
		t.Fatalf("want %v, got %v", "bar", again[0].Args["foo"])
	}
}

func TestServiceOptionsAndFSSourceString(t *testing.T) {
	opts := ServiceOptions{
		Language:  "en-GB",
		Fallback:  "en",
		Formality: FormalityFormal,
		Location:  "workspace",
		Mode:      ModeCollect,
		Debug:     true,
		ExtraFS: []FSSource{
			{FS: fstest.MapFS{}, Dir: "locales"},
		},
	}

	got := opts.String()
	if !strings.Contains(got, `language="en-GB"`) {
		t.Fatalf("expected %q to contain %q", got, `language="en-GB"`)
	}
	if !strings.Contains(got, `fallback="en"`) {
		t.Fatalf("expected %q to contain %q", got, `fallback="en"`)
	}
	if !strings.Contains(got, `formality=formal`) {
		t.Fatalf("expected %q to contain %q", got, `formality=formal`)
	}
	if !strings.Contains(got, `location="workspace"`) {
		t.Fatalf("expected %q to contain %q", got, `location="workspace"`)
	}
	if !strings.Contains(got, `mode=collect`) {
		t.Fatalf("expected %q to contain %q", got, `mode=collect`)
	}
	if !strings.Contains(got, `debug=true`) {
		t.Fatalf("expected %q to contain %q", got, `debug=true`)
	}
	if !strings.Contains(got, `FSSource{fs=fstest.MapFS dir="locales"}`) {
		t.Fatalf("expected %q to contain %q", got, `FSSource{fs=fstest.MapFS dir="locales"}`)
	}

	src := FSSource{FS: fstest.MapFS{}, Dir: "translations"}
	if (`FSSource{fs=fstest.MapFS dir="translations"}`) != (src.String()) {
		t.Fatalf("want %v, got %v", `FSSource{fs=fstest.MapFS dir="translations"}`, src.String())
	}
}

// --- AX-7 canonical triplets ---

func TestCoreService_FSSource_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		src := FSSource{FS: ax7TestFS(), Dir: "locales"}
		got := src.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("FSSource_String was not exercised")
	}
}

func TestCoreService_FSSource_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		src := FSSource{}
		got := src.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("FSSource_String was not exercised")
	}
}

func TestCoreService_FSSource_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		src := FSSource{FS: ax7TestFS()}
		got := src.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("FSSource_String was not exercised")
	}
}

func TestCoreService_ServiceOptions_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		opts := ServiceOptions{Language: "en", ExtraFS: []FSSource{{FS: ax7TestFS(), Dir: "locales"}}}
		got := opts.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("ServiceOptions_String was not exercised")
	}
}

func TestCoreService_ServiceOptions_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		opts := ServiceOptions{}
		got := opts.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("ServiceOptions_String was not exercised")
	}
}

func TestCoreService_ServiceOptions_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		opts := ServiceOptions{Debug: true, Mode: ModeCollect}
		got := opts.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("ServiceOptions_String was not exercised")
	}
}

func TestCoreService_NewCoreService_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		factory := NewCoreService(ServiceOptions{})
		got, err := factory(nil)
		if err != nil || got == nil {
			t.Fatalf("got=%v err=%v", got, err)
		}
	})
	if !called {
		t.Fatal("NewCoreService was not exercised")
	}
}

func TestCoreService_NewCoreService_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		factory := NewCoreService(ServiceOptions{Language: "zz"})
		got, err := factory(nil)
		if err == nil || got != nil {
			t.Fatalf("got=%v err=%v", got, err)
		}
	})
	if !called {
		t.Fatal("NewCoreService was not exercised")
	}
}

func TestCoreService_NewCoreService_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		factory := NewCoreService(ServiceOptions{ExtraFS: []FSSource{{FS: ax7TestFS(), Dir: "locales"}}})
		got, err := factory(nil)
		if err != nil || got == nil {
			t.Fatalf("got=%v err=%v", got, err)
		}
	})
	if !called {
		t.Fatal("NewCoreService was not exercised")
	}
}

func TestCoreService_CoreService_OnStartup_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		r := svc.OnStartup(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnStartup was not exercised")
	}
}

func TestCoreService_CoreService_OnStartup_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		r := svc.OnStartup(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnStartup was not exercised")
	}
}

func TestCoreService_CoreService_OnStartup_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetMode(ModeCollect)
		r := svc.OnStartup(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnStartup was not exercised")
	}
}

func TestCoreService_CoreService_OnShutdown_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		r := svc.OnShutdown(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnShutdown was not exercised")
	}
}

func TestCoreService_CoreService_OnShutdown_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		r := svc.OnShutdown(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnShutdown was not exercised")
	}
}

func TestCoreService_CoreService_OnShutdown_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		r := svc.OnShutdown(nil)
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_OnShutdown was not exercised")
	}
}

func TestCoreService_CoreService_MissingKeys_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		keys := svc.MissingKeys()
		if len(keys) != 0 {
			t.Fatalf("got %v", keys)
		}
	})
	if !called {
		t.Fatal("CoreService_MissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_MissingKeys_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		keys := svc.MissingKeys()
		if len(keys) != 0 {
			t.Fatalf("got %v", keys)
		}
	})
	if !called {
		t.Fatal("CoreService_MissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_MissingKeys_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.missingKeys = []MissingKey{{Key: "x"}}
		keys := svc.MissingKeys()
		keys[0].Key = "mutated"
		if svc.MissingKeys()[0].Key == "mutated" {
			t.Fatal("keys not copied")
		}
	})
	if !called {
		t.Fatal("CoreService_MissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_ClearMissingKeys_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.missingKeys = []MissingKey{{Key: "x"}}
		svc.ClearMissingKeys()
		if len(svc.MissingKeys()) != 0 {
			t.Fatal("expected clear")
		}
	})
	if !called {
		t.Fatal("CoreService_ClearMissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_ClearMissingKeys_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.ClearMissingKeys()
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_ClearMissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_ClearMissingKeys_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.ClearMissingKeys()
		svc.ClearMissingKeys()
		if len(svc.MissingKeys()) != 0 {
			t.Fatal("expected clear")
		}
	})
	if !called {
		t.Fatal("CoreService_ClearMissingKeys was not exercised")
	}
}

func TestCoreService_CoreService_SetMode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetMode(ModeCollect)
		if svc.Mode() != ModeCollect {
			t.Fatal("delegation failed")
		}
	})
	if !called {
		t.Fatal("CoreService_SetMode was not exercised")
	}
}

func TestCoreService_CoreService_SetMode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetMode(ModeCollect)
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetMode was not exercised")
	}
}

func TestCoreService_CoreService_SetMode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetMode(ModeCollect)
		_ = svc.State()
	})
	if !called {
		t.Fatal("CoreService_SetMode was not exercised")
	}
}

func TestCoreService_CoreService_SetFallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetFallback("fr")
		if svc.Fallback() != "fr" {
			t.Fatal("delegation failed")
		}
	})
	if !called {
		t.Fatal("CoreService_SetFallback was not exercised")
	}
}

func TestCoreService_CoreService_SetFallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetFallback("fr")
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetFallback was not exercised")
	}
}

func TestCoreService_CoreService_SetFallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetFallback("fr")
		_ = svc.State()
	})
	if !called {
		t.Fatal("CoreService_SetFallback was not exercised")
	}
}

func TestCoreService_CoreService_SetFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetFormality(FormalityFormal)
		if svc.Formality() != FormalityFormal {
			t.Fatal("delegation failed")
		}
	})
	if !called {
		t.Fatal("CoreService_SetFormality was not exercised")
	}
}

func TestCoreService_CoreService_SetFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetFormality(FormalityFormal)
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetFormality was not exercised")
	}
}

func TestCoreService_CoreService_SetFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetFormality(FormalityFormal)
		_ = svc.State()
	})
	if !called {
		t.Fatal("CoreService_SetFormality was not exercised")
	}
}

func TestCoreService_CoreService_SetLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetLocation("workspace")
		if svc.Location() != "workspace" {
			t.Fatal("delegation failed")
		}
	})
	if !called {
		t.Fatal("CoreService_SetLocation was not exercised")
	}
}

func TestCoreService_CoreService_SetLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetLocation("workspace")
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetLocation was not exercised")
	}
}

func TestCoreService_CoreService_SetLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetLocation("workspace")
		_ = svc.State()
	})
	if !called {
		t.Fatal("CoreService_SetLocation was not exercised")
	}
}

func TestCoreService_CoreService_SetDebug_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetDebug(true)
		if !svc.Debug() {
			t.Fatal("delegation failed")
		}
	})
	if !called {
		t.Fatal("CoreService_SetDebug was not exercised")
	}
}

func TestCoreService_CoreService_SetDebug_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetDebug(true)
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetDebug was not exercised")
	}
}

func TestCoreService_CoreService_SetDebug_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetDebug(true)
		_ = svc.State()
	})
	if !called {
		t.Fatal("CoreService_SetDebug was not exercised")
	}
}

func TestCoreService_CoreService_Mode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Mode()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Mode was not exercised")
	}
}

func TestCoreService_CoreService_Mode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Mode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Mode was not exercised")
	}
}

func TestCoreService_CoreService_Mode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Mode()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Mode was not exercised")
	}
}

func TestCoreService_CoreService_CurrentMode_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentMode()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentMode was not exercised")
	}
}

func TestCoreService_CoreService_CurrentMode_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentMode()
		if got != ModeNormal {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentMode was not exercised")
	}
}

func TestCoreService_CoreService_CurrentMode_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentMode()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentMode was not exercised")
	}
}

func TestCoreService_CoreService_Language_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Language()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Language was not exercised")
	}
}

func TestCoreService_CoreService_Language_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Language()
		if got != "en" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Language was not exercised")
	}
}

func TestCoreService_CoreService_Language_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Language()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Language was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLanguage()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLanguage was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentLanguage()
		if got != "en" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentLanguage was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLanguage()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLanguage was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLang_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLang()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLang was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLang_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentLang()
		if got != "en" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentLang was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLang_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLang()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLang was not exercised")
	}
}

func TestCoreService_CoreService_Fallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Fallback()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Fallback was not exercised")
	}
}

func TestCoreService_CoreService_Fallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Fallback()
		if got != "en" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Fallback was not exercised")
	}
}

func TestCoreService_CoreService_Fallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Fallback()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Fallback was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFallback_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentFallback()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentFallback was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFallback_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentFallback()
		if got != "en" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentFallback was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFallback_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentFallback()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentFallback was not exercised")
	}
}

func TestCoreService_CoreService_Formality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Formality()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Formality was not exercised")
	}
}

func TestCoreService_CoreService_Formality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Formality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Formality was not exercised")
	}
}

func TestCoreService_CoreService_Formality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Formality()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Formality was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFormality_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentFormality()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentFormality was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFormality_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentFormality()
		if got != FormalityNeutral {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentFormality was not exercised")
	}
}

func TestCoreService_CoreService_CurrentFormality_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentFormality()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentFormality was not exercised")
	}
}

func TestCoreService_CoreService_Location_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Location()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Location was not exercised")
	}
}

func TestCoreService_CoreService_Location_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Location()
		if got != "" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Location was not exercised")
	}
}

func TestCoreService_CoreService_Location_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Location()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Location was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLocation_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLocation()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLocation was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLocation_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentLocation()
		if got != "" {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentLocation was not exercised")
	}
}

func TestCoreService_CoreService_CurrentLocation_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentLocation()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentLocation was not exercised")
	}
}

func TestCoreService_CoreService_Debug_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Debug()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Debug was not exercised")
	}
}

func TestCoreService_CoreService_Debug_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Debug()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Debug was not exercised")
	}
}

func TestCoreService_CoreService_Debug_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Debug()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Debug was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDebug_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentDebug()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentDebug was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDebug_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentDebug()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentDebug was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDebug_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentDebug()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentDebug was not exercised")
	}
}

func TestCoreService_CoreService_Direction_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Direction()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Direction was not exercised")
	}
}

func TestCoreService_CoreService_Direction_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Direction()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Direction was not exercised")
	}
}

func TestCoreService_CoreService_Direction_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Direction()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Direction was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentDirection()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentDirection was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentDirection was not exercised")
	}
}

func TestCoreService_CoreService_CurrentDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentDirection()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentDirection was not exercised")
	}
}

func TestCoreService_CoreService_CurrentTextDirection_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentTextDirection()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentTextDirection was not exercised")
	}
}

func TestCoreService_CoreService_CurrentTextDirection_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentTextDirection()
		if got != DirLTR {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentTextDirection was not exercised")
	}
}

func TestCoreService_CoreService_CurrentTextDirection_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentTextDirection()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentTextDirection was not exercised")
	}
}

func TestCoreService_CoreService_IsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.IsRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_IsRTL was not exercised")
	}
}

func TestCoreService_CoreService_IsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.IsRTL()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_IsRTL was not exercised")
	}
}

func TestCoreService_CoreService_IsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.IsRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_IsRTL was not exercised")
	}
}

func TestCoreService_CoreService_RTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.RTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_RTL was not exercised")
	}
}

func TestCoreService_CoreService_RTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.RTL()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_RTL was not exercised")
	}
}

func TestCoreService_CoreService_RTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.RTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_RTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentIsRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentIsRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentIsRTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentIsRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentIsRTL()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentIsRTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentIsRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentIsRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentIsRTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentRTL_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentRTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentRTL_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentRTL()
		if got != false {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentRTL was not exercised")
	}
}

func TestCoreService_CoreService_CurrentRTL_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentRTL()
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentRTL was not exercised")
	}
}

func TestCoreService_CoreService_T_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.T("prompt.yes")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("CoreService_T was not exercised")
	}
}

func TestCoreService_CoreService_T_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.T("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_T was not exercised")
	}
}

func TestCoreService_CoreService_T_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.T("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_T was not exercised")
	}
}

func TestCoreService_CoreService_Raw_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Raw("prompt.yes")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("CoreService_Raw was not exercised")
	}
}

func TestCoreService_CoreService_Raw_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Raw("missing")
		if got != "missing" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Raw was not exercised")
	}
}

func TestCoreService_CoreService_Raw_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Raw("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Raw was not exercised")
	}
}

func TestCoreService_CoreService_Compose_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Compose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Compose was not exercised")
	}
}

func TestCoreService_CoreService_Compose_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Compose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Compose was not exercised")
	}
}

func TestCoreService_CoreService_Compose_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Compose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_Compose was not exercised")
	}
}

func TestCoreService_CoreService_CurrentCompose_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentCompose("core.delete", S("file", "config.yaml"))
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentCompose was not exercised")
	}
}

func TestCoreService_CoreService_CurrentCompose_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentCompose("missing", nil)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentCompose was not exercised")
	}
}

func TestCoreService_CoreService_CurrentCompose_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentCompose("", nil)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentCompose was not exercised")
	}
}

func TestCoreService_CoreService_Translate_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		r := svc.Translate("prompt.yes")
		if !r.OK {
			t.Fatalf("got %v", r)
		}
	})
	if !called {
		t.Fatal("CoreService_Translate was not exercised")
	}
}

func TestCoreService_CoreService_Translate_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		r := svc.Translate("missing")
		if r.OK {
			t.Fatal("expected failure")
		}
	})
	if !called {
		t.Fatal("CoreService_Translate was not exercised")
	}
}

func TestCoreService_CoreService_Translate_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		r := svc.Translate("missing")
		if r.OK {
			t.Fatal("expected missing")
		}
	})
	if !called {
		t.Fatal("CoreService_Translate was not exercised")
	}
}

func TestCoreService_CoreService_Prompt_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Prompt("yes")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("CoreService_Prompt was not exercised")
	}
}

func TestCoreService_CoreService_Prompt_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Prompt("yes")
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("CoreService_Prompt was not exercised")
	}
}

func TestCoreService_CoreService_Prompt_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Prompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Prompt was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPrompt_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentPrompt was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPrompt_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentPrompt("yes")
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentPrompt was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPrompt_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentPrompt("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentPrompt was not exercised")
	}
}

func TestCoreService_CoreService_Lang_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Lang("en")
		if got == "" {
			t.Fatal("expected text")
		}
	})
	if !called {
		t.Fatal("CoreService_Lang was not exercised")
	}
}

func TestCoreService_CoreService_Lang_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.Lang("en")
		if got == "" {
			t.Fatal("expected fallback")
		}
	})
	if !called {
		t.Fatal("CoreService_Lang was not exercised")
	}
}

func TestCoreService_CoreService_Lang_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.Lang("")
		if got != "" {
			t.Fatalf("got %q", got)
		}
	})
	if !called {
		t.Fatal("CoreService_Lang was not exercised")
	}
}

func TestCoreService_CoreService_State_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		state := svc.State()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CoreService_State was not exercised")
	}
}

func TestCoreService_CoreService_State_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		state := svc.State()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("CoreService_State was not exercised")
	}
}

func TestCoreService_CoreService_State_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetDebug(true)
		state := svc.State()
		if !state.Debug {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("CoreService_State was not exercised")
	}
}

func TestCoreService_CoreService_CurrentState_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		state := svc.CurrentState()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentState was not exercised")
	}
}

func TestCoreService_CoreService_CurrentState_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		state := svc.CurrentState()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentState was not exercised")
	}
}

func TestCoreService_CoreService_CurrentState_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetDebug(true)
		state := svc.CurrentState()
		if !state.Debug {
			t.Fatal("expected debug")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentState was not exercised")
	}
}

func TestCoreService_CoreService_String_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("CoreService_String was not exercised")
	}
}

func TestCoreService_CoreService_String_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("CoreService_String was not exercised")
	}
}

func TestCoreService_CoreService_String_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetLocation("workspace")
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("CoreService_String was not exercised")
	}
}

func TestCoreService_CoreService_AddHandler_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.AddHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_AddHandler was not exercised")
	}
}

func TestCoreService_CoreService_AddHandler_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.AddHandler(ax7Handler{match: true})
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_AddHandler was not exercised")
	}
}

func TestCoreService_CoreService_AddHandler_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.AddHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_AddHandler was not exercised")
	}
}

func TestCoreService_CoreService_SetHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetHandlers(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_SetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_SetHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.SetHandlers(ax7Handler{match: true})
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_SetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_SetHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.SetHandlers(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_SetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_PrependHandler_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.PrependHandler(ax7Handler{match: true})
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_PrependHandler was not exercised")
	}
}

func TestCoreService_CoreService_PrependHandler_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.PrependHandler(ax7Handler{match: true})
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_PrependHandler was not exercised")
	}
}

func TestCoreService_CoreService_PrependHandler_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.PrependHandler(ax7Handler{match: true}, nil)
		if len(svc.Handlers()) == 0 {
			t.Fatal("expected handler")
		}
	})
	if !called {
		t.Fatal("CoreService_PrependHandler was not exercised")
	}
}

func TestCoreService_CoreService_ClearHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.ClearHandlers()
		_ = svc.Handlers()
	})
	if !called {
		t.Fatal("CoreService_ClearHandlers was not exercised")
	}
}

func TestCoreService_CoreService_ClearHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.ClearHandlers()
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_ClearHandlers was not exercised")
	}
}

func TestCoreService_CoreService_ClearHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.ClearHandlers()
		svc.ClearHandlers()
		_ = svc.Handlers()
	})
	if !called {
		t.Fatal("CoreService_ClearHandlers was not exercised")
	}
}

func TestCoreService_CoreService_ResetHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.ResetHandlers()
		_ = svc.Handlers()
	})
	if !called {
		t.Fatal("CoreService_ResetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_ResetHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.ResetHandlers()
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_ResetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_ResetHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.ResetHandlers()
		svc.ResetHandlers()
		_ = svc.Handlers()
	})
	if !called {
		t.Fatal("CoreService_ResetHandlers was not exercised")
	}
}

func TestCoreService_CoreService_Handlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		handlers := svc.Handlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("CoreService_Handlers was not exercised")
	}
}

func TestCoreService_CoreService_Handlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		handlers := svc.Handlers()
		if len(handlers) != 0 {
			t.Fatalf("got %d", len(handlers))
		}
	})
	if !called {
		t.Fatal("CoreService_Handlers was not exercised")
	}
}

func TestCoreService_CoreService_Handlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		handlers := svc.Handlers()
		handlers[0] = nil
		if svc.Handlers()[0] == nil {
			t.Fatal("handlers not copied")
		}
	})
	if !called {
		t.Fatal("CoreService_Handlers was not exercised")
	}
}

func TestCoreService_CoreService_CurrentHandlers_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		handlers := svc.CurrentHandlers()
		if len(handlers) == 0 {
			t.Fatal("expected handlers")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentHandlers was not exercised")
	}
}

func TestCoreService_CoreService_CurrentHandlers_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		handlers := svc.CurrentHandlers()
		if len(handlers) != 0 {
			t.Fatalf("got %d", len(handlers))
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentHandlers was not exercised")
	}
}

func TestCoreService_CoreService_CurrentHandlers_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		handlers := svc.CurrentHandlers()
		handlers[0] = nil
		if svc.CurrentHandlers()[0] == nil {
			t.Fatal("handlers not copied")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentHandlers was not exercised")
	}
}

func TestCoreService_CoreService_AvailableLanguages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		langs := svc.AvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("CoreService_AvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_AvailableLanguages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		langs := svc.AvailableLanguages()
		if len(langs) != 0 {
			t.Fatalf("got %v", langs)
		}
	})
	if !called {
		t.Fatal("CoreService_AvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_AvailableLanguages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		langs := svc.AvailableLanguages()
		langs[0] = "mutated"
		if svc.AvailableLanguages()[0] == "mutated" {
			t.Fatal("languages not copied")
		}
	})
	if !called {
		t.Fatal("CoreService_AvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_CurrentAvailableLanguages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		langs := svc.CurrentAvailableLanguages()
		if len(langs) == 0 {
			t.Fatal("expected languages")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentAvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_CurrentAvailableLanguages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		langs := svc.CurrentAvailableLanguages()
		if len(langs) != 0 {
			t.Fatalf("got %v", langs)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentAvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_CurrentAvailableLanguages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		langs := svc.CurrentAvailableLanguages()
		langs[0] = "mutated"
		if svc.CurrentAvailableLanguages()[0] == "mutated" {
			t.Fatal("languages not copied")
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentAvailableLanguages was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.PluralCategory(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_PluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.PluralCategory(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_PluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.PluralCategory(-1)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_PluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPluralCategory_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentPluralCategory(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentPluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPluralCategory_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.CurrentPluralCategory(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_CurrentPluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_CurrentPluralCategory_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.CurrentPluralCategory(-1)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_CurrentPluralCategory was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategoryOf_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.PluralCategoryOf(1)
		if got != PluralOne {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_PluralCategoryOf was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategoryOf_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		got := svc.PluralCategoryOf(1)
		if got != PluralOther {
			t.Fatalf("got %v", got)
		}
	})
	if !called {
		t.Fatal("CoreService_PluralCategoryOf was not exercised")
	}
}

func TestCoreService_CoreService_PluralCategoryOf_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		got := svc.PluralCategoryOf(-1)
		_ = got
	})
	if !called {
		t.Fatal("CoreService_PluralCategoryOf was not exercised")
	}
}

func TestCoreService_CoreService_AddMessages_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.AddMessages("en", map[string]string{"ax7.core": "ready"})
		if svc.T("ax7.core") != "ready" {
			t.Fatal("message not added")
		}
	})
	if !called {
		t.Fatal("CoreService_AddMessages was not exercised")
	}
}

func TestCoreService_CoreService_AddMessages_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		svc.AddMessages("en", nil)
		if svc != nil {
			t.Fatal("unexpected receiver")
		}
	})
	if !called {
		t.Fatal("CoreService_AddMessages was not exercised")
	}
}

func TestCoreService_CoreService_AddMessages_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		svc.AddMessages("en", map[string]string{})
		_ = svc.AvailableLanguages()
	})
	if !called {
		t.Fatal("CoreService_AddMessages was not exercised")
	}
}

func TestCoreService_CoreService_SetLanguage_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.SetLanguage("fr")
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("CoreService_SetLanguage was not exercised")
	}
}

func TestCoreService_CoreService_SetLanguage_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		err := svc.SetLanguage("fr")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_SetLanguage was not exercised")
	}
}

func TestCoreService_CoreService_SetLanguage_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.SetLanguage("zz")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_SetLanguage was not exercised")
	}
}

func TestCoreService_CoreService_AddLoader_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("CoreService_AddLoader was not exercised")
	}
}

func TestCoreService_CoreService_AddLoader_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		err := svc.AddLoader(NewFSLoader(ax7TestFS(), "locales"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_AddLoader was not exercised")
	}
}

func TestCoreService_CoreService_AddLoader_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.AddLoader(nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_AddLoader was not exercised")
	}
}

func TestCoreService_CoreService_LoadFS_Good(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.LoadFS(ax7TestFS(), "locales")
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if !called {
		t.Fatal("CoreService_LoadFS was not exercised")
	}
}

func TestCoreService_CoreService_LoadFS_Bad(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		var svc *CoreService
		err := svc.LoadFS(ax7TestFS(), "locales")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_LoadFS was not exercised")
	}
}

func TestCoreService_CoreService_LoadFS_Ugly(t *testing.T) {
	called := false
	ax7NoPanic(t, func() {
		called = true
		svc := ax7CoreService(t)
		err := svc.LoadFS(ax7TestFS(), "missing")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	if !called {
		t.Fatal("CoreService_LoadFS was not exercised")
	}
}
