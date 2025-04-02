// Reads, analyzes, parses, and caches repository context
package factories

import (
	c "context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	ch "github.com/jabafett/quill/internal/utils/cache"
	"github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/internal/utils/debug"
)

// AnalyzerOptions contains configuration for analyzers
type AnalyzerOptions struct {
	MaxConcurrency int    // Maximum number of concurrent analysis routines
	CacheEnabled   bool   // Whether to use caching
	BasePath       string // Base path for relative imports (usually repo root)
	CachePath      string // Path to the cache directory
	RepoRootPath   string // Absolute path to the repository root for cache key namespacing
}

// ContextEngine manages codebase context extraction and caching
type ContextEngine struct {
	cache        *ch.Cache
	analyzer     *context.DefaultAnalyzer
	options      AnalyzerOptions
	mu           sync.RWMutex
	analyzerPool sync.Pool
}

// NewContextEngine creates a new context engine instance
func NewContextEngine(opts ...func(*AnalyzerOptions)) (*ContextEngine, error) {
	var cache *ch.Cache
	var err error

	// Default options
	options := AnalyzerOptions{
		MaxConcurrency: runtime.NumCPU(),
		CacheEnabled:   true,
		BasePath:       "",
		CachePath:      "",
	}

	// Apply option overrides
	for _, opt := range opts {
		opt(&options)
	}

	if options.CacheEnabled {
		if options.CachePath != "" {
			cache, err = ch.NewCacheWithPath(options.CachePath)
		} else {
			cache, err = ch.NewCache()
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}
	}

	engine := &ContextEngine{
		cache:    cache,
		analyzer: context.NewDefaultAnalyzer(),
		options:  options,
		mu:       sync.RWMutex{},
	}

	// Initialize analyzer pool
	engine.analyzerPool = sync.Pool{
		New: func() interface{} {
			return context.NewDefaultAnalyzer()
		},
	}

	return engine, nil
}

// ExtractContext analyzes files and builds context
// forceReindex: If true, bypasses the cache check for individual files.
func (e *ContextEngine) ExtractContext(files []string, forceReindex bool) (*context.RepositoryContext, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to analyze")
	}

	ctx := &context.RepositoryContext{
		Files:  make(map[string]*context.FileContext),
		Errors: make([]context.Error, 0),
	}

	// Create buffered channels
	filesChan := make(chan string, len(files))
	resultsChan := make(chan *analyzeResult, len(files))
	errorsChan := make(chan context.Error, len(files))

	// Start worker pool
	var wg sync.WaitGroup
	workerCount := min(e.options.MaxConcurrency, len(files))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		// Pass forceReindex to the worker
		go e.analyzeWorker(filesChan, resultsChan, errorsChan, &wg, forceReindex)
	}

	// Send files to workers
	for _, file := range files {
		if shouldSkipDir(file) || shouldSkipFile(file) {
			continue
		}
		filesChan <- file
	}
	close(filesChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Track unique imports and dependencies
	importSet := context.NewImportSet()

	// Collect results
	validFiles := 0

	for result := range resultsChan {
		if result.fileCtx != nil {
			// Normalize imports for this file
			fileImports := context.NewImportSet()
			for _, imp := range result.fileCtx.Imports {
				fileImports.Add(imp)
				importSet.Add(imp) // Add to global set
			}

			// Update file context with normalized imports
			result.fileCtx.Imports = fileImports.Imports()

			// Add to repository context
			ctx.Files[result.path] = result.fileCtx
			validFiles++
		}
	}

	// Set repository-wide dependencies
	ctx.Dependencies = make([]context.Dependency, 0)
	for _, imp := range importSet.Dependencies() {
		ctx.Dependencies = append(ctx.Dependencies, context.Dependency{
			Name: imp,
		})
	}

	// Collect errors
	for err := range errorsChan {
		ctx.Errors = append(ctx.Errors, err)
	}

	// Get language statistics and metrics from the analyzer
	ctx.Languages.Primary, ctx.Languages.Others = e.analyzer.GetLanguages()
	ctx.Metrics.TotalFiles = e.analyzer.GetTotalFiles()
	ctx.Metrics.TotalLines = e.analyzer.GetTotalLines()

	return ctx, nil
}

type analyzeResult struct {
	path    string
	fileCtx *context.FileContext
	_       error
}

func (e *ContextEngine) analyzeWorker(files <-chan string, results chan<- *analyzeResult, errors chan<- context.Error, wg *sync.WaitGroup, forceReindex bool) {
	defer wg.Done()

	// Get analyzer from pool
	analyzer := e.analyzerPool.Get().(*context.DefaultAnalyzer)
	defer func() {
		// Merge stats back to main analyzer before returning to pool
		e.mu.Lock()
		for lang, count := range analyzer.GetLanguageStats() {
			e.analyzer.AddLanguageStats(lang, count)
		}
		e.analyzer.AddTotalFiles(analyzer.GetTotalFiles())
		e.analyzer.AddTotalLines(analyzer.GetTotalLines())
		e.mu.Unlock()
		e.analyzerPool.Put(analyzer)
	}()

	for path := range files {
		result := &analyzeResult{path: path}

		// Check cache first if enabled AND not forcing reindex
		var cacheKey string
		cacheHit := false
		if e.options.CacheEnabled && !forceReindex { // Skip cache read if forceReindex is true
			var cached context.FileContext
			// Use repoRootPath and relative path for namespaced cache key
			// Ensure RepoRootPath is set, otherwise skip caching for this file
			if e.options.RepoRootPath != "" {
				cacheKey = fmt.Sprintf("file_context:%s:%s", e.options.RepoRootPath, path)
				if err := e.cache.Get(cacheKey, &cached); err == nil {
					debug.Log("ContextFactory.analyzeWorker: Cache HIT for %s (Key: %s). Cached UpdatedAt: %s", path, cacheKey, cached.UpdatedAt.Format(time.RFC3339Nano))
					result.fileCtx = &cached
					results <- result
					cacheHit = true
					// continue // Don't continue here, let the loop finish naturally
				} else {
					debug.Log("ContextFactory.analyzeWorker: Cache MISS for %s (Key: %s)", path, cacheKey)
				}
			} else {
				// Log if RepoRootPath is missing when cache is enabled
				log.Printf("Warning: RepoRootPath not set in AnalyzerOptions for %s, skipping cache check.", path)
			}
		}

		// If cache hit (and not forcing), skip analysis
		if cacheHit {
			continue
		}

		// If cache miss or forceReindex is true, proceed with analysis
		debug.Log("ContextFactory.analyzeWorker: Analyzing %s (forceReindex: %v)", path, forceReindex)

		// Create new context for each file
		ctx, cancel := c.WithCancel(c.Background())

		// Analyze file with proper error handling
		fileCtx, err := analyzer.Analyze(ctx, path)
		cancel() // Cancel context after analysis

		if err != nil {
			errors <- context.Error{
				Message: err.Error(),
			}
			continue
		}

		// If no error, proceed with logging and setting the result context
		debug.Log("ContextFactory.analyzeWorker: Analyzed %s. Received UpdatedAt: %s", path, fileCtx.UpdatedAt.Format(time.RFC3339Nano))
		result.fileCtx = fileCtx

		// Cache the result if enabled and analysis was successful
		if e.options.CacheEnabled && cacheKey != "" && result.fileCtx != nil {
			debug.Log("ContextFactory.analyzeWorker: Caching context for %s (Key: %s) with UpdatedAt: %s", path, cacheKey, result.fileCtx.UpdatedAt.Format(time.RFC3339Nano))
			if err := e.cache.Set(cacheKey, result.fileCtx); err != nil {
				// Log caching error, but don't block the main flow
				log.Printf("Warning: Failed to cache context for %s (key: %s): %v", path, cacheKey, err)
			}
		}

		// Send result only if analysis was successful
		if result.fileCtx != nil {
			results <- result
		}
	}
}

// 

// GetCachedContext retrieves a cached value if valid and deserializes it into the provided type
func (e *ContextEngine) GetCachedContext(key string, value interface{}) error {
	return e.cache.Get(key, value)
}

// SetCachedContext caches a value
func (e *ContextEngine) SetCachedContext(key string, value interface{}) error {
	return e.cache.Set(key, value)
}

func shouldSkipDir(path string) bool {
	base := filepath.Base(path)
	// Let git handle ignoring dotfiles via .gitignore
	return base == "node_modules" || base == ".git" || base == "vendor" ||
		base == "build" || base == "dist" || base == "target"
}

func shouldSkipFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, ".min.js") ||
		strings.HasSuffix(base, ".min.css") ||
		strings.HasSuffix(base, ".map")
}

// Option functions for configuring the engine
func WithMaxConcurrency(n int) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		if n > 0 {
			opts.MaxConcurrency = n
		}
	}
}

// Exposed for test
func WithCache(enabled bool) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.CacheEnabled = enabled
	}
}

// Exposed for test
func WithBasePath(path string) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.BasePath = path
	}
}

// Exposed for test
func WithCachePath(path string) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.CachePath = path
	}
}

// WithRepoRootPath sets the repository root path for cache namespacing
func WithRepoRootPath(path string) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.RepoRootPath = path
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
