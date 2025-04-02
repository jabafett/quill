package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/jabafett/quill/internal/factories"
	"github.com/dgraph-io/badger/v4"
	"github.com/jabafett/quill/internal/factories"

	c "github.com/jabafett/quill/internal/utils/context"

	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/internal/utils/config"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/git"
	"github.com/jabafett/quill/internal/utils/helpers"
)

// GenerateFactory handles the generation of commit messages
type GenerateFactory struct {
	config          *config.Config
	repo            *git.Repository
	templates       *factories.TemplateFactory
	provider        factories.Provider
	contextProvider *factories.ContextProvider
}

// NewGenerateFactory creates a new factory specifically for the generate command
func NewGenerateFactory(opts factories.ProviderOptions) (*GenerateFactory, error) {
	var (
		cfg       *config.Config
		repo      *git.Repository
		templates *factories.TemplateFactory
		provider  factories.Provider
		errChan   = make(chan error, 4)
		wg        sync.WaitGroup
	)

	debug.Log("Starting generate factory")

	// Load components concurrently
	wg.Add(3)

	// Load configuration
	go func() {
		defer wg.Done()
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			errChan <- fmt.Errorf("failed to load config: %w", err)
		}
		debug.Log("Finished loading configuration")
	}()

	// Initialize git object
	go func() {
		defer wg.Done()
		var err error
		repo, err = git.NewRepository(".")
		if err != nil {
			errChan <- err
		}
		debug.Log("Finished initializing git repository")
	}()

	// Initialize template factory
	go func() {
		defer wg.Done()
		var err error
		templates, err = factories.NewTemplateFactory()
		if err != nil {
			errChan <- fmt.Errorf("failed to create template factory: %w", err)
		}
		debug.Log("Finished initializing template factory")
	}()

	// Wait for config to be loaded before creating provider
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		return nil, err
	}

	// Create provider with the loaded config
	provider, err := factories.NewProvider(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	debug.Log("Finished creating provider")

	// Initialize context engine
	repoRootPath, err := repo.GetRepoRootPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo root path: %w", err)
	}

	contextEngine, err := factories.NewContextEngine(
		factories.WithBasePath(repoRootPath),
		factories.WithRepoRootPath(repoRootPath),
	)
	if err != nil {
		debug.Log("Warning: Failed to create context engine: %v. Proceeding without repository context.", err)
	}

	factory := &GenerateFactory{
		config:        cfg,
		repo:          repo,
		templates:     templates,
		provider:      provider,
		contextEngine: contextEngine,
	}

	// Attempt to load cached context
	if factory.contextEngine != nil {
		var loadedCtx c.RepositoryContext
		repoCacheKey := fmt.Sprintf("repo_context:%s", repoRootPath)
		err := factory.contextEngine.GetCachedContext(repoCacheKey, &loadedCtx)
		if err == nil {
			debug.Log("Successfully loaded cached repository context for generate command.")
			factory.repoCtx = &loadedCtx
		} else if err == badger.ErrKeyNotFound {
			debug.Log("No cached repository context found (key: %s). Run 'quill index' first for context-aware generation.", repoCacheKey)
		} else {
			debug.Log("Warning: Error loading cached repository context: %v", err)
		}
	}

	return factory, nil
}

// Generate generates commit messages based on staged changes
func (f *GenerateFactory) Generate(ctx context.Context) ([]string, error) {
	// Check for staged changes
	_, err := f.repo.HasStagedChanges()
	if err != nil {
		if _, ok := err.(helpers.ErrNoStagedChanges); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to check staged changes: %w", err)
	}

	// Get diff and stats
	diff, err := f.repo.GetStagedDiff()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff: %w", err)
	}

	added, deleted, files, err := f.repo.GetStagedDiffStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff stats: %w", err)
	}

	debug.Log("Diff stats - Added: %d, Deleted: %d, Files: %d", added, deleted, len(files))

	// Prepare template data
	data := map[string]interface{}{
		"Diff":           diff,
		"AddedLines":     added,
		"DeletedLines":   deleted,
		"Files":          files,
		"RelatedContext": "", // Default to empty string
	}

	// If graph exists, extract context based on changed files
	if f.depGraph != nil && len(files) > 0 {
		debug.Log("Extracting related context from graph for changed files: %v", files)
		// Use a reasonable depth, e.g., 1 or 2
		contextString, err := f.depGraph.GetContextForFiles(files, 2)
		if err != nil {
			debug.Log("Warning: Failed to get context from graph: %v", err)
		} else {
			data["RelatedContext"] = contextString
			debug.Log("Added related context to prompt data\nContext: %s", contextString)
		}
	} else if f.repoCtx != nil && f.depGraph == nil {
		debug.Log("Repository context loaded, but dependency graph failed to build or is nil")
	} else {
		debug.Log("No dependency graph available for context extraction")
	}

	// Generate prompt from template
	prompt, err := f.templates.Generate(factories.CommitMessageType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit prompt: %w", err)
	}

	// Generate messages using the AI provider
	opts := ai.GenerateOptions{
		MaxCandidates: f.config.Core.DefaultCandidates,
	}
	if temp := f.config.Providers[f.config.Core.DefaultProvider].Temperature; temp > 0 {
		opts.Temperature = &temp
	}
	return f.provider.Generate(ctx, prompt, opts)
}
