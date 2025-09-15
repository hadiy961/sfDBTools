package template

import (
	"context"
	"fmt"
	"os"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb/discovery"
)

func LoadConfigurationTemplateWithInstallation(ctx context.Context, installation *mariadb_utils.MariaDBInstallation) (*MariaDBConfigTemplate, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Loading MariaDB configuration template")

	template := &MariaDBConfigTemplate{
		Placeholders:  make(map[string]string),
		DefaultValues: make(map[string]string),
	}

	// Step 2: Prefer template path from application config (config_dir.mariadb_config)
	templatePath := "/etc/sfDBTools/server.cnf"
	if cfg, err := config.Get(); err == nil && cfg != nil {
		if cfg.ConfigDir.MariaDBConfig != "" {
			templatePath = cfg.ConfigDir.MariaDBConfig
		}
	}

	if err := LoadTemplateFile(template, templatePath); err != nil {
		return nil, fmt.Errorf("failed to load template file: %w", err)
	}

	currentConfigPath, err := FindCurrentConfigFileFromInstallation(installation)
	if err != nil {
		lg.Warn("Failed to find current config file, will use default", logger.Error(err))
		template.CurrentPath = "/etc/my.cnf.d/50-server.cnf"
	} else {
		template.CurrentPath = currentConfigPath
		lg.Info("Found current config file", logger.String("path", currentConfigPath))
	}

	if err := LoadCurrentConfig(template); err != nil {
		lg.Warn("Failed to load current config, using template defaults", logger.Error(err))
	}

	ParsePlaceholders(template)
	SetDefaultValues(template)

	lg.Info("Configuration template loaded successfully",
		logger.String("template_path", template.TemplatePath),
		logger.String("current_config_path", template.CurrentPath),
		logger.Int("placeholders", len(template.Placeholders)),
	)

	return template, nil
}

func LoadTemplateFile(template *MariaDBConfigTemplate, templatePath string) error {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s. Please ensure template is installed", templatePath)
	}
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	template.TemplatePath = templatePath
	template.Content = string(content)
	if err := ParsePlaceholders(template); err != nil {
		return fmt.Errorf("failed to parse template placeholders: %w", err)
	}
	return nil
}

func FindCurrentConfigFileFromInstallation(installation *mariadb_utils.MariaDBInstallation) (string, error) {
	if len(installation.ConfigPaths) > 0 {
		return installation.ConfigPaths[0], nil
	}
	standardPaths := []string{
		"/etc/my.cnf.d/50-server.cnf",
		"/etc/my.cnf.d/server.cnf",
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
	}
	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no MariaDB configuration file found in standard locations")
}

func LoadCurrentConfig(template *MariaDBConfigTemplate) error {
	if template.CurrentPath == "" {
		return fmt.Errorf("no current config path specified")
	}
	if _, err := os.Stat(template.CurrentPath); os.IsNotExist(err) {
		return fmt.Errorf("current config file does not exist: %s", template.CurrentPath)
	}
	content, err := os.ReadFile(template.CurrentPath)
	if err != nil {
		return fmt.Errorf("failed to read current config file %s: %w", template.CurrentPath, err)
	}
	template.CurrentConfig = string(content)
	return nil
}
