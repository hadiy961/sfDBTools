package generate

import (
	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/structs"
	"sfDBTools/utils/terminal"
)

// Processor handles generate operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new generate processor
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

// ProcessGenerate handles the core generate operation logic and delegates to
// either automated or interactive flows.
func ProcessGenerate(dbcfg *structs.DBConfigGenerateOptions, lg *logger.Logger) error {
	// Clear screen and show header
	terminal.Headers("Buat Konfigurasi DB")
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration generation", "")
	return processor.processInteractiveMode(dbcfg, lg)
}
