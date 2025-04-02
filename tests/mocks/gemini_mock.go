package mocks

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/jabafett/quill/internal/utils/ai"
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

// mocks.EnsureParentDir ensures that the parent directory of the given file path exists
func EnsureParentDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}
