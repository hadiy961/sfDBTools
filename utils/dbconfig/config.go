package dbconfig

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// ResolveConfig resolves dbconfig configuration from various sources
func ResDBConfigFlag(cmd *cobra.Command) (*Config, error) {
	config := &Config{}

	// Get file path from flag
	if cmd.Flags().Changed("file") {
		filePath, _ := cmd.Flags().GetString("file")
		config.FilePath = filePath
	}

	// Get delete flags if they exist
	if cmd.Flags().Lookup("force") != nil {
		config.ForceDelete, _ = cmd.Flags().GetBool("force")
	}
	if cmd.Flags().Lookup("all") != nil {
		config.DeleteAll, _ = cmd.Flags().GetBool("all")
	}

	// Get generate flags if they exist
	if cmd.Flags().Lookup("name") != nil {
		config.ConfigName, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Lookup("host") != nil {
		config.Host, _ = cmd.Flags().GetString("host")
	}
	if cmd.Flags().Lookup("port") != nil {
		config.Port, _ = cmd.Flags().GetInt("port")
	}
	if cmd.Flags().Lookup("user") != nil {
		config.User, _ = cmd.Flags().GetString("user")
	}
	if cmd.Flags().Lookup("auto") != nil {
		config.AutoMode, _ = cmd.Flags().GetBool("auto")
	}

	return config, nil
}

// ResolveConfigWithInteractiveSelection resolves config with interactive file selection if needed
func ResolveConfigWithInteractiveSelection(cmd *cobra.Command) (*Config, error) {
	config, err := ResDBConfigFlag(cmd)
	if err != nil {
		return nil, err
	}

	// If no file path provided, use interactive selection
	if config.FilePath == "" {
		selectedFile, err := common.SelectConfigFileInteractive()
		if err != nil {
			return nil, fmt.Errorf("failed to select config file: %w", err)
		}
		config.FilePath = selectedFile
	}

	return config, nil
}

// SelectConfigOrUseDefaults tries to select a config file interactively or returns empty for defaults
func SelectConfigOrUseDefaults() (string, error) {
	// Try to find encrypted config files
	configDir, err := config.GetDatabaseConfigDirectory()
	if err != nil {
		return "", fmt.Errorf("failed to get database config directory: %w", err)
	}

	encFiles, err := common.FindEncryptedConfigFiles(configDir)
	if err != nil {
		return "", fmt.Errorf("failed to find encrypted config files: %w", err)
	}

	if len(encFiles) == 0 {
		fmt.Println("ℹ️  No encrypted configuration files found.")
		fmt.Println("   You can use --config flag to specify a config file,")
		fmt.Println("   or use individual connection flags (--source_host, --source_user, etc.)")
		return "", nil
	}

	// Show selection
	return common.SelectConfigFileInteractive()
}
