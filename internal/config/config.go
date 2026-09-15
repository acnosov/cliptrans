package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	NotifyOnError   *bool   `yaml:"notify_on_error,omitempty"`
	NotifyOnSuccess *bool   `yaml:"notify_on_success,omitempty"`
	BaseURL         string  `yaml:"base_url"`
	Model           string  `yaml:"model"`
	APIKeyFallback  string  `yaml:"api_key,omitempty"`
	To              string  `yaml:"to,omitempty"`
	SystemPrompt    string  `yaml:"system_prompt,omitempty"`
	Temperature     float64 `yaml:"temperature"`
	MaxChars        int     `yaml:"max_chars"`
}

const (
	DefaultMaxChars    = 8000
	DefaultTemperature = 0
	DefaultTarget      = "en"
	TargetPlaceholder  = "{{target}}"
	maxTargetLength    = 32
	configDirPerm      = 0o755
	configFilePerm     = 0o600
)

const KeyEnvName = "CLIPTRANS_API_KEY"

func Load(path string) (*Config, error) {
	path, err := resolvePath(path)
	if err != nil {
		return nil, err
	}
	raw, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return decode(path, raw)
}

func resolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	resolved, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func readFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no config at %s — create it or pass --config <path>: %w", path, err)
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	return raw, nil
}

func decode(path string, raw []byte) (*Config, error) {
	cfg := &Config{
		Temperature: DefaultTemperature,
		MaxChars:    DefaultMaxChars,
	}
	if err := yaml.Load(raw, cfg, yaml.WithSingleDocument()); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.MaxChars <= 0 {
		return nil, fmt.Errorf("invalid config %s: max_chars must be > 0 (got %d)", path, cfg.MaxChars)
	}
	if err := applyTargetDefault(cfg, path); err != nil {
		return nil, err
	}
	defaultPrompts(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return cfg, nil
}

func applyTargetDefault(cfg *Config, path string) error {
	target, err := NormalizeTarget(cfg.To)
	if err != nil {
		return fmt.Errorf("invalid config %s: bad `to`: %w", path, err)
	}
	cfg.To = target
	return nil
}

func NormalizeTarget(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return DefaultTarget, nil
	}
	if len(trimmed) > maxTargetLength {
		return "", fmt.Errorf("invalid language %q: too long", target)
	}
	return trimmed, nil
}

func defaultPrompts(cfg *Config) {
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = DefaultPrompt
	}
}

func (c *Config) Validate() error {
	if err := validateEndpoint(c); err != nil {
		return err
	}
	if err := validateTarget(c); err != nil {
		return err
	}
	return validatePrompts(c)
}

func validateEndpoint(c *Config) error {
	if c.BaseURL == "" {
		return errors.New("missing required field `base_url`")
	}
	if c.Model == "" {
		return errors.New("missing required field `model`")
	}
	return nil
}

func validateTarget(c *Config) error {
	if strings.TrimSpace(c.To) == "" {
		return errors.New("missing required field `to`")
	}
	return nil
}

func validatePrompts(c *Config) error {
	if c.SystemPrompt == "" {
		return errors.New("missing required field `system_prompt`")
	}
	return nil
}

func (c *Config) RenderPrompt(target string) string {
	return strings.ReplaceAll(c.SystemPrompt, TargetPlaceholder, LanguageLabel(target))
}

func (c *Config) NotifyOnErrorEnabled() bool {
	return c.NotifyOnError == nil || *c.NotifyOnError
}

func (c *Config) NotifyOnSuccessEnabled() bool {
	return c.NotifyOnSuccess == nil || *c.NotifyOnSuccess
}

func (c *Config) APIKey() string {
	if k := os.Getenv(KeyEnvName); k != "" {
		return k
	}
	return c.APIKeyFallback
}
