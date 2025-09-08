package delete

import (
	coredbconfig "sfDBTools/internal/core/dbconfig"
)

// Processor handles delete operations for database configurations
type Processor struct {
	*coredbconfig.BaseProcessor
	configHelper *coredbconfig.ConfigHelper
}

// NewProcessor creates a new delete processor
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
