package configure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// performDataMigration melakukan migrasi data antar direktori
// Sesuai dengan Step 19 dalam flow implementasi
func performDataMigration(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting data migration process")

	// Deteksi current installation untuk mendapatkan direktori saat ini
	installation, err := mariadb_utils.DiscoverMariaDBInstallation()
	if err != nil {
		return fmt.Errorf("failed to discover current installation: %w", err)
	}

	// Check apakah perlu migrasi
	needsMigration := false
	migrations := []dataMigration{}

	// Check data directory migration
	if installation.DataDir != config.DataDir {
		migrations = append(migrations, dataMigration{
			Type:        "data",
			Source:      installation.DataDir,
			Destination: config.DataDir,
			Critical:    true,
		})
		needsMigration = true
	}

	// Check log directory migration (tidak critical)
	currentLogDir := filepath.Dir(installation.DataDir) // Approximation
	if currentLogDir != config.LogDir {
		migrations = append(migrations, dataMigration{
			Type:        "logs",
			Source:      currentLogDir,
			Destination: config.LogDir,
			Critical:    false,
		})
		needsMigration = true
	}

	// Check binlog directory migration
	currentBinlogDir := "/var/lib/mysqlbinlogs" // Default assumption
	if currentBinlogDir != config.BinlogDir {
		migrations = append(migrations, dataMigration{
			Type:        "binlogs",
			Source:      currentBinlogDir,
			Destination: config.BinlogDir,
			Critical:    false,
		})
		needsMigration = true
	}

	if !needsMigration {
		lg.Info("No data migration required")
		return nil
	}

	// Show migration plan
	showMigrationPlan(migrations)

	// Stop MariaDB service sebelum migrasi
	lg.Info("Stopping MariaDB service for data migration")
	sm := system.NewServiceManager()
	if err := sm.Stop(installation.ServiceName); err != nil {
		return fmt.Errorf("failed to stop MariaDB service: %w", err)
	}

	// Perform migrations
	for _, migration := range migrations {
		if err := performSingleMigration(migration); err != nil {
			if migration.Critical {
				return fmt.Errorf("critical migration failed: %w", err)
			}
			lg.Warn("Non-critical migration failed",
				logger.String("type", migration.Type),
				logger.Error(err))
		}
	}

	lg.Info("Data migration completed successfully")
	return nil
}

// dataMigration berisi informasi migrasi data
type dataMigration struct {
	Type        string // "data", "logs", "binlogs"
	Source      string
	Destination string
	Critical    bool // Apakah migration ini critical untuk startup
}

// showMigrationPlan menampilkan rencana migrasi ke user
func showMigrationPlan(migrations []dataMigration) {
	terminal.PrintInfo("Data Migration Plan:")
	terminal.PrintInfo("====================")

	for _, migration := range migrations {
		criticalText := ""
		if migration.Critical {
			criticalText = " (CRITICAL)"
		}

		fmt.Printf("- %s%s: %s -> %s\n",
			migration.Type, criticalText, migration.Source, migration.Destination)
	}

	fmt.Println()
	terminal.PrintWarning("This process will stop MariaDB service temporarily")
}

// performSingleMigration melakukan migrasi untuk satu direktori
func performSingleMigration(migration dataMigration) error {
	lg, _ := logger.Get()
	lg.Info("Performing migration",
		logger.String("type", migration.Type),
		logger.String("source", migration.Source),
		logger.String("destination", migration.Destination))

	// Check if source exists
	if _, err := os.Stat(migration.Source); os.IsNotExist(err) {
		if migration.Critical {
			return fmt.Errorf("source directory does not exist: %s", migration.Source)
		}
		lg.Warn("Source directory does not exist, skipping migration",
			logger.String("source", migration.Source))
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(migration.Destination, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy data
	if err := copyDirectory(migration.Source, migration.Destination); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Verify migration if critical
	if migration.Critical && migration.Type == "data" {
		if err := verifyDataMigration(migration.Source, migration.Destination); err != nil {
			return fmt.Errorf("data verification failed: %w", err)
		}
	}

	lg.Info("Migration completed successfully",
		logger.String("type", migration.Type))

	return nil
}

// copyDirectory menyalin isi direktori dari source ke destination
func copyDirectory(source, destination string) error {
	// Implementasi sederhana menggunakan os/exec
	// TODO: Bisa diperbaiki dengan implementasi native Go

	cmd := fmt.Sprintf("cp -R %s/* %s/", source, destination)

	// Use exec to run command
	execCmd := exec.Command("sh", "-c", cmd)
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy directory: %w", err)
	}

	return nil
}

// verifyDataMigration memverifikasi integritas data setelah migrasi
func verifyDataMigration(source, destination string) error {
	lg, _ := logger.Get()
	lg.Info("Verifying data migration integrity")

	// Check critical files exist
	criticalFiles := []string{
		"ibdata1",     // InnoDB system tablespace
		"ib_logfile0", // InnoDB log file
		"mysql",       // System database directory
	}

	for _, file := range criticalFiles {
		sourcePath := filepath.Join(source, file)
		destPath := filepath.Join(destination, file)

		// Check if file exists in source
		if _, err := os.Stat(sourcePath); err == nil {
			// File exists in source, must exist in destination
			if _, err := os.Stat(destPath); err != nil {
				return fmt.Errorf("critical file missing after migration: %s", file)
			}
		}
	}

	lg.Info("Data migration verification passed")
	return nil
}

// applyConfiguration menerapkan konfigurasi dan backup yang sudah ada
func applyConfiguration(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig, template *MariaDBConfigTemplate) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Applying MariaDB configuration")

	// Step 15: Backup current config
	if config.BackupCurrentConfig {
		backupPath, err := template.BackupCurrentConfig(config.BackupDir)
		if err != nil {
			return fmt.Errorf("failed to backup current config: %w", err)
		}
		lg.Info("Current configuration backed up", logger.String("backup_path", backupPath))
	}

	// Step 16-17: Generate new config from template
	configValues := buildConfigValues(config)
	newConfig, err := template.GenerateConfigFromTemplate(configValues)
	if err != nil {
		return fmt.Errorf("failed to generate config from template: %w", err)
	}

	// Step 18: Write new configuration
	if err := writeConfiguration(template.CurrentPath, newConfig); err != nil {
		return fmt.Errorf("failed to write new configuration: %w", err)
	}

	lg.Info("MariaDB configuration applied successfully",
		logger.String("config_path", template.CurrentPath))

	return nil
}

// buildConfigValues membangun map values untuk template
func buildConfigValues(config *mariadb_utils.MariaDBConfigureConfig) map[string]string {
	values := make(map[string]string)

	values["server_id"] = fmt.Sprintf("%d", config.ServerID)
	values["port"] = fmt.Sprintf("%d", config.Port)
	values["datadir"] = config.DataDir
	values["socket"] = filepath.Join(config.DataDir, "mysql.sock")
	values["log_bin"] = filepath.Join(config.BinlogDir, "mysql-bin")
	values["log_error"] = filepath.Join(config.LogDir, "mysql_error.log")
	values["slow_query_log_file"] = filepath.Join(config.LogDir, "mysql_slow.log")
	values["innodb_data_home_dir"] = config.DataDir
	values["innodb_log_group_home_dir"] = config.DataDir
	values["innodb_buffer_pool_size"] = config.InnodbBufferPoolSize
	values["innodb_buffer_pool_instances"] = fmt.Sprintf("%d", config.InnodbBufferPoolInstances)

	// Encryption settings
	if config.InnodbEncryptTables {
		values["innodb-encrypt-tables"] = "ON"
		values["file_key_management_encryption_key_file"] = config.EncryptionKeyFile
		values["file_key_management_encryption_algorithm"] = "AES_CTR"
	} else {
		values["innodb-encrypt-tables"] = "OFF"
	}

	return values
}

// writeConfiguration menulis konfigurasi baru ke file
func writeConfiguration(configPath, content string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write configuration
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
