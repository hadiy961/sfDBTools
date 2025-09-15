package interactive

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
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
func GatherInteractiveInput(ctx context.Context, mariadbConfig *mariadb_utils.MariaDBConfigureConfig, template *MariaDBConfigTemplate) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting interactive configuration input")

	// Task 1: Load application config sebagai fallback defaults

	// Show welcome message
	terminal.PrintInfo("MariaDB Configuration Setup")
	terminal.PrintInfo("Please provide the following configuration values.")
	terminal.PrintInfo("Press Enter to use default values shown in brackets.")
	fmt.Println()

	// Create config defaults helper dengan prioritas: template -> appConfig -> hardcoded
	defaults := &ConfigDefaults{
		Template: template,
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

	// Show summary (Task 2: modular function)
	ShowConfigurationSummary(mariadbConfig)

	lg.Info("Interactive configuration input completed")
	return nil
}

// RequestUserConfirmationForConfig meminta konfirmasi user untuk menerapkan konfigurasi
// Task 2: Wrapper function untuk modular confirmation
func RequestUserConfirmationForConfig(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, _ := logger.Get()
	lg.Info("Requesting user confirmation for configuration")

	if err := RequestUserConfirmation(); err != nil {
		return err
	}

	lg.Info("User confirmed configuration changes")
	return nil
}
