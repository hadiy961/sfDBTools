package backup_restore_utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sfDBTools/utils/common"
	"sfDBTools/utils/crypto"

	"github.com/spf13/cobra"
)

// BackupRestoreConfig represents the configuration for backup restore operation
type BackupRestoreConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Target            string
	Account           string
	Encrypt           bool
	DryRun            bool
	SkipConfirmation  bool
	ProductionDB      string
	ProductionDmartDB string
	TargetDB          string
	TargetDmartDB     string
	Users             []string
}

// ResolveBackupRestoreConfig resolves backup restore configuration from various sources
func ResolveBackupRestoreConfig(cmd *cobra.Command) (*BackupRestoreConfig, error) {
	config := &BackupRestoreConfig{}

	// Get target and account from flags
	target, err := cmd.Flags().GetString("target")
	if err != nil || target == "" {
		return nil, fmt.Errorf("target flag is required")
	}
	config.Target = target

	account, err := cmd.Flags().GetString("acc")
	if err != nil || account == "" {
		return nil, fmt.Errorf("acc flag is required")
	}
	config.Account = account

	// Get encryption flag
	config.Encrypt, _ = cmd.Flags().GetBool("encrypt")
	config.DryRun, _ = cmd.Flags().GetBool("dry-run")
	config.SkipConfirmation, _ = cmd.Flags().GetBool("yes")

	// Resolve database connection from config or flags
	configFile, _ := cmd.Flags().GetString("config")

	var host string
	var port int
	var user, password string

	if configFile != "" {
		// Use specified config file
		host, port, user, password, err = common.GetDatabaseConfigFromEncrypted(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load encrypted config: %w", err)
		}
	} else {
		// Interactive config file selection
		configFile, err = common.SelectConfigFileInteractive()
		if err != nil {
			return nil, fmt.Errorf("failed to select config file: %w", err)
		}

		// Get encryption password
		encryptionPassword, err := crypto.GetEncryptionPassword("Enter encryption password: ")
		if err != nil {
			return nil, fmt.Errorf("failed to get encryption password: %w", err)
		}

		// Load config
		dbConfig, err := common.LoadEncryptedConfigFromFile(configFile, encryptionPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load encrypted config: %w", err)
		}

		host = dbConfig.Host
		port = dbConfig.Port
		user = dbConfig.User
		password = dbConfig.Password
	}

	config.Host = host
	config.Port = port
	config.User = user
	config.Password = password

	fmt.Printf("👤 Using interactively selected configuration\n")
	fmt.Printf("   Host: %s:%d\n", host, port)
	fmt.Printf("   User: %s\n", user)

	// Generate database names based on account and target
	config.ProductionDB = fmt.Sprintf("dbsf_nbc_%s", account)
	config.ProductionDmartDB = fmt.Sprintf("dbsf_nbc_%s_dmart", account)
	config.TargetDB = fmt.Sprintf("dbsf_nbc_%s_secondary_%s", account, target)
	config.TargetDmartDB = fmt.Sprintf("dbsf_nbc_%s_secondary_%s_dmart", account, target)

	// Generate user names
	config.Users = []string{
		fmt.Sprintf("sfnbc_%s_admin", account),
		fmt.Sprintf("sfnbc_%s_fin", account),
		fmt.Sprintf("sfnbc_%s_user", account),
	}

	return config, nil
}

// DisplayBackupRestoreConfig displays the resolved configuration
func DisplayBackupRestoreConfig(config *BackupRestoreConfig) {
	fmt.Printf("\n=== Backup Restore Configuration ===\n")
	fmt.Printf("Account:              %s\n", config.Account)
	fmt.Printf("Target Environment:   %s\n", config.Target)
	fmt.Printf("Production DB:        %s\n", config.ProductionDB)
	fmt.Printf("Production Dmart DB:  %s\n", config.ProductionDmartDB)
	fmt.Printf("Target DB:            %s\n", config.TargetDB)
	fmt.Printf("Target Dmart DB:      %s\n", config.TargetDmartDB)
	fmt.Printf("Users:                %s\n", strings.Join(config.Users, ", "))
	fmt.Printf("Encryption:           %t\n", config.Encrypt)
	fmt.Printf("Dry Run:              %t\n", config.DryRun)
	fmt.Printf("====================================\n\n")
}

// PromptBackupRestoreConfirmation prompts user for confirmation before executing
func PromptBackupRestoreConfirmation(config *BackupRestoreConfig) error {
	if config.SkipConfirmation {
		fmt.Println("✅ Skipping confirmation (--yes flag)")
		return nil
	}

	fmt.Printf("⚠️  BACKUP RESTORE CONFIRMATION\n")
	fmt.Printf("===============================\n")
	fmt.Printf("You are about to copy production databases to secondary:\n")
	fmt.Printf("  Source: %s, %s\n", config.ProductionDB, config.ProductionDmartDB)
	fmt.Printf("  Target: %s, %s\n", config.TargetDB, config.TargetDmartDB)
	fmt.Printf("  Host:   %s:%d\n", config.Host, config.Port)
	fmt.Printf("  User:   %s\n", config.User)
	fmt.Printf("\n🚨 WARNING: This will overwrite existing data in target databases!\n\n")

	if config.DryRun {
		fmt.Printf("🔍 DRY RUN MODE: No actual changes will be made\n\n")
	}

	fmt.Printf("Do you want to continue with the backup restore operation? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("operation cancelled by user")
	}

	fmt.Println("✅ Proceeding with backup restore...")
	return nil
}
