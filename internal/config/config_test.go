package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestLoadEmptyPathUsesDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := Load(""); err == nil {
		t.Error("expected error for missing default config, got nil")
	}
}

func TestLoadNonPositiveMaxCharsRejected(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"0", "-5"} {
		path := writeTempConfig(
			t,
			"base_url: https://example.com/v1\nmodel: m1\nmax_chars: "+value+"\n",
		)
		if _, err := Load(path); err == nil {
			t.Errorf("expected error for max_chars: %s, got nil", value)
		}
	}
}

func TestValidateMissingBaseURL(t *testing.T) {
	t.Parallel()
	cfg := &Config{Model: "m1", MaxChars: DefaultMaxChars}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing base_url, got nil")
	}
	cfg = &Config{BaseURL: "https://example.com/v1", MaxChars: DefaultMaxChars}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing model, got nil")
	}
}

func TestDefaultPathNonEmpty(t *testing.T) {
	t.Parallel()
	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("expected non-empty default path")
	}
}

func TestLoadMissing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "missing.yaml")
	if _, err := Load(path); err == nil {
		t.Error("expected error for missing config, got nil")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Parallel()
	path := writeTempConfig(t, "base_url: https://example.com/v1\nmodel: m1\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxChars != DefaultMaxChars {
		t.Errorf("MaxChars = %d, want %d", cfg.MaxChars, DefaultMaxChars)
	}
	if cfg.Temperature != DefaultTemperature {
		t.Errorf("Temperature = %v, want %v", cfg.Temperature, DefaultTemperature)
	}
}

func TestLoadInvalid(t *testing.T) {
	t.Parallel()
	path := writeTempConfig(t, "base_url: https://example.com/v1\n")
	if _, err := Load(path); err == nil {
		t.Error("expected error for missing model, got nil")
	}
	path = writeTempConfig(t, "not: [valid, yaml\n")
	if _, err := Load(path); err == nil {
		t.Error("expected error for malformed yaml, got nil")
	}
}

func TestEnsureDefaultCreatesValid(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, created, err := EnsureDefault()
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("expected EnsureDefault to create the file on first call")
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after EnsureDefault: %v", err)
	}
	if cfg.Model == "" || cfg.BaseURL == "" {
		t.Error("expected auto-created config to validate")
	}
	if _, created, err := EnsureDefault(); err != nil {
		t.Fatal(err)
	} else if created {
		t.Error("expected second EnsureDefault to report created=false")
	}
}

func TestAPIKeyEnvWins(t *testing.T) {
	path := writeTempConfig(t, "base_url: x\nmodel: y\napi_key: fallback\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(KeyEnvName, "env-key")
	if got := cfg.APIKey(); got != "env-key" {
		t.Errorf("APIKey = %q, want env-key", got)
	}
}

func TestAPIKeyFallback(t *testing.T) {
	path := writeTempConfig(t, "base_url: x\nmodel: y\napi_key: fallback\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(KeyEnvName, "")
	if got := cfg.APIKey(); got != "fallback" {
		t.Errorf("APIKey = %q, want fallback", got)
	}
}
