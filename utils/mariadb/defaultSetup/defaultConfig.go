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
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/terminal"
)

// membuat konfigurasi standart perusahaan
func RunStandardConfiguration(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig, installation *discovery.MariaDBInstallation) error {
	terminal.Headers("MariaDB Configuration Process")
	fmt.Println()
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}
	lg.Info("Reading Existing Configurations from Application Config")

	// Guard against nil installation or empty ConfigPaths to avoid panics
	installPath := "unknown"
	if installation != nil && len(installation.ConfigPaths) > 0 {
		installPath = installation.ConfigPaths[0]
	}
	lg.Info("Reading Existing Configurations from MariaDB Installation (" + installPath + ")")

	// Step 2-4: Template dan konfigurasi discovery - gunakan hasil discovery yang sudah ada
	lg.Info("Loading configuration template and current settings")
	template, err := template.LoadConfigurationTemplateWithInstallation(ctx, installation)
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
		{"data_dir", installation.DataDir, config.DataDir},
		{"log_dir", installation.LogDir, config.LogDir},
		{"binlog_dir", installation.BinlogDir, config.BinlogDir},
		{"port", fmt.Sprintf("%d", installation.Port), fmt.Sprintf("%d", config.Port)},
		{"server_id", fmt.Sprintf("%d", installation.ServerID), fmt.Sprintf("%d", config.ServerID)},
		{"innodb_encrypt_tables", fmt.Sprintf("%t", installation.InnodbEncryptTables), fmt.Sprintf("%t", config.InnodbEncryptTables)},
		{"encryption_key_file", installation.EncryptionKeyFile, config.EncryptionKeyFile},
		{"innodb_buffer_pool_size", installation.InnodbBufferPoolSize, config.InnodbBufferPoolSize},
		{"innodb_buffer_pool_instances", fmt.Sprintf("%d", installation.InnodbBufferPoolInstances), fmt.Sprintf("%d", config.InnodbBufferPoolInstances)},
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
	lg.Info("Validating configuration and system requirements")
	if err := validation.ValidateSystemRequirements(ctx, config); err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	// Langkah 5 : Restart mariadb service
	lg.Info("Restarting MariaDB service and verifying configuration")
	if err := service.RestartAndVerifyService(ctx, config, installation); err != nil {
		return fmt.Errorf("service restart/verification failed: %w", err)
	}

	service.ShowSuccessSummary(config)

	return nil
}
