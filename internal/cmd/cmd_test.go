package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const translateName = "translate"

func TestCheckLimit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		chars   int
		limit   int
		wantErr bool
	}{
		{"under limit", 5, 8000, false},
		{"at limit", 8000, 8000, false},
		{"over limit", 8001, 8000, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := checkLimit(tc.chars, tc.limit); (err != nil) != tc.wantErr {
				t.Errorf(
					"checkLimit(%d, %d) error = %v, wantErr %v",
					tc.chars,
					tc.limit,
					err,
					tc.wantErr,
				)
			}
		})
	}
}

func TestDetailOf(t *testing.T) {
	t.Parallel()
	if got := detailOf(errors.New("boom"), "summary"); got != "boom" {
		t.Errorf("detailOf(err) = %q, want boom", got)
	}
	if got := detailOf(nil, "summary"); got != "summary" {
		t.Errorf("detailOf(nil) = %q, want summary", got)
	}
}

func TestPrintDryRun(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		target string
		want   string
	}{
		{"english", "en", "target: en\nmodel: m1\nchars: 5\n"},
		{"ukrainian", "uk", "target: uk\nmodel: m1\nchars: 5\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := NewRoot(t.Context())
			var buf bytes.Buffer
			root.SetOut(&buf)
			translateCmd, _, err := root.Find([]string{translateName})
			if err != nil {
				t.Fatal(err)
			}
			translateCmd.SetOut(&buf)
			printDryRun(translateCmd, tc.target, "m1", 5)
			if buf.String() != tc.want {
				t.Errorf("printDryRun output = %q, want %q", buf.String(), tc.want)
			}
		})
	}
}

func TestPreview(t *testing.T) {
	t.Parallel()
	if got := preview("Hola"); got != "Hola" {
		t.Errorf("preview(short) = %q, want Hola", got)
	}
	long := strings.Repeat("ы", successPreviewLimit+1)
	got := preview(long)
	if got != strings.Repeat("ы", successPreviewLimit)+"…" {
		t.Errorf(
			"preview(long) has len %d, want %d + ellipsis",
			len([]rune(got)),
			successPreviewLimit,
		)
	}
}

func TestNewTranslateCmdFlags(t *testing.T) {
	t.Parallel()
	var configFile string
	cmd := newTranslateCmd(t.Context(), &configFile)
	if cmd.Use != translateName {
		t.Errorf("Use = %q, want %q", cmd.Use, translateName)
	}
	for _, flag := range []string{"to", "model", "dry-run"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag %q to exist", flag)
		}
	}
	if cmd.Flags().Lookup("from") != nil {
		t.Error(`unexpected flag "from" to exist`)
	}
}

func TestNewRootSubcommands(t *testing.T) {
	t.Parallel()
	root := NewRoot(t.Context())
	found := map[string]bool{}
	for _, sub := range root.Commands() {
		found[sub.Name()] = true
	}
	for _, want := range []string{translateName, "version"} {
		if !found[want] {
			t.Errorf("expected subcommand %q to be registered", want)
		}
	}
}

func TestBareRootPrintsHelp(t *testing.T) {
	t.Parallel()
	root := NewRoot(t.Context())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bare root Execute error = %v", err)
	}
	if !strings.Contains(buf.String(), translateName) {
		t.Errorf("bare root help missing translate, got %q", buf.String())
	}
}

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	root := NewRoot(t.Context())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version Execute error = %v", err)
	}
	if strings.TrimSpace(buf.String()) != Version {
		t.Errorf("version output = %q, want %q", buf.String(), Version)
	}
}

func TestLoadEffectiveConfigExplicitPath(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	if _, err := loadEffectiveConfig(missing); err == nil {
		t.Error("expected error for missing explicit config, got nil")
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: https://example.com/v1\nmodel: m1\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadEffectiveConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "m1" {
		t.Errorf("Model = %q, want m1", cfg.Model)
	}
}
