package translate

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const successPayload = `{"choices":[{"message":{"content":"Hola"}}]}`

func writePayload(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("write mock response: %v", err)
	}
}

func TestTranslateWithoutHTTPClient(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writePayload(t, w, successPayload)
	}))
	defer srv.Close()

	client := &Client{BaseURL: srv.URL, Model: "m1", APIKey: "k", SystemPrompt: testSystemPrompt}
	out, err := client.Translate(t.Context(), "Hello")
	require.NoError(t, err)
	require.Equal(t, "Hola", out)
}

func TestTranslateWithoutAuthHeader(t *testing.T) {
	t.Parallel()

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		writePayload(t, w, successPayload)
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:      srv.URL,
		Model:        "m1",
		HTTP:         srv.Client(),
		SystemPrompt: testSystemPrompt,
	}
	out, err := client.Translate(t.Context(), "Hello")
	require.NoError(t, err)
	require.Equal(t, "Hola", out)
	require.Empty(t, gotAuth)
}

func TestTranslateTruncatedErrorPreview(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		writePayload(t, w, strings.Repeat("x", 600))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:      srv.URL,
		Model:        "m1",
		HTTP:         srv.Client(),
		SystemPrompt: testSystemPrompt,
	}
	_, err := client.Translate(t.Context(), "hi")
	require.ErrorContains(t, err, "…")
}

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

type failingTransport struct{}

func (failingTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(failingReader{}),
	}, nil
}

func TestTranslateReadError(t *testing.T) {
	t.Parallel()

	client := &Client{
		BaseURL:      "http://example.com",
		Model:        "m1",
		HTTP:         &http.Client{Transport: failingTransport{}},
		SystemPrompt: testSystemPrompt,
	}
	_, err := client.Translate(t.Context(), "hi")
	require.ErrorContains(t, err, "read translate response")
}
