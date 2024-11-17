package cmd

import (
	"context"
	"fmt"

	"github.com/jabafett/quill/internal/ai"
	"github.com/jabafett/quill/internal/config"
	"github.com/jabafett/quill/internal/git"
	"github.com/jabafett/quill/internal/ui"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generate commit messages from staged changes",
	RunE:    runGenerate,
}

func init() {
	generateCmd.Flags().BoolP("interactive", "i", false, "Interactive mode")
	generateCmd.Flags().StringP("provider", "p", "", "Override default AI provider")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	debug, _ := cmd.Flags().GetBool("debug")
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Show progress spinner
	spinner := ui.NewProgressSpinner()
	spinner.Start("Analyzing repository")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize git repository
	if !git.IsGitRepo() {
		spinner.Error(fmt.Errorf("not a git repository"))
		return fmt.Errorf("not a git repository")
	}

	repo, err := git.NewRepository()
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Check for staged changes
	hasChanges, err := repo.HasStagedChanges()
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	if !hasChanges {
		spinner.Error(fmt.Errorf("no staged changes found"))
		return fmt.Errorf("no staged changes found. Use 'git add' to stage changes")
	}

	// Get git diff
	spinner.Start("Generating diff")
	diff, err := repo.GetStagedDiff()
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Create AI provider
	providerName, _ := cmd.Flags().GetString("provider")
	if providerName == "" {
		providerName = cfg.Core.DefaultProvider
	}

	provider, err := ai.NewProvider(providerName, cfg)
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to create AI provider: %w", err)
	}

	if debug {
		fmt.Println("Using provider:", providerName)
	}

	spinner.Start("Generating commit message")

	// Generate commit message
	msg, err := provider.GenerateCommitMessage(context.Background(), diff)
	if err != nil {
		spinner.Error(err)
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	spinner.Success("Generated commit message")

	if interactive {
		// Create an interactive model for message selection
		model := ui.NewCommitMessageModel([]string{msg})
		p := tea.NewProgram(model)
		
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run interactive UI: %w", err)
		}
		return nil
	}

	// Print the message
	if debug {
		fmt.Println("\nCommit message:")
	}
	fmt.Println(msg)

	return nil
}