package generate

import (
	coredbconfig "sfDBTools/internal/core/dbconfig"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/dbconfig"
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
func ProcessGenerate(cfg *dbconfig.Config, lg *logger.Logger) error {
	// Clear screen and show header
	terminal.Headers("Buat Konfigurasi DB")
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration generation", "")

	// If all required params for automated mode are present, use it.
	autoMode := cfg.AutoMode && cfg.ConfigName != "" && cfg.Host != "" && cfg.Port != 0 && cfg.User != ""
	if autoMode {
		return processor.processAutoMode(cfg)
	}

	return processor.processInteractiveMode()
}
