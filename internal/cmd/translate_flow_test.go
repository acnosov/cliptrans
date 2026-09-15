package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acnosov/cliptrans/internal/config"

	"github.com/stretchr/testify/require"
)

const (
	flowConfigFlag = "--config"
	flowDryRunFlag = "--dry-run"
	flowToFlag     = "--to"
	flowModelFlag  = "--model"

	flowPasteName = "wl-paste"
	flowCopyName  = "wl-copy"
	flowNotify    = "notify-send"
	flowPathKey   = "PATH"

	flowNotifyLogEnv    = "CLIPTRANS_TEST_NOTIFY_LOG"
	flowNotifyLogScript = "printf '%s\\n' \"$@\" > \"$CLIPTRANS_TEST_NOTIFY_LOG\"\n"
	flowShellHeader     = "#!/bin/sh\n"
	flowPasteHello      = "printf '%s' 'Hello'\n"
	flowPasteBig        = "printf '%s' 'Hello world'\n"
	flowPasteEmpty      = "printf '%s' ''\n"
	flowPasteCyr        = "printf '%s' 'Привет'\n"
	flowExitFail        = "exit 1\n"
	flowExitOK          = "exit 0\n"
	flowCopyOutEnv      = "CLIPTRANS_TEST_COPY_OUT"
	flowCopySave        = "printf '%s' \"$1\" > \"$CLIPTRANS_TEST_COPY_OUT\"\n"

	flowExampleBaseURL = "https://example.com/v1"
	flowSuccessBody    = `{"choices":[{"message":{"content":"Hola"}}]}`
	flowErrorBody      = `{"error":{"message":"nope"}}`
)

func writeFlowStubs(t *testing.T, scripts map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, script := range scripts {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(flowShellHeader+script), 0o600))
		require.NoError(t, os.Chmod(path, 0o700))
	}

	return dir
}

func useFlowStubs(t *testing.T, scripts map[string]string) {
	t.Helper()
	t.Setenv(
		flowPathKey,
		writeFlowStubs(t, scripts)+string(os.PathListSeparator)+os.Getenv(flowPathKey),
	)
}

func writeFlowConfig(t *testing.T, baseURL, extra string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "base_url: " + baseURL + "\nmodel: m1\n" + extra
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func startFlowServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write mock response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	return srv
}

func runRoot(args ...string) (string, error) {
	root := NewRoot(context.Background())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return buf.String(), fmt.Errorf("execute cliptrans: %w", err)
	}

	return buf.String(), nil
}

func TestTranslateDryRun(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	out, err := runRoot(translateName, flowConfigFlag, path, flowDryRunFlag)
	require.NoError(t, err)
	require.Contains(t, out, "target: en")
	require.Contains(t, out, "model: m1")
	require.Contains(t, out, "chars: 5")
}

func TestTranslateDryRunOverrides(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteCyr, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	out, err := runRoot("--verbose", translateName,
		flowConfigFlag, path, flowDryRunFlag, flowToFlag, "uk", flowModelFlag, "other")
	require.NoError(t, err)
	require.Contains(t, out, "target: uk")
	require.Contains(t, out, "model: other")
	require.Contains(t, out, "chars: 6")
}

func TestTranslateEmptyClipboard(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteEmpty, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "clipboard is empty")
}

func TestTranslatePasteFailure(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowExitFail, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "wl-paste")
}

func TestTranslateBadToFlag(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path, flowToFlag, strings.Repeat("x", 33))
	require.ErrorContains(t, err, "too long")
}

func TestTranslateFromFlagRejected(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path, "--from", "en")
	require.ErrorContains(t, err, "unknown flag")
}

func TestTranslateToFlag(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteCyr, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	out, err := runRoot(translateName, flowConfigFlag, path, flowDryRunFlag, flowToFlag, "uk")
	require.NoError(t, err)
	require.Contains(t, out, "target: uk")
}

func TestTranslateTargetFromConfig(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "to: uk\n")
	t.Setenv(config.KeyEnvName, "")

	out, err := runRoot(translateName, flowConfigFlag, path, flowDryRunFlag)
	require.NoError(t, err)
	require.Contains(t, out, "target: uk")
}

func TestTranslateDefaultTargetIsEn(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	out, err := runRoot(translateName, flowConfigFlag, path, flowDryRunFlag)
	require.NoError(t, err)
	require.Contains(t, out, "target: en")
}

func TestTranslateMissingConfig(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	t.Setenv(config.KeyEnvName, "")
	missing := filepath.Join(t.TempDir(), "no-such.yaml")

	_, err := runRoot(translateName, flowConfigFlag, missing)
	require.ErrorContains(t, err, "no config")
}

func TestTranslateOverLimit(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteBig, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "max_chars: 5\n")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "limit")
}

func TestTranslateMissingAPIKey(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, config.KeyEnvName)
}

func TestTranslateAPIError(t *testing.T) {
	useFlowStubs(t, map[string]string{flowPasteName: flowPasteHello, flowNotify: flowExitOK})
	srv := startFlowServer(t, http.StatusUnauthorized, flowErrorBody)
	path := writeFlowConfig(t, srv.URL, "")
	t.Setenv(config.KeyEnvName, "test-key")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "401")
}

func TestTranslateCopyFailure(t *testing.T) {
	useFlowStubs(
		t,
		map[string]string{
			flowPasteName: flowPasteHello,
			flowCopyName:  flowExitFail,
			flowNotify:    flowExitOK,
		},
	)
	srv := startFlowServer(t, http.StatusOK, flowSuccessBody)
	path := writeFlowConfig(t, srv.URL, "")
	t.Setenv(config.KeyEnvName, "test-key")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "wl-copy")
}

func TestTranslateSuccess(t *testing.T) {
	useFlowStubs(
		t,
		map[string]string{
			flowPasteName: flowPasteHello,
			flowCopyName:  flowCopySave,
			flowNotify:    flowExitOK,
		},
	)
	srv := startFlowServer(t, http.StatusOK, flowSuccessBody)
	path := writeFlowConfig(t, srv.URL, "")
	t.Setenv(config.KeyEnvName, "test-key")

	capture := filepath.Join(t.TempDir(), "copied.txt")
	t.Setenv(flowCopyOutEnv, capture)

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.NoError(t, err)

	raw, readErr := os.ReadFile(capture)
	require.NoError(t, readErr)
	require.Equal(t, "Hola", string(raw))
}

func TestTranslateSuccessNotifiesWhenEnabled(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(flowNotifyLogEnv, log)
	useFlowStubs(
		t,
		map[string]string{
			flowPasteName: flowPasteHello,
			flowCopyName:  flowCopySave,
			flowNotify:    flowNotifyLogScript,
		},
	)
	srv := startFlowServer(t, http.StatusOK, flowSuccessBody)
	path := writeFlowConfig(t, srv.URL, "")
	t.Setenv(config.KeyEnvName, "test-key")

	capture := filepath.Join(t.TempDir(), "copied.txt")
	t.Setenv(flowCopyOutEnv, capture)

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.NoError(t, err)

	raw, readErr := os.ReadFile(log)
	require.NoError(t, readErr)
	require.Contains(t, string(raw), "translated to")
	require.Contains(t, string(raw), "Hola")
}

func TestTranslateSuccessSilentWhenDisabled(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(flowNotifyLogEnv, log)
	useFlowStubs(
		t,
		map[string]string{
			flowPasteName: flowPasteHello,
			flowCopyName:  flowCopySave,
			flowNotify:    flowNotifyLogScript,
		},
	)
	srv := startFlowServer(t, http.StatusOK, flowSuccessBody)
	path := writeFlowConfig(t, srv.URL, "notify_on_success: false\n")
	t.Setenv(config.KeyEnvName, "test-key")

	capture := filepath.Join(t.TempDir(), "copied.txt")
	t.Setenv(flowCopyOutEnv, capture)

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.NoError(t, err)
	require.NoFileExists(t, log)
}

func TestTranslateErrorNotifiesByDefault(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(flowNotifyLogEnv, log)
	useFlowStubs(
		t,
		map[string]string{flowPasteName: flowPasteEmpty, flowNotify: flowNotifyLogScript},
	)
	path := writeFlowConfig(t, flowExampleBaseURL, "")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "clipboard is empty")

	raw, readErr := os.ReadFile(log)
	require.NoError(t, readErr)
	require.Contains(t, string(raw), "empty clipboard")
}

func TestTranslateErrorSilentWhenDisabled(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(flowNotifyLogEnv, log)
	useFlowStubs(
		t,
		map[string]string{flowPasteName: flowPasteEmpty, flowNotify: flowNotifyLogScript},
	)
	path := writeFlowConfig(t, flowExampleBaseURL, "notify_on_error: false\n")
	t.Setenv(config.KeyEnvName, "")

	_, err := runRoot(translateName, flowConfigFlag, path)
	require.ErrorContains(t, err, "clipboard is empty")
	require.NoFileExists(t, log)
}

func TestLoadEffectiveConfigDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := loadEffectiveConfig("")
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Model)
	require.NotEmpty(t, cfg.BaseURL)

	again, err := loadEffectiveConfig("")
	require.NoError(t, err)
	require.Equal(t, cfg.Model, again.Model)
}
