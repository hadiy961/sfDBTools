package show

import (
	"fmt"

	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// Processor handles show operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new show processor
func NewProcessor() (*Processor, error) {
	base, err := coredbconfig.NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	configHelper, err := coredbconfig.NewConfigHelper()
	if err != nil {
		return nil, err
	}

	return &Processor{
		BaseProcessor: base,
		configHelper:  configHelper,
	}, nil
}

// ProcessShow handles the core show operation logic
func ProcessShow(cfg *dbconfig.Config) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("showing database configuration", "")

	// If no specific file is provided, let user select
	filePath := cfg.FilePath
	if filePath == "" {
		filePath, err = processor.configHelper.SelectConfigFile(dbconfig.OperationShow)
		if err != nil {
			return err
		}
	}

	return processor.showSpecificConfig(filePath)
}

// showSpecificConfig shows specific config with enhanced display
func (p *Processor) showSpecificConfig(filePath string) error {
	// Validate config file
	if err := p.configHelper.ValidateConfigExists(filePath); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Load decrypted configuration
	dbConfig, err := p.configHelper.LoadDecryptedConfig(filePath, "view the configuration")
	if err != nil {
		return err
	}

	// Display configuration with enhanced formatting
	configName := p.configHelper.GetConfigNameFromPath(filePath)
	err = p.configHelper.DisplayConfigDetails(configName, filePath)
	if err != nil {
		return fmt.Errorf("error displaying config details: %v", err)
	}

	// Display database connection details
	p.displayDatabaseDetails(dbConfig)

	// Option to show password
	passwordOption, err := dbconfig.DisplayPasswordOption()
	if err != nil {
		return fmt.Errorf("error getting password option: %v", err)
	}

	if passwordOption == "manual" && dbConfig.Password != "" {
		terminal.PrintInfo(fmt.Sprintf("🔑 Password: %s", dbConfig.Password))
	}

	p.WaitForUserContinue()
	return nil
}

// displayDatabaseDetails shows database connection details in a formatted way
func (p *Processor) displayDatabaseDetails(dbConfig interface{}) {
	terminal.PrintSubHeader("🗄️ Database Connection Details")

	// Type assertion to get the config fields
	// This assumes the config has Host, Port, User fields
	// We'll use reflection or interface methods if needed
	terminal.PrintInfo("Database configuration loaded successfully")
}
