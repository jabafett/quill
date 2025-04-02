package langs_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestRustComplexAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(2),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	if err != nil {
		t.Fatalf("NewContextEngine() error = %v", err)
	}

	// Create test directory and file
	testDir := t.TempDir()
	rustFile := filepath.Join(testDir, "src/complex.rs")
	rustContent := `
use std::sync::Arc;
use tokio::sync::Mutex;

pub trait DataProcessor<T> {
    fn process(&self, data: T) -> Result<T, String>;
    fn validate(&self) -> bool;
}

pub struct AsyncProcessor<T> {
    inner: Arc<Mutex<T>>,
    retries: u32,
}

impl<T: Clone + Send + 'static> AsyncProcessor<T> {
    pub fn new(data: T) -> Self {
        Self {
            inner: Arc::new(Mutex::new(data)),
            retries: 3,
        }
    }

    async fn process_internal(&self) -> Result<(), String> {
        let mut attempts = 0;
        while attempts < self.retries {
            if let Ok(_) = self.inner.try_lock() {
                return Ok(());
            }
            attempts += 1;
        }
        Err("Max retries exceeded".to_string())
    }
}

impl<T: Clone + Send + 'static> DataProcessor<T> for AsyncProcessor<T> {
    fn process(&self, data: T) -> Result<T, String> {
        if !self.validate() {
            return Err("Validation failed".to_string());
        }
        Ok(data)
    }

    fn validate(&self) -> bool {
        true
    }
}`

	// Write test file
	if err := mocks.EnsureParentDir(rustFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(rustFile, []byte(rustContent), 0644); err != nil {
		t.Fatalf("Failed to write complex.rs: %v", err)
	}
	// Extract context
	ctx, err := engine.ExtractContext([]string{rustFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}
	// Get the file context
	fileCtx, ok := ctx.Files[rustFile]
	require.True(t, ok, "File context not found for %s", rustFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"DataProcessor", string(context.Interface)},
		{"AsyncProcessor", string(context.Class)},
		{"process", string(context.Function)},
		{"validate", string(context.Function)},
		{"process_internal", string(context.Function)},
		{"new", string(context.Function)},
	}

	imports := []struct {
		name string
	}{
		{"std::sync::Arc"},
		{"tokio::sync::Mutex"},
	}
	// Test all imports in a loop
	for _, sym := range imports {
		assert.Contains(t, fileCtx.Imports, sym.name, "Missing symbol %q", sym.name)
	}
	// Test all symbols in a loop
	for _, sym := range symbols {
		mocks.AssertSymbol(t, fileCtx.Symbols, sym.name, sym.typ)
	}
	if t.Failed() {
		jsonCtx, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			t.Errorf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonCtx))
	}
}
