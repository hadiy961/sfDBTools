package mariadb

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"sfDBTools/internal/core/mariadb/templates"
	"sfDBTools/internal/logger"
)

// GenerateServerConfig generates MariaDB server configuration from template
func GenerateServerConfig(outputPath string, params *templates.MariaDBConfigParams) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Generating MariaDB server configuration from template")

	// Use default params if none provided
	if params == nil {
		defaultParams, err := templates.GetDefaultParams()
		if err != nil {
			lg.Error("Failed to get default parameters", logger.Error(err))
			return err
		}
		params = &defaultParams
	}

	// Parse the template
	tmpl, err := template.New("mariadb-server").Parse(templates.MariaDBServerTemplate)
	if err != nil {
		lg.Error("Failed to parse MariaDB server template", logger.Error(err))
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with parameters
	var configBuffer bytes.Buffer
	if err := tmpl.Execute(&configBuffer, params); err != nil {
		lg.Error("Failed to execute template", logger.Error(err))
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write configuration to file
	if err := os.WriteFile(outputPath, configBuffer.Bytes(), 0644); err != nil {
		lg.Error("Failed to write configuration file",
			logger.String("path", outputPath),
			logger.Error(err))
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	lg.Info("MariaDB server configuration generated successfully",
		logger.String("output_path", outputPath))

	return nil
}

// GenerateServerConfigInteractive generates config with interactive parameter input
func GenerateServerConfigInteractive(outputPath string) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting interactive MariaDB configuration generation")

	// Get default parameters
	params, err := templates.GetDefaultParams()
	if err != nil {
		lg.Error("Failed to get default parameters", logger.Error(err))
		return err
	}

	// Here you can add interactive prompts to customize parameters
	// For now, we'll use defaults with some customization
	fmt.Println("🔧 MariaDB Configuration Generator")
	fmt.Println("==================================")
	fmt.Printf("Using default configuration parameters:\n")
	fmt.Printf("  Server ID: %s\n", params.ServerID)
	fmt.Printf("  Port: %s\n", params.Port)
	fmt.Printf("  Data Directory: %s\n", params.DataDir)
	fmt.Printf("  Max Connections: %s\n", params.MaxConnections)
	fmt.Printf("  Buffer Pool Size: %s\n", params.BufferPoolSize)
	fmt.Printf("  Bind Address: %s\n", params.BindAddress)
	fmt.Println()

	// Generate the configuration
	return GenerateServerConfig(outputPath, &params)
}

// ValidateConfigParams validates the configuration parameters
func ValidateConfigParams(params *templates.MariaDBConfigParams) error {
	if params == nil {
		return fmt.Errorf("configuration parameters cannot be nil")
	}

	if params.ServerID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}

	if params.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	if params.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}

	return nil
}
