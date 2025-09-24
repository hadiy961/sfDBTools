package validate

import (
	"fmt"

	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// Processor handles validate operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper     *coredbconfig.ConfigHelper
	validationHelper *coredbconfig.ValidationHelper
}

// NewProcessor creates a new validate processor
func NewProcessor() (*Processor, error) {
	base, err := coredbconfig.NewBaseProcessor()
	if err != nil {
		return nil, err
	}

	configHelper, err := coredbconfig.NewConfigHelper()
	if err != nil {
		return nil, err
	}

	validationHelper, err := coredbconfig.NewValidationHelper()
	if err != nil {
		return nil, err
	}

	return &Processor{
		BaseProcessor:    base,
		configHelper:     configHelper,
		validationHelper: validationHelper,
	}, nil
}

// ProcessValidate handles the core validation operation logic
func ProcessValidate(cfg *dbconfig.Config) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}
	terminal.Headers("🔍 Validate Database Configuration")
	processor.LogOperation("database configuration validation", "")

	// If no specific file is provided, let user select
	filePath := cfg.FilePath
	if filePath == "" {
		filePath, err = processor.configHelper.SelectConfigFile(dbconfig.OperationValidate)
		if err != nil {
			return err
		}
	}

	return processor.validateSpecificConfig(filePath)
}

// validateSpecificConfig validates specific config with comprehensive checks
func (p *Processor) validateSpecificConfig(filePath string) error {
	// Use the validation module to validate the file
	result, err := p.validationHelper.ValidateConfigFile(filePath)
	if err != nil {
		return fmt.Errorf("validation error: %v", err)
	}

	// Display basic validation result
	dbconfig.DisplayValidationResult(result)

	// Additional validation with decryption if basic validation passes
	if result.IsValid {
		if err := p.validationHelper.ValidateWithDecryption(filePath, result); err != nil {
			return err
		}

		// Display final results
		dbconfig.DisplayValidationResult(result)
		terminal.PrintSuccess("Configuration validation completed successfully!")
	} else {
		terminal.PrintError("Configuration validation failed")
		// return fmt.Errorf("configuration validation failed")
	}

	p.WaitForUserContinue()
	return nil
}
