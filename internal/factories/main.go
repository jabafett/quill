package factories

import (
	"fmt"
	"github.com/jabafett/quill/internal/utils/git"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/config"
)

// Factory is the main factory that coordinates AI operations
type Factory struct {
	Config     *config.Config
	Repo       *git.Repository
	Templates  *TemplateFactory
	Provider   Provider
}

// NewFactory creates a new factory instance
func NewFactory() (*Factory, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create template factory
	templateFactory, err := NewTemplateFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create template factory: %w", err)
	}

	// Create provider factory
	provider, err := NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider factory: %w", err)
	}

	// Initialize git repository
	if !git.IsGitRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	debug.Log("Git repository detected")

	repo, err := git.NewRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	if repo == nil {
		return nil, fmt.Errorf("git repository is nil")
	}

	f := &Factory{
		Config:     cfg,
		Repo:       repo,
		Templates:  templateFactory,
		Provider:   provider,
	}

	return f, nil
}

// GenerateCommitPrompt creates a commit message prompt from staged changes
func (f *Factory) GenerateCommitPrompt() (string, error) {
	if f.Repo == nil {
		return "", fmt.Errorf("git repository not initialized")
	}

	// Get diff and stats
	diff, err := f.Repo.GetStagedDiff()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	if diff == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	added, deleted, files, err := f.Repo.GetStagedDiffStats()
	if err != nil {
		return "", fmt.Errorf("failed to get diff stats: %w", err)
	}

	// For simple file deletions, return a standard message
	if len(files) == 1 && added == 0 && deleted > 0 {
		return fmt.Sprintf("chore: remove %s", files[0]), nil
	}

	data := TemplateData{
		Files:   files,
		Added:   added,
		Deleted: deleted,
		Diff:    diff,
	}

	if debug.IsDebug() {
		prompt, err := f.Templates.Generate(CommitMessageType, data)
		if err != nil {
			return "", fmt.Errorf("failed to generate commit message: %w", err)
		}
		fmt.Println(prompt)
		return prompt, nil
	}

	return f.Templates.Generate(CommitMessageType, data)
}

// CreateProvider creates a new AI provider instance
func (f *Factory) CreateProvider(name string) (Provider, error) {
	return NewProvider(f.Config)
}
