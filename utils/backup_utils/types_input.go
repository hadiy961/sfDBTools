package backup_utils

import (
	"sfDBTools/utils/common/structs"
	"time"
)

// Pattern variables yang bisa dipakai
var PatternVars = map[string]string{
	"{year}":        "2006",            // 2025
	"{month}":       "01",              // 01-12
	"{day}":         "02",              // 01-31
	"{hour}":        "15",              // 00-23
	"{minute}":      "04",              // 00-59
	"{database}":    "",                // nama database
	"{hostname}":    "",                // hostname server
	"{backup_type}": "",                // full, incremental, differential
	"{timestamp}":   "20060102_150405", // full timestamp
}

// OutputOptions - Output and file handling flags
type OutputOptions struct {
	BaseDir              string `default:"/opt/backups"`
	DirectoryPattern     string `default:"{year}/{month}/{day}"`
	FileNamingPattern    string `default:"{backup_type}_{database}_{hostname}_{timestamp}"`
	TimestampFormat      string `default:"20060102_150405"`
	CreateDateDirs       bool   `default:"true"`
	PreserveDirStructure bool   `default:"true"`
	CreateMetadataFile   bool   `default:"true"`
	MetadataPattern      string `default:"metadata_{database}_{timestamp}.json"`
}

// BackupScopeOptions - What to backup
type BackupScopeOptions struct {
	// Input sources (multiple ways to specify databases)
	Databases        []string `flag:"databases" env:"BACKUP_DATABASES"`
	DatabasesFile    string   `flag:"databases-file" env:"DATABASES_FILE"`
	DatabasesPattern string   `flag:"databases-pattern" env:"DATABASES_PATTERN"`

	// Exclusions
	ExcludeDatabases []string `flag:"exclude-databases" env:"EXCLUDE_DATABASES"`
	ExcludeFile      string   `flag:"exclude-file" env:"EXCLUDE_FILE"`
	ExcludePattern   string   `flag:"exclude-pattern" env:"EXCLUDE_PATTERN"`

	// Auto-discovery options
	AutoDiscover  bool `flag:"auto-discover" env:"AUTO_DISCOVER" default:"false"`
	IncludeSystem bool `flag:"include-system" env:"INCLUDE_SYSTEM" default:"false"`
	SkipEmpty     bool `flag:"skip-empty" env:"SKIP_EMPTY" default:"true"`

	// Table-level scope
	IncludeTables []string `flag:"include-tables" env:"INCLUDE_TABLES"`
	ExcludeTables []string `flag:"exclude-tables" env:"EXCLUDE_TABLES"`
	IncludeData   bool     `flag:"include-data" env:"INCLUDE_DATA" default:"true"`
	IncludeUser   bool     `flag:"include-users" env:"INCLUDE_USERS"`

	// Internal field - hasil resolve dari semua sources di atas
	ResolvedDatabases []string `json:"-"` // Hidden dari config file
}

// CompressionOptions - Compression related flags
type CompressionOptions struct {
	Compress         bool   `flag:"compress" env:"BACKUP_COMPRESS"`
	Compression      string `flag:"compression-type" env:"COMPRESSION_TYPE" default:"gzip"`
	CompressionLevel string `flag:"compression-level" env:"COMPRESSION_LEVEL" default:"6"`
}

// SecurityOptions - Security and encryption flags
type SecurityOptions struct {
	Encrypt       bool   `flag:"encrypt" env:"BACKUP_ENCRYPT"`
	EncryptionKey string `flag:"encryption-key" env:"ENCRYPTION_KEY" sensitive:"true"`
	VerifyDisk    bool   `flag:"verify-disk" env:"VERIFY_DISK" default:"true"`
}

type RetentionOptions struct {
	RetentionDays int  `flag:"retention-days" env:"RETENTION_DAYS" default:"7"`
	AutoCleanup   bool `flag:"auto-cleanup" env:"AUTO_CLEANUP" default:"true"`
}

type VerificationOptions struct {
	CalculateChecksum bool `flag:"calculate-checksum" env:"CALCULATE_CHECKSUM" default:"false"`
	ValidateBackup    bool `flag:"validate-backup" env:"VALIDATE_BACKUP" default:"false"`
}

type IncrementalBackupOptions struct {
	IncrementalBackup bool   `flag:"incremental-backup" env:"INCREMENTAL_BACKUP" default:"false"`
	BinlogFile        string `flag:"binlog-file" env:"BINLOG_FILE"`
	BinlogPosition    string `flag:"binlog-position" env:"BINLOG_POSITION"`
}

type DatabaseInfoMeta struct {
	Version      string
	TotalSize    int64
	Tables       int
	Views        int
	StoredProcs  int
	Functions    int
	Triggers     int
	Users        []string
	CharacterSet string
	Collation    string
}

type ReplicationInfo struct {
	HasGTID      bool
	GTIDExecuted string
	GTIDPurged   string
	ServerUUID   string
	HasBinlog    bool
	LogFile      string
	LogPosition  int64
	GTIDPosition string // From BINLOG_GTID_POS function
}

// BackupConfig - Full backup configuration structure
type BackupConfig struct {
	structs.ConnectionOptions
	OutputOptions
	BackupScopeOptions
	CompressionOptions
	SecurityOptions
	RetentionOptions
	VerificationOptions
	IncrementalBackupOptions
	ReplicationInfo
	DatabaseInfoMeta
	Background bool `flag:"background" env:"BACKGROUND" default:"false"`
	LastBackup time.Time
}
