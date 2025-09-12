package configure

import (
	"context"
	"fmt"

	"sfDBTools/internal/core/mariadb/configure/interactive"
	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
)

// RunMariaDBConfigure adalah entry point utama untuk konfigurasi MariaDB
// Mengikuti flow implementasi yang telah ditentukan dalam dokumentasi
func RunMariaDBConfigure(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB configuration process",
		logger.String("data_dir", config.DataDir),
		logger.String("log_dir", config.LogDir),
		logger.String("binlog_dir", config.BinlogDir),
		logger.Int("port", config.Port),
		logger.Int("server_id", config.ServerID),
		logger.Bool("non_interactive", config.NonInteractive),
	)

	// Step 1: Installation Checks - simpan hasil discovery untuk digunakan kembali
	lg.Info("Performing installation and privilege checks")
	mariadbInstallation, err := performPreChecks(ctx, config)
	if err != nil {
		return fmt.Errorf("pre-checks failed: %w", err)
	}

	// Step 2-4: Template dan konfigurasi discovery - gunakan hasil discovery yang sudah ada
	// lg.Info("Loading MariaDB configuration template")
	template, err := loadConfigurationTemplateWithInstallation(ctx, mariadbInstallation)
	if err != nil {
		return fmt.Errorf("failed to load configuration template: %w", err)
	}

	// Step 5: Baca konfigurasi aplikasi (sudah dilakukan di ResolveMariaDBConfigureConfig)
	lg.Info("Configuration loaded from application config")

	// Step 6: Interactive input gathering (jika tidak non-interactive)
	if !config.NonInteractive {
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
		if err := interactive.GatherInteractiveInput(ctx, config, interactiveTemplate); err != nil {
			return fmt.Errorf("failed to gather interactive input: %w", err)
		}
	}

	// Step 7-11: Validasi input dan sistem
	lg.Info("Validating configuration and system requirements")
	if err := validateSystemRequirements(ctx, config); err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	// Step 12-14: Hardware checks dan auto-tuning
	if config.AutoTune {
		lg.Info("Performing hardware-based auto-tuning")
		if err := performAutoTuning(ctx, config); err != nil {
			return fmt.Errorf("auto-tuning failed: %w", err)
		}
	}

	// Step 11: Konfirmasi user (jika tidak non-interactive)
	if !config.NonInteractive {
		lg.Info("Requesting user confirmation for configuration changes")
		if err := interactive.RequestUserConfirmationForConfig(ctx, config); err != nil {
			return fmt.Errorf("user confirmation failed: %w", err)
		}
	}

	// Step 15-18: Backup dan konfigurasi
	lg.Info("Backing up current configuration and applying new settings")
	if err := applyConfiguration(ctx, config, template); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Step 19: Data Migration (jika diperlukan)
	if config.MigrateData {
		lg.Info("Starting data migration process")
		if err := performDataMigration(ctx, config); err != nil {
			return fmt.Errorf("data migration failed: %w", err)
		}
	}

	// Step 20-23: Service restart dan verifikasi
	lg.Info("Restarting MariaDB service and verifying configuration")
	if err := restartAndVerifyService(ctx, config); err != nil {
		return fmt.Errorf("service restart/verification failed: %w", err)
	}

	// Step 24-25: Cleanup dan update konfigurasi aplikasi
	lg.Info("Finalizing configuration and updating application settings")
	if err := finalizeConfiguration(ctx, config); err != nil {
		return fmt.Errorf("failed to finalize configuration: %w", err)
	}

	lg.Info("MariaDB configuration completed successfully")
	return nil
}
