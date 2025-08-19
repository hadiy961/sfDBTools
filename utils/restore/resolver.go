package restore_utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// ResolveDatabaseConnection resolves database connection from various sources for restore
func ResolveDatabaseConnection(cmd *cobra.Command) (host string, port int, user, password string, source ConfigurationSource, err error) {
	// Check if --config flag is provided
	configFile := common.GetStringFlagOrEnv(cmd, "config", "RESTORE_CONFIG", "")

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
	hasConnectionFlags := cmd.Flags().Changed("target_host") ||
		cmd.Flags().Changed("target_port") ||
		cmd.Flags().Changed("target_user") ||
		cmd.Flags().Changed("target_password")

	if hasConnectionFlags {
		// Use individual flags with defaults
		host := common.GetStringFlagOrEnv(cmd, "target_host", "TARGET_HOST", "localhost")
		port := common.GetIntFlagOrEnv(cmd, "target_port", "TARGET_PORT", 3306)
		user := common.GetStringFlagOrEnv(cmd, "target_user", "TARGET_USER", "root")
		password := common.GetStringFlagOrEnv(cmd, "target_password", "TARGET_PASSWORD", "")

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
	dbName := common.GetStringFlagOrEnv(cmd, "target_db", "TARGET_DB", "")
	if dbName != "" {
		return dbName, nil
	}

	// Check if user wants to create a new database
	createNewDB := common.GetBoolFlagOrEnv(cmd, "create-new-db", "CREATE_NEW_DB", false)
	if createNewDB {
		return resolveDatabaseNameForNewDB(cmd)
	}

	// Show available databases and let user choose
	dbConfig := database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

	selectedDB, err := SelectDatabaseInteractiveWithNewOption(dbConfig)
	if err != nil {
		return "", fmt.Errorf("failed to select database: %w", err)
	}

	return selectedDB, nil
}

// ResolveDatabaseNameWithFile resolves database name with file path available for filename extraction
func ResolveDatabaseNameWithFile(cmd *cobra.Command, host string, port int, user, password, filePath string) (string, error) {
	dbName := common.GetStringFlagOrEnv(cmd, "target_db", "TARGET_DB", "")
	if dbName != "" {
		return dbName, nil
	}

	// Check if user wants to create a new database
	createNewDB := common.GetBoolFlagOrEnv(cmd, "create-new-db", "CREATE_NEW_DB", false)
	if createNewDB {
		return resolveDatabaseNameForNewDBWithFile(cmd, filePath)
	}

	// Show available databases and let user choose
	dbConfig := database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

	selectedDB, err := SelectDatabaseInteractiveWithNewOptionAndFile(dbConfig, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to select database: %w", err)
	}

	return selectedDB, nil
}

// ResolveBackupFile resolves backup file path, with interactive selection if not provided
func ResolveBackupFile(cmd *cobra.Command) (string, error) {
	filePath := common.GetStringFlagOrEnv(cmd, "file", "RESTORE_FILE", "")
	if filePath != "" {
		// Validate the provided file
		if err := ValidateBackupFile(filePath); err != nil {
			return "", fmt.Errorf("invalid backup file: %w", err)
		}
		return filePath, nil
	}

	// Show available backup files and let user choose
	selectedFile, err := SelectBackupFileInteractive("./backup")
	if err != nil {
		return "", fmt.Errorf("failed to select backup file: %w", err)
	}

	// Validate the selected file
	if err := ValidateBackupFile(selectedFile); err != nil {
		return "", fmt.Errorf("invalid selected backup file: %w", err)
	}

	return selectedFile, nil
}

// ResolveGrantsFile resolves grants backup file path, with interactive selection if not provided
func ResolveGrantsFile(cmd *cobra.Command) (string, error) {
	filePath := common.GetStringFlagOrEnv(cmd, "file", "RESTORE_FILE", "")
	if filePath != "" {
		// Validate the provided file
		if err := ValidateBackupFile(filePath); err != nil {
			return "", fmt.Errorf("invalid grants file: %w", err)
		}
		return filePath, nil
	}

	// Show available grants files and let user choose (look in grants directories specifically)
	selectedFile, err := SelectGrantsFileInteractive("./backup")
	if err != nil {
		return "", fmt.Errorf("failed to select grants file: %w", err)
	}

	// Validate the selected file
	if err := ValidateBackupFile(selectedFile); err != nil {
		return "", fmt.Errorf("invalid selected grants file: %w", err)
	}

	return selectedFile, nil
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
		fmt.Println("   or use individual connection flags (--target_host, --target_user, etc.)")
		return "", nil
	}

	// Show selection
	return common.SelectConfigFileInteractive()
}

// resolveDatabaseNameForNewDB handles database name resolution for new database creation
func resolveDatabaseNameForNewDB(cmd *cobra.Command) (string, error) {
	// Check if user wants to use database name from filename
	useFilename := common.GetBoolFlagOrEnv(cmd, "db-from-filename", "DB_FROM_FILENAME", false)

	if useFilename {
		// Get the backup file path first
		filePath := common.GetStringFlagOrEnv(cmd, "file", "RESTORE_FILE", "")
		if filePath == "" {
			return "", fmt.Errorf("backup file must be specified when using --db-from-filename")
		}

		// Extract database name from filename
		dbName := extractDatabaseNameFromFilename(filepath.Base(filePath))
		if dbName == "" {
			return "", fmt.Errorf("failed to extract database name from filename: %s", filePath)
		}

		fmt.Printf("🗂️  Using database name from filename: %s\n", dbName)
		return dbName, nil
	}

	// Prompt user for manual input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter new database name: ")
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

// SelectDatabaseInteractiveWithNewOption displays available databases and option to create new one
func SelectDatabaseInteractiveWithNewOption(config database.Config) (string, error) {
	databases, err := info.ListDatabases(config)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve database list: %w", err)
	}

	// Display available databases with new database option
	fmt.Println("📁 Available Databases:")
	fmt.Println("======================")

	// Option to create new database
	fmt.Printf("   0. 🆕 Create new database\n")

	// List existing databases
	for i, db := range databases {
		fmt.Printf("   %d. %s\n", i+1, db)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect database (0-%d): ", len(databases))
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
		return promptForNewDatabaseName()
	}

	// User chose existing database
	return databases[index-1], nil
}

// promptForNewDatabaseName prompts user for new database name with options
func promptForNewDatabaseName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🆕 Create New Database:")
	fmt.Println("======================")
	fmt.Println("   1. Use database name from backup filename")
	fmt.Println("   2. Enter database name manually")

	fmt.Print("\nSelect option (1-2): ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	switch choice {
	case "1":
		// This will be handled later when we have the file path
		return "USE_FILENAME", nil
	case "2":
		fmt.Print("Enter new database name: ")
		dbName, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read database name: %w", err)
		}

		dbName = strings.TrimSpace(dbName)
		if dbName == "" {
			return "", fmt.Errorf("database name cannot be empty")
		}

		return dbName, nil
	default:
		return "", fmt.Errorf("invalid selection: %s", choice)
	}
}

// resolveDatabaseNameForNewDBWithFile handles database name resolution for new database creation with file path
func resolveDatabaseNameForNewDBWithFile(cmd *cobra.Command, filePath string) (string, error) {
	// Check if user wants to use database name from filename
	useFilename := common.GetBoolFlagOrEnv(cmd, "db-from-filename", "DB_FROM_FILENAME", false)

	if useFilename {
		// Extract database name from filename
		dbName := extractDatabaseNameFromFilename(filepath.Base(filePath))
		if dbName == "" {
			return "", fmt.Errorf("failed to extract database name from filename: %s", filePath)
		}

		fmt.Printf("🗂️  Using database name from filename: %s\n", dbName)
		return dbName, nil
	}

	// Prompt user for manual input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter new database name: ")
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

// SelectDatabaseInteractiveWithNewOptionAndFile displays available databases and option to create new one with file context
func SelectDatabaseInteractiveWithNewOptionAndFile(config database.Config, filePath string) (string, error) {
	databases, err := info.ListDatabases(config)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve database list: %w", err)
	}

	// Extract suggested database name from filename
	suggestedName := extractDatabaseNameFromFilename(filepath.Base(filePath))

	// Display available databases with new database option
	fmt.Println("📁 Available Databases:")
	fmt.Println("======================")

	// Option to create new database
	fmt.Printf("   0. 🆕 Create new database\n")

	// List existing databases
	for i, db := range databases {
		fmt.Printf("   %d. %s\n", i+1, db)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect database (0-%d): ", len(databases))
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
		return promptForNewDatabaseNameWithFile(suggestedName)
	}

	// User chose existing database
	return databases[index-1], nil
}

// promptForNewDatabaseNameWithFile prompts user for new database name with filename suggestion
func promptForNewDatabaseNameWithFile(suggestedName string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🆕 Create New Database:")
	fmt.Println("======================")
	fmt.Printf("   1. Use database name from backup filename (%s)\n", suggestedName)
	fmt.Println("   2. Enter database name manually")

	fmt.Print("\nSelect option (1-2): ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	switch choice {
	case "1":
		if suggestedName == "" {
			return "", fmt.Errorf("cannot extract database name from filename")
		}
		return suggestedName, nil
	case "2":
		fmt.Print("Enter new database name: ")
		dbName, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read database name: %w", err)
		}

		dbName = strings.TrimSpace(dbName)
		if dbName == "" {
			return "", fmt.Errorf("database name cannot be empty")
		}

		return dbName, nil
	default:
		return "", fmt.Errorf("invalid selection: %s", choice)
	}
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
