package clipboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	fakeShellHeader = "#!/bin/sh\n"
	fakeHelloText   = "hello"
)

func writeStub(t *testing.T, dir, name, script string) {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(fakeShellHeader+script), 0o600))
	require.NoError(t, os.Chmod(path, 0o700))
}

func stubClipboardTools(t *testing.T, scripts map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, script := range scripts {
		writeStub(t, dir, name, script)
	}

	return dir
}

func TestRead(t *testing.T) {
	cases := []struct {
		name    string
		script  string
		want    string
		wantErr bool
	}{
		{"ok", "printf '%s' '" + fakeHelloText + "'\n", fakeHelloText, false},
		{"empty output is not a read error", "printf '%s' ''\n", "", false},
		{"failure", "exit 1\n", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := stubClipboardTools(t, map[string]string{"wl-paste": tc.script})
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

			got, err := Read(t.Context())
			if tc.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestWrite(t *testing.T) {
	const captureEnv = "CLIPTRANS_TEST_CAPTURE"

	cases := []struct {
		name    string
		script  string
		text    string
		want    string
		wantErr bool
	}{
		{
			"ok",
			"printf '%s' \"$1\" > \"$CLIPTRANS_TEST_CAPTURE\"\n",
			fakeHelloText,
			fakeHelloText,
			false,
		},
		{"multiline", "printf '%s' \"$1\" > \"$CLIPTRANS_TEST_CAPTURE\"\n", "a\nb", "a\nb", false},
		{"failure", "exit 1\n", fakeHelloText, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := stubClipboardTools(t, map[string]string{"wl-copy": tc.script})
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

			capture := filepath.Join(t.TempDir(), "captured.txt")
			t.Setenv(captureEnv, capture)

			err := Write(t.Context(), tc.text)
			if tc.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)

			raw, readErr := os.ReadFile(capture)
			require.NoError(t, readErr)
			require.Equal(t, tc.want, string(raw))
		})
	}
}
