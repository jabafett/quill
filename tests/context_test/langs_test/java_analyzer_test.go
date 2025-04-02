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
	types "github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/tests/mocks"
)

func TestJavaComplexAnalyzer(t *testing.T) {
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
	javaFile := filepath.Join(testDir, "java/ComplexService.java")
	javaContent := `
package com.example.service;

import java.util.concurrent.CompletableFuture;
import java.util.Optional;
import java.util.function.Function;
import javax.inject.Singleton;

@Singleton
public class ComplexService<T> implements DataProcessor<T> {
    private final Cache<String, T> cache;
    private final MetricsCollector metricsCollector;

    @Inject
    public ComplexService(
            Cache<String, T> cache,
            MetricsCollector metricsCollector) {
        this.cache = cache;
        this.metricsCollector = metricsCollector;
    }

    @Override
    @Transactional(isolation = Isolation.REPEATABLE_READ)
    public CompletableFuture<Optional<T>> processAsync(T data) {
        return CompletableFuture
            .supplyAsync(() -> validate(data))
            .thenApply(valid -> {
                if (!valid) {
                    return Optional.empty();
                }
                return Optional.of(data);
            });
    }
}`

	// Write test file
	if err := mocks.EnsureParentDir(javaFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(javaFile, []byte(javaContent), 0644); err != nil {
		t.Fatalf("Failed to write ComplexService.java: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{javaFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[javaFile]
	require.True(t, ok, "File context not found for %s", javaFile)

	// Define expected symbols
	expectedSymbols := []struct {
		name string
		typ  string
	}{
		{"ComplexService", string(types.Class)},
		{"Singleton", string(types.Modifier)},
		{"Inject", string(types.Modifier)},
		{"Override", string(types.Modifier)},
		{"Transactional", string(types.Modifier)},
		{"processAsync", string(types.Function)},
		{"cache", string(types.Field)},
		{"metricsCollector", string(types.Field)},
	}

	// Test all symbols in a loop
	for _, sym := range expectedSymbols {
		mocks.AssertSymbol(t, fileCtx.Symbols, sym.name, sym.typ)
	}

	// Verify imports are correctly detected
	assert.Contains(t, fileCtx.Imports, "java.util.concurrent.CompletableFuture")
	assert.Contains(t, fileCtx.Imports, "java.util.Optional")
	assert.Contains(t, fileCtx.Imports, "javax.inject.Singleton")
	assert.Contains(t, fileCtx.Imports, "java.util.function.Function")

	if t.Failed() {
		jsonCtx, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			t.Errorf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonCtx))
	}
}
