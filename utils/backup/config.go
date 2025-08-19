package backup_utils

import (
	"fmt"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// BackupConfig represents the resolved backup configuration
type BackupConfig struct {
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
}

// ResolveBackupConfig resolves backup configuration from various sources with proper priority
func ResolveBackupConfig(cmd *cobra.Command) (*BackupConfig, error) {
	// Get default values from config
	_, _, _, defaultOutputDir,
		defaultCompress, defaultCompression, defaultCompressionLevel, defaultIncludeData,
		defaultEncrypt, defaultVerifyDisk, defaultRetentionDays, defaultCalculateChecksum, _ := config.GetBackupDefaults()

	backupConfig := &BackupConfig{}

	// Resolve database connection
	host, port, user, password, source, err := ResolveDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database connection: %w", err)
	}

	backupConfig.Host = host
	backupConfig.Port = port
	backupConfig.User = user
	backupConfig.Password = password

	// Display configuration source
	switch source {
	case SourceConfigFile:
		fmt.Printf("📁 Using configuration file\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceFlags:
		fmt.Printf("🔧 Using command line flags\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	case SourceInteractive:
		fmt.Printf("👤 Using interactively selected configuration\n")
		fmt.Printf("   Host: %s:%d\n", host, port)
		fmt.Printf("   User: %s\n", user)
	}

	// Resolve database name
	dbName, err := ResolveDatabaseName(cmd, host, port, user, password)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database name: %w", err)
	}
	backupConfig.DBName = dbName

	// Resolve other backup options
	backupConfig.OutputDir = common.GetStringFlagOrEnv(cmd, "output-dir", "OUTPUT_DIR", defaultOutputDir)
	backupConfig.Compress = common.GetBoolFlagOrEnv(cmd, "compress", "COMPRESS", defaultCompress)
	backupConfig.IncludeData = common.GetBoolFlagOrEnv(cmd, "data", "INCLUDE_DATA", defaultIncludeData)
	backupConfig.Encrypt = common.GetBoolFlagOrEnv(cmd, "encrypt", "ENCRYPT", defaultEncrypt)
	backupConfig.Compression = common.GetStringFlagOrEnv(cmd, "compression", "COMPRESSION", defaultCompression)
	backupConfig.CompressionLevel = common.GetStringFlagOrEnv(cmd, "compression-level", "COMPRESSION_LEVEL", defaultCompressionLevel)
	backupConfig.VerifyDisk = common.GetBoolFlagOrEnv(cmd, "verify-disk", "VERIFY_DISK", defaultVerifyDisk)
	backupConfig.RetentionDays = common.GetIntFlagOrEnv(cmd, "retention-days", "RETENTION_DAYS", defaultRetentionDays)
	backupConfig.CalculateChecksum = common.GetBoolFlagOrEnv(cmd, "calculate-checksum", "CALCULATE_CHECKSUM", defaultCalculateChecksum)

	if backupConfig.Compression == "" && backupConfig.Compress {
		backupConfig.Compression = "gzip"
	}

	return backupConfig, nil
}

// ConvertToBackupOptions converts BackupConfig to BackupOptions for backward compatibility
func (bc *BackupConfig) ToBackupOptions() BackupOptions {
	return BackupOptions{
		Host:              bc.Host,
		Port:              bc.Port,
		User:              bc.User,
		Password:          bc.Password,
		DBName:            bc.DBName,
		OutputDir:         bc.OutputDir,
		Compress:          bc.Compress,
		Compression:       bc.Compression,
		CompressionLevel:  bc.CompressionLevel,
		IncludeData:       bc.IncludeData,
		Encrypt:           bc.Encrypt,
		VerifyDisk:        bc.VerifyDisk,
		RetentionDays:     bc.RetentionDays,
		CalculateChecksum: bc.CalculateChecksum,
	}
}
