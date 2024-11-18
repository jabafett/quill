package ai

import (
	"context"
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
	if prompt == "" {
		return nil, fmt.Errorf("empty prompt provided")
	}

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

	// Use the prompt directly since it's already formatted
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	responses := make([]string, 0, len(resp.Candidates))
	for _, candidate := range resp.Candidates {
		if len(candidate.Content.Parts) > 0 {
			if textPart, ok := candidate.Content.Parts[0].(genai.Text); ok {
				message := strings.TrimSpace(string(textPart))
				responses = append(responses, message)
			}
		}
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated")
	}

	return responses, nil
}
