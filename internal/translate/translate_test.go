package translate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSystemPrompt = "test prompt"

func TestRequestCarriesPromptAndWrappedText(t *testing.T) {
	t.Parallel()
	got := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad req", http.StatusBadRequest)
			return
		}
		for _, msg := range req.Messages {
			got[msg.Role] = msg.Content
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"Hola"}}]}`)); err != nil {
			t.Errorf("write mock response: %v", err)
		}
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:      srv.URL,
		Model:        "m1",
		APIKey:       "k",
		HTTP:         srv.Client(),
		SystemPrompt: "be terse",
	}
	out, err := c.Translate(t.Context(), "Hello")
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hola" {
		t.Errorf("out = %q, want Hola", out)
	}
	if len(got) != 2 {
		t.Fatalf("got %d messages, want system and user", len(got))
	}
	if got["system"] != "be terse" {
		t.Errorf("system prompt = %q, want %q", got["system"], "be terse")
	}
	if got["user"] != "<text>\nHello\n</text>" {
		t.Errorf("user message = %q, want wrapped text", got["user"])
	}
}

func TestTranslateMissingSystemPrompt(t *testing.T) {
	t.Parallel()
	c := &Client{BaseURL: "http://example.com", Model: "m1", HTTP: &http.Client{}}
	if _, err := c.Translate(t.Context(), "hi"); err == nil {
		t.Error("expected error for missing system prompt, got nil")
	}
}

func TestTranslateSuccess(t *testing.T) {
	t.Parallel()
	var gotAuth, gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad req", http.StatusBadRequest)
			return
		}
		gotModel = req.Model
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"Hola"}}]}`)); err != nil {
			t.Errorf("write mock response: %v", err)
		}
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:      srv.URL,
		Model:        "m1",
		APIKey:       "k",
		HTTP:         srv.Client(),
		SystemPrompt: testSystemPrompt,
	}
	out, err := c.Translate(t.Context(), "Hello")
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hola" {
		t.Errorf("out = %q, want Hola", out)
	}
	if gotAuth != "Bearer k" {
		t.Errorf("auth = %q, want Bearer k", gotAuth)
	}
	if gotModel != "m1" {
		t.Errorf("model = %q, want m1", gotModel)
	}
}

func TestTranslateBadPayloads(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
	}{
		{"invalid json", "not json"},
		{"api error object", `{"error":{"message":"boom"}}`},
		{"api error without message", `{"error":{}}`},
		{"empty content", `{"choices":[{"message":{"content":""}}]}`},
		{"whitespace content", `{"choices":[{"message":{"content":"   "}}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					if _, err := w.Write([]byte(tc.body)); err != nil {
						t.Errorf("write mock response: %v", err)
					}
				}),
			)
			defer srv.Close()

			c := &Client{
				BaseURL:      srv.URL,
				Model:        "m1",
				HTTP:         srv.Client(),
				SystemPrompt: testSystemPrompt,
			}
			if _, err := c.Translate(t.Context(), "hi"); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestTranslateBadURL(t *testing.T) {
	t.Parallel()
	c := &Client{
		BaseURL:      "http://exa mple.com",
		Model:        "m1",
		HTTP:         &http.Client{},
		SystemPrompt: testSystemPrompt,
	}
	if _, err := c.Translate(t.Context(), "hi"); err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestTranslateHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"nope"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "m1", HTTP: srv.Client(), SystemPrompt: testSystemPrompt}
	if _, err := c.Translate(t.Context(), "hi"); err == nil {
		t.Error("expected error for 401, got nil")
	}
}

func TestTranslateEmptyChoices(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"choices":[]}`)); err != nil {
			t.Errorf("write mock response: %v", err)
		}
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "m1", HTTP: srv.Client(), SystemPrompt: testSystemPrompt}
	if _, err := c.Translate(t.Context(), "hi"); err == nil {
		t.Error("expected error for empty choices, got nil")
	}
}

func TestTranslateTimeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:      srv.URL,
		Model:        "m1",
		HTTP:         &http.Client{Timeout: 20 * time.Millisecond},
		SystemPrompt: testSystemPrompt,
	}
	if _, err := c.Translate(t.Context(), "hi"); err == nil {
		t.Error("expected timeout error, got nil")
	}
}
