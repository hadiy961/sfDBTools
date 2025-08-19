package single

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/config"
	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/common"
	"sfDBTools/utils/compression"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
)

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
		Host: options.Host, Port: options.Port, User: options.User,
		Password: options.Password, DBName: options.DBName,
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
		cfgApp, err := config.LoadConfig()
		if err != nil {
			return err
		}

		lg.Debug("Loaded config for decryption",
			logger.String("app_name", cfgApp.General.AppName),
			logger.String("client_code", cfgApp.General.ClientCode),
			logger.String("version", cfgApp.General.Version),
			logger.String("author", cfgApp.General.Author))

		key, err := crypto.DeriveKeyFromAppConfig(cfgApp.General.AppName, cfgApp.General.ClientCode, cfgApp.General.Version, cfgApp.General.Author)
		if err != nil {
			return err
		}

		lg.Debug("Derived decryption key", logger.Int("key_length", len(key)))

		dr, err := crypto.NewGCMDecryptingReader(reader, key)
		if err != nil {
			return fmt.Errorf("failed to create decrypting reader: failed to decrypt data (key derivation or data corruption issue): %w", err)
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

	cmd := exec.Command("mysql", args...)
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if options.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", options.Password))
	}

	lg.Info("Starting restore", logger.String("db", options.DBName), logger.String("file", options.File))
	if err := cmd.Run(); err != nil {
		lg.Error("mysql restore failed", logger.Error(err))
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	for i := len(closers) - 1; i >= 0; i-- {
		closers[i].Close()
	}

	lg.Info("Restore completed", logger.String("db", options.DBName))
	DisplayRestoreSummary(options, startTime, lg, &configDB)

	meta := metadataPath(options.File)
	if meta != "" {
		data, err := os.ReadFile(meta)
		if err == nil {
			var metaInfo backup_utils.BackupMetadata
			if json.Unmarshal(data, &metaInfo) == nil {
				dbInfo, err := info.GetDatabaseInfo(configDB)
				if err == nil {
					DisplayDatabaseComparison(metaInfo, *dbInfo)
				}
			}
		}
	}

	return nil
}

func verifyChecksumIfPossible(filePath string, lg *logger.Logger) {
	meta := metadataPath(filePath)
	if meta == "" {
		lg.Warn("Metadata file not found, skipping checksum verification", logger.String("file", filePath))
		return
	}

	data, err := os.ReadFile(meta)
	if err != nil {
		lg.Warn("Failed to read metadata file", logger.String("metadata", meta), logger.Error(err))
		return
	}

	var metaInfo backup_utils.BackupMetadata
	if err := json.Unmarshal(data, &metaInfo); err != nil {
		lg.Warn("Invalid metadata format", logger.String("metadata", meta), logger.Error(err))
		return
	}

	if metaInfo.Checksum == "" {
		lg.Warn("Checksum not found in metadata, skipping verification", logger.String("metadata", meta))
		return
	}

	sum, err := calculateChecksum(filePath)
	if err != nil {
		lg.Warn("Checksum calculation failed", logger.String("file", filePath), logger.Error(err))
		return
	}

	if strings.EqualFold(sum, metaInfo.Checksum) {
		lg.Info("Checksum verified successfully", logger.String("file", filePath))
	} else {
		lg.Error("Checksum mismatch", logger.String("file", filePath), logger.String("expected", metaInfo.Checksum), logger.String("got", sum))
	}
}

func metadataPath(filePath string) string {
	base := strings.TrimSuffix(filePath, ".enc")
	ext := filepath.Ext(base)
	for _, e := range []string{".gz", ".zst", ".zlib", ".sql"} {
		if ext == e {
			base = strings.TrimSuffix(base, ext)
			ext = filepath.Ext(base)
		}
	}
	base = strings.TrimSuffix(base, ".sql")
	if base == filePath {
		return ""
	}
	return base + ".json"
}

func calculateChecksum(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DisplayRestoreOverview shows restore parameters before execution
func DisplayRestoreOverview(options restoreUtils.RestoreOptions, startTime time.Time, filePath string, lg *logger.Logger) {

	lg.Info("Restore overview",
		logger.String("target_database", options.DBName),
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("backup_file", options.File),
		logger.String("start_time", common.FormatTime(startTime, "2006-01-02 15:04:05")))

	meta := metadataPath(filePath)
	if meta == "" {
		lg.Warn("Metadata file not found, skipping display backup metadata", logger.String("file", filePath))
		return
	} else {
		data, err := os.ReadFile(meta)
		if err != nil {
			lg.Warn("Failed to read metadata file", logger.String("metadata", meta), logger.Error(err))
			return
		}

		var metaInfo backup_utils.BackupMetadata
		if err := json.Unmarshal(data, &metaInfo); err != nil {
			lg.Warn("Invalid metadata format", logger.String("metadata", meta), logger.Error(err))
			return
		}

		lg.Info("Backup metadata",
			logger.String("metadata_file", meta),
			logger.String("backup_date", common.FormatTime(metaInfo.BackupDate, "2006-01-02 15:04:05")),
			logger.Bool("compression", metaInfo.Compressed),
			logger.Bool("encryption", metaInfo.Encrypted),
			logger.Bool("included_data", metaInfo.IncludesData),
			logger.String("source_host", metaInfo.Host),
			logger.Int("source_port", metaInfo.Port),
			logger.String("source_user", metaInfo.User))

		lg.Info("DB source metadata",
			logger.String("db_name", metaInfo.DatabaseName),
			logger.String("db_size", common.FormatSize(metaInfo.DatabaseInfo.SizeBytes)),
			logger.Int("table_count", metaInfo.DatabaseInfo.TableCount),
			logger.Int("view_count", metaInfo.DatabaseInfo.ViewCount),
			logger.Int("routine_count", metaInfo.DatabaseInfo.RoutineCount),
			logger.Int("trigger_count", metaInfo.DatabaseInfo.TriggerCount),
			logger.Int("user_count", metaInfo.DatabaseInfo.UserCount))

	}
}

func DisplayRestoreSummary(options restoreUtils.RestoreOptions, startTime time.Time, lg *logger.Logger, config *database.Config) {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	lg.Info("Restore summary",
		logger.String("target_database", options.DBName),
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("backup_file", options.File),
		logger.String("start_time", common.FormatTime(startTime, "2006-01-02 15:04:05")),
		logger.String("end_time", common.FormatTime(endTime, "2006-01-02 15:04:05")),
		logger.String("duration", duration.String()))

	lg.Info("Collecting database information", logger.String("database", config.DBName))

	dbInfo, err := info.GetDatabaseInfo(*config)
	if err != nil {
		lg.Warn("Failed to collect database information", logger.Error(err))
		return
	}

	lg.Info("Restore database info",
		logger.String("db_size", common.FormatSize(dbInfo.SizeBytes)),
		logger.Int("table_count", dbInfo.TableCount),
		logger.Int("view_count", dbInfo.ViewCount),
		logger.Int("routine_count", dbInfo.RoutineCount),
		logger.Int("trigger_count", dbInfo.TriggerCount),
		logger.Int("user_count", dbInfo.UserCount))
}

func DisplayDatabaseComparison(metaInfo backup_utils.BackupMetadata, dbInfo info.DatabaseInfo) {
	lg, _ := logger.Get()

	lg.Info("Database comparison - Database Size",
		logger.String("source_size", common.FormatSize(metaInfo.DatabaseInfo.SizeBytes)),
		logger.String("restored_size", common.FormatSize(dbInfo.SizeBytes)),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.SizeBytes, dbInfo.SizeBytes)))

	lg.Info("Database comparison - Table Count",
		logger.Int("source_count", metaInfo.DatabaseInfo.TableCount),
		logger.Int("restored_count", dbInfo.TableCount),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.TableCount, dbInfo.TableCount)))

	lg.Info("Database comparison - View Count",
		logger.Int("source_count", metaInfo.DatabaseInfo.ViewCount),
		logger.Int("restored_count", dbInfo.ViewCount),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.ViewCount, dbInfo.ViewCount)))

	lg.Info("Database comparison - Routine Count",
		logger.Int("source_count", metaInfo.DatabaseInfo.RoutineCount),
		logger.Int("restored_count", dbInfo.RoutineCount),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.RoutineCount, dbInfo.RoutineCount)))

	lg.Info("Database comparison - Trigger Count",
		logger.Int("source_count", metaInfo.DatabaseInfo.TriggerCount),
		logger.Int("restored_count", dbInfo.TriggerCount),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.TriggerCount, dbInfo.TriggerCount)))

	lg.Info("Database comparison - User Count",
		logger.Int("source_count", metaInfo.DatabaseInfo.UserCount),
		logger.Int("restored_count", dbInfo.UserCount),
		logger.String("status", compareValues(metaInfo.DatabaseInfo.UserCount, dbInfo.UserCount)))
}

// Helper function to compare values and return "MATCHED" or "MISMATCHED"
func compareValues(sourceValue, restoredValue interface{}) string {
	if sourceValue == restoredValue {
		return "MATCHED"
	}
	return "MISMATCHED"
}
