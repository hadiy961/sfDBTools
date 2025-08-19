package backup_utils

import "time"

// BackupOptions represents the configuration for a single database backup
type BackupOptions struct {
	Host              string
	Port              int
	User              string
	Password          string
	DBName            string
	OutputDir         string
	Compress          bool
	Compression       string
	CompressionLevel  string
	IncludeData       bool
	Encrypt           bool
	VerifyDisk        bool
	RetentionDays     int
	CalculateChecksum bool
	IncludeSystem     bool
	SystemUsers       bool
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	Success         bool
	OutputFile      string
	BackupMetaFile  string
	OutputSize      int64
	CompressionUsed string
	Encrypted       bool
	IncludedData    bool
	Duration        time.Duration
	AverageSpeed    float64
	Checksum        string
	Error           error
}

// BackupMetadata represents metadata about the backup
type BackupMetadata struct {
	DatabaseName    string            `json:"database_name"`
	BackupDate      time.Time         `json:"backup_date"`
	BackupType      string            `json:"backup_type"`
	OutputFile      string            `json:"output_file"`
	FileSize        int64             `json:"file_size"`
	Compressed      bool              `json:"compressed"`
	CompressionType string            `json:"compression_type,omitempty"`
	Encrypted       bool              `json:"encrypted"`
	IncludesData    bool              `json:"includes_data"`
	Duration        string            `json:"duration"`
	Checksum        string            `json:"checksum,omitempty"`
	Host            string            `json:"host"`
	Port            int               `json:"port"`
	User            string            `json:"user"`
	MySQLVersion    string            `json:"mariadb_version,omitempty"`
	DatabaseInfo    *DatabaseInfoMeta `json:"database_info,omitempty"`
	ReplicationInfo *ReplicationMeta  `json:"replication_info,omitempty"`
}

// ReplicationMeta represents replication information in metadata
type ReplicationMeta struct {
	HasGTID      bool   `json:"has_gtid"`
	GTIDExecuted string `json:"gtid_executed,omitempty"`
	GTIDPurged   string `json:"gtid_purged,omitempty"`
	ServerUUID   string `json:"server_uuid,omitempty"`
	HasBinlog    bool   `json:"has_binlog"`
	LogFile      string `json:"log_file,omitempty"`
	LogPosition  int64  `json:"log_position,omitempty"`
}

// DatabaseInfoMeta represents database information in metadata
type DatabaseInfoMeta struct {
	SizeBytes    int64   `json:"size_bytes"`
	SizeMB       float64 `json:"size_mb"`
	TableCount   int     `json:"table_count"`
	ViewCount    int     `json:"view_count"`
	RoutineCount int     `json:"routine_count"`
	TriggerCount int     `json:"trigger_count"`
	UserCount    int     `json:"user_count"`
}
