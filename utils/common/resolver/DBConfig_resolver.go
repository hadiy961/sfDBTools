package resolver

import (
	"fmt"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/cons"
	"sfDBTools/utils/dbconfig"

	"github.com/spf13/cobra"
)

// ResolveDatabaseConnection resolves database connection from various sources
func ResolveDatabaseConnection(cmd *cobra.Command) (host string, port int, user, password string, source cons.ConfigurationSource, err error) {
	// Check if --config flag is provided
	configFile := common.GetStringFlagOrEnv(cmd, "config", "SFDB_CONFIG_FILE", "")

	if configFile != "" {
		// Validate and load from config file
		if err := common.ValidateConfigFile(configFile); err != nil {
			return "", 0, "", "", cons.SourceConfigFile, fmt.Errorf("invalid config file: %w", err)
		}

		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(configFile)
		if err != nil {
			return "", 0, "", "", cons.SourceConfigFile, fmt.Errorf("failed to load config from file: %w", err)
		}

		return host, port, user, password, cons.SourceConfigFile, nil
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

		return host, port, user, password, cons.SourceFlags, nil
	}

	// Try to select config interactively
	selectedFile, err := dbconfig.SelectConfigOrUseDefaults()
	if err != nil {
		return "", 0, "", "", cons.SourceInteractive, err
	}

	if selectedFile != "" {
		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(selectedFile)
		if err != nil {
			return "", 0, "", "", cons.SourceInteractive, fmt.Errorf("failed to load config from file: %w", err)
		}

		return host, port, user, password, cons.SourceInteractive, nil
	}

	// Use defaults
	return "localhost", 3306, "root", "", cons.SourceDefaults, nil
}
