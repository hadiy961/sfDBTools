package interactive

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/mariadb/discovery"
	"sfDBTools/utils/terminal"
)

// MariaDBConfigTemplate adalah struct untuk template configuration
// Dipindahkan ke sini untuk menghindari import cycle
type MariaDBConfigTemplate struct {
	TemplatePath  string            `json:"template_path"`
	Content       string            `json:"content"`
	Placeholders  map[string]string `json:"placeholders"`
	DefaultValues map[string]string `json:"default_values"`
	CurrentConfig string            `json:"current_config"`
	CurrentPath   string            `json:"current_path"`
}

// GatherInteractiveInput mengumpulkan input konfigurasi secara interaktif
// Task 1: Menggunakan config.yaml sebagai fallback defaults
// Task 2: Refactored menjadi modular dengan helper functions
func GatherInteractiveInput(ctx context.Context, mariadbConfig *mariadb_config.MariaDBConfigureConfig, template *MariaDBConfigTemplate, mariadbInstallation *discovery.MariaDBInstallation) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Respect context cancellation early
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	lg.Debug("Starting interactive configuration input")

	// Task 1: Load application config sebagai fallback defaults

	// Show welcome message
	terminal.Headers("MariaDB Configuration Setup")
	terminal.PrintSubHeader("Please provide the following configuration values.")
	terminal.PrintInfo("Press Enter to use default values shown in brackets.")
	fmt.Println()

	// Create config defaults helper dengan prioritas: template -> currentConfig -> Appconfig
	defaults := &ConfigDefaults{
		Template:     template,
		Installation: mariadbInstallation,
		AppConfig:    mariadbConfig, // Optional, jika ingin gunakan config.yaml
	}

	// Create input collector untuk simplify input gathering (Task 2: modular)
	collector := NewInputCollector(defaults)

	// Task 2: Gather each configuration value menggunakan modular functions
	if err := GatherServerID(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather server ID: %w", err)
	}

	if err := GatherPort(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather port: %w", err)
	}

	if err := GatherDataDirectory(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather data directory: %w", err)
	}

	if err := GatherLogDirectory(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather log directory: %w", err)
	}

	if err := GatherBinlogDirectory(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather binlog directory: %w", err)
	}

	if err := GatherEncryptionSettings(mariadbConfig, collector); err != nil {
		return fmt.Errorf("failed to gather encryption settings: %w", err)
	}

	lg.Info("Interactive configuration input completed")
	return nil
}

// RequestUserConfirmationForConfig meminta konfirmasi user untuk menerapkan konfigurasi
// Task 2: Wrapper function untuk modular confirmation
func RequestUserConfirmationForConfig(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Requesting user confirmation for configuration")

	// Respect context cancellation before prompting user
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	if err := RequestUserConfirmation(); err != nil {
		return err
	}

	lg.Info("User confirmed configuration changes")
	return nil
}
