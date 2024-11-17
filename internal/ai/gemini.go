package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/jabafett/quill/internal/prompts"
	"google.golang.org/api/option"
)

// GeminiResponse represents the structure of the Gemini API response
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []string `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type GeminiProvider struct {
	options Options
	client  *genai.Client
}

func NewGeminiProvider(options Options) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(options.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		options: options,
		client:  client,
	}, nil
}

func (p *GeminiProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	model := p.client.GenerativeModel(p.options.Model)
	model.SetTemperature(p.options.Temperature)

	prompt := prompts.GetCommitPrompt(diff)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Convert response to JSON
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	// Parse JSON into our custom response struct
	var geminiResp GeminiResponse
	if err := json.Unmarshal(jsonBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract commit message from response
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no valid response generated")
	}

	message := geminiResp.Candidates[0].Content.Parts[0]
	message = strings.TrimSpace(message)

	return message, nil
}
