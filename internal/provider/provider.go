package provider

import (
	"context"
	"fmt"

	"github.com/fixr-app/fixr/internal/config"
)

// Preferences holds tone and language settings for grammar correction.
type Preferences struct {
	Tone     string
	Language string
}

// Provider is the interface that all AI providers must implement.
type Provider interface {
	// Name returns the provider's identifier (e.g. "anthropic", "openai").
	Name() string

	// Complete sends text and optional images to the AI and returns corrected text.
	// The images slice contains PNG screenshot data for visual context.
	// Prefs contains tone and language preferences for the correction.
	Complete(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error)
}

// New creates a provider instance based on the provider name and config.
func New(name string, cfg *config.Provider) (Provider, error) {
	switch name {
	case "anthropic":
		return NewAnthropic(cfg)
	case "openai":
		return NewOpenAI(cfg)
	case "bedrock":
		return NewBedrock(cfg)
	case "ollama":
		return NewOllama(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
