package backup_utils

import (
	"sfDBTools/internal/config"

	"github.com/spf13/cobra"
)

// AddCommonBackupFlags adds common backup flags to the given command.
func AddCommonBackupFlags(cmd *cobra.Command) {
	_, _, _, defaultOutputDir,
		defaultCompress, defaultCompression, defaultCompressionLevel, defaultIncludeData,
		defaultEncrypt, _, _, _, defaultSystemUser := config.GetBackupDefaults()

	// Configuration options
	cmd.Flags().String("config", "", "encrypted configuration file (.cnf.enc)")

	// Database connection options
	cmd.Flags().String("source_db", "", "database name")
	cmd.Flags().String("source_host", "", "source database host")
	cmd.Flags().Int("source_port", 0, "source database port")
	cmd.Flags().String("source_user", "", "source database user")
	cmd.Flags().String("source_password", "", "source database password")

	// Backup options
	cmd.Flags().Bool("compress", defaultCompress, "compress output")
	cmd.Flags().String("compression", defaultCompression, "compression format (gzip, pgzip, zlib, zstd)")
	cmd.Flags().String("compression-level", defaultCompressionLevel, "compression level (best_speed, fast, default, better, best)")
	cmd.Flags().String("output-dir", defaultOutputDir, "output directory")
	cmd.Flags().Bool("data", defaultIncludeData, "include data in backup")
	cmd.Flags().Bool("encrypt", defaultEncrypt, "encrypt output")
	cmd.Flags().Bool("system-user", defaultSystemUser, "include system users (sst_user, papp, sysadmin, backup_user, dbaDO, maxscale)")
}

// ParseBackupOptionsFromFlags parses backup options from command flags.
// Deprecated: Use ResolveBackupConfig instead for better configuration handling
func ParseBackupOptionsFromFlags(cmd *cobra.Command) (BackupOptions, error) {
	// Use the new configuration resolution method
	backupConfig, err := ResolveBackupConfig(cmd)
	if err != nil {
		return BackupOptions{}, err
	}

	return backupConfig.ToBackupOptions(), nil
}
