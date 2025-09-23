package backup_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
	"time"
)

// createMetadataFile creates a JSON metadata file with backup information
func CreateMetadataFile(
	options BackupOptions,
	result *BackupResult,
	config database.Config,
	dbInfos ...*info.DatabaseInfo,
) error {
	lg, _ := logger.Get()

	// Get MySQL version
	mysqlVersion, _ := database.GetMySQLVersion(config)

	// Get replication information
	// replicationInfo, err := GetReplicationInfoForBackup(config)
	// if err != nil {
	// 	lg.Warn("Failed to get replication information for metadata", logger.Error(err))
	// }

	metadata := BackupMetadata{
		DatabaseName:    options.DBName,
		BackupDate:      time.Now(),
		BackupType:      "single",
		OutputFile:      filepath.Base(result.OutputFile),
		FileSize:        result.OutputSize,
		Compressed:      options.Compress,
		CompressionType: options.Compression,
		Encrypted:       options.Encrypt,
		IncludesData:    options.IncludeData,
		Duration:        result.Duration.String(),
		Checksum:        result.Checksum,
		Host:            options.Host,
		Port:            options.Port,
		User:            options.User,
		MySQLVersion:    mysqlVersion,
	}

	// Helper to convert *info.DatabaseInfo to *utils.DatabaseInfoMeta
	toMeta := func(i *info.DatabaseInfo) *DatabaseInfoMeta {
		if i == nil {
			return nil
		}
		return &DatabaseInfoMeta{
			SizeBytes:    i.SizeBytes,
			SizeMB:       i.SizeMB,
			TableCount:   i.TableCount,
			ViewCount:    i.ViewCount,
			RoutineCount: i.RoutineCount,
			TriggerCount: i.TriggerCount,
			UserCount:    i.UserCount,
		}
	}

	if len(dbInfos) > 0 {
		metadata.DatabaseInfo = toMeta(dbInfos[0])
	}

	metaBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(result.BackupMetaFile, metaBytes, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	lg.Info("Metadata file created", logger.String("file", result.BackupMetaFile))
	return nil
}
