package ai

import "context"

// Provider defines the interface for AI providers
type Provider interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
}

// Options represents common configuration options for AI providers
type Options struct {
	Model       string
	MaxTokens   int
	Temperature float32
	APIKey      string
} 