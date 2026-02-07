package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/fixr-app/fixr/internal/capture"
	"github.com/fixr-app/fixr/internal/clipboard"
	"github.com/fixr-app/fixr/internal/config"
	"github.com/fixr-app/fixr/internal/provider"
)

// Engine orchestrates the grammar fix flow:
// clipboard read -> screenshot -> AI call -> result.
type Engine struct {
	cfg      *config.Config
	provider provider.Provider
}

// New creates a new Engine with the given config.
func New(cfg *config.Config) (*Engine, error) {
	providerCfg, err := cfg.GetProvider("")
	if err != nil {
		return nil, err
	}

	p, err := provider.New(cfg.DefaultProvider, providerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider %q: %w", cfg.DefaultProvider, err)
	}

	return &Engine{
		cfg:      cfg,
		provider: p,
	}, nil
}

// NewWithProvider creates a new Engine with a specific provider name.
func NewWithProvider(cfg *config.Config, providerName string) (*Engine, error) {
	if providerName == "" {
		return New(cfg)
	}

	providerCfg, err := cfg.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	p, err := provider.New(providerName, providerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider %q: %w", providerName, err)
	}

	return &Engine{
		cfg:      cfg,
		provider: p,
	}, nil
}

// FixResult contains the result of a grammar fix operation.
type FixResult struct {
	Original  string
	Corrected string
	Provider  string
}

// FixClipboard reads text from clipboard, optionally captures a screenshot,
// sends both to the AI provider, and returns the corrected text.
func (e *Engine) FixClipboard(ctx context.Context) (*FixResult, error) {
	// Step 1: Read clipboard
	text, err := clipboard.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read clipboard: %w", err)
	}
	if text == "" {
		return nil, fmt.Errorf("clipboard is empty — copy some text first")
	}

	return e.FixText(ctx, text)
}

// FixText fixes the grammar of the given text.
func (e *Engine) FixText(ctx context.Context, text string) (*FixResult, error) {
	// Step 2: Capture screenshot (if enabled)
	var images [][]byte
	if e.cfg.Preferences.AutoScreenshot {
		screenshot, err := capture.ActiveWindow()
		if err != nil {
			// Screenshot failure is non-fatal — proceed without it
			fmt.Fprintf(os.Stderr, "Warning: could not capture screenshot: %v\n", err)
		} else {
			images = append(images, screenshot)
		}
	}

	// Step 3: Send to AI provider
	prefs := provider.Preferences{
		Tone:     e.cfg.Preferences.Tone,
		Language: e.cfg.Preferences.Language,
	}
	corrected, err := e.provider.Complete(ctx, text, images, prefs)
	if err != nil {
		return nil, fmt.Errorf("AI provider %q failed: %w", e.provider.Name(), err)
	}

	return &FixResult{
		Original:  text,
		Corrected: corrected,
		Provider:  e.provider.Name(),
	}, nil
}

// FixFile reads a file, fixes its grammar, and returns the result.
func (e *Engine) FixFile(ctx context.Context, path string) (*FixResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	text := string(data)
	if text == "" {
		return nil, fmt.Errorf("file %q is empty", path)
	}

	// Disable screenshot for file mode — no visual context needed
	origPref := e.cfg.Preferences.AutoScreenshot
	e.cfg.Preferences.AutoScreenshot = false
	defer func() { e.cfg.Preferences.AutoScreenshot = origPref }()

	return e.FixText(ctx, text)
}
