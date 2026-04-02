package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDebug_Good_PackageLevel(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	// Ensure Init() has run first, then override with our service
	_ = Init()
	SetDefault(svc)

	SetDebug(true)
	assert.True(t, svc.Debug())

	SetDebug(false)
	assert.False(t, svc.Debug())
}

func TestSetDebug_Good_ServiceLevel(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	svc.SetDebug(true)
	assert.True(t, svc.Debug())

	svc.SetDebug(false)
	assert.False(t, svc.Debug())
}

func TestCurrentDebug_Good_PackageLevel(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)

	_ = Init()
	SetDefault(svc)

	SetDebug(true)
	assert.True(t, CurrentDebug())

	SetDebug(false)
	assert.False(t, CurrentDebug())
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDebugMode_Good_Integration(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	svc.SetDebug(true)
	defer svc.SetDebug(false)

	// T() should wrap output in debug format
	got := svc.T("prompt.yes")
	assert.Equal(t, "[prompt.yes] y", got)

	// Raw() should also wrap output in debug format
	got = svc.Raw("prompt.yes")
	assert.Equal(t, "[prompt.yes] y", got)
}

func TestTranslate_DebugMode_PreservesOK(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	svc.SetDebug(true)
	defer svc.SetDebug(false)

	translated := svc.Translate("prompt.yes")
	assert.True(t, translated.OK)
	assert.Equal(t, "[prompt.yes] y", translated.Value)

	missing := svc.Translate("missing.translation.key")
	assert.False(t, missing.OK)
	assert.Equal(t, "[missing.translation.key] missing.translation.key", missing.Value)
}
