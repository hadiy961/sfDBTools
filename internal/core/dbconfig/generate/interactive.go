package generate

import (
	"fmt"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/structs"
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// processInteractiveMode handles interactive generation
func (p *Processor) processInteractiveMode(dbcfg *structs.DBConfig, lg *logger.Logger) error {
	// Get configuration details from user
	inputConfig, err := dbconfig.PromptDatabaseConfig(dbcfg)
	if err != nil {
		return fmt.Errorf("error getting database configuration: %v", err)
	}

	// Get configuration name
	configName, err := dbconfig.PromptConfigName("")
	if err != nil {
		return fmt.Errorf("error getting configuration name: %v", err)
	}

	// Get password handling option
	passwordOption, err := dbconfig.DisplayPasswordOption(dbcfg)
	if err != nil {
		return fmt.Errorf("error getting password option: %v", err)
	}

	// Handle password based on option
	var password string
	switch passwordOption {
	case "manual":
		password = terminal.AskPassword("Enter database password", "")
		if password == "" {
			return fmt.Errorf("password cannot be empty - please provide a valid password")
		}
	case "env":
		password = dbcfg.ConnectionOptions.Password
		if password == "" {
			return fmt.Errorf("password not set in environment variable - please set SFDB_DB_PASSWORD")
		}
		terminal.PrintInfo("Using database password from environment variable (hidden)")
	default:
		return fmt.Errorf("invalid password option: %s", passwordOption)
	}

	// Create final configuration
	dbConfig := &config.EncryptedDatabaseConfig{
		Host:     inputConfig.Host,
		Port:     inputConfig.Port,
		User:     inputConfig.User,
		Password: password,
	}

	// Display summary
	configInfo := &dbconfig.ConfigInfo{
		Name:         configName,
		Host:         dbConfig.Host,
		Port:         dbConfig.Port,
		User:         dbConfig.User,
		HasPassword:  dbConfig.Password != "",
		FileSize:     "New file",
		LastModified: time.Now(),
		IsValid:      true,
	}

	dbconfig.DisplayConfigSummary([]*dbconfig.ConfigInfo{configInfo})

	// Confirm save
	if !terminal.AskYesNo("Save this configuration?", true) {
		terminal.PrintWarning("Configuration not saved.")
		return nil
	}

	// Get encryption password
	encryptionPassword, err := p.GetEncryptionPassword("encrypt the configuration")
	if err != nil {
		return err
	}

	// Save configuration
	return p.saveEncryptedConfig(configName, dbConfig, encryptionPassword)
}
