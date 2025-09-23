package single

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/schollz/progressbar/v3"

	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/compression"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/database"
)

// countingReader counts bytes read through it in an atomic counter
type countingReader struct {
	r     io.Reader
	count int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		atomic.AddInt64(&c.count, int64(n))
	}
	return n, err
}

func (c *countingReader) Count() int64 { return atomic.LoadInt64(&c.count) }

// RestoreSingle restores a single database from backup file
func RestoreSingle(options restoreUtils.RestoreOptions) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	startTime := time.Now()

	lg.Info("Starting single database restore",
		logger.String("database", options.DBName),
		logger.String("host", options.Host),
		logger.Int("port", options.Port))
	DisplayRestoreOverview(options, startTime, options.File, lg)
	configDB := database.Config{
		Host:     options.Host,
		Port:     options.Port,
		User:     options.User,
		Password: options.Password,
		DBName:   options.DBName,
	}

	if timeManager, _ := database.SetupMaxStatementTimeManager(configDB, lg); timeManager != nil {
		defer database.CleanupMaxStatementTimeManager(timeManager)
	}

	cfg := database.Config{Host: options.Host, Port: options.Port, User: options.User, Password: options.Password, DBName: options.DBName}
	if err := database.ValidateConnection(cfg); err != nil {
		return err
	}
	if err := database.EnsureDatabase(cfg); err != nil {
		return err
	}

	if options.VerifyChecksum {
		verifyChecksumIfPossible(options.File, lg)
	}

	file, err := os.Open(options.File)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var reader io.ReadCloser = file
	var closers []io.Closer

	pathNoEnc := options.File
	if strings.HasSuffix(strings.ToLower(pathNoEnc), ".enc") {
		// Get encryption password from user (same method as config generate and backup)
		encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password to decrypt backup: ")
		if err != nil {
			return fmt.Errorf("failed to get encryption password: %w", err)
		}

		// Use the same key derivation method as config generate and backup
		key, err := crypto.DeriveKeyWithPassword(encryptionPassword)
		if err != nil {
			return fmt.Errorf("failed to derive decryption key: %w", err)
		}

		lg.Debug("Derived decryption key", logger.Int("key_length", len(key)))

		dr, err := crypto.NewGCMDecryptingReader(reader, key)
		if err != nil {
			return fmt.Errorf("failed to create decrypting reader: failed to decrypt data (incorrect password or data corruption): %w", err)
		}
		reader = io.NopCloser(dr)
		pathNoEnc = strings.TrimSuffix(pathNoEnc, ".enc")
	}

	ctype := compression.DetectCompressionTypeFromFile(pathNoEnc)
	if ctype != compression.CompressionNone {
		dr, err := compression.NewDecompressingReader(reader, ctype)
		if err != nil {
			return fmt.Errorf("failed to create decompressing reader: %w", err)
		}
		reader = dr
		closers = append(closers, dr)
	}

	args := []string{
		fmt.Sprintf("--host=%s", options.Host),
		fmt.Sprintf("--port=%d", options.Port),
		fmt.Sprintf("--user=%s", options.User),
		"--force",
		options.DBName,
	}

	// Wrap the final reader with a counting reader so we can display progress
	counting := &countingReader{r: reader}

	cmd := exec.Command("mysql", args...)
	cmd.Stdin = counting
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	// Determine whether we can compute an accurate total for percentage.
	// Accurate if file is not encrypted and not compressed (we can use raw file size).
	wasEncrypted := strings.HasSuffix(strings.ToLower(options.File), ".enc")
	accuratePercentage := !wasEncrypted && (compression.DetectCompressionTypeFromFile(options.File) == compression.CompressionNone)
	var totalBytes int64 = 0
	if fi, err := os.Stat(options.File); err == nil {
		totalBytes = fi.Size()
	}

	lg.Info("Starting restore", logger.String("db", options.DBName))

	// Setup progressbar (use accurate total only when not compressed and not encrypted)
	var readerForCmd io.Reader = counting
	var bar *progressbar.ProgressBar
	if accuratePercentage && totalBytes > 0 {
		bar = progressbar.NewOptions64(totalBytes,
			progressbar.OptionSetDescription("Restoring"),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetElapsedTime(true),
			progressbar.OptionSetPredictTime(false),
		)
		// wrap counting so bar is updated as bytes flow
		readerForCmd = io.TeeReader(counting, bar)
	} else {
		// Unknown total: show bytes and elapsed only
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Restoring"),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetElapsedTime(true),
			progressbar.OptionSpinnerType(14),
		)
		readerForCmd = io.TeeReader(counting, bar)
	}

	cmd.Stdin = io.NopCloser(readerForCmd)

	if err := cmd.Run(); err != nil {
		// ensure bar finished/cleared
		if bar != nil {
			_ = bar.Finish()
		}
		lg.Error("mysql restore failed", logger.Error(err))
		return err
	}
	if bar != nil {
		_ = bar.Finish()
	}

	for i := len(closers) - 1; i >= 0; i-- {
		// Best-effort close any readers (decompressor/decrypters)
		_ = closers[i].Close()
	}

	lg.Info("Restore completed", logger.String("db", options.DBName))
	// Display summary and collect DB info (single-db restore only)
	dbInfo, _ := DisplayRestoreSummary(options, startTime, lg, &configDB)

	// Process metadata (read metadata file and compare with collected db info)
	ProcessMetadataAfterRestore(options.File, dbInfo, lg)

	return nil
}
