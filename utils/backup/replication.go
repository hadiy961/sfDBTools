package backup_utils

import (
	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// CreateReplicationMetadata creates replication metadata from database replication info
func CreateReplicationMetadata(replicationInfo *database.ReplicationInfo) *ReplicationMeta {
	if replicationInfo == nil {
		return nil
	}

	meta := &ReplicationMeta{}

	// Add GTID information
	if replicationInfo.GTIDInfo != nil {
		meta.HasGTID = replicationInfo.GTIDInfo.HasGTID
		meta.GTIDExecuted = replicationInfo.GTIDInfo.GTIDExecuted
		meta.GTIDPurged = replicationInfo.GTIDInfo.GTIDPurged
		meta.ServerUUID = replicationInfo.GTIDInfo.ServerUUID
		meta.GTIDPosition = replicationInfo.GTIDInfo.GTIDPosition
	}

	// Add Binary Log information
	if replicationInfo.BinaryLogInfo != nil {
		meta.HasBinlog = replicationInfo.BinaryLogInfo.HasBinlog
		meta.LogFile = replicationInfo.BinaryLogInfo.LogFile
		meta.LogPosition = replicationInfo.BinaryLogInfo.LogPosition
	}

	return meta
}

// GetReplicationInfoForBackup retrieves replication information for backup purposes
func GetReplicationInfoForBackup(dbConfig database.Config) (*database.ReplicationInfo, error) {
	lg, _ := logger.Get()

	lg.Info("Collecting replication information for backup",
		logger.String("host", dbConfig.Host),
		logger.Int("port", dbConfig.Port))

	// Get complete replication information
	replicationInfo, err := database.GetReplicationInfo(dbConfig)
	if err != nil {
		lg.Warn("Failed to collect replication information", logger.Error(err))
		return nil, err
	}

	// Log collected information
	if replicationInfo.GTIDInfo != nil && replicationInfo.GTIDInfo.HasGTID {
		lg.Info("GTID information collected for backup",
			logger.String("server_uuid", replicationInfo.GTIDInfo.ServerUUID),
			logger.Bool("has_gtid_executed", len(replicationInfo.GTIDInfo.GTIDExecuted) > 0),
			logger.Bool("has_gtid_purged", len(replicationInfo.GTIDInfo.GTIDPurged) > 0))
	}

	if replicationInfo.BinaryLogInfo != nil && replicationInfo.BinaryLogInfo.HasBinlog {
		lg.Info("Binary log information collected for backup",
			logger.String("log_file", replicationInfo.BinaryLogInfo.LogFile),
			logger.Int64("log_position", replicationInfo.BinaryLogInfo.LogPosition))
	}

	return replicationInfo, nil
}
