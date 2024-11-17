package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	options Options
	client  *openai.Client
}

func NewOpenAIProvider(options Options) (*OpenAIProvider, error) {
	client := openai.NewClient(options.APIKey)
	return &OpenAIProvider{
		options: options,
		client:  client,
	}, nil
}

func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) ([]string, error) {
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

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.options.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   maxTokens,
			Temperature: float32(temperature),
			N:          maxCandidates,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no responses generated")
	}

	responses := make([]string, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		responses = append(responses, choice.Message.Content)
	}

	return responses, nil
}
