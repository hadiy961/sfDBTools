package configure

import (
	"context"
	"fmt"

	"sfDBTools/internal/core/mariadb/configure/interactive"
	"sfDBTools/internal/core/mariadb/configure/migration"
	"sfDBTools/internal/core/mariadb/configure/service"
	"sfDBTools/internal/core/mariadb/configure/template"
	validation "sfDBTools/internal/core/mariadb/configure/validation"
	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// RunMariaDBConfigure adalah entry point utama untuk konfigurasi MariaDB
// Mengikuti flow implementasi yang telah ditentukan dalam dokumentasi
func RunMariaDBConfigure(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	terminal.ClearScreen()
	terminal.PrintHeader("MariaDB Configuration Process")
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Step 1: Installation Checks - simpan hasil discovery untuk digunakan kembali
	lg.Info("Performing installation and privilege checks")
	lg.Info("Gathering existing MariaDB Configuration from application config")
	lg.Info("Gathering existing MariaDB installation details")
	mariadbInstallation, err := PerformPreChecks(ctx, config)
	if err != nil {
		return fmt.Errorf("pre-checks failed: %w", err)
	}

	lg.Info("Loading configuration template and current settings")
	// Step 2-4: Template dan konfigurasi discovery - gunakan hasil discovery yang sudah ada
	template, err := template.LoadConfigurationTemplateWithInstallation(ctx, mariadbInstallation)
	if err != nil {
		return fmt.Errorf("failed to load configuration template: %w", err)
	}

	// Step 6: Interactive input gathering
	lg.Info("Gathering interactive configuration input")
	// Convert template type untuk kompatibilitas
	interactiveTemplate := &interactive.MariaDBConfigTemplate{
		TemplatePath:  template.TemplatePath,
		Content:       template.Content,
		Placeholders:  template.Placeholders,
		DefaultValues: template.DefaultValues,
		CurrentConfig: template.CurrentConfig,
		CurrentPath:   template.CurrentPath,
	}
	if err := interactive.GatherInteractiveInput(ctx, config, interactiveTemplate, mariadbInstallation); err != nil {
		return fmt.Errorf("failed to gather interactive input: %w", err)
	}

	// Step 7-11: Validasi input dan sistem
	lg.Info("Validating configuration and system requirements")
	if err := validation.ValidateSystemRequirements(ctx, config); err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	// Step 12-14: Hardware checks dan auto-tuning
	if config.AutoTune {
		lg.Info("Performing hardware-based auto-tuning")
		if err := PerformAutoTuning(ctx, config); err != nil {
			return fmt.Errorf("auto-tuning failed: %w", err)
		}
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

	// Step 19: Data Migration (jika diperlukan)
	lg.Info("Starting data migration process")
	// Use the already discovered installation to avoid duplicated discovery work
	if err := migration.PerformDataMigrationWithInstallation(ctx, config, mariadbInstallation); err != nil {
		return fmt.Errorf("data migration failed: %w", err)
	}

	// Step 15-18: Backup dan konfigurasi
	lg.Info("Backing up current configuration and applying new settings")
	if err := migration.ApplyConfiguration(ctx, config, template); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Step 20-23: Service restart dan verifikasi
	lg.Info("Restarting MariaDB service and verifying configuration")
	if err := service.RestartAndVerifyService(ctx, config, mariadbInstallation); err != nil {
		return fmt.Errorf("service restart/verification failed: %w", err)
	}

	// Step 24-25: Cleanup dan update konfigurasi aplikasi
	lg.Info("Finalizing configuration and updating application settings")
	if err := service.FinalizeConfiguration(config); err != nil {
		return fmt.Errorf("failed to finalize configuration: %w", err)
	}

	// lg.Info("MariaDB configuration completed successfully")
	return nil
}
