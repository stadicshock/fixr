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

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAI implements the Provider interface for OpenAI's API.
type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(cfg *config.Provider) (*OpenAI, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai: api_key is required (set OPENAI_API_KEY env var)")
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAI{
		apiKey: cfg.APIKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (o *OpenAI) Name() string { return "openai" }

func (o *OpenAI) Complete(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error) {
	// Build content parts for the user message
	var contentParts []map[string]interface{}

	// Add images first (as context)
	for _, img := range images {
		contentParts = append(contentParts, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]interface{}{
				"url":    "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(img),
				"detail": "auto",
			},
		})
	}

	// Add text
	contentParts = append(contentParts, map[string]interface{}{
		"type": "text",
		"text": prompt.BuildUserMessage(text, len(images) > 0),
	})

	reqBody := map[string]interface{}{
		"model":      o.model,
		"max_tokens": 4096,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": prompt.BuildGrammarPrompt(prefs.Tone, prefs.Language),
			},
			{
				"role":    "user",
				"content": contentParts,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openai: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("openai: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("openai: failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}

	return result.Choices[0].Message.Content, nil
}
