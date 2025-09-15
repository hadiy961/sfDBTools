package templates

import (
	"fmt"
	"os"
)

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
