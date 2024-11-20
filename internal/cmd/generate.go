package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/ui"
	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/internal/utils/debug"
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
	// Create main factory
	main, err := factories.NewFactory()
	if err != nil {
		return fmt.Errorf("failed to create main factory: %w", err)
	}

	// Check for staged changes
	hasChanges, err := main.Repo.HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	debug.Log("Staged changes found: %v", hasChanges)

	if !hasChanges {
		return fmt.Errorf("no staged changes found")
	}

	// Create AI provider through factory
	provider, err := main.CreateProvider(providerFlag)
	if err != nil {
		return fmt.Errorf("failed to create AI provider: %w", err)
	}

	debug.Log("Using provider: %s", providerFlag)

	// Get formatted prompt
	prompt, err := main.GenerateCommitPrompt()
	if err != nil {
		return fmt.Errorf("failed to get commit prompt: %w", err)
	}

	// Get number of candidates
	candidates := candidatesFlag
	if candidates <= 0 {
		candidates = main.Config.Core.DefaultCandidates
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

	// Generate messages using the AI provider
	msgs, err := provider.Generate(context.Background(), prompt, opts)
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to generate commit messages: %w", err)
	}

	spinner.Success("Generated commit messages")

	// Create an interactive model for message selection
	model := ui.NewCommitMessageModel(msgs)
	p := tea.NewProgram(model, tea.WithFPS(120))

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
	err = main.Repo.Commit(selectedModel.Selected())
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	cmd.Printf("Successfully created commit: %s\n", selectedModel.Selected())
	return nil
}
