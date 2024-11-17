package ai

import (
	"context"
	"fmt"

	"github.com/jabafett/quill/internal/git"
	"github.com/jabafett/quill/internal/prompts"
)

// CommitMessageGenerator handles generating commit messages using AI providers
type CommitMessageGenerator struct {
	provider Provider
	repo     *git.Repository
}

// NewCommitMessageGenerator creates a new generator instance
func NewCommitMessageGenerator(provider Provider, repo *git.Repository) *CommitMessageGenerator {
	return &CommitMessageGenerator{
		provider: provider,
		repo:     repo,
	}
}

// GenerateStagedCommitMessage creates a commit message for the current staged changes
func (g *CommitMessageGenerator) GenerateStagedCommitMessage(ctx context.Context, opts GenerateOptions) ([]string, error) {
	// Get formatted prompt
	prompt, err := prompts.GetCommitPrompt(g.repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit prompt: %w", err)
	}

	// Generate messages using the AI provider
	messages, err := g.provider.Generate(ctx, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit message: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no commit messages generated")
	}

	return messages, nil
} 