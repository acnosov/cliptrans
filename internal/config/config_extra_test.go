package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultYAMLLoads(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(DefaultYAML()), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.NotEmpty(t, cfg.BaseURL)
	require.NotEmpty(t, cfg.Model)
	require.Equal(t, DefaultMaxChars, cfg.MaxChars)
	require.Equal(t, DefaultTarget, cfg.To)
	require.True(t, cfg.NotifyOnSuccessEnabled())
	require.True(t, cfg.NotifyOnErrorEnabled())
	require.NotEmpty(t, cfg.SystemPrompt)
	require.Contains(t, cfg.SystemPrompt, TargetPlaceholder)
}

func TestLoadTargetNormalization(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\nto: \"  uk \"\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "uk", cfg.To)
}

func TestLoadEmptyTargetDefaults(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, DefaultTarget, cfg.To)
}

func TestLoadLegacyFromIgnored(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\nfrom: en\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, DefaultTarget, cfg.To)
}

func TestLoadInvalidTarget(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\nto: \"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\"\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	_, err := Load(path)
	require.Error(t, err)
}

func TestNormalizeTarget(t *testing.T) {
	t.Parallel()

	got, err := NormalizeTarget("")
	require.NoError(t, err)
	require.Equal(t, DefaultTarget, got)

	got, err = NormalizeTarget("  uk ")
	require.NoError(t, err)
	require.Equal(t, "uk", got)

	_, err = NormalizeTarget("   ")
	require.NoError(t, err)

	_, err = NormalizeTarget("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	require.Error(t, err)
}

func TestPromptsDefaultWhenMissing(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(
		t,
		os.WriteFile(path, []byte("base_url: https://example.com/v1\nmodel: m1\n"), 0o600),
	)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, DefaultPrompt, cfg.SystemPrompt)
}

func TestCustomPromptKept(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\nsystem_prompt: translate into {{target}}, please\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "translate into {{target}}, please", cfg.SystemPrompt)
	require.Equal(t, "translate into Ukrainian (uk), please", cfg.RenderPrompt("uk"))
}

func TestNotifyFlags(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\nnotify_on_success: true\nnotify_on_error: false\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	require.True(t, cfg.NotifyOnSuccessEnabled())
	require.False(t, cfg.NotifyOnErrorEnabled())
}

func TestNotifyOnSuccessDefaultsEnabled(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(
		t,
		os.WriteFile(path, []byte("base_url: https://example.com/v1\nmodel: m1\n"), 0o600),
	)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.True(t, cfg.NotifyOnSuccessEnabled())
}

func TestNotifyOnErrorDefaultsEnabled(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(
		t,
		os.WriteFile(path, []byte("base_url: https://example.com/v1\nmodel: m1\n"), 0o600),
	)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.True(t, cfg.NotifyOnErrorEnabled())
}

func TestValidateSuccess(t *testing.T) {
	t.Parallel()

	cfg := &Config{BaseURL: "https://llm.example/v1", Model: "m1", To: "uk", SystemPrompt: "prompt"}
	require.NoError(t, cfg.Validate())
}

func TestLoadDirectoryPath(t *testing.T) {
	t.Parallel()

	_, err := Load(t.TempDir())
	require.ErrorContains(t, err, "read config")
}

func TestAPIKeyEmpty(t *testing.T) {
	t.Setenv(KeyEnvName, "")

	cfg := &Config{BaseURL: "https://llm.example/v1", Model: "m1"}
	require.Empty(t, cfg.APIKey())
}
