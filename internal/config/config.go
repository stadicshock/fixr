package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for fixr.
type Config struct {
	DefaultProvider string              `yaml:"default_provider"`
	Providers       map[string]Provider `yaml:"providers"`
	Preferences     Preferences         `yaml:"preferences"`
}

// Provider holds configuration for a single AI provider.
type Provider struct {
	APIKey  string `yaml:"api_key,omitempty"`
	Model   string `yaml:"model,omitempty"`
	ModelID string `yaml:"model_id,omitempty"`
	Region  string `yaml:"region,omitempty"`
	Profile string `yaml:"profile,omitempty"`
	Host    string `yaml:"host,omitempty"`
}

// Preferences holds user preferences.
type Preferences struct {
	Tone           string `yaml:"tone"`
	Language       string `yaml:"language"`
	Hotkey         string `yaml:"hotkey"`
	AutoScreenshot bool   `yaml:"auto_screenshot"`
	AutoPaste      bool   `yaml:"auto_paste"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "bedrock",
		Providers: map[string]Provider{
			"anthropic": {
				APIKey: "${ANTHROPIC_API_KEY}",
				Model:  "claude-sonnet-4-20250514",
			},
			"openai": {
				APIKey: "${OPENAI_API_KEY}",
				Model:  "gpt-4o",
			},
			"bedrock": {
				Region:  "us-east-1",
				ModelID: "us.anthropic.claude-3-5-haiku-20241022-v1:0",
			},
			"ollama": {
				Host:  "http://localhost:11434",
				Model: "llama3",
			},
		},
		Preferences: Preferences{
			Tone:           "professional",
			Language:       "en-US",
			Hotkey:         "ctrl+shift+g",
			AutoScreenshot: true,
			AutoPaste:      false,
		},
	}
}

// ConfigDir returns the path to the fixr config directory (~/.fixr).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".fixr"), nil
}

// ConfigPath returns the path to the fixr config file (~/.fixr/config.yaml).
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads and parses the config file from ~/.fixr/config.yaml.
// It resolves ${ENV_VAR} references in string values.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s — run 'fixr config init' to create one", path)
		}
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	// Resolve environment variables before parsing YAML
	resolved := resolveEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(resolved), &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks the config for required fields and valid values.
func (c *Config) Validate() error {
	if c.DefaultProvider == "" {
		return fmt.Errorf("default_provider is required")
	}

	if _, ok := c.Providers[c.DefaultProvider]; !ok {
		available := make([]string, 0, len(c.Providers))
		for k := range c.Providers {
			available = append(available, k)
		}
		return fmt.Errorf("default_provider %q not found in providers (available: %s)",
			c.DefaultProvider, strings.Join(available, ", "))
	}

	validTones := map[string]bool{"professional": true, "casual": true, "formal": true}
	if c.Preferences.Tone != "" && !validTones[c.Preferences.Tone] {
		return fmt.Errorf("invalid tone %q (must be one of: professional, casual, formal)", c.Preferences.Tone)
	}

	return nil
}

// GetProvider returns the provider config for the given name, or the default provider.
func (c *Config) GetProvider(name string) (*Provider, error) {
	if name == "" {
		name = c.DefaultProvider
	}
	p, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not configured", name)
	}
	return &p, nil
}

// WriteDefault creates the config directory and writes the default config file.
func WriteDefault() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")

	// Don't overwrite existing config
	if _, err := os.Stat(path); err == nil {
		return path, fmt.Errorf("config file already exists at %s", path)
	}

	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("could not marshal default config: %w", err)
	}

	header := "# fixr configuration\n# See https://github.com/fixr-app/fixr for documentation\n#\n# API keys can reference environment variables: ${ENV_VAR_NAME}\n\n"

	if err := os.WriteFile(path, []byte(header+string(data)), 0600); err != nil {
		return "", fmt.Errorf("could not write config file: %w", err)
	}

	return path, nil
}

// envVarPattern matches ${VAR_NAME} patterns.
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// resolveEnvVars replaces ${VAR_NAME} with the corresponding environment variable value.
func resolveEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		// Leave unresolved if env var is not set
		return match
	})
}
