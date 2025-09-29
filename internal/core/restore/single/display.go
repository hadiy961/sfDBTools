package single

import (
	"encoding/json"
	"os"
	"time"

	restoreUtils "sfDBTools/internal/core/restore/utils"
	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/format"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
)

// DisplayRestoreOverview shows restore parameters before execution
func DisplayRestoreOverview(options restoreUtils.RestoreOptions, startTime time.Time, filePath string, lg *logger.Logger) {
	lg.Info("Restore overview",
		logger.String("target_database", options.DBName),
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("backup_file", options.File),
		logger.String("start_time", format.FormatTime(startTime, format.UnixTimestamp)))

	meta := metadataPath(filePath)
	if meta == "" {
		lg.Warn("Metadata file not found, skipping display backup metadata", logger.String("file", filePath))
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

	lg.Info("Backup metadata",
		logger.String("metadata_file", meta),
		logger.String("backup_date", format.FormatTime(metaInfo.BackupDate, format.UnixTimestamp)),
		logger.Bool("compression", metaInfo.Compressed),
		logger.Bool("encryption", metaInfo.Encrypted),
		logger.Bool("included_data", metaInfo.IncludesData),
		logger.String("source_host", metaInfo.Host),
		logger.Int("source_port", metaInfo.Port),
		logger.String("source_user", metaInfo.User),
		logger.String("db_name", metaInfo.DatabaseName),
		logger.String("size", common.FormatSize(metaInfo.FileSize)),
	)

	lg.Info("DB source metadata",
		logger.String("db_name", metaInfo.DatabaseName),
		logger.String("db_size", common.FormatSize(metaInfo.DatabaseInfo.SizeBytes)),
		logger.Int("table_count", metaInfo.DatabaseInfo.TableCount),
		logger.Int("view_count", metaInfo.DatabaseInfo.ViewCount),
		logger.Int("routine_count", metaInfo.DatabaseInfo.RoutineCount),
		logger.Int("trigger_count", metaInfo.DatabaseInfo.TriggerCount),
		logger.Int("user_count", metaInfo.DatabaseInfo.UserCount))
}

// DisplayRestoreSummary shows restore summary after completion and returns collected db info
func DisplayRestoreSummary(options restoreUtils.RestoreOptions, startTime time.Time, lg *logger.Logger, config *database.Config) (*info.DatabaseInfo, error) {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	lg.Info("Restore summary",
		logger.String("target_database", options.DBName),
		logger.String("target_host", options.Host),
		logger.Int("target_port", options.Port),
		logger.String("target_user", options.User),
		logger.String("backup_file", options.File),
		logger.String("start_time", format.FormatTime(startTime, format.UnixTimestamp)),
		logger.String("end_time", format.FormatTime(endTime, format.UnixTimestamp)),
		logger.String("duration", duration.String()))

	lg.Info("Collecting database information", logger.String("database", config.DBName))

	// Spinner := terminal.NewLoadingSpinner("Collecting database information...")
	// Spinner.Start()
	dbInfo, err := info.GetDatabaseInfo(*config)
	if err != nil {
		lg.Warn("Failed to collect database information", logger.Error(err))
		return nil, err
	}
	// Spinner.Stop()

	lg.Info("Restore database info",
		logger.String("db_size", common.FormatSize(dbInfo.SizeBytes)),
		logger.Int("table_count", dbInfo.TableCount),
		logger.Int("view_count", dbInfo.ViewCount),
		logger.Int("routine_count", dbInfo.RoutineCount),
		logger.Int("trigger_count", dbInfo.TriggerCount),
		logger.Int("user_count", dbInfo.UserCount))

	return dbInfo, nil
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
