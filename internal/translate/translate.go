package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const RequestTimeout = 30 * time.Second

const (
	maxResponseBytes = 4 << 20
	maxErrorPreview  = 500
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type Client struct {
	HTTP         *http.Client
	BaseURL      string
	Model        string
	APIKey       string
	SystemPrompt string
	Temperature  float64
}

func (c *Client) Translate(ctx context.Context, text string) (string, error) {
	body, err := c.requestBody(text)
	if err != nil {
		return "", err
	}
	respBody, err := c.post(ctx, body)
	if err != nil {
		return "", err
	}
	return decodeTranslation(respBody)
}

func (c *Client) requestBody(text string) ([]byte, error) {
	if c.SystemPrompt == "" {
		return nil, errors.New("missing system prompt")
	}
	body, err := json.Marshal(chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: c.SystemPrompt},
			{Role: "user", Content: "<text>\n" + text + "\n</text>"},
		},
		Temperature: c.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	return body, nil
}

type chatResponseError struct {
	Message string `json:"message"`
}

type chatResponseMessage struct {
	Content string `json:"content"`
}

type chatChoice struct {
	Message chatResponseMessage `json:"message"`
}

type chatResponse struct {
	Error   *chatResponseError `json:"error,omitempty"`
	Choices []chatChoice       `json:"choices"`
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:       RequestTimeout,
		CheckRedirect: stripAuthOnCrossHostRedirect,
	}
}

func stripAuthOnCrossHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 && req.URL.Host != via[0].URL.Host {
		req.Header.Del("Authorization")
	}
	return nil
}

func (c *Client) post(ctx context.Context, body []byte) ([]byte, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = defaultHTTPClient()
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("translate request to %s (model %s): %w", url, c.Model, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read translate response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"translate API error: HTTP %d: %s",
			resp.StatusCode,
			truncate(string(respBody), maxErrorPreview),
		)
	}
	return respBody, nil
}

func decodeTranslation(respBody []byte) (string, error) {
	var decoded chatResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode translate response: %w", err)
	}
	if decoded.Error != nil {
		if decoded.Error.Message != "" {
			return "", fmt.Errorf("translate API error: %s", decoded.Error.Message)
		}
		return "", errors.New("translate API error: unknown error")
	}
	if len(decoded.Choices) == 0 {
		return "", errors.New("translate API error: empty choices")
	}
	out := decoded.Choices[0].Message.Content
	if strings.TrimSpace(out) == "" {
		return "", errors.New("translate API error: empty translation")
	}
	return out, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
