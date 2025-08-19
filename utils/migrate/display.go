package migrate_utils

import (
	"bufio"
	"fmt"
	"os"
	"sfDBTools/internal/logger"
	"strings"
)

// PromptMigrationConfirmation prompts user for confirmation before performing migration
func PromptMigrationConfirmation(config *MigrationConfig) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n⚠️  MIGRATION CONFIRMATION")
	fmt.Println("==========================")
	fmt.Printf("You are about to migrate database:\n")
	fmt.Printf("  Source Database: %s\n", config.SourceDBName)
	fmt.Printf("  Source Host:     %s:%d\n", config.SourceHost, config.SourcePort)
	fmt.Printf("  Source User:     %s\n", config.SourceUser)
	fmt.Printf("  Target Database: %s\n", config.TargetDBName)
	fmt.Printf("  Target Host:     %s:%d\n", config.TargetHost, config.TargetPort)
	fmt.Printf("  Target User:     %s\n", config.TargetUser)

	fmt.Println("\n📋 Migration Options:")
	if config.BackupTarget {
		fmt.Println("  ✅ Backup target database (if exists)")
	}
	if config.DropTarget {
		fmt.Println("  ✅ Drop target database")
	}
	if config.CreateTarget {
		fmt.Println("  ✅ Create target database")
	}
	if config.MigrateStructure {
		fmt.Println("  ✅ Migrate database structure")
	}
	if config.MigrateData {
		fmt.Println("  ✅ Migrate database data")
	}
	if config.MigrateUsers {
		fmt.Println("  ✅ Migrate database users and grants")
	}
	if config.VerifyData {
		fmt.Println("  ✅ Verify data integrity after migration")
	}

	fmt.Println("\n🚨 WARNING: This will overwrite existing data in the target database!")

	fmt.Print("\nDo you want to continue with the migration? [y/N]: ")
	confirmInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	if confirm != "y" && confirm != "yes" {
		return fmt.Errorf("migration operation cancelled by user")
	}

	fmt.Println("✅ Proceeding with migration...")
	return nil
}

// PromptBulkMigrationConfirmation prompts user for confirmation before performing bulk migration
func PromptBulkMigrationConfirmation(sourceConfig, targetConfig *MigrationConfig, databases []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n⚠️  BULK MIGRATION CONFIRMATION")
	fmt.Println("===============================")
	fmt.Printf("You are about to migrate %d database(s):\n", len(databases))

	// Show first few databases and total count
	for i, dbName := range databases {
		if i < 5 { // Show first 5 databases
			fmt.Printf("  %d. %s\n", i+1, dbName)
		} else if i == 5 {
			fmt.Printf("  ... and %d more databases\n", len(databases)-5)
			break
		}
	}

	fmt.Printf("\nSource Server: %s:%d (user: %s)\n", sourceConfig.SourceHost, sourceConfig.SourcePort, sourceConfig.SourceUser)
	fmt.Printf("Target Server: %s:%d (user: %s)\n", targetConfig.TargetHost, targetConfig.TargetPort, targetConfig.TargetUser)

	fmt.Println("\n📋 Migration Options for ALL databases:")
	if sourceConfig.BackupTarget {
		fmt.Println("  ✅ Backup target databases (if they exist)")
	}
	if sourceConfig.DropTarget {
		fmt.Println("  ✅ Drop target databases")
	}
	if sourceConfig.CreateTarget {
		fmt.Println("  ✅ Create target databases")
	}
	if sourceConfig.MigrateStructure {
		fmt.Println("  ✅ Migrate database structures")
	}
	if sourceConfig.MigrateData {
		fmt.Println("  ✅ Migrate database data")
	}
	if sourceConfig.MigrateUsers {
		fmt.Println("  ✅ Migrate database users and grants")
	}
	if sourceConfig.VerifyData {
		fmt.Println("  ✅ Verify data integrity after migration")
	}

	fmt.Println("\n🚨 WARNING: This will overwrite existing data in ALL target databases!")

	fmt.Print("\nDo you want to continue with the bulk migration? [y/N]: ")
	confirmInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	if confirm != "y" && confirm != "yes" {
		return fmt.Errorf("bulk migration operation cancelled by user")
	}

	fmt.Println("✅ Proceeding with bulk migration...")
	return nil
}

// DisplayMigrationSummary shows migration summary after completion
func DisplayMigrationSummary(databases []string, successCount, errorCount int, duration string, lg *logger.Logger) {
	fmt.Println("\n=== Migration Summary ===")
	fmt.Printf("Total Databases:    %d\n", len(databases))
	fmt.Printf("Successful:         %d\n", successCount)
	fmt.Printf("Failed:             %d\n", errorCount)
	fmt.Printf("Duration:           %s\n", duration)
	fmt.Println("=========================")

	if lg != nil {
		lg.Info("Migration summary",
			logger.Int("total_databases", len(databases)),
			logger.Int("successful", successCount),
			logger.Int("failed", errorCount),
			logger.String("duration", duration))
	}
}
