package i18n

import "testing"

func TestLabelHandler(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	h := LabelHandler{}

	if !h.Match("i18n.label.status") {
		t.Error("LabelHandler should match i18n.label.*")
	}
	if h.Match("other.key") {
		t.Error("LabelHandler should not match other.key")
	}

	got := h.Handle("i18n.label.status", nil, nil)
	if got != "Status:" {
		t.Errorf("LabelHandler.Handle(status) = %q, want %q", got, "Status:")
	}
}

func TestProgressHandler(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	h := ProgressHandler{}

	if !h.Match("i18n.progress.build") {
		t.Error("ProgressHandler should match i18n.progress.*")
	}

	// Without subject
	got := h.Handle("i18n.progress.build", nil, nil)
	if got != "Building..." {
		t.Errorf("ProgressHandler.Handle(build) = %q, want %q", got, "Building...")
	}

	// With subject
	got = h.Handle("i18n.progress.build", []any{"project"}, nil)
	if got != "Building project..." {
		t.Errorf("ProgressHandler.Handle(build, project) = %q, want %q", got, "Building project...")
	}
}

func TestCountHandler(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	h := CountHandler{}

	if !h.Match("i18n.count.file") {
		t.Error("CountHandler should match i18n.count.*")
	}

	tests := []struct {
		key  string
		args []any
		want string
	}{
		{"i18n.count.file", []any{1}, "1 file"},
		{"i18n.count.file", []any{5}, "5 files"},
		{"i18n.count.file", []any{0}, "0 files"},
		{"i18n.count.child", []any{3}, "3 children"},
		{"i18n.count.url", []any{2}, "2 URLs"},
		{"i18n.count.api", []any{2}, "2 APIs"},
		{"i18n.count.cpus", []any{2}, "2 CPUs"},
		{"i18n.count.file", nil, "file"},
		{"i18n.count.url", nil, "URL"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := h.Handle(tt.key, tt.args, nil)
			if got != tt.want {
				t.Errorf("CountHandler.Handle(%q, %v) = %q, want %q", tt.key, tt.args, got, tt.want)
			}
		})
	}
}

func TestDoneHandler(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	h := DoneHandler{}

	if !h.Match("i18n.done.delete") {
		t.Error("DoneHandler should match i18n.done.*")
	}

	// With subject
	got := h.Handle("i18n.done.delete", []any{"config.yaml"}, nil)
	if got != "Config.yaml deleted" {
		t.Errorf("DoneHandler.Handle(delete, config.yaml) = %q, want %q", got, "Config.yaml deleted")
	}

	// Without subject — just past tense
	got = h.Handle("i18n.done.delete", nil, nil)
	if got != "Deleted" {
		t.Errorf("DoneHandler.Handle(delete) = %q, want %q", got, "Deleted")
	}
}

func TestFailHandler(t *testing.T) {
	h := FailHandler{}

	if !h.Match("i18n.fail.push") {
		t.Error("FailHandler should match i18n.fail.*")
	}

	got := h.Handle("i18n.fail.push", []any{"commits"}, nil)
	if got != "Failed to push commits" {
		t.Errorf("FailHandler.Handle(push, commits) = %q, want %q", got, "Failed to push commits")
	}

	got = h.Handle("i18n.fail.push", nil, nil)
	if got != "Failed to push" {
		t.Errorf("FailHandler.Handle(push) = %q, want %q", got, "Failed to push")
	}
}

func TestNumericHandler(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)

	h := NumericHandler{}

	tests := []struct {
		key  string
		args []any
		want string
	}{
		{"i18n.numeric.number", []any{int64(1234567)}, "1,234,567"},
		{"i18n.numeric.ordinal", []any{1}, "1st"},
		{"i18n.numeric.ordinal", []any{2}, "2nd"},
		{"i18n.numeric.ordinal", []any{3}, "3rd"},
		{"i18n.numeric.ordinal", []any{11}, "11th"},
		{"i18n.numeric.percent", []any{0.85}, "85%"},
		{"i18n.numeric.bytes", []any{int64(1536000)}, "1.5 MB"},
		{"i18n.numeric.ago", []any{5, "minutes"}, "5 minutes ago"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := h.Handle(tt.key, tt.args, func() string { return "fallback" })
			if got != tt.want {
				t.Errorf("NumericHandler.Handle(%q, %v) = %q, want %q", tt.key, tt.args, got, tt.want)
			}
		})
	}

	// No args falls through to next
	got := h.Handle("i18n.numeric.number", nil, func() string { return "fallback" })
	if got != "fallback" {
		t.Errorf("NumericHandler with no args should fallback, got %q", got)
	}
}

func TestCountHandler_UsesLocaleNumberFormat(t *testing.T) {
	prev := Default()
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prev)
	})

	if err := SetLanguage("fr"); err != nil {
		t.Fatalf("SetLanguage(fr) failed: %v", err)
	}

	h := CountHandler{}
	got := h.Handle("i18n.count.file", []any{1234}, nil)
	want := "1 234 files"
	if got != want {
		t.Errorf("CountHandler.Handle(locale format) = %q, want %q", got, want)
	}
}

func TestRunHandlerChain(t *testing.T) {
	handlers := DefaultHandlers()
	fallback := func() string { return "missed" }

	// Label handler catches it
	got := RunHandlerChain(handlers, "i18n.label.status", nil, fallback)
	if got != "Status:" {
		t.Errorf("chain label = %q, want %q", got, "Status:")
	}

	// Non-matching key falls through to fallback
	got = RunHandlerChain(handlers, "some.other.key", nil, fallback)
	if got != "missed" {
		t.Errorf("chain miss = %q, want %q", got, "missed")
	}
}

func TestDefaultHandlers(t *testing.T) {
	handlers := DefaultHandlers()
	if len(handlers) != 6 {
		t.Errorf("DefaultHandlers() returned %d handlers, want 6", len(handlers))
	}
}
