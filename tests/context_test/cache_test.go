package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jabafett/quill/internal/utils/cache"
	"github.com/jabafett/quill/internal/utils/context"
)

func TestCache(t *testing.T) {
	cache, err := cache.NewCache()
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	defer cache.Close()

	// Test basic set/get
	t.Run("Basic Set/Get", func(t *testing.T) {
		ctx := &context.FileContext{
			Path:          "test.go",
			Type:          "go",
			Symbols:       []context.SymbolContext{{Name: "TestFunc"}},
		}

		if err := cache.Set("test.go", ctx); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var got context.FileContext
		if err := cache.Get("test.go", &got); err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		// Verify fields
		if got.Path != ctx.Path {
			t.Errorf("Path = %v, want %v", got.Path, ctx.Path)
		}
		if got.Type != ctx.Type {
			t.Errorf("Type = %v, want %v", got.Type, ctx.Type)
		}
		if len(got.Symbols) != len(ctx.Symbols) {
			t.Errorf("Got %d symbols, want %d", len(got.Symbols), len(ctx.Symbols))
		}
	})

	// Test non-existent key
	t.Run("Non-existent Key", func(t *testing.T) {
		var got context.FileContext
		if err := cache.Get("nonexistent.go", &got); err == nil {
			t.Error("Expected error for non-existent key")
		}
	})

	// Test TTL
	t.Run("TTL", func(t *testing.T) {
		ctx := &context.FileContext{
			Path:          "expire.go",
			Type:          "go",
		}

		if err := cache.Set("expire.go", ctx); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Verify immediately available
		var got context.FileContext
		if err := cache.Get("expire.go", &got); err != nil {
			t.Error("Expected cache hit immediately after set")
		}

		// Wait a bit and verify still available (shouldn't expire too quickly)
		time.Sleep(1 * time.Second)
		if err := cache.Get("expire.go", &got); err != nil {
			t.Error("Cache expired too quickly")
		}
	})

	// Test concurrent access
	t.Run("Concurrent Access", func(t *testing.T) {
		const numGoroutines = 10
		errCh := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				key := fmt.Sprintf("concurrent-%d.go", i)
				ctx := &context.FileContext{
					Path:          key,
					Type:          "go",
				}

				if err := cache.Set(key, ctx); err != nil {
					errCh <- fmt.Errorf("Set() error = %v", err)
					return
				}

				var got context.FileContext
				if err := cache.Get(key, &got); err != nil {
					errCh <- fmt.Errorf("Get() error = %v", err)
					return
				}
				errCh <- nil
			}(i)
		}

		// Collect errors
		for i := 0; i < numGoroutines; i++ {
			if err := <-errCh; err != nil {
				t.Error(err)
			}
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete-test.go"
		ctx := &context.FileContext{
			Path: key,
			Type: "go",
		}

		// Set and verify
		if err := cache.Set(key, ctx); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var got context.FileContext
		if err := cache.Get(key, &got); err != nil {
			t.Fatalf("Get() error before delete = %v", err)
		}

		// Delete and verify
		if err := cache.Delete(key); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if err := cache.Get(key, &got); err == nil {
			t.Error("Expected error after delete")
		}
	})
}
