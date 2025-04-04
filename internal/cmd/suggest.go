package cmd

import (
        "context"
        "fmt"
        "os/exec"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/jabafett/quill/internal/factories"
        "github.com/jabafett/quill/internal/providers"
        "github.com/jabafett/quill/internal/ui"
        "github.com/jabafett/quill/internal/utils/debug"
        "github.com/jabafett/quill/internal/utils/helpers"
        "github.com/spf13/cobra"
)

var suggestCmd = &cobra.Command{
        Use:   "suggest",
        Short: "Suggest commit groupings for changes",
        Long: `Suggest logical commit groupings for staged and unstaged changes using AI.
This command analyzes your repository changes and suggests how to group them into commits.

Examples:
  # Suggest commit groupings with default settings
  quill suggest

  # Only consider staged changes
  quill suggest --staged-only

  # Only consider unstaged changes
  quill suggest --unstaged-only

  # Use a specific AI provider
  quill suggest --provider anthropic

  # Generate multiple grouping suggestions
  quill suggest --candidates 3

  # Adjust generation temperature
  quill suggest --temperature 0.7`,
        RunE: runSuggest,
}

func init() {
        suggestCmd.Flags().StringP("provider", "p", "", "Override default AI provider (gemini, anthropic, openai, ollama)")
        suggestCmd.Flags().IntP("candidates", "c", 2, "Number of grouping suggestions to generate (1-3)")
        suggestCmd.Flags().Float32P("temperature", "t", 0, "Generation temperature (0.0-1.0, 0 for default)")
        suggestCmd.Flags().BoolP("staged-only", "s", false, "Only consider staged changes")
        suggestCmd.Flags().BoolP("unstaged-only", "u", false, "Only consider unstaged changes")

        suggestCmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
                return []string{"gemini", "anthropic", "openai", "ollama"}, cobra.ShellCompDirectiveNoFileComp
        })
}

func runSuggest(cmd *cobra.Command, args []string) error {
        debug.Log("Starting suggest command")

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

        stagedOnly, err := cmd.Flags().GetBool("staged-only")
        if err != nil {
                return fmt.Errorf("failed to get staged-only flag: %w", err)
        }

        unstagedOnly, err := cmd.Flags().GetBool("unstaged-only")
        if err != nil {
                return fmt.Errorf("failed to get unstaged-only flag: %w", err)
        }

        // Validate flags
        if stagedOnly && unstagedOnly {
                return fmt.Errorf("cannot use both --staged-only and --unstaged-only flags")
        }

        // Create suggest factory with options
        suggester, err := providers.NewSuggestFactory(factories.ProviderOptions{
                Provider:     providerVal,
                Candidates:   candidatesVal,
                Temperature:  temperatureVal,
                StagedOnly:   stagedOnly,
                UnstagedOnly: unstagedOnly,
        })
        if err != nil {
                return fmt.Errorf("failed to create suggest factory: %w", err)
        }

        // Generate suggestions
        suggestions, err := suggester.Suggest(context.Background())
        if err != nil {
                if _, ok := err.(helpers.ErrNoChanges); ok {
                        return fmt.Errorf("no changes found to suggest groupings for")
                }
                return fmt.Errorf("failed to generate suggestions: %w", err)
        }

        // Create an interactive model for suggestion selection
        model := ui.NewSuggestModel(suggestions)
        p := tea.NewProgram(
                model,
                tea.WithMouseCellMotion(), // Enable mouse support
        )

        finalModel, err := p.Run()
        if err != nil {
                return fmt.Errorf("failed to run interactive UI: %w", err)
        }

        // Get selected suggestion
        selectedModel := finalModel.(ui.SuggestModel)
        if selectedModel.Quitting() {
                fmt.Print("\033[H\033[2J")
                return fmt.Errorf("operation cancelled")
        }

        // If the user selected a suggestion, apply it
        if selectedModel.HasSelection() {
                selected := selectedModel.Selected()
                cmd.Printf("Selected grouping: %s\n", selected.Description)
                
                // If the user wants to stage the suggested files
                if selected.ShouldStage {
                        for _, file := range selected.Files {
                                cmd.Printf("Staging file: %s\n", file)
                                // Execute git add for each file
                                stageCmd := exec.Command("git", "add", file)
                                err := stageCmd.Run()
                                if err != nil {
                                        fmt.Printf("Warning: Failed to stage file %s: %v\n", file, err)
                                }
                        }
                        fmt.Println("\nFiles staged successfully. You can now run 'git status' to verify.")
                        if selected.Message != "" {
                                fmt.Printf("\nSuggested commit message: %s\n", selected.Message)
                                fmt.Println("\nTo commit with this message, run:")
                                fmt.Printf("  git commit -m \"%s\"\n", selected.Message)
                        }
                } else {
                        // Just show the suggested message
                        if selected.Message != "" {
                                fmt.Printf("\nSuggested commit message: %s\n", selected.Message)
                                fmt.Println("\nTo stage and commit these changes, run:")
                                for _, file := range selected.Files {
                                        fmt.Printf("  git add %s\n", file)
                                }
                                fmt.Printf("  git commit -m \"%s\"\n", selected.Message)
                        }
                }
        }

        return nil
}