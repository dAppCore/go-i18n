package i18n

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestRFCIntentAndCommonAliases(t *testing.T) {
	loaderFS := fstest.MapFS{
		"en.json": &fstest.MapFile{
			Data: []byte(`{
				"common": {
					"verb": {
						"delete": { "base": "delete", "past": "deleted", "gerund": "deleting" }
					},
					"noun": {
						"file": { "one": "file", "other": "files" }
					},
					"article": {
						"indefinite": { "default": "a", "vowel": "an" },
						"definite": "the"
					},
					"prompt": {
						"yes": "y",
						"no": "n"
					}
				},
				"core": {
					"delete": {
						"_meta": {
							"type": "action",
							"verb": "common.verb.delete",
							"dangerous": true,
							"default": "no",
							"supports": ["all", "skip"]
						},
						"question": "Delete {{.Subject}}?",
						"confirm": "Really delete {{.Subject}}?",
						"success": "{{.Subject | title}} deleted",
						"failure": "Failed to delete {{.Subject}}"
					}
				}
			}`),
		},
	}

	svc, err := NewWithLoader(NewFSLoader(loaderFS, "."))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prevDefault := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prevDefault)
	})
	if err := svc.SetLanguage("en"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subj := S("file", "config.yaml")
	if ("y") != (Prompt("yes")) {
		t.Fatalf("want %v, got %v", "y", Prompt("yes"))
	}
	if ("y") != (svc.Prompt("yes")) {
		t.Fatalf("want %v, got %v", "y", svc.Prompt("yes"))
	}
	if ("deleted") != (PastTense("delete")) {
		t.Fatalf("want %v, got %v", "deleted", PastTense("delete"))
	}
	if ("files") != (Pluralize("file", 2)) {
		t.Fatalf("want %v, got %v", "files", Pluralize("file", 2))
	}
	if ("an") != (Article("item")) {
		t.Fatalf("want %v, got %v", "an", Article("item"))
	}

	composed := Compose("core.delete", subj)
	if ("Delete config.yaml?") != (composed.Question) {
		t.Fatalf("want %v, got %v", "Delete config.yaml?", composed.Question)
	}
	if ("Really delete config.yaml?") != (composed.Confirm) {
		t.Fatalf("want %v, got %v", "Really delete config.yaml?", composed.Confirm)
	}
	if ("Config.yaml deleted") != (composed.Success) {
		t.Fatalf("want %v, got %v", "Config.yaml deleted", composed.Success)
	}
	if ("Failed to delete config.yaml") != (composed.Failure) {
		t.Fatalf("want %v, got %v", "Failed to delete config.yaml", composed.Failure)
	}
	if !reflect.DeepEqual(IntentMeta{
		Type:      "action",
		Verb:      "common.verb.delete",
		Dangerous: true,
		Default:   "no",
		Supports:  []string{"all", "skip"},
	}, composed.Meta) {
		t.Fatalf("want %v, got %v", IntentMeta{
			Type:      "action",
			Verb:      "common.verb.delete",
			Dangerous: true,
			Default:   "no",
			Supports:  []string{"all", "skip"},
		}, composed.Meta)
	}
	if !reflect.DeepEqual(composed, svc.Compose("core.delete", subj)) {
		t.Fatalf("want %v, got %v", composed, svc.Compose("core.delete", subj))
	}
	if ("Delete config.yaml?") != (T("core.delete", subj)) {
		t.Fatalf("want %v, got %v", "Delete config.yaml?", T("core.delete", subj))
	}
	if ("Delete config.yaml?") != (svc.T("core.delete", subj)) {
		t.Fatalf("want %v, got %v", "Delete config.yaml?", svc.T("core.delete", subj))
	}

	data := GetGrammarData("en")
	if (data) == (nil) {
		t.Fatalf("expected non-nil")
	}
	intent, ok := data.Intents["core.delete"]
	if !(ok) {
		t.Fatal("expected true")
	}
	if !reflect.DeepEqual(composed.Meta, intent.Meta) {
		t.Fatalf("want %v, got %v", composed.Meta, intent.Meta)
	}
	if ("Delete {{.Subject}}?") != (intent.Question) {
		t.Fatalf("want %v, got %v", "Delete {{.Subject}}?", intent.Question)
	}
	if ("Really delete {{.Subject}}?") != (intent.Confirm) {
		t.Fatalf("want %v, got %v", "Really delete {{.Subject}}?", intent.Confirm)
	}
}
