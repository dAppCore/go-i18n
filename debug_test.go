package i18n

import (
	"testing"
)

func TestSetDebug_Good_PackageLevel(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ensure Init() has run first, then override with our service
	_ = Init()
	SetDefault(svc)

	SetDebug(true)
	if !(svc.Debug()) {
		t.Fatal("expected true")
	}

	SetDebug(false)
	if svc.Debug() {
		t.Fatal("expected false")
	}
}

func TestSetDebug_Good_ServiceLevel(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.SetDebug(true)
	if !(svc.Debug()) {
		t.Fatal("expected true")
	}

	svc.SetDebug(false)
	if svc.Debug() {
		t.Fatal("expected false")
	}
}

func TestCurrentDebug_Good_PackageLevel(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = Init()
	SetDefault(svc)

	SetDebug(true)
	if !(CurrentDebug()) {
		t.Fatal("expected true")
	}

	SetDebug(false)
	if CurrentDebug() {
		t.Fatal("expected false")
	}
}

func TestDebugFormat_Good(t *testing.T) {
	tests := []struct {
		name string
		key  string
		text string
		want string
	}{
		{"simple", "greeting", "Hello", "[greeting] Hello"},
		{"dotted_key", "i18n.label.status", "Status:", "[i18n.label.status] Status:"},
		{"empty_text", "key", "", "[key] "},
		{"empty_key", "", "text", "[] text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := debugFormat(tt.key, tt.text)
			if tt.want != got {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDebugMode_Good_Integration(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	svc.SetDebug(true)
	defer svc.SetDebug(false)

	// T() should wrap output in debug format
	got := svc.T("prompt.yes")
	if "[prompt.yes] y" != got {
		t.Fatalf("want %v, got %v", "[prompt.yes] y", got)
	}

	got = svc.Raw("prompt.yes")
	if "[prompt.yes] y" != got {
		t.Fatalf("want %v, got %v", "[prompt.yes] y", got)
	}
}

func TestTranslate_DebugMode_PreservesOK(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	SetDefault(svc)

	svc.SetDebug(true)
	defer svc.SetDebug(false)

	translated := svc.Translate("prompt.yes")
	if !(translated.OK) {
		t.Fatal("expected true")
	}
	if "[prompt.yes] y" != translated.Value {
		t.Fatalf("want %v, got %v", "[prompt.yes] y", translated.Value)
	}

	missing := svc.Translate("missing.translation.key")
	if missing.OK {
		t.Fatal("expected false")
	}
	if "[missing.translation.key] missing.translation.key" != missing.Value {
		t.Fatalf("want %v, got %v", "[missing.translation.key] missing.translation.key", missing.Value)
	}
}
