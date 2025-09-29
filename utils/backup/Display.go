package backup_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/common/format"
)

// DisplayBackupParameters logs backup parameters before execution (simplified)
func DisplayBackupParameters(options BackupOptions) {
	lg, _ := logger.Get()

	lg.Info("Starting backup",
		logger.String("database", options.DBName),
		logger.String("host", fmt.Sprintf("%s:%d", options.Host, options.Port)),
		logger.String("user", options.User),
		logger.String("output_dir", options.OutputDir),
		logger.Bool("compress", options.Compress),
		logger.String("compression", options.Compression),
		logger.String("compression_level", options.CompressionLevel),
		logger.Bool("encrypt", options.Encrypt),
		logger.Bool("include_data", options.IncludeData))
}

// DisplayBackupResults logs comprehensive backup results with all information
func DisplayBackupResults(result *BackupResult, options BackupOptions, title string) {
	// lg, _ := logger.Get()

	if title == "" {
		title = "Backup"
	}

	// Read database metadata if available
	var dbInfo map[string]interface{}
	if result.BackupMetaFile != "" {
		if data, err := os.ReadFile(result.BackupMetaFile); err == nil {
			var metadata BackupMetadata
			if json.Unmarshal(data, &metadata) == nil && metadata.DatabaseInfo != nil {
				dbInfo = map[string]interface{}{
					"database_size": common.FormatSize(metadata.DatabaseInfo.SizeBytes),
					"table_count":   format.FormatNumber(metadata.DatabaseInfo.TableCount),
					"view_count":    format.FormatNumber(metadata.DatabaseInfo.ViewCount),
					"routine_count": format.FormatNumber(metadata.DatabaseInfo.RoutineCount),
					"trigger_count": format.FormatNumber(metadata.DatabaseInfo.TriggerCount),
					"user_count":    format.FormatNumber(metadata.DatabaseInfo.UserCount),
				}
			}
		}
	}

	// Log comprehensive backup information
	fields := []logger.Field{
		logger.String("operation", title),
		logger.String("database", options.DBName),
		logger.String("host", fmt.Sprintf("%s:%d", options.Host, options.Port)),
		logger.String("user", options.User),
		logger.String("output_file", result.OutputFile),
		logger.String("output_size", common.FormatSize(result.OutputSize)),
		logger.String("metadata_file", result.BackupMetaFile),
		logger.Bool("compressed", options.Compress),
		logger.String("compression_type", result.CompressionUsed),
		logger.Bool("encrypted", result.Encrypted),
		logger.Bool("include_data", result.IncludedData),
		logger.String("duration", format.FormatDuration(result.Duration, "shorts")),
		logger.String("average_speed", common.FormatSize(int64(result.AverageSpeed))+"/s"),
		logger.String("checksum_sha256", result.Checksum),
	}

	// Add database info if available
	if dbInfo != nil {
		fields = append(fields,
			logger.String("db_size", dbInfo["database_size"].(string)),
			logger.String("db_size_mb", dbInfo["database_size_mb"].(string)),
			logger.String("tables", dbInfo["table_count"].(string)),
			logger.String("views", dbInfo["view_count"].(string)),
			logger.String("routines", dbInfo["routine_count"].(string)),
			logger.String("triggers", dbInfo["trigger_count"].(string)),
			logger.String("users", dbInfo["user_count"].(string)),
		)
	}

	// lg.Info("Backup completed successfully", fields...)
}
