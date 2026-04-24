package i18n

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"dappco.re/go/core"
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
			if (core.Result{Value: "hello", OK: false}) != (svc.Translate("hello")) {
				t.Fatalf("want %v, got %v", core.Result{Value: "hello", OK: false}, svc.Translate("hello"))
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
	if (core.Result{OK: true}) != (svc.OnStartup(nil)) {
		t.Fatalf("want %v, got %v", core.Result{OK: true}, svc.OnStartup(nil))
	}
	if (core.Result{OK: true}) != (svc.OnShutdown(nil)) {
		t.Fatalf("want %v, got %v", core.Result{OK: true}, svc.OnShutdown(nil))
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
