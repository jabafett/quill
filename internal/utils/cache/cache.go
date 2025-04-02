package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	defaultTTL = 72 * time.Hour // 72 hour TTL
)

// badgerLogger implements the badger.Logger interface
type badgerLogger struct {
	*log.Logger
}

func (l *badgerLogger) Errorf(f string, v ...interface{})   { l.Printf(f, v...) }
func (l *badgerLogger) Warningf(f string, v ...interface{}) { l.Printf(f, v...) }
func (l *badgerLogger) Infof(f string, v ...interface{})    { l.Printf(f, v...) }
func (l *badgerLogger) Debugf(f string, v ...interface{})   { l.Printf(f, v...) }

// Cache provides TTL-based caching using Badger
type Cache struct {
	db      *badger.DB
	logFile *os.File
	path    string // Store the cache path
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
// Checks if the cache directory exists, if not it creates it
// Creates a badger-log directory in the tmp dir if it doesn't exist
// Creates a temp log file in the badger-log directory if it doesn't exist
// Opens the badger db with the given path
// Runs the garbage collector periodically
// Returns a new cache instance
func NewCacheWithPath(path string) (*Cache, error) {
	if err := os.MkdirAll("/tmp/badger-logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create badger-logs directory: %w", err)
	}
	logFile, err := os.CreateTemp("/tmp/badger-logs", "badger-*.log")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp log file: %w", err)
	}

	logger := &badgerLogger{log.New(logFile, "", log.LstdFlags)}

	opts := badger.DefaultOptions(path)
	opts.NumMemtables = 2
	opts.NumLevelZeroTables = 2
	opts.NumLevelZeroTablesStall = 3
	opts.ValueLogFileSize = 10 << 20 // 10MB
	opts.BaseTableSize = 20 << 20    // 20MB
	opts.SyncWrites = true           // Ensure writes are synced
	opts.Logger = logger

	db, err := badger.Open(opts)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to open cache db: %w", err)
	}

	cache := &Cache{
		db:      db,
		logFile: logFile,
		path:    path, // Store the path
	}
	go cache.runGC()
	return cache, nil
}

// Get retrieves a cached value if valid and deserializes it into the provided type
func (c *Cache) Get(key string, value interface{}) error {
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err // Key not found or other error
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, value)
		})
	})
	if err != nil {
		// Don't wrap badger.ErrKeyNotFound, let caller handle it
		if err == badger.ErrKeyNotFound {
			return err
		}
		return fmt.Errorf("failed to get value for key '%s': %w", key, err)
	}
	return nil
}

// Set adds or updates a cached value with the default TTL
func (c *Cache) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key '%s': %w", key, err)
	}

	err = c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), data).WithTTL(defaultTTL)
		return txn.SetEntry(e)
	})
	if err != nil {
		return fmt.Errorf("failed to set value for key '%s': %w", key, err)
	}
	return nil
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		// Don't wrap badger.ErrKeyNotFound
		if err == badger.ErrKeyNotFound {
			return nil // Deleting a non-existent key is not an error
		}
		return fmt.Errorf("failed to delete key '%s': %w", key, err)
	}
	return nil
}

// Close closes the underlying Badger database and log file
func (c *Cache) Close() error {
	// Close DB first
	dbErr := c.db.Close()
	// Always attempt to close log file
	logErr := c.logFile.Close()

	if dbErr != nil {
		return fmt.Errorf("failed to close cache db: %w", dbErr)
	}
	if logErr != nil {
		// Log or return this error if DB closed successfully?
		// For now, prioritize DB error.
		return fmt.Errorf("failed to close cache log file: %w", logErr)
	}
	return nil
}

// GetPath returns the path used for the cache directory
func (c *Cache) GetPath() string {
	return c.path
}

// runGC periodically runs Badger's garbage collection
func (c *Cache) runGC() {
	// Run initial GC shortly after startup
	time.Sleep(1 * time.Minute)
	c.performGC()

	// Then run periodically
	ticker := time.NewTicker(1 * time.Hour) // Run GC less frequently
	defer ticker.Stop()
	for range ticker.C {
		c.performGC()
	}
}

// performGC runs the Badger value log garbage collection
func (c *Cache) performGC() {
	// Keep running GC until Badger returns an error (likely indicating completion)
	for {
		err := c.db.RunValueLogGC(0.5) // Lower threshold for more frequent cleanup if needed
		if err != nil {
			// badger.ErrNoRewrite is expected when GC completes
			if err == badger.ErrNoRewrite {
				log.Printf("Badger GC completed for %s", c.path)
			} else {
				log.Printf("Badger GC error for %s: %v", c.path, err)
			}
			break // Exit loop on error or completion
		}
	}
}
