package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fixr-app/fixr/internal/config"
	"github.com/fixr-app/fixr/internal/prompt"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// Anthropic implements the Provider interface for Anthropic's Claude API.
type Anthropic struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropic creates a new Anthropic provider.
func NewAnthropic(cfg *config.Provider) (*Anthropic, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic: api_key is required (set ANTHROPIC_API_KEY env var)")
	}
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Anthropic{
		apiKey: cfg.APIKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Complete(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error) {
	// Build content blocks for the user message
	var contentBlocks []map[string]interface{}

	// Add images first (as context)
	for _, img := range images {
		contentBlocks = append(contentBlocks, map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": "image/png",
				"data":       base64.StdEncoding.EncodeToString(img),
			},
		})
	}

	// Add text
	contentBlocks = append(contentBlocks, map[string]interface{}{
		"type": "text",
		"text": prompt.BuildUserMessage(text, len(images) > 0),
	})

	reqBody := map[string]interface{}{
		"model":      a.model,
		"max_tokens": 4096,
		"system":     prompt.BuildGrammarPrompt(prefs.Tone, prefs.Language),
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": contentBlocks,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("anthropic: failed to parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}

	return result.Content[0].Text, nil
}
