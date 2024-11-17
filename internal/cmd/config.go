package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Quill configuration",
	Long:  `View and modify Quill configuration settings`,
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runSet,
}

func init() {
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := viper.Get(key)
	if value == nil {
		return fmt.Errorf("key '%s' not found in configuration", key)
	}
	fmt.Printf("%v\n", value)
	return nil
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load existing config
	configFile := filepath.Join(os.Getenv("HOME"), ".config", ".quill.toml")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Set new value
	viper.Set(key, value)

	// Save config
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
} 