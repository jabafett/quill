package cmd

import (
	"github.com/spf13/cobra"
	"github.com/jabafett/quill/internal/utils/debug"
)

var rootCmd = &cobra.Command{
	Use:   "quill",
	Short: "Quill - AI-powered git commit message generator",
	Long:  `Quill is a CLI tool that leverages git diff and AI to intelligently generate commit messages.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		debugMode, _ := cmd.Flags().GetBool("debug")
		debug.Initialize(debugMode)
	}

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
}

// GetRootCmd exposes the root command for testing
func GetRootCmd() *cobra.Command {
	return rootCmd
}
