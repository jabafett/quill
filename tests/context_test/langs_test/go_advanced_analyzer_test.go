package langs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/factories"
	types "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestGoAdvancedAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(1),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	if err != nil {
		t.Fatalf("NewContextEngine() error = %v", err)
	}

	// Create test directory and file
	testDir := t.TempDir()
	goFile := filepath.Join(testDir, "go/advanced.go")
	goContent := `
package services

import (
	"context"
	"sync"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
)

// Configuration holds service configuration
type Configuration struct {
	Level LogLevel
	Path  string
}

// Service represents a generic service interface
type Service[T any] interface {
	Process(ctx context.Context, data T) (T, error)
}

// singleton is a generic singleton implementation
type singleton[T any] struct {
	once     sync.Once
	instance T
}

func (s *singleton[T]) getInstance(constructor func() T) T {
	s.once.Do(func() {
		s.instance = constructor()
	})
	return s.instance
}

// ComplexService implements Service with caching capabilities
type ComplexService[T comparable] struct {
	singleton[*ComplexService[T]]
	cache map[string]T
	mu    sync.RWMutex
}

func NewComplexService[T comparable]() *ComplexService[T] {
	return &ComplexService[T]{
		cache: make(map[string]T),
	}
}

func (s *ComplexService[T]) Process(ctx context.Context, data T) (T, error) {
	// Implementation
	return data, nil
}

func (s *ComplexService[T]) CachedCompute(key string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	val, ok := s.cache[key]
	return val, ok
}`

	// Write test file
	if err := mocks.EnsureParentDir(goFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to write advanced.go: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{goFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[goFile]
	require.True(t, ok, "File context not found for %s", goFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"LogLevel", string(types.Type)},
		{"Configuration", string(types.Class)},
		{"Service", string(types.Interface)},
		{"singleton", string(types.Class)},
		{"ComplexService", string(types.Class)},
		{"NewComplexService", string(types.Function)},
		{"Process", string(types.Function)},
		{"CachedCompute", string(types.Function)},
		{"Debug", string(types.Constant)},
		{"Info", string(types.Constant)},
		{"Warning", string(types.Constant)},
		{"Error", string(types.Constant)},
	}

	// Verify all expected symbols are present
	for _, expected := range symbols {
		found := false
		for _, symbol := range fileCtx.Symbols {
			if symbol.Name == expected.name && string(symbol.Type) == expected.typ {
				found = true
				break
			}
		}
		assert.True(t, found, "Symbol %s of type %s not found", expected.name, expected.typ)
	}

	if t.Failed() {
		jsonCtx, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal context: %v", err)
		}
		t.Logf("Context:\n%s", string(jsonCtx))
	}
}
