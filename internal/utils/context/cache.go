package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	defaultTTL = 24 * time.Hour // 24 hour TTL
)

// Cache provides TTL-based caching using Badger
type Cache struct {
	db *badger.DB
}

// NewCache creates a new cache instance using Badger
func NewCache() (*Cache, error) {
	cacheDir := filepath.Join(os.TempDir(), "quill-cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return NewCacheWithPath(cacheDir)
}

// NewCacheWithPath creates a new cache instance using Badger with a custom path
func NewCacheWithPath(path string) (*Cache, error) {
	opts := badger.DefaultOptions(path)
	opts.NumMemtables = 2
	opts.NumLevelZeroTables = 2
	opts.NumLevelZeroTablesStall = 3
	opts.ValueLogFileSize = 10 << 20
	opts.BaseTableSize = 20 << 20
	opts.SyncWrites = true

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache db: %w", err)
	}

	cache := &Cache{db: db}
	go cache.runGC()
	return cache, nil
}

// Get retrieves a cached context if valid
func (c *Cache) Get(path string) *FileContext {
	var ctx FileContext

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(path))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &ctx)
		})
	})
	if err != nil {
		return nil
	}

	return &ctx
}

// Set adds or updates a cached context
func (c *Cache) Set(path string, ctx *FileContext) error {
	data, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(path), data).WithTTL(defaultTTL)
		return txn.SetEntry(e)
	})
}

// Close closes the underlying Badger database
func (c *Cache) Close() error {
	return c.db.Close()
}

// runGC periodically runs Badger's garbage collection
func (c *Cache) runGC() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
	again:
		err := c.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}