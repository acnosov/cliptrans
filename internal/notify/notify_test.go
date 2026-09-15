package notify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	notifyShellHeader = "#!/bin/sh\n"
	notifyLogEnv      = "CLIPTRANS_TEST_NOTIFY_LOG"
	notifyLogScript   = "printf '%s\\n' \"$@\" > \"$CLIPTRANS_TEST_NOTIFY_LOG\"\n"
	pathEnvKey        = "PATH"
)

func stubNotifyDir(t *testing.T, script string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "notify-send")
	require.NoError(t, os.WriteFile(path, []byte(notifyShellHeader+script), 0o600))
	require.NoError(t, os.Chmod(path, 0o700))

	return dir
}

func stubNotifyPATH(t *testing.T, script string) string {
	t.Helper()

	return stubNotifyDir(t, script) + string(os.PathListSeparator) + os.Getenv(pathEnvKey)
}

func readNotifyLog(t *testing.T, log string) string {
	t.Helper()

	raw, err := os.ReadFile(log)
	require.NoError(t, err)

	return string(raw)
}

func TestFailureWithoutNotifySend(t *testing.T) {
	t.Setenv(pathEnvKey, t.TempDir())
	require.NotPanics(t, func() {
		Failure(t.Context(), "cliptrans: boom", "detail")
	})
}

func TestFailureSendsNotification(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(notifyLogEnv, log)
	t.Setenv(pathEnvKey, stubNotifyPATH(t, notifyLogScript))

	Failure(t.Context(), "cliptrans: boom", "some detail")

	body := readNotifyLog(t, log)
	require.Contains(t, body, "cliptrans: boom")
	require.Contains(t, body, "some detail")
	require.Contains(t, body, "--urgency=critical")
}

func TestFailureWithoutBody(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(notifyLogEnv, log)
	t.Setenv(pathEnvKey, stubNotifyPATH(t, notifyLogScript))

	Failure(t.Context(), "cliptrans: boom", "")

	lines := strings.Split(strings.TrimSpace(readNotifyLog(t, log)), "\n")
	require.Len(t, lines, 3)
}

func TestFailureNotifySendError(t *testing.T) {
	t.Setenv(pathEnvKey, stubNotifyPATH(t, "exit 1\n"))
	require.NotPanics(t, func() {
		Failure(t.Context(), "cliptrans: boom", "detail")
	})
}

func TestSuccessWithoutNotifySend(t *testing.T) {
	t.Setenv(pathEnvKey, t.TempDir())
	require.NotPanics(t, func() {
		Success(t.Context(), "cliptrans: translated to English", "")
	})
}

func TestSuccessSendsNotification(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(notifyLogEnv, log)
	t.Setenv(pathEnvKey, stubNotifyPATH(t, notifyLogScript))

	Success(t.Context(), "cliptrans: translated to English", "some detail")

	body := readNotifyLog(t, log)
	require.Contains(t, body, "cliptrans: translated to English")
	require.Contains(t, body, "some detail")
	require.Contains(t, body, "--urgency=normal")
}

func TestSuccessWithoutBody(t *testing.T) {
	log := filepath.Join(t.TempDir(), "notify.log")
	t.Setenv(notifyLogEnv, log)
	t.Setenv(pathEnvKey, stubNotifyPATH(t, notifyLogScript))

	Success(t.Context(), "cliptrans: translated to English", "")

	lines := strings.Split(strings.TrimSpace(readNotifyLog(t, log)), "\n")
	require.Len(t, lines, 3)
}

func TestSuccessNotifySendError(t *testing.T) {
	t.Setenv(pathEnvKey, stubNotifyPATH(t, "exit 1\n"))
	require.NotPanics(t, func() {
		Success(t.Context(), "cliptrans: translated to English", "detail")
	})
}
