package i18n

import "testing"

func TestMapTokenToDomain(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"technical", "technical"},
		{"Technical", "technical"},
		{"tech", "technical"},
		{"creative", "creative"},
		{"Creative", "creative"},
		{"cre", "creative"},
		{"ethical", "ethical"},
		{"Ethical", "ethical"},
		{"eth", "ethical"},
		{"casual", "casual"},
		{"Casual", "casual"},
		{"cas", "casual"},
		{"unknown", "unknown"},
		{"", "unknown"},
		{"foo", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := mapTokenToDomain(tt.token)
			if got != tt.want {
				t.Errorf("mapTokenToDomain(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}
