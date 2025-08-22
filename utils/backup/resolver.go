package backup_utils

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// ConfigurationSource represents the source of database configuration
type ConfigurationSource int

const (
	SourceConfigFile ConfigurationSource = iota
	SourceFlags
	SourceDefaults
	SourceInteractive
)

// ResolveDatabaseConnection resolves database connection from various sources
func ResolveDatabaseConnection(cmd *cobra.Command) (host string, port int, user, password string, source ConfigurationSource, err error) {
	// Check if --config flag is provided
	configFile := common.GetStringFlagOrEnv(cmd, "config", "BACKUP_CONFIG", "")

	if configFile != "" {
		// Validate and load from config file
		if err := common.ValidateConfigFile(configFile); err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("invalid config file: %w", err)
		}

		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(configFile)
		if err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("failed to load config from file: %w", err)
		}

		return host, port, user, password, SourceConfigFile, nil
	}

	// Check if individual connection flags are provided
	hasConnectionFlags := cmd.Flags().Changed("source_host") ||
		cmd.Flags().Changed("source_port") ||
		cmd.Flags().Changed("source_user") ||
		cmd.Flags().Changed("source_password")

	if hasConnectionFlags {
		// Use individual flags with defaults
		host := common.GetStringFlagOrEnv(cmd, "source_host", "SOURCE_HOST", "localhost")
		port := common.GetIntFlagOrEnv(cmd, "source_port", "SOURCE_PORT", 3306)
		user := common.GetStringFlagOrEnv(cmd, "source_user", "SOURCE_USER", "root")
		password := common.GetStringFlagOrEnv(cmd, "source_password", "SOURCE_PASSWORD", "")

		return host, port, user, password, SourceFlags, nil
	}

	// Try to select config interactively
	selectedFile, err := selectConfigOrUseDefaults()
	if err != nil {
		return "", 0, "", "", SourceInteractive, err
	}

	if selectedFile != "" {
		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(selectedFile)
		if err != nil {
			return "", 0, "", "", SourceInteractive, fmt.Errorf("failed to load config from file: %w", err)
		}

		return host, port, user, password, SourceInteractive, nil
	}

	// Use defaults
	return "localhost", 3306, "root", "", SourceDefaults, nil
}

// ResolveDatabaseName resolves database name, with interactive selection if not provided
func ResolveDatabaseName(cmd *cobra.Command, host string, port int, user, password string) (string, error) {
	dbName := common.GetStringFlagOrEnv(cmd, "source_db", "SOURCE_DB", "")
	if dbName != "" {
		return dbName, nil
	}

	// Show available databases and let user choose
	dbConfig := database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

	selectedDB, err := info.SelectDatabaseInteractive(dbConfig)
	if err != nil {
		return "", fmt.Errorf("failed to select database: %w", err)
	}

	return selectedDB, nil
}

// selectConfigOrUseDefaults tries to select a config file interactively or returns empty for defaults
func selectConfigOrUseDefaults() (string, error) {
	// Try to find encrypted config files
	configDir := config.GetDatabaseConfigDirectory()
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

// DisplayConfigurationSource shows which configuration source is being used
func DisplayConfigurationSource(source ConfigurationSource, details string) {
	switch source {
	case SourceConfigFile:
		fmt.Printf("📁 Using configuration file: %s\n", details)
	case SourceFlags:
		fmt.Printf("🔧 Using command line flags\n")
	case SourceDefaults:
		fmt.Printf("⚙️  Using default configuration from config.yaml\n")
	case SourceInteractive:
		fmt.Printf("👤 Using interactively selected configuration: %s\n", details)
	}
}
