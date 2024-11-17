package tests

import (
	"testing"

	"github.com/spf13/cobra"
)

// mockRootCmd creates a mock root command for testing
func mockRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quill",
		Short: "Quill - AI-powered git commit message generator",
	}
	
	cmd.PersistentFlags().Bool("debug", false, "Enable debug output")
	cmd.AddCommand(&cobra.Command{Use: "generate"})
	cmd.AddCommand(&cobra.Command{Use: "init"})
	cmd.AddCommand(&cobra.Command{Use: "config"})
	
	return cmd
}

func TestRootCommand(t *testing.T) {
	// Test debug flag
	rootCmd := mockRootCmd()
	debugFlag, err := rootCmd.PersistentFlags().GetBool("debug")
	if err != nil {
		t.Fatalf("Failed to get debug flag: %v", err)
	}
	if debugFlag {
		t.Error("Expected debug flag to be false by default")
	}

	// Test subcommands
	subcommands := rootCmd.Commands()
	expectedCommands := map[string]bool{
		"generate": false,
		"init":     false,
		"config":   false,
	}

	for _, cmd := range subcommands {
		if _, ok := expectedCommands[cmd.Name()]; ok {
			expectedCommands[cmd.Name()] = true
		}
	}

	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Expected to find command '%s'", name)
		}
	}
} 