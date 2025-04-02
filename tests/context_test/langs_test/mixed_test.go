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

func TestContextEngineMixed(t *testing.T) {
	tmpDir := t.TempDir()
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(1),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	require.NoError(t, err)

	// Test file paths
	pythonFile := filepath.Join(tmpDir, "data_processor.py")
	rustFile := filepath.Join(tmpDir, "processor.rs")

	// Create test files
	err = mocks.EnsureParentDir(pythonFile)
	require.NoError(t, err)
	err = mocks.EnsureParentDir(rustFile)
	require.NoError(t, err)

	// Write test content for Python file
	pythonContent := `
		from typing import List, Optional
		from dataclasses import dataclass
		from abc import ABC, abstractmethod

		@dataclass
		class ProcessConfig:
			batch_size: int
			timeout: Optional[float] = None

		class DataProcessor(ABC):
			@abstractmethod
			def process(self, data: List[str]) -> List[str]:
				pass

		class BatchProcessor(DataProcessor):
			def __init__(self, config: ProcessConfig):
				self.config = config

			def process(self, data: List[str]) -> List[str]:
				return [item.upper() for item in data]

			def validate(self, data: List[str]) -> bool:
				return all(isinstance(item, str) for item in data)
	`
	err = os.WriteFile(pythonFile, []byte(pythonContent), 0644)
	require.NoError(t, err)

	// Write test content for Rust file
	rustContent := `
		use std::collections::HashMap;
		use std::sync::Arc;
		use tokio::sync::Mutex;

		pub trait DataProcessor {
			fn process(&self, data: Vec<String>) -> Vec<String>;
		}

		pub struct BatchProcessor {
			config: ProcessConfig,
			cache: Arc<Mutex<HashMap<String, String>>>,
		}

		impl BatchProcessor {
			pub fn new(config: ProcessConfig) -> Self {
				Self {
					config,
					cache: Arc::new(Mutex::new(HashMap::new())),
				}
			}

			pub async fn process_cached(&self, key: &str, data: Vec<String>) -> Vec<String> {
				let mut cache = self.cache.lock().await;
				if let Some(result) = cache.get(key) {
					return vec![result.clone()];
				}
				let result = self.process(data);
				cache.insert(key.to_string(), result[0].clone());
				result
			}
		}

		impl DataProcessor for BatchProcessor {
			fn process(&self, data: Vec<String>) -> Vec<String> {
				data.into_iter()
					.map(|s| s.to_uppercase())
					.collect()
			}
		}
	`
	err = os.WriteFile(rustFile, []byte(rustContent), 0644)
	require.NoError(t, err)

	// Analyze files
	ctx, err := engine.ExtractContext([]string{pythonFile, rustFile}, false)
	require.NoError(t, err)

	// Get Python file context
	pythonFileCtx, ok := ctx.Files[pythonFile]
	require.True(t, ok, "File context not found for %s", pythonFile)

	// Test Python Imports
	pythonImports := []string{
		"typing",
		"dataclasses",
		"abc",
	}

	// Test all Python imports
	for _, imp := range pythonImports {
		assert.Contains(t, pythonFileCtx.Imports, imp)
	}

	// Test Python symbols
	pythonSymbols := []struct {
		name string
		typ  string
	}{
		{"ProcessConfig", string(types.Class)},
		{"DataProcessor", string(types.Class)},
		{"BatchProcessor", string(types.Class)},
		{"process", string(types.Function)},
		{"validate", string(types.Function)},
	}

	// Test all Python symbols
	for _, sym := range pythonSymbols {
		mocks.AssertSymbol(t, pythonFileCtx.Symbols, sym.name, sym.typ)
	}

	// Get Rust file context
	rustFileCtx, ok := ctx.Files[rustFile]
	require.True(t, ok, "File context not found for %s", rustFile)

	// Test Rust imports
	rustImports := []string{
		"std::collections::HashMap",
		"std::sync::Arc",
		"tokio::sync::Mutex",
	}

	// Test all Rust imports
	for _, imp := range rustImports {
		assert.Contains(t, rustFileCtx.Imports, imp)
	}

	// Test Rust symbols
	rustSymbols := []struct {
		name string
		typ  string
	}{
		{"DataProcessor", string(types.Interface)},
		{"BatchProcessor", string(types.Class)},
		{"process", string(types.Function)},
		{"process_cached", string(types.Function)},
		{"new", string(types.Function)},
	}

	// Test all Rust symbols
	for _, sym := range rustSymbols {
		mocks.AssertSymbol(t, rustFileCtx.Symbols, sym.name, sym.typ)
	}
}
