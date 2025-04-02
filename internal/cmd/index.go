package cmd

import (
        "context"
        "fmt"

        "github.com/jabafett/quill/internal/providers"
        "github.com/jabafett/quill/internal/utils/debug"
        "github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
        Use:   "index",
        Short: "Generate repository summary",
        Long: `Analyzes the repository and generates an AI-powered summary.

This command examines your git repository structure, files, and languages
to create a concise summary of the repository. This summary is used by
the 'generate' command to provide more context-aware commit messages.

By default, the summary is generated once and cached. Use the --force
flag to regenerate the summary.`,
        RunE: runIndex,
}

func init() {
        indexCmd.Flags().Bool("force", false, "Force regeneration of repository summary")
        rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
        debug.Log("Starting index command")

        forceReindex, _ := cmd.Flags().GetBool("force")
        debug.Log("Force regeneration: %v", forceReindex)

        fmt.Println("Initializing index provider...")

        // Instantiate IndexProvider
        indexProvider, err := providers.NewIndexProvider()
        if err != nil {
                return fmt.Errorf("failed to create index provider: %w", err)
        }

        // Check if summary already exists
        if !forceReindex && indexProvider.HasSummary() {
                fmt.Println("Repository summary already exists. Use --force to regenerate.")
                return nil
        }

        fmt.Println("Analyzing repository and generating summary...")

        // Generate repository summary
        err = indexProvider.IndexRepository(context.Background(), forceReindex)
        if err != nil {
                return fmt.Errorf("failed to generate repository summary: %w", err)
        }

        fmt.Println("Repository summary generated successfully.")
        
        // Display the summary
        if indexProvider.HasSummary() {
                fmt.Println("\nRepository Summary:")
                fmt.Println("-------------------")
                fmt.Println(indexProvider.GetRepoSummary())
        }

        return nil
}
