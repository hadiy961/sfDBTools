package backup_cmd

import (
	"fmt"
	"os"

	"sfDBTools/internal/config"
	user_grants_backup "sfDBTools/internal/core/backup/user_grants"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"

	"github.com/spf13/cobra"
)

var BackupUserCMD = &cobra.Command{
	Use:   "user",
	Short: "Backup all user grants from MySQL/MariaDB server",
	Long:  `This command backs up all user grants from a MySQL/MariaDB server using the SHOW GRANTS method. The backup will be saved as a separate SQL file with user privileges.`,
	Example: `sfDBTools backup user --source_host localhost --source_user root
sfDBTools backup user --config ./config/mydb.cnf.enc
sfDBTools backup user --source_host localhost --source_user root --output-dir ./backups`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeUserGrantsBackup(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("User grants backup failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// executeUserGrantsBackup handles the user grants backup execution using the new package
func executeUserGrantsBackup(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting user grants backup process")

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

	// 3. Create backup options
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
	}

	// 4. Execute user grants backup using the new package
	result, err := user_grants_backup.BackupUserGrants(options)
	if err != nil {
		return fmt.Errorf("user grants backup failed: %w", err)
	}

	// 5. Display results
	lg.Info("User grants backup completed successfully",
		logger.String("output_file", result.OutputFile),
		logger.Int64("file_size", result.OutputSize),
		logger.String("duration", result.Duration.String()),
		logger.Int("total_users", result.TotalUsers))

	fmt.Printf("User grants backup completed successfully:\n")
	fmt.Printf("  Output file: %s\n", result.OutputFile)
	fmt.Printf("  File size: %d bytes\n", result.OutputSize)
	fmt.Printf("  Duration: %s\n", result.Duration.String())
	fmt.Printf("  Total users: %d\n", result.TotalUsers)
	fmt.Printf("  Compression: %v\n", result.CompressionUsed)
	fmt.Printf("  Encryption: %v\n", result.EncryptionUsed)

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
}
