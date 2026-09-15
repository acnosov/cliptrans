package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		target string
		want   string
	}{
		{"en", "English (en)"},
		{"ru", "Russian (ru)"},
		{"uk", "Ukrainian (uk)"},
		{"de", "German (de)"},
		{"EN", "English (EN)"},
		{" uk ", "Ukrainian (uk)"},
		{"en-US", "English (en-US)"},
		{"pt_BR", "Portuguese (pt_BR)"},
		{"xx", "xx"},
		{"klingon", "klingon"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, LanguageLabel(tc.target))
		})
	}
}

func TestRenderPromptExpandsTargetLabel(t *testing.T) {
	t.Parallel()

	cfg := &Config{SystemPrompt: "translate into {{target}}, {{target}}!"}
	require.Equal(t,
		"translate into Ukrainian (uk), Ukrainian (uk)!",
		cfg.RenderPrompt("uk"))
	require.Equal(t,
		"translate into xx, xx!",
		(&Config{SystemPrompt: "translate into {{target}}, {{target}}!"}).RenderPrompt("xx"))
}
