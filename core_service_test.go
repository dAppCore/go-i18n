package i18n

import (
	"testing"

	"dappco.re/go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoreServiceNilSafe(t *testing.T) {
	var svc *CoreService

	assert.NotPanics(t, func() {
		assert.Equal(t, ModeNormal, svc.Mode())
		assert.Equal(t, "en", svc.Language())
		assert.Equal(t, "en", svc.Fallback())
		assert.Equal(t, FormalityNeutral, svc.Formality())
		assert.Equal(t, "", svc.Location())
		assert.False(t, svc.Debug())
		assert.Equal(t, DirLTR, svc.Direction())
		assert.False(t, svc.IsRTL())
		assert.Equal(t, PluralOther, svc.PluralCategory(3))
		assert.Empty(t, svc.AvailableLanguages())
		assert.Empty(t, svc.Handlers())
		assert.Equal(t, "prompt.confirm", svc.Prompt("confirm"))
		assert.Equal(t, "lang.fr", svc.Lang("fr"))
		assert.Equal(t, "hello", svc.T("hello"))
		assert.Equal(t, "hello", svc.Raw("hello"))
		assert.Equal(t, core.Result{Value: "hello", OK: false}, svc.Translate("hello"))
		assert.Equal(t, State(), svc.State())
		assert.Equal(t, State(), svc.CurrentState())
	})

	assert.NoError(t, svc.OnStartup(nil))
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

	require.ErrorIs(t, svc.SetLanguage("fr"), ErrServiceNotInitialised)
	require.ErrorIs(t, svc.AddLoader(nil), ErrServiceNotInitialised)
	require.ErrorIs(t, svc.LoadFS(nil, "locales"), ErrServiceNotInitialised)
}

func TestCoreServiceMissingKeysReturnsCopies(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	coreSvc := &CoreService{svc: svc}

	coreSvc.SetMode(ModeCollect)
	_ = svc.T("missing.copy.key", map[string]any{"foo": "bar"})

	missing := coreSvc.MissingKeys()
	require.Len(t, missing, 1)
	require.Equal(t, "bar", missing[0].Args["foo"])

	missing[0].Args["foo"] = "mutated"

	again := coreSvc.MissingKeys()
	require.Len(t, again, 1)
	assert.Equal(t, "bar", again[0].Args["foo"])
}
