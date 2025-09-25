package backup_utils

import "time"

// Supporting types
type BackupStatus string

const (
	StatusSuccess BackupStatus = "success"
	StatusPartial BackupStatus = "partial"
	StatusFailed  BackupStatus = "failed"
	StatusSkipped BackupStatus = "skipped"
	StatusRunning BackupStatus = "running"
)

type BackupWarning struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Database  string    `json:"database,omitempty"`
	Table     string    `json:"table,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type BackupError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Database  string    `json:"database,omitempty"`
	Table     string    `json:"table,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Fatal     bool      `json:"fatal"`
}

// BackupResult - Untuk MariaDB backup dengan mariadb-dump
type BackupResult struct {
	// Basic Info
	BackupID   string        `json:"backup_id"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Status     BackupStatus  `json:"status"`
	BackupType string        `json:"backup_type"` // "full"

	// MariaDB Command Info
	Command            string `json:"command_executed"`
	MariaDBDumpVersion string `json:"mariadb_dump_version"`

	// Single Output File (karena --all-databases)
	OutputFile       string  `json:"output_file"`
	OutputSize       int64   `json:"output_size_bytes"`
	OutputSizeHuman  string  `json:"output_size_human"`
	CompressedSize   int64   `json:"compressed_size_bytes,omitempty"`
	CompressionRatio float64 `json:"compression_ratio,omitempty"`

	// Database Discovery
	DatabasesFound    []string `json:"databases_found"`
	DatabaseCount     int      `json:"database_count"`
	SystemDBsIncluded bool     `json:"system_dbs_included"`

	// MariaDB-specific info
	GTIDInfo   *GTIDInfo   `json:"gtid_info,omitempty"`
	BinlogInfo *BinlogInfo `json:"binlog_info,omitempty"`

	// Pre-backup Analysis
	TotalSizeEstimate int64            `json:"total_size_estimate_bytes,omitempty"`
	DatabaseSizes     map[string]int64 `json:"database_sizes,omitempty"`

	// Security & Validation
	Checksum     string `json:"checksum,omitempty"`
	ChecksumAlgo string `json:"checksum_algorithm,omitempty"`
	IsEncrypted  bool   `json:"is_encrypted"`

	// Performance Metrics
	ThroughputMBps float64 `json:"throughput_mbps"`

	// Error Handling
	ExitCode int      `json:"exit_code"`
	StdErr   string   `json:"stderr,omitempty"`
	Warnings []string `json:"warnings,omitempty"`

	// Metadata
	Hostname       string `json:"hostname"`
	MariaDBVersion string `json:"mariadb_version"`
	BackupVersion  string `json:"backup_version"`
	ConfigUsed     string `json:"config_used,omitempty"`
}

// MariaDB-specific structures
type GTIDInfo struct {
	GTID          string `json:"gtid"`
	BinlogGTIDPos string `json:"binlog_gtid_pos"`
	Captured      bool   `json:"captured"`
}

type BinlogInfo struct {
	File     string `json:"file"`
	Position string `json:"position"`
	Captured bool   `json:"captured"`
}
