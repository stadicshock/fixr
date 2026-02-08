package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"

	"github.com/fixr-app/fixr/internal/config"
	"github.com/fixr-app/fixr/internal/prompt"
)

// Bedrock implements the Provider interface for AWS Bedrock.
type Bedrock struct {
	modelID string
	region  string
	profile string
	client  *bedrockruntime.BedrockRuntime
}

// NewBedrock creates a new AWS Bedrock provider.
// It uses AWS credentials from the environment, ~/.aws/credentials, or a named profile.
func NewBedrock(cfg *config.Provider) (*Bedrock, error) {
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	modelID := cfg.ModelID
	if modelID == "" {
		modelID = "us.anthropic.claude-3-5-haiku-20241022-v1:0"
	}

	return &Bedrock{
		modelID: modelID,
		region:  region,
		profile: cfg.Profile,
	}, nil
}

func (b *Bedrock) Name() string { return "bedrock" }

// initClient lazily initializes the Bedrock runtime client.
func (b *Bedrock) initClient() error {
	if b.client != nil {
		return nil
	}

	opts := session.Options{
		Config: aws.Config{
			Region: aws.String(b.region),
		},
		SharedConfigState: session.SharedConfigEnable,
	}

	// Use named AWS profile if configured
	if b.profile != "" {
		opts.Profile = b.profile
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return fmt.Errorf("bedrock: failed to create AWS session: %w", err)
	}
	b.client = bedrockruntime.New(sess)
	return nil
}

func (b *Bedrock) Complete(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error) {
	if err := b.initClient(); err != nil {
		return "", err
	}

	// Try with images first; if the model doesn't support vision, retry without
	result, err := b.invoke(ctx, text, images, prefs)
	if err != nil && len(images) > 0 && isImageNotSupportedError(err) {
		fmt.Fprintf(os.Stderr, "Warning: model does not support image input, retrying without screenshot\n")
		return b.invoke(ctx, text, nil, prefs)
	}
	return result, err
}

// invoke sends the request to Bedrock and returns the corrected text.
func (b *Bedrock) invoke(ctx context.Context, text string, images [][]byte, prefs Preferences) (string, error) {
	// Build the Anthropic Messages API payload (Bedrock uses the same format for Claude models)
	var contentBlocks []map[string]interface{}

	for _, img := range images {
		contentBlocks = append(contentBlocks, map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": "image/jpeg",
				"data":       base64.StdEncoding.EncodeToString(img),
			},
		})
	}

	contentBlocks = append(contentBlocks, map[string]interface{}{
		"type": "text",
		"text": prompt.BuildUserMessage(text, len(images) > 0),
	})

	payload := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        4096,
		"system":            prompt.BuildGrammarPrompt(prefs.Tone, prefs.Language),
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": contentBlocks,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("bedrock: failed to marshal payload: %w", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        payloadBytes,
	}

	output, err := b.client.InvokeModelWithContext(ctx, input)
	if err != nil {
		return "", fmt.Errorf("bedrock: invoke model failed: %w", err)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(bytes.NewReader(output.Body)).Decode(&result); err != nil {
		return "", fmt.Errorf("bedrock: failed to parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("bedrock: empty response")
	}

	return result.Content[0].Text, nil
}

// isImageNotSupportedError checks if the error is due to the model not supporting image input.
func isImageNotSupportedError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "does not support image input") ||
		strings.Contains(msg, "image is not supported")
}
