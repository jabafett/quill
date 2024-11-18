// Reads, analyzes, parses, and caches repository context
package factories

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	c "context"
	"github.com/jabafett/quill/internal/utils/context"
)

// AnalyzerOptions contains configuration for analyzers
type AnalyzerOptions struct {
	MaxConcurrency int    // Maximum number of concurrent analysis routines
	CacheEnabled   bool   // Whether to use caching
	BasePath       string // Base path for relative imports
	CachePath      string // Add cache path option
}

// ContextEngine manages codebase context extraction and caching
type ContextEngine struct {
	cache     *context.Cache
	analyzer  *context.DefaultAnalyzer
	options   AnalyzerOptions
	mu        sync.RWMutex
}

// NewContextEngine creates a new context engine instance
func NewContextEngine(opts ...func(*AnalyzerOptions)) (*ContextEngine, error) {
	var cache *context.Cache
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
			cache, err = context.NewCacheWithPath(options.CachePath)
		} else {
			cache, err = context.NewCache()
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}
	}

	return &ContextEngine{
		cache:    cache,
		analyzer: context.NewDefaultAnalyzer(),
		options:  options,
		mu:       sync.RWMutex{},
	}, nil
}

// ExtractContext analyzes files and builds context
func (e *ContextEngine) ExtractContext(files []string) (*context.Context, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to analyze")
	}

	ctx := &context.Context{
		Files:      make(map[string]*context.FileContext),
		UpdatedAt:  time.Now(),
		Complexity: 0,
		References: make(map[string][]string),
		Errors:     make([]context.Error, 0),
	}

	// Create buffered channels
	filesChan := make(chan string, len(files))
	resultsChan := make(chan *analyzeResult, len(files))
	errorsChan := make(chan context.Error, len(files))

	// Start workers with limited concurrency
	numWorkers := min(e.options.MaxConcurrency, len(files))
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go e.analyzeWorker(filesChan, resultsChan, errorsChan, &wg)
	}

	// Send files to workers
	for _, file := range files {
		filesChan <- file
	}
	close(filesChan)

	// Wait for completion in a separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Collect results and errors
	for result := range resultsChan {
		if result.fileCtx != nil {
			ctx.Files[result.path] = result.fileCtx
			ctx.Complexity += result.fileCtx.Complexity
		}
	}

	for err := range errorsChan {
		ctx.Errors = append(ctx.Errors, err)
	}

	// Build relationships
	e.buildRelationships(ctx)
	
	return ctx, nil
}

type analyzeResult struct {
	path    string
	fileCtx *context.FileContext
	_     error
}

func (e *ContextEngine) analyzeWorker(files <-chan string, results chan<- *analyzeResult, errors chan<- context.Error, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create new context for each analysis
	ctx, cancel := c.WithCancel(c.Background())
	defer cancel()

	for path := range files {
		result := &analyzeResult{path: path}

		// Check cache first if enabled
		if e.options.CacheEnabled {
			if cached := e.cache.Get(path); cached != nil {
				result.fileCtx = cached
				results <- result
				continue
			}
		}


		// Analyze file with proper error handling
		fileCtx, err := e.analyzer.Analyze(ctx, path)
		if err != nil {
			errors <- context.Error{
				Path:    path,
				Message: err.Error(),
			}
			cancel()
			continue
		}

		// Cache the result if enabled
		if e.options.CacheEnabled && fileCtx != nil {
			if err := e.cache.Set(path, fileCtx); err != nil {
				errors <- context.Error{
					Path:    path,
					Message: fmt.Sprintf("failed to cache: %v", err),
				}
			}
		}

		result.fileCtx = fileCtx
		results <- result
	}
}

func (e *ContextEngine) buildRelationships(ctx *context.Context) {
	// Use a mutex to safely build relationships
	var relMu sync.Mutex
	var wg sync.WaitGroup

	// Process files concurrently
	semaphore := make(chan struct{}, e.options.MaxConcurrency)
	
	for path, fileCtx := range ctx.Files {
		wg.Add(1)
		go func(p string, fc *context.FileContext) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			// Process imports
			for _, imp := range fc.Imports {
				relMu.Lock()
				ctx.References[imp] = append(ctx.References[imp], p)
				relMu.Unlock()
			}
		}(path, fileCtx)
	}

	wg.Wait()
}

// Option functions for configuring the engine
		func WithMaxConcurrency(n int) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		if n > 0 {
			opts.MaxConcurrency = n
		}
	}
}

func WithCache(enabled bool) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.CacheEnabled = enabled
	}
}

func WithBasePath(path string) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.BasePath = path
	}
}

func WithCachePath(path string) func(*AnalyzerOptions) {
	return func(opts *AnalyzerOptions) {
		opts.CachePath = path
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 