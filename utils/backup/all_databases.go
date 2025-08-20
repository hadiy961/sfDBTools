package backup_utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

// AllDatabasesBackupOptions represents options for backing up all databases to a single file
type AllDatabasesBackupOptions struct {
	BackupOptions
	ExcludeSystemDatabases bool
	IncludeUser            bool // Include user grants for replication using SHOW GRANTS method
	CaptureGTID            bool // Capture GTID information including BINLOG_GTID_POS
	IncludeDatabaseName    bool // Include database name as comments in the output
}

// AllDatabasesBackupResult represents the result of all databases backup
type AllDatabasesBackupResult struct {
	BackupResult
	ProcessedDatabases []string
	SkippedDatabases   []string
	TotalDatabases     int
	GTIDPosition       string // GTID position from BINLOG_GTID_POS
}

// ExecuteAllDatabasesBackup executes backup for all databases into a single file
func ExecuteAllDatabasesBackup(
	cmd *cobra.Command,
	backupFunc func(AllDatabasesBackupOptions, []string) (*AllDatabasesBackupResult, error),
) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	lg.Info("Starting all databases backup to single file")

	// 1. Resolve backup configuration
	backupConfig, err := ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve backup configuration: %w", err)
	}

	// 2. Create database config and test connection
	dbConfig := CreateDatabaseConfig(backupConfig)
	if err := TestDatabaseConnection(dbConfig); err != nil {
		return err
	}

	// 3. Get databases ONCE based on system database inclusion preference
	includeSystemDatabases, _ := cmd.Flags().GetBool("include-system-databases")
	includeUser, _ := cmd.Flags().GetBool("include-user")
	captureGTID, _ := cmd.Flags().GetBool("capture-gtid")

	availableDatabases, err := GetAllDatabasesList(dbConfig, !includeSystemDatabases)
	if err != nil {
		return fmt.Errorf("failed to get available databases: %w", err)
	}

	if len(availableDatabases) == 0 {
		return fmt.Errorf("no databases found to backup")
	}

	lg.Info("Found databases for backup",
		logger.Int("count", len(availableDatabases)),
		logger.Strings("databases", availableDatabases),
		logger.Bool("exclude_system", !includeSystemDatabases))

	// 4. Create all databases backup options
	options := AllDatabasesBackupOptions{
		BackupOptions: BackupOptions{
			Host:              backupConfig.Host,
			Port:              backupConfig.Port,
			User:              backupConfig.User,
			Password:          backupConfig.Password,
			OutputDir:         backupConfig.OutputDir,
			Compress:          backupConfig.Compress,
			Compression:       backupConfig.Compression,
			CompressionLevel:  backupConfig.CompressionLevel,
			IncludeData:       backupConfig.IncludeData,
			Encrypt:           backupConfig.Encrypt,
			VerifyDisk:        backupConfig.VerifyDisk,
			RetentionDays:     backupConfig.RetentionDays,
			CalculateChecksum: backupConfig.CalculateChecksum,
		},
		ExcludeSystemDatabases: !includeSystemDatabases,
		IncludeUser:            includeUser,
		CaptureGTID:            captureGTID,
		IncludeDatabaseName:    true,
	}

	// Set a special database name for all databases backup
	options.DBName = "all_databases"

	// 5. Execute backup with pre-loaded database list
	result, err := backupFunc(options, availableDatabases)
	if err != nil {
		return fmt.Errorf("all databases backup failed: %w", err)
	}

	// 6. Display results
	DisplayAllDatabasesBackupResults(result, options)

	return nil
}

// GetAllDatabasesList retrieves all databases excluding or including system databases as specified
func GetAllDatabasesList(dbConfig database.Config, excludeSystem bool) ([]string, error) {
	lg, _ := logger.Get()

	if excludeSystem {
		// Get user databases only (system databases already excluded by info.ListDatabases)
		databases, err := info.ListDatabases(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to list databases: %w", err)
		}

		lg.Info("Retrieved databases list (system databases excluded)",
			logger.Int("count", len(databases)),
			logger.Strings("databases", databases))

		return databases, nil
	}

	// Get all databases including system databases
	allDatabases, err := info.ListAllDatabases(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to list all databases: %w", err)
	}

	lg.Info("Retrieved databases list (including system databases)",
		logger.Int("count", len(allDatabases)),
		logger.Strings("databases", allDatabases))

	return allDatabases, nil
}

// GenerateAllDatabasesOutputPaths generates output paths for all databases backup
func GenerateAllDatabasesOutputPaths(options AllDatabasesBackupOptions) (string, string) {
	timestamp := time.Now().Format("2006_01_02")
	timeDetail := time.Now().Format("20060102_150405")

	// Create output directory structure: outputDir/YYYY_MM_DD/all_databases/
	outputDir := filepath.Join(options.OutputDir, timestamp, "all_databases")

	// Generate filename: all_databases_YYYYMMDD_HHMMSS.sql[.compression][.enc]
	filename := fmt.Sprintf("all_databases_%s.sql", timeDetail)

	// Add compression extension
	if options.Compress && options.Compression != "" {
		switch strings.ToLower(options.Compression) {
		case "gzip", "pgzip":
			filename += ".gz"
		case "zstd":
			filename += ".zst"
		}
	}

	// Add encryption extension
	if options.Encrypt {
		filename += ".enc"
	}

	outputFile := filepath.Join(outputDir, filename)
	metaFile := filepath.Join(outputDir, fmt.Sprintf("all_databases_%s.meta.json", timeDetail))

	return outputFile, metaFile
}

// DisplayAllDatabasesBackupResults displays the backup results for all databases
func DisplayAllDatabasesBackupResults(result *AllDatabasesBackupResult, options AllDatabasesBackupOptions) {
	lg, _ := logger.Get()

	lg.Info("All databases backup summary",
		logger.String("output_file", result.OutputFile),
		logger.String("meta_file", result.BackupMetaFile),
		logger.Int64("file_size", result.OutputSize),
		logger.String("duration", result.Duration.String()),
		logger.String("compression", result.CompressionUsed),
		logger.Bool("encrypted", result.Encrypted),
		logger.Bool("includes_data", result.IncludedData),
		logger.Float64("average_speed_mbps", result.AverageSpeed),
		logger.String("checksum", result.Checksum))

	if len(result.SkippedDatabases) > 0 {
		lg.Warn("Skipped databases",
			logger.Int("count", len(result.SkippedDatabases)),
			logger.Strings("databases", result.SkippedDatabases))
	}
}

// ValidateAllDatabasesBackupOptions validates options for all databases backup
func ValidateAllDatabasesBackupOptions(options AllDatabasesBackupOptions) error {
	// Validate base backup options
	if err := ValidateBackupOptions(options.BackupOptions); err != nil {
		return fmt.Errorf("invalid backup options: %w", err)
	}

	// Additional validations specific to all databases backup
	if options.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	return nil
}

// CreateAllDatabasesMetadata creates metadata for all databases backup
func CreateAllDatabasesMetadata(
	options AllDatabasesBackupOptions,
	result *AllDatabasesBackupResult,
	dbConfig database.Config,
	replicationInfo *database.ReplicationInfo,
) *BackupMetadata {
	// Get MySQL version only
	mysqlVersion, _ := database.GetMySQLVersion(dbConfig)

	metadata := &BackupMetadata{
		DatabaseName:    "all_databases",
		BackupDate:      time.Now(),
		BackupType:      "all_databases",
		OutputFile:      result.OutputFile,
		FileSize:        result.OutputSize,
		Compressed:      options.Compress,
		CompressionType: result.CompressionUsed,
		Encrypted:       result.Encrypted,
		IncludesData:    result.IncludedData,
		Duration:        result.Duration.String(),
		Checksum:        result.Checksum,
		Host:            options.Host,
		Port:            options.Port,
		User:            options.User,
		MySQLVersion:    mysqlVersion,
		ReplicationInfo: CreateReplicationMetadata(replicationInfo),
		DatabaseInfo: &DatabaseInfoMeta{
			SizeBytes:    result.OutputSize,
			TableCount:   result.TotalDatabases, // Use total databases count
			ViewCount:    0,
			RoutineCount: 0,
			TriggerCount: 0,
			UserCount:    0,
		},
	}

	// Add custom metadata for all databases backup
	if metadata.DatabaseInfo != nil {
		// Store processed databases count in TableCount field for reference
		metadata.DatabaseInfo.TableCount = len(result.ProcessedDatabases)
	}

	return metadata
}
