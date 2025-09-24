package restore_utils

import (
	"fmt"

	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// ResolveRestoreConfig resolves restore configuration from various sources with proper priority
func ResolveRestoreConfig(cmd *cobra.Command) (*RestoreConfig, error) {
	restoreConfig := &RestoreConfig{}

	// Resolve database connection
	host, port, user, password, source, err := ResolveDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database connection: %w", err)
	}

	restoreConfig.Host = host
	restoreConfig.Port = port
	restoreConfig.User = user
	restoreConfig.Password = password

	// Display configuration source
	switch source {
	case SourceConfigFile:
		fmt.Printf("📁 Using configuration file\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceFlags:
		fmt.Printf("🔧 Using command line flags\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceInteractive:
		terminal.Headers("Restore Tools - Restore Single Database")
		fmt.Printf("👤 Using interactively selected configuration\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	}

	// Resolve file path first (needed for database name extraction from filename)
	filePath, err := ResolveBackupFile(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve backup file: %w", err)
	}
	restoreConfig.File = filePath

	// Resolve database name (may depend on file path for filename extraction)
	// If this command is the "all" restore mode, skip interactive database selection
	// because "all" operates on all databases and shouldn't prompt for a single DB.
	if cmd != nil && cmd.Name() == "all" {
		// Explicitly leave DBName empty for all-mode restores
		restoreConfig.DBName = ""
	} else {
		dbName, err := ResolveDatabaseNameWithFile(cmd, host, port, user, password, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve database name: %w", err)
		}
		restoreConfig.DBName = dbName
	}

	// Resolve other restore options
	restoreConfig.VerifyChecksum = common.GetBoolFlagOrEnv(cmd, "verify-checksum", "VERIFY_CHECKSUM", false)

	return restoreConfig, nil
}

// ResolveRestoreUserConfig resolves restore user grants configuration from various sources with proper priority
func ResolveRestoreUserConfig(cmd *cobra.Command) (*RestoreUserConfig, error) {
	restoreConfig := &RestoreUserConfig{}

	// Resolve database connection
	host, port, user, password, source, err := ResolveDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database connection: %w", err)
	}

	restoreConfig.Host = host
	restoreConfig.Port = port
	restoreConfig.User = user
	restoreConfig.Password = password

	// Display configuration source
	switch source {
	case SourceConfigFile:
		fmt.Printf("📁 Using configuration file\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceFlags:
		fmt.Printf("🔧 Using command line flags\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceInteractive:
		fmt.Printf("👤 Using interactively selected configuration\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	}

	// Resolve grants file path (may be from backup/grants directory)
	filePath, err := ResolveGrantsFile(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve grants file: %w", err)
	}
	restoreConfig.File = filePath

	// Resolve other restore options
	restoreConfig.VerifyChecksum = common.GetBoolFlagOrEnv(cmd, "verify-checksum", "VERIFY_CHECKSUM", false)

	return restoreConfig, nil
}

// AddCommonRestoreFlags adds common restore flags to the given command
func AddCommonRestoreFlags(cmd *cobra.Command) {
	// Configuration options
	cmd.Flags().String("config", "", "encrypted configuration file (.cnf.enc)")

	// Database connection options
	cmd.Flags().String("target_db", "", "target database name")
	cmd.Flags().String("target_host", "", "target database host")
	cmd.Flags().Int("target_port", 0, "target database port")
	cmd.Flags().String("target_user", "", "target database user")
	cmd.Flags().String("target_password", "", "target database password")

	// Database creation options
	cmd.Flags().Bool("create-new-db", false, "create new database instead of selecting existing one")
	cmd.Flags().Bool("db-from-filename", false, "use database name from backup filename (requires --create-new-db)")

	// Restore options
	cmd.Flags().String("file", "", "backup file to restore")
	cmd.Flags().Bool("verify-checksum", false, "verify checksum after restore")
}

// AddCommonRestoreUserFlags adds common restore user grants flags to the given command
func AddCommonRestoreUserFlags(cmd *cobra.Command) {
	// Configuration options
	cmd.Flags().String("config", "", "encrypted configuration file (.cnf.enc)")

	// Database connection options (no target_db needed for user grants restore)
	cmd.Flags().String("target_host", "", "target database host")
	cmd.Flags().Int("target_port", 0, "target database port")
	cmd.Flags().String("target_user", "", "target database user")
	cmd.Flags().String("target_password", "", "target database password")

	// Restore options
	cmd.Flags().String("file", "", "grants backup file to restore")
	cmd.Flags().Bool("verify-checksum", false, "verify checksum after restore")
}

// ParseRestoreOptionsFromFlags parses restore options from command flags.
// Deprecated: Use ResolveRestoreConfig instead for better configuration handling
func ParseRestoreOptionsFromFlags(cmd *cobra.Command) (RestoreOptions, error) {
	// Use the new configuration resolution method
	restoreConfig, err := ResolveRestoreConfig(cmd)
	if err != nil {
		return RestoreOptions{}, err
	}

	return restoreConfig.ToRestoreOptions(), nil
}
