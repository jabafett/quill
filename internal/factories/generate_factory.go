package factories

import (
	"context"
	"fmt"
	"sync"

	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/internal/utils/config"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/git"
	"github.com/jabafett/quill/internal/utils/helpers"
)

// GenerateFactory handles the generation of commit messages
type GenerateFactory struct {
	config    *config.Config
	repo      *git.Repository
	templates *TemplateFactory
	provider  Provider
}

// GenerateOptions contains options for the generate factory
type GenerateOptions struct {
	Provider    string
	Candidates  int
	Temperature float32
}

// NewGenerateFactory creates a new factory specifically for the generate command
func NewGenerateFactory(opts GenerateOptions) (*GenerateFactory, error) {
	var (
		cfg       *config.Config
		repo      *git.Repository
		templates *TemplateFactory
		provider  Provider
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

	// Initialize git repository
	go func() {
		defer wg.Done()
		var err error
		repo, err = git.NewRepository()
		if err != nil {
			errChan <- err
		}
		debug.Log("Finished initializing git repository")
	}()

	// Initialize template factory
	go func() {
		defer wg.Done()
		var err error
		templates, err = NewTemplateFactory()
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

	// Update the

	// Create provider with the loaded config
	provider, err := NewProvider(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	debug.Log("Finished creating provider")

	return &GenerateFactory{
		config:    cfg,
		repo:      repo,
		templates: templates,
		provider:  provider,
	}, nil
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
		"Diff":         diff,
		"AddedLines":   added,
		"DeletedLines": deleted,
		"Files":        files,
	}

	// Generate prompt from template
	prompt, err := f.templates.Generate(CommitMessageType, data)
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
