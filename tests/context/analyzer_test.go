package context_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	c "context"

	"github.com/jabafett/quill/internal/utils/context"
)

func TestAnalyzer(t *testing.T) {
	analyzer := context.NewDefaultAnalyzer()

	// Create test files
	tmpDir := t.TempDir()
	testFiles := map[string]string{
		"test.go": `package main
		
func main() {
    println("Hello")
}

type User struct {
    Name string
}`,
		"test.js": `
// Function declaration
function greet(name) {
    console.log("Hello " + name);
}

// Class declaration
class User {
    constructor(name) {
        this.name = name;
    }
    
    sayHello() {
        console.log("Hello from " + this.name);
    }
}

// Variable function
const handler = function(event) {
    console.log(event);
};

// Arrow function
const process = (data) => {
    return data.map(x => x * 2);
};
`,
		"test.py": `def greet(name):
    print(f"Hello {name}")

class User:
    def __init__(self, name):
        self.name = name`,
	}

	// Write test files
	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	tests := []struct {
		name           string
		file           string
		minSymbols     int
		wantType       string
		wantComplexity bool
	}{
		{
			name:           "Go file analysis",
			file:           "test.go",
			minSymbols:     2,
			wantType:       "go",
			wantComplexity: true,
		},
		{
			name:           "JavaScript file analysis",
			file:           "test.js",
			minSymbols:     2,
			wantType:       "javascript",
			wantComplexity: true,
		},
		{
			name:           "Python file analysis",
			file:           "test.py",
			minSymbols:     2,
			wantType:       "python",
			wantComplexity: true,
		},
	}

	ctx, cancel := c.WithCancel(c.Background())
	defer cancel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.file)
			result, err := analyzer.Analyze(ctx, path)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if result.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", result.Type, tt.wantType)
			}

			if len(result.Symbols) < tt.minSymbols {
				t.Errorf("Got %d symbols, want at least %d", len(result.Symbols), tt.minSymbols)
			}

			if tt.wantComplexity && result.Complexity == 0 {
				t.Error("Expected non-zero complexity")
			}

			// Verify symbol details
			for _, symbol := range result.Symbols {
				if symbol.Name == "" {
					t.Error("Symbol name should not be empty")
				}
				if symbol.StartLine == 0 || symbol.EndLine == 0 {
					t.Error("Symbol should have valid line numbers")
				}
			}
		})
	}
}

func TestAnalyzerConcurrency(t *testing.T) {
	analyzer := context.NewDefaultAnalyzer()
	defer analyzer.Close()

	ctx := c.Background()
	
	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := []byte(`package main
	
	func main() {
		println("Hello")
	}`)
	
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Test concurrent analysis
	errs := make(chan error, 10)
	var wg sync.WaitGroup
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			result, err := analyzer.Analyze(ctx, testFile)
			if err != nil {
				errs <- err
				return
			}
			if result == nil {
				errs <- fmt.Errorf("expected non-nil result")
				return
			}
		}()
	}

	// Wait for all goroutines and check for errors
	go func() {
		wg.Wait()
		close(errs)
	}()

	// Collect any errors
	for err := range errs {
		t.Error(err)
	}
}

func TestAnalyzerContextCancellation(t *testing.T) {
	analyzer := context.NewDefaultAnalyzer()
	ctx, cancel := c.WithCancel(c.Background())
	defer cancel()

	// Create a large test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.go")
	var content strings.Builder
	for i := 0; i < 10000; i++ {
		content.WriteString("func test() {}\n")
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatal(err)
	}

	// Cancel context immediately
	cancel()

	_, err := analyzer.Analyze(ctx, testFile)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}
