//go:build integration

package translate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranslateLive(t *testing.T) {
	t.Parallel()
	dotenv := readDotEnv(t)
	baseURL := firstNonEmpty(os.Getenv("CLIPTRANS_BASE_URL"), dotenv["CLIPTRANS_BASE_URL"])
	apiKey := firstNonEmpty(os.Getenv("CLIPTRANS_API_KEY"), dotenv["CLIPTRANS_API_KEY"])
	model := firstNonEmpty(os.Getenv("CLIPTRANS_MODEL"), dotenv["CLIPTRANS_MODEL"], "kilo-free")
	if baseURL == "" || apiKey == "" {
		t.Fatal("CLIPTRANS_BASE_URL and CLIPTRANS_API_KEY must be set (env or repo-root .env)")
	}
	client := &Client{
		BaseURL:      baseURL,
		Model:        model,
		APIKey:       apiKey,
		SystemPrompt: "Translate the following text into German. Return only the translation.",
	}
	out, err := client.Translate(t.Context(), "Hello, how are you?")
	require.NoError(t, err)
	require.NotEmpty(t, strings.TrimSpace(out))
	t.Logf("model %s translated to %q", model, out)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type dotenvEntry struct {
	key   string
	value string
}

func readDotEnv(t *testing.T) map[string]string {
	t.Helper()
	values := map[string]string{}
	path := findDotEnv(t)
	if path == "" {
		return values
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	for line := range strings.Lines(string(raw)) {
		entry, ok := parseDotEnvLine(line)
		if !ok {
			continue
		}
		values[entry.key] = entry.value
	}
	return values
}

func findDotEnv(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for range 20 {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

func parseDotEnvLine(line string) (dotenvEntry, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return dotenvEntry{}, false
	}
	key, value, found := strings.Cut(trimmed, "=")
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if !found || key == "" {
		return dotenvEntry{}, false
	}
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if first == last && (first == '"' || first == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return dotenvEntry{key: key, value: value}, true
}
