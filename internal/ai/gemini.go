package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

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

func (p *GeminiProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error) {
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

	model := p.client.GenerativeModel(p.options.Model)
	model.SetTemperature(temperature)
	model.SetCandidateCount(int32(maxCandidates))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Convert response to JSON
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// Parse JSON into our custom response struct
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []string `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(jsonBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no valid response generated")
	}

	responses := make([]string, 0, len(geminiResp.Candidates))
	for _, candidate := range geminiResp.Candidates {
		if len(candidate.Content.Parts) > 0 {
			message := strings.TrimSpace(candidate.Content.Parts[0])
			responses = append(responses, message)
		}
	}

	return responses, nil
}
