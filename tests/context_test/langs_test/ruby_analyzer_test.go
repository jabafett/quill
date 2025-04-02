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

func TestRubyComplexAnalyzer(t *testing.T) {
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
	rubyFile := filepath.Join(testDir, "ruby/advanced.rb")
	rubyContent := `
require 'singleton'
require 'logger'

module Services
  # Define log levels
  module LogLevel
    DEBUG = 0
    INFO = 1
    WARNING = 2
    ERROR = 3
  end

  # Configuration class for services
  class Configuration
    attr_accessor :level, :path

    def initialize(level: LogLevel::INFO, path: nil)
      @level = level
      @path = path
    end
  end

  # Generic service module
  module ServiceInterface
    def process(data)
      raise NotImplementedError, "#{self.class} has not implemented method '#{__method__}'"
    end
  end

  # Service decorator for logging
  module LoggingDecorator
    def self.included(base)
      base.class_eval do
        alias_method :original_process, :process

        def process(data)
          puts "Processing data in #{self.class}"
          result = original_process(data)
          puts "Completed processing in #{self.class}"
          result
        end
      end
    end
  end

  # Complex service implementation
  class ComplexService
    include Singleton
    include ServiceInterface
    include LoggingDecorator
    
    def initialize
      @cache = {}
    end

    def process(data)
      # Implementation
      data
    end

    def cached_compute(key)
      @cache[key]
    end

    private

    def cache_result(key, value)
      @cache[key] = value
    end
  end

  # Custom error types
  class ServiceError < StandardError; end
  class ValidationError < ServiceError; end
end`

	// Write test file
	if err := mocks.EnsureParentDir(rubyFile); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(rubyFile, []byte(rubyContent), 0644); err != nil {
		t.Fatalf("Failed to write advanced.rb: %v", err)
	}

	// Extract context
	ctx, err := engine.ExtractContext([]string{rubyFile}, false)
	if err != nil {
		t.Fatalf("ExtractContext() error = %v", err)
	}

	// Get the file context
	fileCtx, ok := ctx.Files[rubyFile]
	require.True(t, ok, "File context not found for %s", rubyFile)

	// Define expected symbols
	symbols := []struct {
		name string
		typ  string
	}{
		{"Services", string(types.Module)},
		{"LogLevel", string(types.Module)},
		{"Configuration", string(types.Class)},
		{"ServiceInterface", string(types.Module)},
		{"LoggingDecorator", string(types.Module)},
		{"ComplexService", string(types.Class)},
		{"ServiceError", string(types.Class)},
		{"ValidationError", string(types.Class)},
		{"process", string(types.Function)},
		{"cached_compute", string(types.Function)},
		{"DEBUG", string(types.Constant)},
		{"INFO", string(types.Constant)},
		{"WARNING", string(types.Constant)},
		{"ERROR", string(types.Constant)},
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
