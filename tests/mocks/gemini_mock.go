package mocks

import (
	"context"
	"sync/atomic"

	"github.com/jabafett/quill/internal/ai"
)

type MockGeminiProvider struct {
	GenerateFunc func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error)
	RetryCount   int32
}

func (m *MockGeminiProvider) Generate(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
	atomic.AddInt32(&m.RetryCount, 1)
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, prompt, opts)
	}
	return []string{"feat(test): add mock implementation"}, nil
} 