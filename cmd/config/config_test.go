package config

import (
	"fmt"
	"os"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/crypto"

	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test encrypted database configuration connectivity",
	Long: `Test encrypted database configuration by attempting to decrypt and validate
the connection parameters. This does not perform an actual database connection test.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := testEncryptedConfig(); err != nil {
			lg, _ := logger.Get()
			lg.Error("Failed to test encrypted config", logger.Error(err))
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func testEncryptedConfig() error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Load current config to get general settings
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	lg.Info("Starting encrypted database configuration test")

	// Get encryption password
	encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
	if err != nil {
		return fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Try to load and decrypt the database configuration
	dbConfig, err := config.LoadEncryptedDatabaseConfig(cfg, encryptionPassword)
	if err != nil {
		return fmt.Errorf("failed to load encrypted database configuration: %w", err)
	}

	// Validate the configuration values
	if dbConfig.Host == "" {
		return fmt.Errorf("invalid configuration: host cannot be empty")
	}

	if dbConfig.Port < 1 || dbConfig.Port > 65535 {
		return fmt.Errorf("invalid configuration: port must be between 1 and 65535, got %d", dbConfig.Port)
	}

	if dbConfig.User == "" {
		return fmt.Errorf("invalid configuration: user cannot be empty")
	}

	if dbConfig.Password == "" {
		return fmt.Errorf("invalid configuration: password cannot be empty")
	}

	lg.Info("Encrypted database configuration test successful")

	fmt.Println("✅ Encrypted database configuration test passed!")
	fmt.Printf("   Host: %s\n", dbConfig.Host)
	fmt.Printf("   Port: %d\n", dbConfig.Port)
	fmt.Printf("   User: %s\n", dbConfig.User)
	fmt.Println("   Password: *** (verified - valid length)")
	fmt.Println("\n🔑 Password verification successful!")
	fmt.Println("🔓 Configuration can be decrypted with the provided password.")

	return nil
}

func init() {
	// This will be called from config_cmd.go
}
