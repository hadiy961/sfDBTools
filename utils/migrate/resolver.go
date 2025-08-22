package migrate_utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// ResolveSourceDatabaseConnection resolves source database connection from various sources
func ResolveSourceDatabaseConnection(cmd *cobra.Command) (host string, port int, user, password string, source ConfigurationSource, err error) {
	// Check if --source-config flag is provided
	configFile := common.GetStringFlagOrEnv(cmd, "source-config", "SOURCE_CONFIG", "")

	if configFile != "" {
		// Validate and load from config file
		if err := common.ValidateConfigFile(configFile); err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("invalid source config file: %w", err)
		}

		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(configFile)
		if err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("failed to load source config from file: %w", err)
		}

		return host, port, user, password, SourceConfigFile, nil
	}

	// Check if individual connection flags are provided
	hasConnectionFlags := cmd.Flags().Changed("source-host") ||
		cmd.Flags().Changed("source-port") ||
		cmd.Flags().Changed("source-user") ||
		cmd.Flags().Changed("source-password")

	if hasConnectionFlags {
		// Use individual flags with defaults
		host := common.GetStringFlagOrEnv(cmd, "source-host", "SOURCE_HOST", "localhost")
		port := common.GetIntFlagOrEnv(cmd, "source-port", "SOURCE_PORT", 3306)
		user := common.GetStringFlagOrEnv(cmd, "source-user", "SOURCE_USER", "root")
		password := common.GetStringFlagOrEnv(cmd, "source-password", "SOURCE_PASSWORD", "")

		return host, port, user, password, SourceFlags, nil
	}

	// Try to select config interactively
	fmt.Println("\n🔧 Select Source Database Configuration:")
	fmt.Println("=======================================")
	selectedFile, err := selectConfigOrUseDefaults("source")
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

// ResolveTargetDatabaseConnection resolves target database connection from various sources
func ResolveTargetDatabaseConnection(cmd *cobra.Command) (host string, port int, user, password string, source ConfigurationSource, err error) {
	// Check if --target-config flag is provided
	configFile := common.GetStringFlagOrEnv(cmd, "target-config", "TARGET_CONFIG", "")

	if configFile != "" {
		// Validate and load from config file
		if err := common.ValidateConfigFile(configFile); err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("invalid target config file: %w", err)
		}

		host, port, user, password, err := common.GetDatabaseConfigFromEncrypted(configFile)
		if err != nil {
			return "", 0, "", "", SourceConfigFile, fmt.Errorf("failed to load target config from file: %w", err)
		}

		return host, port, user, password, SourceConfigFile, nil
	}

	// Check if individual connection flags are provided
	hasConnectionFlags := cmd.Flags().Changed("target-host") ||
		cmd.Flags().Changed("target-port") ||
		cmd.Flags().Changed("target-user") ||
		cmd.Flags().Changed("target-password")

	if hasConnectionFlags {
		// Use individual flags with defaults
		host := common.GetStringFlagOrEnv(cmd, "target-host", "TARGET_HOST", "localhost")
		port := common.GetIntFlagOrEnv(cmd, "target-port", "TARGET_PORT", 3306)
		user := common.GetStringFlagOrEnv(cmd, "target-user", "TARGET_USER", "root")
		password := common.GetStringFlagOrEnv(cmd, "target-password", "TARGET_PASSWORD", "")

		return host, port, user, password, SourceFlags, nil
	}

	// Try to select config interactively
	fmt.Println("\n🎯 Select Target Database Configuration:")
	fmt.Println("=======================================")
	selectedFile, err := selectConfigOrUseDefaults("target")
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

// ResolveSourceDatabaseName resolves source database name, with interactive selection if not provided
func ResolveSourceDatabaseName(cmd *cobra.Command, host string, port int, user, password string) (string, error) {
	dbName := common.GetStringFlagOrEnv(cmd, "source-db", "SOURCE_DB", "")
	if dbName != "" {
		return dbName, nil
	}

	// Show available databases and let user choose
	fmt.Println("\n📁 Select Source Database:")
	fmt.Println("=========================")
	dbConfig := database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

	selectedDB, err := info.SelectDatabaseInteractive(dbConfig)
	if err != nil {
		return "", fmt.Errorf("failed to select source database: %w", err)
	}

	return selectedDB, nil
}

// ResolveTargetDatabaseName resolves target database name, automatically using source DB name if not specified
func ResolveTargetDatabaseName(cmd *cobra.Command, host string, port int, user, password, sourceDBName string) (string, error) {
	dbName := common.GetStringFlagOrEnv(cmd, "target-db", "TARGET_DB", "")
	if dbName != "" {
		return dbName, nil
	}

	// Automatically use the source database name for target
	fmt.Printf("\n🎯 Target Database:\n")
	fmt.Printf("===================\n")
	fmt.Printf("ℹ️  Using source database name for target: %s\n", sourceDBName)
	fmt.Printf("   (Use --target-db flag to specify a different name)\n")

	return sourceDBName, nil
}

// selectConfigOrUseDefaults tries to select a config file interactively or returns empty for defaults
func selectConfigOrUseDefaults(configType string) (string, error) {
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
		fmt.Printf("ℹ️  No encrypted configuration files found for %s.\n", configType)
		fmt.Printf("   You can use --%s-config flag to specify a config file,\n", configType)
		fmt.Printf("   or use individual connection flags (--%s-host, --%s-user, etc.)\n", configType, configType)
		return "", nil
	}

	// Show selection
	return common.SelectConfigFileInteractive()
}

// SelectTargetDatabaseInteractive displays available databases and option to create new one for target
func SelectTargetDatabaseInteractive(config database.Config) (string, error) {
	databases, err := info.ListDatabases(config)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve database list: %w", err)
	}

	// Display available databases with new database option
	fmt.Println("📁 Available Target Databases:")
	fmt.Println("=============================")

	// Option to create new database
	fmt.Printf("   0. 🆕 Create new database\n")

	// List existing databases
	for i, db := range databases {
		fmt.Printf("   %d. %s\n", i+1, db)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect target database (0-%d): ", len(databases))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 0 || index > len(databases) {
		return "", fmt.Errorf("invalid selection: %s", choice)
	}

	if index == 0 {
		// User chose to create new database
		return promptForNewTargetDatabaseName()
	}

	// User chose existing database
	return databases[index-1], nil
}

// promptForNewTargetDatabaseName prompts user for new target database name
func promptForNewTargetDatabaseName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new target database name: ")
	dbName, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read database name: %w", err)
	}

	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return "", fmt.Errorf("database name cannot be empty")
	}

	return dbName, nil
}
