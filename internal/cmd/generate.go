package cmd

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/ui"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/helpers"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate commit messages for staged changes",
	Long: `Generate commit messages for staged changes using AI.
This command analyzes your staged changes and generates appropriate commit messages.

Examples:
  # Generate commit messages with default settings
  quill generate

  # Use a specific AI provider
  quill generate --provider anthropic

  # Generate multiple commit message variations
  quill generate --candidates 3

  # Adjust generation temperature
  quill generate --temperature 0.7`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringP("provider", "p", "", "Override default AI provider (gemini, anthropic, openai, ollama)")
	generateCmd.Flags().IntP("candidates", "c", 2, "Number of commit message variations to generate (1-3)")
	generateCmd.Flags().Float32P("temperature", "t", 0, "Generation temperature (0.0-1.0, 0 for default)")

	generateCmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"gemini", "anthropic", "openai", "ollama"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func runGenerate(cmd *cobra.Command, args []string) error {
	debug.Log("Starting generate command")

	// Get flag values
	providerVal, candidatesVal, temperatureVal, err := helpers.SetGenerateFlagValues[string, int, float32](
		cmd,
		"provider",
		"candidates",
		"temperature",
	)
	if err != nil {
		return fmt.Errorf("failed to get flags: %w", err)
	}

	// Create generate factory with options
	generator, err := factories.NewGenerateFactory(factories.GenerateOptions{
		Provider:    providerVal,
		Candidates:  candidatesVal,
		Temperature: temperatureVal,
	})
	if err != nil {
		if strings.Contains(err.Error(), "no git repository found") {
			return fmt.Errorf("no git repository found")
		}
		return fmt.Errorf("failed to create generate factory: %w", err)
	}

	spinner := ui.NewProgressSpinner()
	spinner.Start("Generating commit messages")

	// Generate messages
	msgs, err := generator.Generate(context.Background())
	if err != nil {
		spinner.Error("No staged changes found")
		if _, ok := err.(helpers.ErrNoStagedChanges); ok {
			return fmt.Errorf("no staged changes found")
		}
		return fmt.Errorf("failed to generate commit messages: %w", err)
	}

	spinner.Success("Generated commit messages")

	// Create an interactive model for message selection
	model := ui.NewCommitMessageModel(msgs)
	p := tea.NewProgram(
		model,
		tea.WithMouseCellMotion(), // Enable mouse support
	)

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

	cmd.Printf("Successfully created commit: %s\n", selectedModel.Selected())
	return nil
}
