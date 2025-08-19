package command_backup

import (
	"database/sql"
	"fmt"
	"os"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/connection"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

var BackupUserCMD = &cobra.Command{
	Use:   "user",
	Short: "Backup user grants for selected databases or system users",
	Long:  `This command allows you to backup user grants either for specific databases (interactive selection) or for system users. Use --system-user flag to backup system user grants.`,
	Example: `sfDBTools backup user --source_host localhost --source_user root
sfDBTools backup user --config ./config/mydb.cnf.enc
sfDBTools backup user --system-user --source_host localhost --source_user root
sfDBTools backup user --db_list config/db_list/db_list_1.txt --source_host localhost --source_user root`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the value of the db_list flag
		dbListPath, err := cmd.Flags().GetString("db_list")
		if err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to get db_list flag", logger.Error(err))
			os.Exit(1)
		}

		// Check if system user backup is requested
		systemUser, err := cmd.Flags().GetBool("system-user")
		if err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to get system-user flag", logger.Error(err))
			os.Exit(1)
		}

		// System user backup doesn't support db_list
		if systemUser && dbListPath != "" {
			fmt.Printf("Error: --db_list flag cannot be used with --system-user flag\n")
			os.Exit(1)
		}

		// Execute based on flags
		if systemUser {
			if err := executeGrantBackup(cmd); err != nil {
				lg, _ := logger.Get()
				lg.Error("System user grant backup failed", logger.Error(err))
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else if dbListPath != "" {
			if err := executeListGrantBackup(cmd); err != nil {
				lg, _ := logger.Get()
				lg.Error("List grant backup failed", logger.Error(err))
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := executeGrantBackup(cmd); err != nil {
				lg, _ := logger.Get()
				lg.Error("Grant backup failed", logger.Error(err))
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// executeGrantBackup handles the main grant backup execution logic
func executeGrantBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting grant backup process")

	// 1. Resolve backup configuration
	backupConfig, err := backup_utils.ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	// 2. Create database config and test connection
	dbConfig := backup_utils.CreateDatabaseConfig(backupConfig)
	if err := backup_utils.TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 3. Check if system user backup is requested
	systemUser, err := cmd.Flags().GetBool("system-user")
	if err != nil {
		return fmt.Errorf("failed to get system-user flag: %w", err)
	}

	// 4. Create database connection
	db, err := connection.GetWithoutDB(database.Config{
		Host:     backupConfig.Host,
		Port:     backupConfig.Port,
		User:     backupConfig.User,
		Password: backupConfig.Password,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 5. Execute grant backup based on flag
	if systemUser {
		return executeSystemUserGrantBackup(db, backupConfig, lg)
	} else {
		return executeDatabaseGrantBackup(db, backupConfig, dbConfig, lg)
	}
}

// executeSystemUserGrantBackup backs up system user grants
func executeSystemUserGrantBackup(db *sql.DB, backupConfig *backup_utils.BackupConfig, lg *logger.Logger) error {
	lg.Info("Starting system user grants backup")

	// Create grant backup options
	options := backup_utils.BackupOptions{
		Host:              backupConfig.Host,
		Port:              backupConfig.Port,
		User:              backupConfig.User,
		Password:          backupConfig.Password,
		OutputDir:         backupConfig.OutputDir,
		Compress:          backupConfig.Compress,
		Compression:       backupConfig.Compression,
		CompressionLevel:  backupConfig.CompressionLevel,
		Encrypt:           backupConfig.Encrypt,
		VerifyDisk:        backupConfig.VerifyDisk,
		RetentionDays:     backupConfig.RetentionDays,
		CalculateChecksum: backupConfig.CalculateChecksum,
		SystemUsers:       true,
	}

	// Execute system user grants backup
	result, err := backup_utils.BackupSystemUserGrants(db, options)
	if err != nil {
		return fmt.Errorf("system user grants backup failed: %w", err)
	}

	lg.Info("System user grants backup completed successfully",
		logger.String("output_file", result.OutputFile),
		logger.Int64("file_size", result.OutputSize),
		logger.String("duration", result.Duration.String()))

	fmt.Printf("System user grants backup completed successfully:\n")
	fmt.Printf("  Output file: %s\n", result.OutputFile)
	fmt.Printf("  File size: %d bytes\n", result.OutputSize)
	fmt.Printf("  Duration: %s\n", result.Duration.String())

	return nil
}

// executeDatabaseGrantBackup backs up grants for selected databases
func executeDatabaseGrantBackup(db *sql.DB, backupConfig *backup_utils.BackupConfig, dbConfig database.Config, lg *logger.Logger) error {
	lg.Info("Starting database grants backup")

	// 3. Get available databases and let user select multiple
	selectedDatabases, err := info.SelectMultipleDatabasesInteractive(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to select databases: %w", err)
	}

	if len(selectedDatabases) == 0 {
		return fmt.Errorf("no databases selected for backup")
	}

	lg.Info("Databases selected for grant backup",
		logger.Int("count", len(selectedDatabases)),
		logger.Strings("databases", selectedDatabases))

	// Execute grant backup for each selected database
	var successCount int
	var failures []string

	for _, dbName := range selectedDatabases {
		lg.Info("Processing grants for database", logger.String("database", dbName))

		// Create grant backup options for this database
		options := backup_utils.BackupOptions{
			Host:              backupConfig.Host,
			Port:              backupConfig.Port,
			User:              backupConfig.User,
			Password:          backupConfig.Password,
			DBName:            dbName,
			OutputDir:         backupConfig.OutputDir,
			Compress:          backupConfig.Compress,
			Compression:       backupConfig.Compression,
			CompressionLevel:  backupConfig.CompressionLevel,
			Encrypt:           backupConfig.Encrypt,
			VerifyDisk:        backupConfig.VerifyDisk,
			RetentionDays:     backupConfig.RetentionDays,
			CalculateChecksum: backupConfig.CalculateChecksum,
			SystemUsers:       false,
		}

		// Execute database grants backup
		_, err := backup_utils.BackupDatabaseGrants(db, options)
		if err != nil {
			lg.Error("Database grants backup failed",
				logger.String("database", dbName),
				logger.Error(err))
			failures = append(failures, fmt.Sprintf("%s: %v", dbName, err))
			continue
		}

		successCount++

	}

	if len(failures) > 0 {
		lg.Error("Some database grants backup failed",
			logger.Int("total_processed", len(selectedDatabases)),
			logger.Int("successful", successCount),
			logger.Int("failed", len(failures)),
			logger.Strings("failures", failures))

		return fmt.Errorf("some grant backups failed")
	}

	lg.Info("All database grants backup completed successfully",
		logger.Int("total_processed", len(selectedDatabases)),
		logger.Int("successful", successCount))

	return nil
}

// executeListGrantBackup handles grant backup for databases from a list file
func executeListGrantBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting list grant backup process")

	// 1. Resolve backup configuration
	backupConfig, err := backup_utils.ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	// 2. Resolve database list file
	dbListFile, err := backup_utils.ResolveDBListFile(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve db_list file: %w", err)
	}

	// 3. Create database config and test connection
	dbConfig := backup_utils.CreateDatabaseConfig(backupConfig)
	if err := backup_utils.TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 4. Process and validate database list
	dbListResult, err := backup_utils.ProcessDatabaseList(dbListFile, dbConfig)
	if err != nil {
		return err
	}

	if len(dbListResult.ValidDatabases) == 0 {
		return fmt.Errorf("no valid databases found in the list")
	}

	lg.Info("Databases loaded from list for grant backup",
		logger.Int("total_found", dbListResult.TotalFound),
		logger.Int("valid", len(dbListResult.ValidDatabases)),
		logger.Int("invalid", len(dbListResult.InvalidDatabases)),
		logger.Strings("valid_databases", dbListResult.ValidDatabases))

	// 5. Create database connection
	db, err := connection.GetWithoutDB(database.Config{
		Host:     backupConfig.Host,
		Port:     backupConfig.Port,
		User:     backupConfig.User,
		Password: backupConfig.Password,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 6. Execute grant backup for each database in the list
	var successCount int
	var failures []string

	for _, dbName := range dbListResult.ValidDatabases {
		lg.Info("Processing grants for database from list", logger.String("database", dbName))

		// Create grant backup options for this database
		options := backup_utils.BackupOptions{
			Host:              backupConfig.Host,
			Port:              backupConfig.Port,
			User:              backupConfig.User,
			Password:          backupConfig.Password,
			DBName:            dbName,
			OutputDir:         backupConfig.OutputDir,
			Compress:          backupConfig.Compress,
			Compression:       backupConfig.Compression,
			CompressionLevel:  backupConfig.CompressionLevel,
			Encrypt:           backupConfig.Encrypt,
			VerifyDisk:        backupConfig.VerifyDisk,
			RetentionDays:     backupConfig.RetentionDays,
			CalculateChecksum: backupConfig.CalculateChecksum,
			SystemUsers:       false,
		}

		// Execute database grants backup
		result, err := backup_utils.BackupDatabaseGrants(db, options)
		if err != nil {
			lg.Error("Database grants backup failed",
				logger.String("database", dbName),
				logger.Error(err))
			failures = append(failures, fmt.Sprintf("%s: %v", dbName, err))
			continue
		}

		successCount++
		lg.Info("Database grants backup completed",
			logger.String("database", dbName),
			logger.String("output_file", result.OutputFile),
			logger.Int64("file_size", result.OutputSize),
			logger.String("duration", result.Duration.String()))

		fmt.Printf("Database grants backup completed for '%s':\n", dbName)
		fmt.Printf("  Output file: %s\n", result.OutputFile)
		fmt.Printf("  File size: %d bytes\n", result.OutputSize)
		fmt.Printf("  Duration: %s\n", result.Duration.String())
		fmt.Println()
	}

	// Summary
	fmt.Printf("List grant backup summary:\n")
	fmt.Printf("  Total databases in list: %d\n", dbListResult.TotalFound)
	fmt.Printf("  Valid databases: %d\n", len(dbListResult.ValidDatabases))
	fmt.Printf("  Invalid databases: %d\n", len(dbListResult.InvalidDatabases))
	fmt.Printf("  Successful backups: %d\n", successCount)
	fmt.Printf("  Failed backups: %d\n", len(failures))

	if len(dbListResult.InvalidDatabases) > 0 {
		fmt.Printf("  Invalid databases:\n")
		for _, invalid := range dbListResult.InvalidDatabases {
			fmt.Printf("    - %s\n", invalid)
		}
	}

	if len(failures) > 0 {
		lg.Error("Some database grants backup failed",
			logger.Int("total_processed", len(dbListResult.ValidDatabases)),
			logger.Int("successful", successCount),
			logger.Int("failed", len(failures)),
			logger.Strings("failures", failures))

		fmt.Printf("  Failures:\n")
		for _, failure := range failures {
			fmt.Printf("    - %s\n", failure)
		}
		return fmt.Errorf("some grant backups failed")
	}

	lg.Info("All database grants backup from list completed successfully",
		logger.Int("total_processed", len(dbListResult.ValidDatabases)),
		logger.Int("successful", successCount))

	return nil
}

func init() {
	backup_utils.AddCommonBackupFlags(BackupUserCMD)

	// Additional backup options
	_, _, _, _,
		_, _, _, _,
		_, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	BackupUserCMD.Flags().Bool("verify-disk", defaultVerifyDisk, "verify available disk space before backup")
	BackupUserCMD.Flags().Int("retention-days", defaultRetentionDays, "retention period in days")
	BackupUserCMD.Flags().Bool("calculate-checksum", defaultCalculateChecksum, "calculate SHA256 checksum of backup file")

	// Database list flag for grant backup
	BackupUserCMD.Flags().String("db_list", "", "path to text file containing list of database names (optional, will show selection if not provided)")

	// Note: system-user flag is already added by AddCommonBackupFlags
}
