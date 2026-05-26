// Package llm je klient pro OpenAI-kompatibilní LiteLLM endpoint (llm.ol1n.com) za CF Access.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL      string
	Model        string
	AccessID     string
	AccessSecret string
	HTTP         *http.Client
}

func New(baseURL, model, accessID, accessSecret string) *Client {
	return &Client{
		BaseURL:      baseURL,
		Model:        model,
		AccessID:     accessID,
		AccessSecret: accessSecret,
		HTTP:         &http.Client{},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Complete pošle chat completion request a vrátí obsah odpovědi + latenci v ms.
func (c *Client) Complete(ctx context.Context, model, system, user string, maxTokens int, temp float64) (string, int64, error) {
	if model == "" {
		model = c.Model
	}
	body, _ := json.Marshal(chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens:   maxTokens,
		Temperature: temp,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AccessID != "" {
		req.Header.Set("CF-Access-Client-Id", c.AccessID)
		req.Header.Set("CF-Access-Client-Secret", c.AccessSecret)
	}

	start := time.Now()
	resp, err := c.HTTP.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return "", latency, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", latency, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(raw))
	}
	var cr chatResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", latency, fmt.Errorf("dekódování odpovědi: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", latency, fmt.Errorf("prázdná odpověď LLM")
	}
	return cr.Choices[0].Message.Content, latency, nil
}
