package interactive

import (
	"fmt"

	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/terminal"
)

// ShowConfigurationSummary menampilkan ringkasan konfigurasi - Task 2: modular function
func ShowConfigurationSummary(config *mariadb_config.MariaDBConfigureConfig) {
	fmt.Println()
	terminal.PrintInfo("Configuration Summary:")
	terminal.PrintInfo("====================================")

	fmt.Printf("Server ID: %d\n", config.ServerID)
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Data Directory: %s\n", config.DataDir)
	fmt.Printf("Log Directory: %s\n", config.LogDir)
	fmt.Printf("Binlog Directory: %s\n", config.BinlogDir)
	fmt.Printf("Table Encryption: %t\n", config.InnodbEncryptTables)
	if config.InnodbEncryptTables {
		fmt.Printf("Encryption Key File: %s\n", config.EncryptionKeyFile)
	}

	fmt.Println()
}

// RequestUserConfirmation meminta konfirmasi user untuk menerapkan konfigurasi - Task 2: modular function
func RequestUserConfirmation() error {
	terminal.PrintWarning("The above configuration will be applied to your MariaDB server.")
	terminal.PrintWarning("This may require stopping the service and migrating data.")

	question := "Do you want to proceed?"
	confirmed := terminal.AskYesNo(question, false)

	if !confirmed {
		return fmt.Errorf("configuration cancelled by user")
	}

	return nil
}
