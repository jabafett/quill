package providers

import (
	c "context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/internal/utils/config"
	"github.com/jabafett/quill/internal/utils/context"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/git"
)

// IndexProvider handles the repository indexing process
type IndexProvider struct {
	config          *config.Config
	repo            *git.Repository
	contextProvider *factories.ContextProvider
	repoRootPath    string
	aiProvider      factories.Provider
}

type IndexProviderOptions struct {
	RepoRootPath string
}

// NewIndexProvider creates a new provider for the index command
func NewIndexProvider(opts ...func(*IndexProviderOptions)) (*IndexProvider, error) {
	var (
		cfg             *config.Config
		repo            *git.Repository
		contextProvider *factories.ContextProvider
		repoRootPath    string
		errChan         = make(chan error, 3)
		wg              sync.WaitGroup
	)

	// Default options
	options := IndexProviderOptions{
		RepoRootPath: "",
	}

	// Apply option overrides
	for _, opt := range opts {
		opt(&options)
	}

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

	// Create context provider
	var err error
	contextProvider, err = factories.NewContextProvider(
		factories.WithRepoRootPath(repoRootPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create context provider: %w", err)
	}
	debug.Log("Finished initializing context provider")

	// Create AI provider
	aiProvider, err := factories.NewProvider(cfg, factories.ProviderOptions{
		Provider: cfg.Core.DefaultProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	return &IndexProvider{
		config:          cfg,
		repo:            repo,
		contextProvider: contextProvider,
		repoRootPath:    repoRootPath,
		aiProvider:      aiProvider,
	}, nil
}

// IndexRepository generates a summary of the repository using AI
func (p *IndexProvider) IndexRepository(ctx c.Context, forceReindex bool) error {
	debug.Log("Starting repository indexing for: %s", p.repoRootPath)
	startTime := time.Now()

	// Skip if summary exists and not forcing reindex
	if !forceReindex && p.contextProvider.HasSummary() {
		debug.Log("Repository summary already exists. Use --force to regenerate.")
		return nil
	}

	// Get repository information
	repoName, err := p.repo.GetRepoName()
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	// Get directory tree
	dirTree, err := p.contextProvider.GetDirectoryTree(3) // Limit to 3 levels deep
	if err != nil {
		return fmt.Errorf("failed to get directory tree: %w", err)
	}

	// Get file info
	fileInfo, err := p.contextProvider.GetFileInfo()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Count files
	fileCount := 0
	for _, count := range fileInfo {
		fileCount += count
	}

	// Get language info
	languages, err := p.contextProvider.GetLanguageInfo()
	if err != nil {
		return fmt.Errorf("failed to get language info: %w", err)
	}

	// Get README content if available, searching subfolders case-insensitively for the first match
	// find <path> -type f \( -iname README.md -o -iname README \) -print -quit
	// Explanation:
	// p.repoRootPath:  Start searching here
	// -type f:         Find only files
	// \( ... \):       Group conditions
	// -iname README.md: Find README.md (case-insensitive)
	// -o:              OR
	// -iname README:   Find README (case-insensitive)
	// -print:          Print the path
	// -quit:           Exit immediately after the first match is found
	readmeCmd := exec.Command("find", p.repoRootPath, "-type", "f", `(`, "-iname", "README.md", "-o", "-iname", "README", `)`, "-print", "-quit")
	readmePathBytes, err := readmeCmd.Output()
	var readmeContent string
	if err == nil && len(readmePathBytes) > 0 {
		readmePath := strings.TrimSpace(string(readmePathBytes))
		if readmePath != "" {
			debug.Log("Attempting to read README content from: %s", readmePath)
			catCmd := exec.Command("cat", readmePath)
			readmeBytes, catErr := catCmd.Output()
			if catErr == nil {
				readmeContent = string(readmeBytes)
				debug.Log("Successfully read README content.")
			} else {
				debug.Log("Failed to 'cat' README file '%s': %v", readmePath, catErr)
			}
		}
	} else {
		if err != nil {
			debug.Log("Error executing find command for README: %v", err)
		} else {
			debug.Log("No README file found in the repository.")
		}
	}

	prompt := fmt.Sprintf(`Generate a concise summary of this repository:

Repository Name: %s
File Count: %d
Languages: %v

Directory Structure:
%s

Readme Content:
%s

The summary should describe what this repository is for, its main components, and its purpose.
Focus on the technical aspects and be specific about what the code does based on the directory structure and languages used.
`, repoName, fileCount, languages, dirTree, readmeContent)

	// Generate summary using AI
	debug.Log("Generating repository summary using AI...\nPrompt: %s", prompt)
	summary, err := p.aiProvider.Generate(ctx, prompt, ai.GenerateOptions{
		MaxCandidates: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to generate repository summary: %w", err)
	}

	if len(summary) == 0 {
		return fmt.Errorf("AI provider returned empty summary")
	}

	// Create repository summary
	repoSummary := &context.RepoSummary{
		Name:        repoName,
		Description: summary[0],
		Files:       fileCount,
		Directories: len(dirTree),
		Languages:   languages,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save summary
	err = p.contextProvider.SaveSummary(repoSummary)
	if err != nil {
		return fmt.Errorf("failed to save repository summary: %w", err)
	}

	debug.Log("Repository indexing finished successfully in %s", time.Since(startTime))
	return nil
}

// GetRepoSummary returns the repository summary
func (p *IndexProvider) GetRepoSummary() string {
	return p.contextProvider.GetRepoSummary()
}

// HasSummary checks if a repository summary exists
func (p *IndexProvider) HasSummary() bool {
	return p.contextProvider.HasSummary()
}

// WithRepoRootPath sets the repository root path
func WithRepoRootPath(path string) func(*IndexProviderOptions) {
	return func(opts *IndexProviderOptions) {
		opts.RepoRootPath = path
	}
}
