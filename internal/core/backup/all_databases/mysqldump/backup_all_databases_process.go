package backup_all_databases_mysqldump

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/common"
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

	// Write header with backup information
	if options.IncludeDatabaseName {
		header := fmt.Sprintf("-- All Databases Backup\n-- Generated on: %s\n-- Host: %s:%d\n-- User: %s\n-- Total Databases: %d\n\n",
			time.Now().Format("2006-01-02 15:04:05"),
			options.Host,
			options.Port,
			options.User,
			len(databases))
		writer.Write([]byte(header))
	}

	var processedDatabases []string
	var skippedDatabases []string

	// Process each database
	for i, dbName := range databases {
		lg.Info("Processing database",
			logger.String("database", dbName),
			logger.Int("current", i+1),
			logger.Int("total", len(databases)))

		// Write database separator comment
		if options.IncludeDatabaseName {
			separator := fmt.Sprintf("\n-- ================================\n-- Database: %s\n-- ================================\n\n", dbName)
			writer.Write([]byte(separator))
		}

		// Execute mysqldump for this database
		success, err := executeSingleDatabaseDump(options, writer, dbName)
		if err != nil {
			lg.Error("Failed to dump database",
				logger.String("database", dbName),
				logger.Error(err))
			skippedDatabases = append(skippedDatabases, dbName)
			continue
		}

		if success {
			processedDatabases = append(processedDatabases, dbName)
			lg.Info("Successfully processed database", logger.String("database", dbName))
		} else {
			skippedDatabases = append(skippedDatabases, dbName)
			lg.Warn("Database was skipped", logger.String("database", dbName))
		}
	}

	// Close writers in reverse order (inner to outer)
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].Close(); err != nil {
			lg.Warn("Failed to close writer", logger.Error(err))
		}
	}

	lg.Info("All databases backup completed",
		logger.Int("total", len(databases)),
		logger.Int("processed", len(processedDatabases)),
		logger.Int("skipped", len(skippedDatabases)))

	return processedDatabases, skippedDatabases, nil
}

// executeSingleDatabaseDump executes mysqldump for a single database and writes to the provided writer
func executeSingleDatabaseDump(options backup_utils.AllDatabasesBackupOptions, writer io.Writer, dbName string) (bool, error) {
	lg, _ := logger.Get()

	// Build mysqldump command arguments
	args := getOptimizedMysqldumpArgsForAllDatabases(options, dbName)

	lg.Debug("Executing mysqldump for database",
		logger.String("database", dbName),
		logger.String("host", options.Host),
		logger.Int("port", options.Port),
		logger.Bool("compress", options.Compress),
		logger.Bool("is_remote", common.IsRemoteConnection(options.Host)))

	// Execute mysqldump command
	cmd := exec.Command("mysqldump", args...)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr // Capture stderr for error diagnostics

	// Set environment variable for password
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	// Start the command execution
	startTime := time.Now()

	err := cmd.Run()

	if err != nil {
		lg.Error("mysqldump command failed for database",
			logger.Error(err),
			logger.String("database", dbName),
			logger.String("host", options.Host),
			logger.Int("port", options.Port),
			logger.String("user", options.User))
		return false, fmt.Errorf("mysqldump failed for database %s: %w", dbName, err)
	}

	duration := time.Since(startTime)
	lg.Debug("mysqldump completed for database",
		logger.String("database", dbName),
		logger.String("duration", duration.String()))

	return true, nil
}

// getOptimizedMysqldumpArgsForAllDatabases builds optimized mysqldump arguments for a single database in all databases backup
func getOptimizedMysqldumpArgsForAllDatabases(options backup_utils.AllDatabasesBackupOptions, dbName string) []string {
	cfg, err := config.Get()
	lg, _ := logger.Get()
	if err != nil || cfg == nil || cfg.Mysqldump.Args == "" {
		lg.Fatal("Config/Mysqldump args is required but not found", logger.Error(err))
		return nil
	}

	args := []string{
		fmt.Sprintf("--host=%s", options.Host),
		fmt.Sprintf("--port=%d", options.Port),
		fmt.Sprintf("--user=%s", options.User),
	}

	// Parse and add config args
	configArgs := common.ParseArgsString(cfg.Mysqldump.Args)
	args = append(args, configArgs...)

	// Handle data inclusion
	if !options.IncludeData {
		args = append(common.RemoveDataFlags(args), "--no-data")
	}

	// Add database name as the last argument
	args = append(args, dbName)

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
