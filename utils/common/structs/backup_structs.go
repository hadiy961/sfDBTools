package structs

import "time"

// DefaultBackupOptions holds the default backup options
var DefaultBackupOptions = BackupOptions{
	Host:              "localhost",
	Port:              3306,
	User:              "root",
	OutputDir:         "./backups",
	Compress:          true,
	Compression:       "gzip",
	CompressionLevel:  "default",
	IncludeData:       true,
	Encrypt:           false,
	VerifyDisk:        true,
	RetentionDays:     7,
	CalculateChecksum: true,
	SystemUser:        true,
	IncludeUser:       true,
	IncludeGrants:     true,
	BackupType:        "full", // full, Single DB, user_grants,
	DatabaseName:      "",
}

// BackupOptions defines the options for a database backup
type BackupOptions struct {
	Host              string
	Port              int
	User              string
	Password          string
	OutputDir         string
	Compress          bool
	Compression       string
	CompressionLevel  string
	IncludeData       bool
	Encrypt           bool
	VerifyDisk        bool
	RetentionDays     int
	CalculateChecksum bool
	SystemUser        bool
	IncludeUser       bool // Specific to user grants backup
	IncludeGrants     bool // Specific to user grants backup
	BackupType        string
	DatabaseName      string
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
