package generate

import (
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/utils/common/structs"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// processAutoMode handles automated generation using provided parameters
func (p *Processor) processAutoMode(cfg *structs.DBConfig) error {
	terminal.PrintInfo("Automated mode - using provided parameters")

	// Create configuration from provided parameters
	dbConfig := &config.EncryptedDatabaseConfig{
		Host:     cfg.ConnectionOptions.Host,
		Port:     cfg.ConnectionOptions.Port,
		User:     cfg.ConnectionOptions.User,
		Password: cfg.ConnectionOptions.Password,
	}

	// Final name provided by caller
	finalConfigName := cfg.ConfigName

	// Display summary
	configInfo := &dbconfig.ConfigInfo{
		Name:         finalConfigName,
		Host:         dbConfig.Host,
		Port:         dbConfig.Port,
		User:         dbConfig.User,
		HasPassword:  dbConfig.Password != "",
		FileSize:     "New file",
		LastModified: time.Now(),
		IsValid:      true,
	}

	dbconfig.DisplayConfigSummary([]*dbconfig.ConfigInfo{configInfo})

	// Get encryption password and save
	encryptionPassword, err := p.GetEncryptionPassword("encrypt the configuration")
	if err != nil {
		return err
	}

	return p.saveEncryptedConfig(finalConfigName, dbConfig, encryptionPassword)
}
