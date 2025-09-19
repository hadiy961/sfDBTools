package defaultsetup

import (
	"context"
	"fmt"
	"sfDBTools/internal/core/mariadb/configure"
	"sfDBTools/internal/core/mariadb/configure/interactive"
	"sfDBTools/internal/core/mariadb/configure/migration"
	"sfDBTools/internal/core/mariadb/configure/service"
	"sfDBTools/internal/core/mariadb/configure/template"
	"sfDBTools/internal/core/mariadb/configure/validation"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// membuat konfigurasi standart perusahaan
func RunStandardConfiguration(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	terminal.PrintHeader("MariaDB Configuration Process")
	terminal.PrintSubHeader("Reading Existing Configurations from Application Config")
	fmt.Println()
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	headers := []string{"Dir", "Value"}
	rows := [][]string{
		{"data_dir", config.DataDir},
		{"log_dir", config.LogDir},
		{"binlog_dir", config.BinlogDir},
		{"port", fmt.Sprintf("%d", config.Port)},
		{"server_id", fmt.Sprintf("%d", config.ServerID)},
		{"innodb_encrypt_tables", fmt.Sprintf("%t", config.InnodbEncryptTables)},
		{"encryption_key_file", config.EncryptionKeyFile},
		{"backup_dir", config.BackupDir},
	}
	terminal.FormatTable(headers, rows)

	lg.Info("Performing installation and privilege checks")
	mariadbInstallation, err := configure.PerformPreChecks(ctx, config)
	if err != nil {
		return fmt.Errorf("pre-checks failed: %w", err)
	}
	terminal.PrintSubHeader("Reading Existing Configurations from MariaDB Installation (" + mariadbInstallation.ConfigPaths[0] + ")")

	headers1 := []string{"Dir", "Value"}
	rows1 := [][]string{
		{"binary", mariadbInstallation.BinaryPath},
		{"data_dir", mariadbInstallation.DataDir},
		{"log_dir", mariadbInstallation.LogDir},
		{"binlog_dir", mariadbInstallation.BinlogDir},
		{"port", fmt.Sprintf("%d", mariadbInstallation.Port)},
		{"server_id", fmt.Sprintf("%d", mariadbInstallation.ServerID)},
		{"innodb_encrypt_tables", fmt.Sprintf("%t", mariadbInstallation.InnodbEncryptTables)},
		{"encryption_key_file", mariadbInstallation.EncryptionKeyFile},
		{"innodb_buffer_pool_size", mariadbInstallation.InnodbBufferPoolSize},
		{"innodb_buffer_pool_instances", fmt.Sprintf("%d", mariadbInstallation.InnodbBufferPoolInstances)},
	}
	terminal.FormatTable(headers1, rows1)

	// Step 2-4: Template dan konfigurasi discovery - gunakan hasil discovery yang sudah ada
	template, err := template.LoadConfigurationTemplateWithInstallation(ctx, mariadbInstallation)
	if err != nil {
		return fmt.Errorf("failed to load configuration template: %w", err)
	}

	lg.Info("Performing hardware-based auto-tuning")
	if err := configure.PerformAutoTuning(ctx, config); err != nil {
		return fmt.Errorf("auto-tuning failed: %w", err)
	}

	lg.Info("Validating configuration and system requirements")
	if err := validation.ValidateSystemRequirements(ctx, config); err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	terminal.PrintSubHeader("MariaDB Configurations Comparison Summary")
	fmt.Println()

	headersNew := []string{"Dir", "Existing", "New Value"}
	rowsNew := [][]string{
		{"data_dir", mariadbInstallation.DataDir, config.DataDir},
		{"log_dir", mariadbInstallation.LogDir, config.LogDir},
		{"binlog_dir", mariadbInstallation.BinlogDir, config.BinlogDir},
		{"port", fmt.Sprintf("%d", mariadbInstallation.Port), fmt.Sprintf("%d", config.Port)},
		{"server_id", fmt.Sprintf("%d", mariadbInstallation.ServerID), fmt.Sprintf("%d", config.ServerID)},
		{"innodb_encrypt_tables", fmt.Sprintf("%t", mariadbInstallation.InnodbEncryptTables), fmt.Sprintf("%t", config.InnodbEncryptTables)},
		{"encryption_key_file", mariadbInstallation.EncryptionKeyFile, config.EncryptionKeyFile},
		{"innodb_buffer_pool_size", mariadbInstallation.InnodbBufferPoolSize, config.InnodbBufferPoolSize},
		{"innodb_buffer_pool_instances", fmt.Sprintf("%d", mariadbInstallation.InnodbBufferPoolInstances), fmt.Sprintf("%d", config.InnodbBufferPoolInstances)},
	}
	terminal.FormatTable(headersNew, rowsNew)

	// Step 11: Konfirmasi user
	lg.Info("Requesting user confirmation for configuration changes")
	if err := interactive.RequestUserConfirmationForConfig(ctx, config); err != nil {
		return fmt.Errorf("user confirmation failed: %w", err)
	}
	lg.Info("Backing up current configuration and applying new settings")
	if err := migration.ApplyConfiguration(ctx, config, template); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Langkah 5 : Restart mariadb service
	lg.Info("Restarting MariaDB service and verifying configuration")
	if err := service.RestartAndVerifyService(ctx, config, mariadbInstallation); err != nil {
		return fmt.Errorf("service restart/verification failed: %w", err)
	}

	service.ShowSuccessSummary(config)

	return nil
}
