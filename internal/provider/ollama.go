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

// Ollama implements the Provider interface for local Ollama models.
type Ollama struct {
	host   string
	model  string
	client *http.Client
}

// NewOllama creates a new Ollama provider.
func NewOllama(cfg *config.Provider) (*Ollama, error) {
	host := cfg.Host
	if host == "" {
		host = "http://localhost:11434"
	}
	model := cfg.Model
	if model == "" {
		model = "llama3"
	}
	return &Ollama{
		host:   host,
		model:  model,
		client: &http.Client{},
	}, nil
}

func (o *Ollama) Name() string { return "ollama" }

func (o *Ollama) Complete(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error) {
	systemPrompt := prompt.BuildGrammarPrompt(prefs.Tone, prefs.Language)
	userMessage := prompt.BuildUserMessage(text, len(images) > 0)

	// Encode images as base64 strings for Ollama's API
	var imageStrings []string
	for _, img := range images {
		imageStrings = append(imageStrings, base64.StdEncoding.EncodeToString(img))
	}

	reqBody := map[string]interface{}{
		"model":  o.model,
		"stream": false,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userMessage,
				"images":  imageStrings,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: failed to marshal request: %w", err)
	}

	apiURL := o.host + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("ollama: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: request failed (is Ollama running at %s?): %w", o.host, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("ollama: failed to parse response: %w", err)
	}

	return result.Message.Content, nil
}
