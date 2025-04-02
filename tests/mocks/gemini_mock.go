package mocks

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/jabafett/quill/internal/utils/context"

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

// mocks.AssertSymbol checks if a symbol with the given name and type exists in the symbols list
func AssertSymbol(t *testing.T, symbols []types.SymbolContext, name string, symbolType string) {
	t.Helper()
	found := false
	for _, sym := range symbols {
		if sym.Name == name && sym.Type == symbolType {
			found = true
			break
		}
	}
	assert.True(t, found, "Symbol %q of type %q not found", name, symbolType)
}
