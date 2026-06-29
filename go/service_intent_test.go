package i18n

import (
	"testing"

	"dappco.re/go"
)

// templateIntentLoader registers a language whose intent definitions carry
// explicit {{.Subject}} template fields. This drives resolveIntentFieldLocked
// down the renderIntentTemplate branch, which the Meta.Verb auto-composition
// path (exercised elsewhere) never reaches.
type templateIntentLoader struct{}

func (templateIntentLoader) Languages() []string { return []string{"qaa-x-tmpl"} }

func (templateIntentLoader) Load(string) core.Result {
	return localeLoadResult(map[string]Message{}, &GrammarData{
		Intents: map[string]Intent{
			"app.delete": {
				Meta:     IntentMeta{Type: "action"},
				Question: "Delete {{.Subject}}?",
				Confirm:  "Really delete {{.Subject}}?",
				Success:  "{{.Subject}} deleted",
				Failure:  "Failed to delete {{.Subject}}",
			},
		},
	})
}

// TestServiceIntent_renderIntentTemplate_Subject drives renderIntentTemplate via
// Compose with a *Subject, exercising the *Subject branch on every field.
//
//	svc.Compose("app.delete", S("file", "config.yaml")).Question // "Delete config.yaml?"
func TestServiceIntent_renderIntentTemplate_Subject(t *testing.T) {
	svc, err := serviceFromResult(NewWithLoader(templateIntentLoader{}, WithFallback("en")))
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	if err := errorFromResult(svc.SetLanguage("qaa-x-tmpl")); err != nil {
		t.Fatalf("SetLanguage failed: %v", err)
	}

	composed := svc.Compose("app.delete", S("file", "config.yaml"))
	if composed.Question != "Delete config.yaml?" {
		t.Errorf("Question = %q, want %q", composed.Question, "Delete config.yaml?")
	}
	if composed.Confirm != "Really delete config.yaml?" {
		t.Errorf("Confirm = %q, want %q", composed.Confirm, "Really delete config.yaml?")
	}
	if composed.Success != "config.yaml deleted" {
		t.Errorf("Success = %q, want %q", composed.Success, "config.yaml deleted")
	}
	if composed.Failure != "Failed to delete config.yaml" {
		t.Errorf("Failure = %q, want %q", composed.Failure, "Failed to delete config.yaml")
	}
}

// TestServiceIntent_renderIntentTemplate_Nil drives the nil-data branch of
// renderIntentTemplate: the template still renders, but {{.Subject}} resolves to
// empty so only the surrounding literal text survives.
func TestServiceIntent_renderIntentTemplate_Nil(t *testing.T) {
	svc, err := serviceFromResult(NewWithLoader(templateIntentLoader{}, WithFallback("en")))
	if err != nil {
		t.Fatalf("NewWithLoader() failed: %v", err)
	}
	if err := errorFromResult(svc.SetLanguage("qaa-x-tmpl")); err != nil {
		t.Fatalf("SetLanguage failed: %v", err)
	}

	composed := svc.Compose("app.delete", nil)
	if composed.Question != "Delete ?" {
		t.Errorf("Question (nil subject) = %q, want %q", composed.Question, "Delete ?")
	}
}

// TestServiceIntent_renderIntentTemplate_Default exercises renderIntentTemplate's
// default (non-Subject, non-nil) data branch directly with a map, routing
// through applyTemplate.
func TestServiceIntent_renderIntentTemplate_Default(t *testing.T) {
	got := renderIntentTemplate("Delete {{.Name}}?", map[string]any{"Name": "config.yaml"})
	if got != "Delete config.yaml?" {
		t.Errorf("renderIntentTemplate(map) = %q, want %q", got, "Delete config.yaml?")
	}
}

// TestServiceIntent_renderIntentTemplate_Ugly confirms the empty-text guard
// short-circuits before any template parsing.
func TestServiceIntent_renderIntentTemplate_Ugly(t *testing.T) {
	if got := renderIntentTemplate("", S("file", "x")); got != "" {
		t.Errorf("renderIntentTemplate(empty) = %q, want empty", got)
	}
	if got := renderIntentTemplate("", nil); got != "" {
		t.Errorf("renderIntentTemplate(empty, nil) = %q, want empty", got)
	}
}
