package i18n

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	prevDefault := Default()
	SetDefault(svc)
	t.Cleanup(func() {
		SetDefault(prevDefault)
	})

	require.NoError(t, svc.SetLanguage("en"))

	subj := S("file", "config.yaml")

	assert.Equal(t, "y", Prompt("yes"))
	assert.Equal(t, "y", svc.Prompt("yes"))
	assert.Equal(t, "deleted", PastTense("delete"))
	assert.Equal(t, "files", Pluralize("file", 2))
	assert.Equal(t, "an", Article("item"))

	composed := Compose("core.delete", subj)
	assert.Equal(t, "Delete config.yaml?", composed.Question)
	assert.Equal(t, "Really delete config.yaml?", composed.Confirm)
	assert.Equal(t, "Config.yaml deleted", composed.Success)
	assert.Equal(t, "Failed to delete config.yaml", composed.Failure)
	assert.Equal(t, IntentMeta{
		Type:      "action",
		Verb:      "common.verb.delete",
		Dangerous: true,
		Default:   "no",
		Supports:  []string{"all", "skip"},
	}, composed.Meta)

	assert.Equal(t, composed, svc.Compose("core.delete", subj))
	assert.Equal(t, "Delete config.yaml?", T("core.delete", subj))
	assert.Equal(t, "Delete config.yaml?", svc.T("core.delete", subj))

	data := GetGrammarData("en")
	require.NotNil(t, data)
	intent, ok := data.Intents["core.delete"]
	require.True(t, ok)
	assert.Equal(t, composed.Meta, intent.Meta)
	assert.Equal(t, "Delete {{.Subject}}?", intent.Question)
	assert.Equal(t, "Really delete {{.Subject}}?", intent.Confirm)
}
