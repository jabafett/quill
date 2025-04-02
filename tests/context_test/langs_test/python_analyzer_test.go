package langs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/factories"
	types "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestPythonComplexAnalyzer(t *testing.T) {
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
	pythonFile := filepath.Join(testDir, "python/advanced.py")
	pythonContent := `
import functools
from typing import TypeVar, Generic, Optional
from abc import ABC, abstractmethod

T = TypeVar('T')

def singleton(cls):
    @functools.wraps(cls)
    def wrapper(*args, **kwargs):
        if not wrapper.instance:
            wrapper.instance = cls(*args, **kwargs)
        return wrapper.instance
    wrapper.instance = None
    return wrapper

class MetaLogger(type):
    def __new__(mcs, name, bases, namespace):
        for key, value in namespace.items():
            if callable(value):
                namespace[key] = mcs.log_calls(value)
        return super().__new__(mcs, name, bases, namespace)
    
    @staticmethod
    def log_calls(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            print(f"Calling {func.__name__}")
            result = func(*args, **kwargs)
            print(f"Finished {func.__name__}")
            return result
        return wrapper

@singleton
class ComplexService(Generic[T], ABC, metaclass=MetaLogger):
    def __init__(self):
        self._cache = {}
    
    @abstractmethod
    def process(self, data: T) -> Optional[T]:
        pass
    
    def cached_compute(self, key: str) -> Optional[T]:
        if key in self._cache:
            return self._cache[key]
        return None`

	// Write test file
	if err := mocks.EnsureParentDir(pythonFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(pythonFile, []byte(pythonContent), 0644); err != nil {
		t.Fatalf("Failed to write advanced.py: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{pythonFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[pythonFile]
	require.True(t, ok, "File context not found for %s", pythonFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"MetaLogger", string(types.Class)},
		{"ComplexService", string(types.Class)},
		{"singleton", string(types.Modifier)},
		{"process", string(types.Function)},
		{"cached_compute", string(types.Function)},
		{"log_calls", string(types.Function)},
	}

	// Test all symbols in a loop
	for _, sym := range symbols {
		mocks.AssertSymbol(t, fileCtx.Symbols, sym.name, sym.typ)
	}

	// Verify imports are correctly detected
	assert.Contains(t, fileCtx.Imports, "functools")
	assert.Contains(t, fileCtx.Imports, "typing")
	assert.Contains(t, fileCtx.Imports, "abc")
}
