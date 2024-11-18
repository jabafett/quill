package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jabafett/quill/internal/ai"
	"github.com/jabafett/quill/internal/templates"
	"github.com/jabafett/quill/internal/config"
	"github.com/jabafett/quill/internal/debug"
	"github.com/jabafett/quill/internal/git"
	"github.com/jabafett/quill/internal/ui"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generate commit messages from staged changes",
	Long: `Generate commit messages by analyzing your staged git changes using AI.
    
The command analyzes your git diff and generates appropriate conventional commit 
messages based on the changes in your staging area. You can select from multiple
generated variations in an interactive interface.

Features:
- Analyzes staged changes only (git add)
- Supports multiple AI providers (Gemini, Claude, GPT-4)
- Interactive selection of generated messages
- Follows Conventional Commits specification
- Configurable generation parameters

Conventional Commit Types:
- feat:     A new feature
- fix:      A bug fix
- docs:     Documentation changes
- style:    Code style/formatting changes
- refactor: Code refactoring
- test:     Adding/modifying tests
- chore:    Maintenance tasks
- perf:     Performance improvements
- ci:       CI/CD changes
- build:    Build system changes
- revert:   Revert a previous commit

The scope (optional) describes the affected component or module.`,
	Example: `  # Generate with default settings
  quill generate

  # Use a specific provider
  quill gen --provider gemini

  # Generate more variations
  quill g --candidates 3

  # Adjust generation temperature
  quill generate --temperature 0.7`,
	RunE: runGenerate,
}

var (
	providerFlag    string
	candidatesFlag  int
	temperatureFlag float32
)

func init() {
	generateCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Override default AI provider (gemini, anthropic, openai)")
	generateCmd.Flags().IntVarP(&candidatesFlag, "candidates", "c", 2, "Number of commit message variations to generate (1-3)")
	generateCmd.Flags().Float32VarP(&temperatureFlag, "temperature", "t", 0, "Generation temperature (0.0-1.0, 0 for default)")

	generateCmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"gemini", "anthropic", "openai"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func runGenerate(cmd *cobra.Command, args []string) error {
	debug.Log("Starting generate command")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	debug.Dump("config", cfg)

	// Initialize git repository
	if !git.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}
	debug.Log("Git repository detected")

	repo, err := git.NewRepository()
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Check for staged changes
	hasChanges, err := repo.HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	debug.Log("Staged changes found: %v", hasChanges)

	if !hasChanges {
		return fmt.Errorf("no staged changes found")
	}

	// Create AI provider
	providerName := providerFlag
	if providerName == "" {
		providerName = cfg.Core.DefaultProvider
	}

	provider, err := ai.NewProvider(providerName, cfg)
	if err != nil {
		return fmt.Errorf("failed to create AI provider: %w", err)
	}

	debug.Log("Using provider: %s", providerName)

	// Create commit message generator
	generator := templates.NewCommitMessageGenerator(provider, repo)

	// Get number of candidates
	candidates := candidatesFlag
	if candidates <= 0 {
		candidates = cfg.Core.DefaultCandidates
	}
	if candidates > 3 {
		candidates = 3 // enforce maximum
	}

	// Prepare generation options
	opts := ai.GenerateOptions{
		MaxCandidates: candidates,
	}
	if temperatureFlag > 0 {
		opts.Temperature = &temperatureFlag
	}

	spinner := ui.NewProgressSpinner()
	spinner.Start("Generating commit messages")

	// Generate commit messages
	msgs, err := generator.GenerateStagedCommitMessage(context.Background(), opts)
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to generate commit messages: %w", err)
	}

	spinner.Success("Generated commit messages")

	// Create an interactive model for message selection
	model := ui.NewCommitMessageModel(msgs)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run interactive UI: %w", err)
	}

	// Get selected message
	selectedModel := finalModel.(ui.CommitMessageModel)
	if selectedModel.Selected() == "" {
		return fmt.Errorf("no commit message selected")
	}

	// Create git commit
	err = repo.Commit(selectedModel.Selected())
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	cmd.Printf("Successfully created commit: %s\n", selectedModel.Selected())
	return nil
}
