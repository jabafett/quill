package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jabafett/quill/internal/utils/context"
)

func TestCache(t *testing.T) {
	// Create temporary directory for cache
	tmpDir := t.TempDir()
	
	// Override NewCache to use temp directory
	cache, err := context.NewCacheWithPath(tmpDir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	defer cache.Close() // Ensure cleanup

	// Test data
	testCtx := &context.FileContext{
		Path:       "test.go",
		Type:       "go",
		Complexity: 5,
		UpdatedAt:  time.Now(),
		Symbols: []context.Symbol{
			{
				Name:      "TestFunc",
				Type:      "function",
				StartLine: 1,
				EndLine:   10,
			},
		},
	}

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		cache.Set("test.go", testCtx)
		got := cache.Get("test.go")
		if got == nil {
			t.Fatal("Get() returned nil")
		}
		if got.Path != testCtx.Path {
			t.Errorf("Path = %v, want %v", got.Path, testCtx.Path)
		}
	})

	// Test TTL expiration
	t.Run("TTL Expiration", func(t *testing.T) {
		cache.Set("expire.go", testCtx)
		time.Sleep(2 * time.Second) // Wait for potential TTL
		if got := cache.Get("expire.go"); got == nil {
			t.Error("Entry expired too quickly")
		}
	})

	// Test concurrent access
	t.Run("Concurrent Access", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				key := fmt.Sprintf("concurrent-%d", i)
				cache.Set(key, testCtx)
				got := cache.Get(key)
				if got == nil {
					t.Errorf("Get() returned nil for key %s", key)
				}
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
} 