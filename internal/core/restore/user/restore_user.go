package user

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/compression"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/database"
	restore_utils "sfDBTools/utils/restore"
)

// RestoreUserGrants restores user grants from backup file
func RestoreUserGrants(options restore_utils.RestoreUserOptions) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	startTime := time.Now()

	lg.Info("Starting user grants restore",
		logger.String("host", options.Host),
		logger.Int("port", options.Port),
		logger.String("file", options.File))

	DisplayRestoreUserOverview(options, startTime, lg)

	// Create database config for connection validation
	cfg := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
		DBName:   "", // No specific database needed for grants restore
	}

	// Setup max statement time manager for long operations
	if timeManager, _ := database.SetupMaxStatementTimeManager(cfg, lg); timeManager != nil {
		defer database.CleanupMaxStatementTimeManager(timeManager)
	}

	// Validate connection (without specific database)
	if err := database.ValidateConnection(cfg); err != nil {
		return fmt.Errorf("database connection validation failed: %w", err)
	}

	// Verify checksum if requested
	if options.VerifyChecksum {
		verifyChecksumIfPossible(options.File, lg)
	}

	// Open and process the grants file
	file, err := os.Open(options.File)
	if err != nil {
		return fmt.Errorf("failed to open grants file: %w", err)
	}
	defer file.Close()

	var reader io.ReadCloser = file
	var closers []io.Closer

	pathNoEnc := options.File

	// Handle decryption if file is encrypted
	if strings.HasSuffix(strings.ToLower(pathNoEnc), ".enc") {
		cfgApp, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config for decryption: %w", err)
		}

		lg.Debug("Loaded config for decryption",
			logger.String("app_name", cfgApp.General.AppName),
			logger.String("client_code", cfgApp.General.ClientCode),
			logger.String("version", cfgApp.General.Version),
			logger.String("author", cfgApp.General.Author))

		key, err := crypto.DeriveKeyFromAppConfig(cfgApp.General.AppName, cfgApp.General.ClientCode, cfgApp.General.Version, cfgApp.General.Author)
		if err != nil {
			return fmt.Errorf("failed to derive decryption key: %w", err)
		}

		lg.Debug("Derived decryption key", logger.Int("key_length", len(key)))

		dr, err := crypto.NewGCMDecryptingReader(reader, key)
		if err != nil {
			return fmt.Errorf("failed to create decrypting reader: failed to decrypt data (key derivation or data corruption issue): %w", err)
		}
		reader = io.NopCloser(dr)
		pathNoEnc = strings.TrimSuffix(pathNoEnc, ".enc")
	}

	// Handle decompression if file is compressed
	ctype := compression.DetectCompressionTypeFromFile(pathNoEnc)
	if ctype != compression.CompressionNone {
		dr, err := compression.NewDecompressingReader(reader, ctype)
		if err != nil {
			return fmt.Errorf("failed to create decompressing reader: %w", err)
		}
		reader = dr
		closers = append(closers, dr)
	}

	// Execute grants restore using mysql command
	args := []string{
		fmt.Sprintf("--host=%s", options.Host),
		fmt.Sprintf("--port=%d", options.Port),
		fmt.Sprintf("--user=%s", options.User),
		"--force", // Continue on errors
	}

	cmd := exec.Command("mysql", args...)
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	lg.Info("Starting grants restore execution", logger.String("file", options.File))
	if err := cmd.Run(); err != nil {
		lg.Error("mysql grants restore failed", logger.Error(err))
		return fmt.Errorf("mysql grants restore failed: %w", err)
	}

	// Close all readers properly
	for i := len(closers) - 1; i >= 0; i-- {
		closers[i].Close()
	}

	lg.Info("User grants restore completed", logger.String("file", options.File))
	DisplayRestoreUserSummary(options, startTime, lg)

	return nil
}

// verifyChecksumIfPossible verifies the checksum of the grants file if metadata exists
func verifyChecksumIfPossible(filePath string, lg *logger.Logger) {
	// Try to find metadata file
	metaFile := strings.TrimSuffix(filePath, ".enc")
	metaFile = strings.TrimSuffix(metaFile, ".gz")
	metaFile = strings.TrimSuffix(metaFile, ".zst")
	metaFile = strings.TrimSuffix(metaFile, ".sql")
	metaFile = metaFile + ".json"

	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		lg.Warn("Metadata file not found for checksum verification", logger.String("file", filePath))
		return
	}

	lg.Info("Checksum verification requested but not implemented for grants files yet",
		logger.String("file", filePath))
	// TODO: Implement checksum verification for grants files
}

// DisplayRestoreUserOverview shows user grants restore parameters before execution
func DisplayRestoreUserOverview(options restore_utils.RestoreUserOptions, startTime time.Time, lg *logger.Logger) {
	lg.Info("User grants restore overview",
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("grants_file", options.File),
		logger.String("start_time", startTime.Format("2006-01-02 15:04:05")))

	fmt.Println("\n=== User Grants Restore Overview ===")
	fmt.Printf("Target Host:      %s:%d\n", options.Host, options.Port)
	fmt.Printf("Target User:      %s\n", options.User)
	fmt.Printf("Grants File:      %s\n", options.File)
	fmt.Printf("Start Time:       %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Println("=====================================")
}

// DisplayRestoreUserSummary shows user grants restore summary after completion
func DisplayRestoreUserSummary(options restore_utils.RestoreUserOptions, startTime time.Time, lg *logger.Logger) {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	lg.Info("User grants restore summary",
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("grants_file", options.File),
		logger.String("start_time", startTime.Format("2006-01-02 15:04:05")),
		logger.String("end_time", endTime.Format("2006-01-02 15:04:05")),
		logger.String("duration", duration.String()))

	fmt.Println("\n=== User Grants Restore Summary ===")
	fmt.Printf("Target Host:      %s:%d\n", options.Host, options.Port)
	fmt.Printf("Target User:      %s\n", options.User)
	fmt.Printf("Grants File:      %s\n", options.File)
	fmt.Printf("Start Time:       %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time:         %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration:         %s\n", duration.String())
	fmt.Printf("Verify Checksum:  %t\n", options.VerifyChecksum)
	fmt.Println("====================================")
}
