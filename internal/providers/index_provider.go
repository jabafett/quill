package providers

import (
	c "context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/utils/config"
	"github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/git"
)

// IndexProvider handles the repository indexing process
type IndexProvider struct {
	config        *config.Config
	repo          *git.Repository
	contextEngine *factories.ContextEngine
	repoRootPath  string
}

type IndexProviderOptions struct {
	RepoRootPath string
	CachePath    string
	BasePath     string
}

// NewIndexProvider creates a new provider for the index command
func NewIndexProvider(opts ...func(*IndexProviderOptions)) (*IndexProvider, error) {
	var (
		cfg           *config.Config
		repo          *git.Repository
		contextEngine *factories.ContextEngine
		repoRootPath  string
		errChan       = make(chan error, 4) // Increased buffer for potential repoRootPath error
		wg            sync.WaitGroup
	)

	// Default options
	options := IndexProviderOptions{
		RepoRootPath: "",
		BasePath:     "",
		CachePath:    "",
	}

	// Apply option overrides
	for _, opt := range opts {
		opt(&options)
	}

	debug.Dump("options", options)

	debug.Log("Starting index provider initialization")

	// Load components concurrently
	wg.Add(2)

	// Load configuration
	go func() {
		defer wg.Done()
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			errChan <- fmt.Errorf("failed to load config: %w", err)
			return
		}
		debug.Log("Finished loading configuration")
	}()

	// Initialize git object and get root path
	go func() {
		defer wg.Done()
		var err error
		repo, err = git.NewRepository(options.RepoRootPath)
		if err != nil {
			errChan <- fmt.Errorf("failed to initialize git repository: %w", err)
			return
		}
		// Get repo root path after repo is initialized
		repoRootPath, err = repo.GetRepoRootPath()
		if err != nil {
			errChan <- fmt.Errorf("failed to get repo root path: %w", err)
			return
		}
	}()

	// Wait for concurrent tasks
	wg.Wait()
	close(errChan)

	// Check for any errors during initialization
	for err := range errChan {
		return nil, err
	}

	var err error
	contextEngine, err = factories.NewContextEngine(
		factories.WithCachePath(options.CachePath), // Assuming GetPath() exists or is added
		factories.WithBasePath(repoRootPath),       // Use repo root as base path
		factories.WithRepoRootPath(repoRootPath),   // Pass repo root for cache key namespacing
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create context engine: %w", err)
	}
	debug.Log("Finished initializing context engine")

	return &IndexProvider{
		config:        cfg,
		repo:          repo,
		contextEngine: contextEngine,
		repoRootPath:  repoRootPath,
	}, nil
}

// IndexRepository analyzes the repository and updates the cached context
func (p *IndexProvider) IndexRepository(ctx c.Context, forceReindex bool) error {
	debug.Log("Starting repository indexing for: %s", p.repoRootPath)
	startTime := time.Now()

	// 1. Retrieve Previous Context
	var previousContext context.RepositoryContext
	repoCacheKey := fmt.Sprintf("repo_context:%s", p.repoRootPath)
	err := p.contextEngine.GetCachedContext(repoCacheKey, &previousContext)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			debug.Log("No previous repository context found in cache for key: %s", repoCacheKey)
			// Initialize empty context if not found
			previousContext = context.RepositoryContext{
				Files: make(map[string]*context.FileContext),
			}
		} else {
			return fmt.Errorf("failed to get previous repository context from cache: %w", err)
		}
	} else {
		debug.Log("Loaded previous repository context from cache (key: %s)", repoCacheKey)
	}

	// 2. List Files
	trackedFiles := p.repo.GetNonIgnoredFiles()
	debug.Log("Found %d tracked files", len(trackedFiles))

	// 3. Determine Files to Analyze
	filesToAnalyze := make([]string, 0)
	skippedFiles := make(map[string]*context.FileContext) // Store context of skipped files

	for _, fileRelPath := range trackedFiles {
		// Construct absolute path for os.Stat
		fileAbsPath := filepath.Join(p.repoRootPath, fileRelPath)

		// Check if we should force analysis
		if forceReindex {
			filesToAnalyze = append(filesToAnalyze, fileRelPath)
			continue
		}

		// Get file modification time
		fileInfo, err := os.Stat(fileAbsPath)
		if err != nil {
			// Log error but continue; maybe the file was deleted?
			debug.Log("Warning: Failed to stat file %s: %v", fileAbsPath, err)
			// If file doesn't exist, we don't need to analyze it, but also remove from previous context if present
			delete(previousContext.Files, fileRelPath)
			continue
		}
		modTime := fileInfo.ModTime()

		// Check against previous context
		if cachedFileCtx, exists := previousContext.Files[fileRelPath]; exists {
			// Compare file system mod time with cached mod time directly.
			// If the filesystem timestamp hasn't changed, we skip.
			if modTime.Equal(cachedFileCtx.ModTime) {
				// File hasn't changed, skip analysis and keep cached context
				skippedFiles[fileRelPath] = cachedFileCtx
				// debug.Log("Skipping unchanged file: %s (FS: %s, Cache: %s)", fileRelPath, modTime, cachedFileCtx.ModTime) // Can be noisy
				continue
			}
			debug.Log("File changed, needs re-analysis: %s (FS: %s, Cache: %s)", fileRelPath, modTime, cachedFileCtx.ModTime)
		} else {
			debug.Log("New file detected, needs analysis: %s", fileRelPath)
		}

		// If not skipped, add to analysis list
		filesToAnalyze = append(filesToAnalyze, fileRelPath)
	}

	debug.Log("Files to analyze: %d, Skipped files: %d", len(filesToAnalyze), len(skippedFiles))

	// 4. Analyze Files
	var analyzedContext *context.RepositoryContext
	if len(filesToAnalyze) > 0 {
		debug.Log("Analyzing %d files... (forceReindex: %v)", len(filesToAnalyze), forceReindex)
		var analysisErr error
		// Pass forceReindex flag to ExtractContext
		analyzedContext, analysisErr = p.contextEngine.ExtractContext(filesToAnalyze, forceReindex)
		if analysisErr != nil {
			// Log error but potentially continue to save partial context?
			// For now, let's return the error.
			return fmt.Errorf("error during file analysis: %w", analysisErr)
		}
		debug.Log("Analysis complete. Analyzed context has %d files.", len(analyzedContext.Files))
		// TODO: Handle analyzedContext.Errors if needed
	} else {
		debug.Log("No files needed analysis.")
		// Initialize empty if nothing was analyzed
		analyzedContext = &context.RepositoryContext{
			Files: make(map[string]*context.FileContext),
		}
	}

	// 5. Merge Contexts
	finalContext := context.RepositoryContext{
		Name:          previousContext.Name, // Preserve previous metadata for now
		Description:   previousContext.Description,
		RepositoryURL: previousContext.RepositoryURL,
		Visibility:    previousContext.Visibility,
		Files:         make(map[string]*context.FileContext),
		// Metrics and Languages will be overwritten by the latest analysis run's aggregate
		Metrics:   analyzedContext.Metrics,
		Languages: analyzedContext.Languages,
		// Preserve previous errors? Or only show new ones? Let's keep new ones for now.
		Errors: analyzedContext.Errors,
		// Dependencies will be recalculated based on merged files
	}

	// Add newly analyzed files
	for path, fileCtx := range analyzedContext.Files {
		finalContext.Files[path] = fileCtx
	}

	// Add skipped files (from cache)
	for path, fileCtx := range skippedFiles {
		finalContext.Files[path] = fileCtx
	}

	// Recalculate overall dependencies from the final merged file set
	importSet := context.NewImportSet()
	for _, fileCtx := range finalContext.Files {
		if fileCtx != nil {
			for _, imp := range fileCtx.Imports {
				importSet.Add(imp)
			}
		}
	}
	finalContext.Dependencies = make([]context.Dependency, 0, len(importSet.Imports()))
	for _, depName := range importSet.Dependencies() {
		finalContext.Dependencies = append(finalContext.Dependencies, context.Dependency{Name: depName})
	}

	// TODO: 6. Add/Update Repository Metadata (missing some fields from RepositoryContext)
	finalContext.Name, err = p.repo.GetRepoName()
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}
	finalContext.VersionControl.Branch, err = p.repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// 7. Update Metrics/Languages (already done by taking from analyzedContext)
	// Note: This assumes ContextEngine correctly aggregates metrics for the analyzed subset.
	// If we skipped all files, we should probably retain previous metrics/languages.
	if len(filesToAnalyze) == 0 {
		finalContext.Metrics = previousContext.Metrics
		finalContext.Languages = previousContext.Languages
	} else {
		// If some files were analyzed, update total counts based on the final merged set
		finalContext.Metrics.TotalFiles = len(finalContext.Files)
		// Recalculating total lines would require reading skipped files, skip for now.
		// finalContext.Metrics.TotalLines = calculateTotalLines(finalContext.Files)
	}

	// 8. Persist Context
	debug.Log("Persisting final repository context to cache (key: %s)", repoCacheKey)
	err = p.contextEngine.SetCachedContext(repoCacheKey, finalContext)
	if err != nil {
		return fmt.Errorf("failed to persist final repository context to cache: %w", err)
	}

	debug.Dump("Final context", finalContext)

	// Build context graph
	err = p.BuildContextGraph(&finalContext)
	if err != nil {
		return fmt.Errorf("failed to build context graph: %w", err)
	}

	debug.Log("Repository indexing finished successfully in %s", time.Since(startTime))
	return nil
}

// BuildContextGraph builds the context graph from the repository context
func (p *IndexProvider) BuildContextGraph(finalContext *context.RepositoryContext) error {
	graph, err := context.BuildDependencyGraph(finalContext)
	if err != nil {
		return fmt.Errorf("failed to build context graph: %w", err)
	}
	repoGraphKey := fmt.Sprintf("context_graph:%s", p.repoRootPath)
	err = p.contextEngine.SetCachedContext(repoGraphKey, graph)
	if err != nil {
		return fmt.Errorf("failed to persist context graph to cache: %w", err)
	}
	return nil
}

// GetCachedContext retrieves a cached value if valid and deserializes it into the provided type
func (p *IndexProvider) GetCachedContext(key string, value interface{}) error {
	return p.contextEngine.GetCachedContext(key, value)
}

// SetCachedContext caches a value
func (p *IndexProvider) SetCachedContext(key string, value interface{}) error {
	return p.contextEngine.SetCachedContext(key, value)
}

// Exposed for test
func WithRepoRootPath(path string) func(*IndexProviderOptions) {
	return func(opts *IndexProviderOptions) {
		opts.RepoRootPath = path
	}
}

// Exposed for test
func WithCachePath(path string) func(*IndexProviderOptions) {
	return func(opts *IndexProviderOptions) {
		opts.CachePath = path
	}
}

// Exposed for test
func WithBasePath(path string) func(*IndexProviderOptions) {
	return func(opts *IndexProviderOptions) {
		opts.BasePath = path
	}
}
