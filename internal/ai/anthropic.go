package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AnthropicProvider struct {
	options Options
	baseURL string
}

type anthropicRequest struct {
	Model       string               `json:"model"`
	MaxTokens   int                  `json:"max_tokens"`
	Messages    []anthropicMessage   `json:"messages"`
	Temperature float32              `json:"temperature"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func NewAnthropicProvider(options Options) (*AnthropicProvider, error) {
	return &AnthropicProvider{
		options: options,
		baseURL: "https://api.anthropic.com/v1/messages",
	}, nil
}

func (p *AnthropicProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error) {
	// Use provider defaults if not specified in opts
	maxCandidates := opts.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = p.options.CandidateCount
	}
	if maxCandidates > 3 {
		maxCandidates = 3
	}

	temperature := p.options.Temperature
	if opts.Temperature != nil {
		temperature = *opts.Temperature
	}

	maxTokens := p.options.MaxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	reqBody := anthropicRequest{
		Model:       p.options.Model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Make multiple requests if multiple candidates are requested
	var responses []string
	for i := 0; i < maxCandidates; i++ {
		resp, err := p.makeRequest(ctx, reqBody)
		if err != nil {
			return responses, fmt.Errorf("failed to generate candidate %d: %w", i+1, err)
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// makeRequest handles the actual API call
func (p *AnthropicProvider) makeRequest(ctx context.Context, reqBody anthropicRequest) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.options.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from the first content block
	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	for _, content := range result.Content {
		if content.Type == "text" {
			return content.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}
