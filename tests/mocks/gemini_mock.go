package mocks

import (
	"context"
	"github.com/jabafett/quill/internal/ai"
)

type MockGeminiProvider struct {
	GenerateFunc func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error)
	RetryCount   int
}

func (m *MockGeminiProvider) Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
	m.RetryCount++
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, prompt, opts)
	}
	return []string{"feat(test): add mock implementation"}, nil
} 