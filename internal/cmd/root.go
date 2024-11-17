package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "quill",
	Short: "Quill - AI-powered git commit message generator",
	Long: `Quill is a CLI tool that leverages git diff and AI to intelligently 
generate commit messages. It supports multiple AI providers with seamless 
configuration-driven functionality.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
}
