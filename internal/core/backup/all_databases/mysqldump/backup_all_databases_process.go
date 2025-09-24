package mysqldump

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"sfDBTools/internal/config"
	user_grants_backup "sfDBTools/internal/core/backup/user_grants"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
)

// executeAllDatabasesMysqldump executes mysqldump for all databases and writes to a single file
func executeAllDatabasesMysqldump(options backup_utils.AllDatabasesBackupOptions, outputFile string, databases []string) ([]string, []string, error) {
	lg, _ := logger.Get()

	// Validate backup options
	if err := backup_utils.ValidateBackupOptions(options.BackupOptions); err != nil {
		lg.Error("Invalid backup options", logger.Error(err))
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Always use single mysqldump command for replication consistency
	return executeAllDatabasesWithSingleCommand(options, outputFile, databases)
}

// executeAllDatabasesWithSingleCommand executes a single mysqldump command for all databases (for replication consistency)
func executeAllDatabasesWithSingleCommand(options backup_utils.AllDatabasesBackupOptions, outputFile string, databases []string) ([]string, []string, error) {
	lg, _ := logger.Get()

	lg.Info("Using single mysqldump command for replication consistency",
		logger.Int("database_count", len(databases)),
		logger.Bool("capture_gtid", options.CaptureGTID))

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Set up writer chain: compression -> encryption -> file
	var writer io.WriteCloser
	var closers []io.Closer

	writer, closers, err = backup_utils.BuildWriterChain(outFile, options.BackupOptions, lg)
	if err != nil {
		lg.Error("Failed to set up writer chain", logger.Error(err))
		return nil, nil, err
	}
	defer func() {
		// Close writers in reverse order (inner to outer)
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				lg.Warn("Failed to close writer", logger.Error(err))
			}
		}
	}()

	// Build mysqldump command for all databases
	args := getReplicationMysqldumpArgs(options, databases)

	lg.Info("Executing single mysqldump command for replication",
		logger.Strings("databases", databases),
		logger.String("host", options.Host),
		logger.Int("port", options.Port))

	// Execute mysqldump command
	cmd := exec.Command("mysqldump", args...)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	// Set environment variable for password
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	startTime := time.Now()
	err = cmd.Run()

	if err != nil {
		lg.Error("Single mysqldump command failed", logger.Error(err))
		return nil, databases, fmt.Errorf("mysqldump failed: %w", err)
	}

	duration := time.Since(startTime)
	lg.Info("Single mysqldump command completed successfully",
		logger.String("duration", duration.String()),
		logger.Int("databases_count", len(databases)))

	// Handle user grants backup if requested - save to separate file
	if options.IncludeUser {
		if err := createSeparateUserGrantsBackup(options); err != nil {
			lg.Error("Failed to create separate user grants backup", logger.Error(err))
			// Don't fail the entire backup, just log the error
		}
	}

	return databases, []string{}, nil
}

// createSeparateUserGrantsBackup creates user grants backup in separate file
func createSeparateUserGrantsBackup(options backup_utils.AllDatabasesBackupOptions) error {
	lg, _ := logger.Get()

	// Convert AllDatabasesBackupOptions to BackupOptions
	backupOptions := backup_utils.BackupOptions{
		Host:              options.Host,
		Port:              options.Port,
		User:              options.User,
		Password:          options.Password,
		OutputDir:         options.OutputDir,
		Compress:          options.Compress,
		Compression:       options.Compression,
		CompressionLevel:  options.CompressionLevel,
		Encrypt:           options.Encrypt,
		VerifyDisk:        options.VerifyDisk,
		RetentionDays:     options.RetentionDays,
		CalculateChecksum: options.CalculateChecksum,
	}

	// Call the BackupUserGrants function from the separate package
	result, err := user_grants_backup.BackupUserGrants(backupOptions)
	if err != nil {
		return fmt.Errorf("failed to create separate user grants backup: %w", err)
	}

	lg.Info("Separate user grants backup created successfully",
		logger.String("output_file", result.OutputFile),
		logger.Int64("file_size", result.OutputSize),
		logger.Int("total_users", result.TotalUsers))

	return nil
}

// getReplicationMysqldumpArgs builds mysqldump arguments optimized for replication
func getReplicationMysqldumpArgs(options backup_utils.AllDatabasesBackupOptions, databases []string) []string {
	cfg, err := config.Get()
	lg, _ := logger.Get()
	if err != nil || cfg == nil {
		lg.Fatal("Config is required but not found", logger.Error(err))
		return nil
	}

	args := []string{
		fmt.Sprintf("--host=%s", options.Host),
		fmt.Sprintf("--port=%d", options.Port),
		fmt.Sprintf("--user=%s", options.User),
	}

	// Essential replication flags
	args = append(args, "--single-transaction") // Ensures consistency

	// Check if binary logging is enabled before adding --master-data flags
	dbConfig := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
	}

	binlogInfo, err := database.GetBinaryLogInfo(dbConfig)
	if err != nil {
		lg.Warn("Failed to check binary log status, skipping --master-data flag", logger.Error(err))
	} else if binlogInfo != nil && binlogInfo.HasBinlog {
		// Only add --master-data if binary logging is enabled
		if options.CaptureGTID {
			// For GTID-based replication, use --master-data=2 (commented CHANGE MASTER TO)
			args = append(args, "--master-data=2")
			lg.Info("Added --master-data=2 for GTID replication")
		} else {
			// For traditional binlog replication, use --master-data=1 (executable CHANGE MASTER TO)
			args = append(args, "--master-data=1")
			lg.Info("Added --master-data=1 for traditional replication")
		}
	} else {
		lg.Info("Binary logging is not enabled, skipping --master-data flag for mysqldump")
	}

	// Parse and add config args, excluding conflicting ones
	if cfg.Mysqldump.Args != "" {
		configArgs := common.ParseArgsString(cfg.Mysqldump.Args)
		for _, arg := range configArgs {
			// Skip flags that we're handling explicitly or that conflict
			if arg != "--master-data" && arg != "--single-transaction" &&
				arg != "--databases" && arg != "-B" && arg != "--all-databases" && arg != "-A" {
				args = append(args, arg)
			}
		}
	}

	// Handle data inclusion
	if !options.IncludeData {
		args = append(common.RemoveDataFlags(args), "--no-data")
	}

	// Add database specification
	if options.ExcludeSystemDatabases {
		// Use --databases flag for specific databases
		args = append(args, "--databases")
		args = append(args, databases...)
	} else {
		// Use --all-databases for complete server backup
		args = append(args, "--all-databases")
		lg.Info("Using --all-databases for complete server backup including system databases")
	}

	return args
}

// saveAllDatabasesMetadata saves metadata for all databases backup
func saveAllDatabasesMetadata(metaFile string, metadata *backup_utils.BackupMetadata) error {
	// Create metadata directory
	if err := os.MkdirAll(filepath.Dir(metaFile), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Convert metadata to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to file
	if err := os.WriteFile(metaFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}
