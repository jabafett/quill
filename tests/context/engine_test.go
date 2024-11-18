package context_test

import (
	"path/filepath"
	"testing"
	"time"
	"os"

	"github.com/jabafett/quill/internal/factories"
)

func TestContextEngine(t *testing.T) {
	// Create temp directory for cache
	tmpDir := t.TempDir()
	
	// Create engine with test options
	engine, err := factories.NewContextEngine(
		factories.WithMaxConcurrency(2),
		factories.WithCache(true),
		factories.WithCachePath(tmpDir),
	)
	if err != nil {
		t.Fatalf("NewContextEngine() error = %v", err)
	}

	// Create test directory structure
	testDir := t.TempDir()
	testFiles := map[string]string{
		"main.go": `package main
import (
	"fmt"
	"strings"
)
func main() { fmt.Println("test") }`,
		"lib/util.go": `package lib
import (
	"fmt"
	"path/filepath"
)
func Util() { fmt.Println("util") }`,
		"lib/helper.go": `package lib
import "fmt"
func Helper() { fmt.Println("helper") }`,
	}

	var files []string
	for name, content := range testFiles {
		path := filepath.Join(testDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		files = append(files, path)
	}

	// Test context extraction
	t.Run("Extract Context", func(t *testing.T) {
		ctx, err := engine.ExtractContext(files)
		if err != nil {
			t.Fatalf("ExtractContext() error = %v", err)
		}

		// Verify file contexts
		if len(ctx.Files) != len(files) {
			t.Errorf("Got %d files, want %d", len(ctx.Files), len(files))
		}

		// Verify imports
		fmtUsers := ctx.References["fmt"]
		if len(fmtUsers) != 3 { // all three files use fmt
			t.Errorf("Expected 3 files using fmt, got %d", len(fmtUsers))
		}

		// Verify other imports
		stringsUsers := ctx.References["strings"]
		if len(stringsUsers) != 1 {
			t.Errorf("Expected 1 file using strings, got %d", len(stringsUsers))
		}

		filepathUsers := ctx.References["path/filepath"]
		if len(filepathUsers) != 1 {
			t.Errorf("Expected 1 file using path/filepath, got %d", len(filepathUsers))
		}
	})

	// Test caching
	t.Run("Cache Usage", func(t *testing.T) {
		// First extraction
		ctx1, _ := engine.ExtractContext(files)
		
		// Second extraction should use cache
		ctx2, _ := engine.ExtractContext(files)

		// Verify cache effectiveness
		for path, file := range ctx1.Files {
			cached := ctx2.Files[path]
			if cached.UpdatedAt != file.UpdatedAt {
				t.Errorf("Cache miss for %s", path)
			}
		}
	})

	// Test concurrent processing
	t.Run("Concurrent Processing", func(t *testing.T) {
		start := time.Now()
		_, err := engine.ExtractContext(files)
		if err != nil {
			t.Fatalf("ExtractContext() error = %v", err)
		}
		duration := time.Since(start)

		// With concurrency of 2, should be faster than sequential
		sequential := time.Duration(len(files) * 100) * time.Millisecond // Rough estimate
		if duration > sequential {
			t.Errorf("Concurrent processing took %v, expected less than %v", duration, sequential)
		}
	})
} 