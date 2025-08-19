package backup_single_mysqldump

import (
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

// performBackup performs the actual database backup using mysqldump
func performBackup(options backup_utils.BackupOptions, outputFile string) error {
	lg, _ := logger.Get()

	if err := backup_utils.ValidateBackupOptions(options); err != nil {
		lg.Error("Invalid backup options", logger.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build mysqldump command with optimizations
	args := getOptimizedMysqldumpArgs(options)

	lg.Debug("Executing mysqldump with compression",
		logger.String("database", options.DBName),
		logger.String("output", outputFile),
		logger.String("compression", options.Compression),
		logger.Bool("compress", options.Compress),
		logger.Bool("is_remote", common.IsRemoteConnection(options.Host)))

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Set up writer chain: compression -> encryption -> file
	var writer io.WriteCloser
	var closers []io.Closer

	writer, closers, err = backup_utils.BuildWriterChain(outFile, options, lg)
	if err != nil {
		lg.Error("Failed to set up writer chain", logger.Error(err))
		return err
	}

	// Execute mysqldump command
	cmd := exec.Command("mysqldump", args...)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr // Capture stderr for error diagnostics

	// Set environment variable for password
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	// Execute the command with retry logic for remote connections
	lg.Info("Starting mysqldump execution")

	// Start the command execution
	startTime := time.Now()

	err = cmd.Run()

	if err != nil {
		lg.Error("mysqldump command failed",
			logger.Error(err),
			logger.String("database", options.DBName),
			logger.String("host", options.Host),
			logger.Int("port", options.Port),
			logger.String("user", options.User))
		return fmt.Errorf("mysqldump failed: %w", err)
	}

	duration := time.Since(startTime)
	lg.Info("mysqldump completed successfully",
		logger.String("duration", duration.String()))

	// Close writers in reverse order (inner to outer)
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].Close(); err != nil {
			lg.Warn("Failed to close writer", logger.Error(err))
			return fmt.Errorf("failed to close writer: %w", err)
		}
	}

	lg.Info("Backup file created successfully",
		logger.String("file", outputFile),
		logger.Bool("compressed", options.Compress),
		logger.Bool("encrypted", options.Encrypt))

	return nil
}

func getOptimizedMysqldumpArgs(options backup_utils.BackupOptions) []string {
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
	args = append(args, common.ParseArgsString(cfg.Mysqldump.Args)...)
	if !options.IncludeData {
		args = append(common.RemoveDataFlags(args), "--no-data")
	}
	args = append(args, options.DBName)
	return args
}
