package cmd

import (
	"context" // Add context import
	"fmt"

	"github.com/jabafett/quill/internal/providers"
	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index repository context",
	Long: `Analyzes the repository files to build a context index.

This command scans the files in your git repository, analyzes their structure
and content (respecting .gitignore), and stores an aggregated context representation
in the cache. This index is used by other commands like 'generate' and 'suggest'
to provide more context-aware results.

By default, indexing is incremental and only analyzes changed files. Use the --force
flag to re-analyze all files.`,
	RunE: runIndex,
}

func init() {
	indexCmd.Flags().Bool("force", false, "Force re-indexing of all files, ignoring cache")
	rootCmd.AddCommand(indexCmd) // Add indexCmd to rootCmd
}

func runIndex(cmd *cobra.Command, args []string) error {
	debug.Log("Starting index command")

	forceReindex, _ := cmd.Flags().GetBool("force")
	debug.Log("Force re-index: %v", forceReindex)

	fmt.Println("Initializing index provider...") // Simple feedback

	// Instantiate IndexProvider
	indexProvider, err := providers.NewIndexProvider()
	if err != nil {
		return fmt.Errorf("failed to create index provider: %w", err)
	}

	fmt.Println("Starting repository indexing...") // Simple feedback

	// Call IndexRepository
	err = indexProvider.IndexRepository(context.Background(), forceReindex)
	if err != nil {
		// Consider more specific error handling or logging here
		return fmt.Errorf("failed to index repository: %w", err)
	}

	fmt.Println("Repository indexing completed successfully.") // Simple feedback


	return nil
}
