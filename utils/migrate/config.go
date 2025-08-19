package migrate_utils

import (
	"fmt"

	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// ResolveMigrationConfig resolves migration configuration from various sources with proper priority
func ResolveMigrationConfig(cmd *cobra.Command) (*MigrationConfig, error) {
	migrationConfig := &MigrationConfig{}

	// Resolve source database connection
	sourceHost, sourcePort, sourceUser, sourcePassword, sourceSource, err := ResolveSourceDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source database connection: %w", err)
	}

	migrationConfig.SourceHost = sourceHost
	migrationConfig.SourcePort = sourcePort
	migrationConfig.SourceUser = sourceUser
	migrationConfig.SourcePassword = sourcePassword

	// Display source configuration
	switch sourceSource {
	case SourceConfigFile:
		fmt.Printf("📁 Source: Using configuration file\n")
		fmt.Printf("   Host: %s:%d\n", sourceHost, sourcePort)
		fmt.Printf("   User: %s\n", sourceUser)
	case SourceFlags:
		fmt.Printf("🔧 Source: Using command line flags\n")
		fmt.Printf("   Host: %s:%d\n", sourceHost, sourcePort)
		fmt.Printf("   User: %s\n", sourceUser)
	case SourceInteractive:
		fmt.Printf("👤 Source: Using interactively selected configuration\n")
		fmt.Printf("   Host: %s:%d\n", sourceHost, sourcePort)
		fmt.Printf("   User: %s\n", sourceUser)
	}

	// Resolve target database connection
	targetHost, targetPort, targetUser, targetPassword, targetSource, err := ResolveTargetDatabaseConnection(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target database connection: %w", err)
	}

	migrationConfig.TargetHost = targetHost
	migrationConfig.TargetPort = targetPort
	migrationConfig.TargetUser = targetUser
	migrationConfig.TargetPassword = targetPassword

	// Display target configuration
	switch targetSource {
	case SourceConfigFile:
		fmt.Printf("📁 Target: Using configuration file\n")
		fmt.Printf("   Host: %s:%d\n", targetHost, targetPort)
		fmt.Printf("   User: %s\n", targetUser)
	case SourceFlags:
		fmt.Printf("🔧 Target: Using command line flags\n")
		fmt.Printf("   Host: %s:%d\n", targetHost, targetPort)
		fmt.Printf("   User: %s\n", targetUser)
	case SourceInteractive:
		fmt.Printf("👤 Target: Using interactively selected configuration\n")
		fmt.Printf("   Host: %s:%d\n", targetHost, targetPort)
		fmt.Printf("   User: %s\n", targetUser)
	}

	// Resolve source database name
	sourceDBName, err := ResolveSourceDatabaseName(cmd, sourceHost, sourcePort, sourceUser, sourcePassword)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source database name: %w", err)
	}
	migrationConfig.SourceDBName = sourceDBName

	// Resolve target database name (automatically uses source DB name if not specified)
	targetDBName, err := ResolveTargetDatabaseName(cmd, targetHost, targetPort, targetUser, targetPassword, sourceDBName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target database name: %w", err)
	}
	migrationConfig.TargetDBName = targetDBName

	// Migration options
	migrationConfig.MigrateUsers = common.GetBoolFlagOrEnv(cmd, "migrate-users", "MIGRATE_USERS", true)
	migrationConfig.MigrateData = common.GetBoolFlagOrEnv(cmd, "migrate-data", "MIGRATE_DATA", true)
	migrationConfig.MigrateStructure = common.GetBoolFlagOrEnv(cmd, "migrate-structure", "MIGRATE_STRUCTURE", true)
	migrationConfig.VerifyData = common.GetBoolFlagOrEnv(cmd, "verify-data", "VERIFY_DATA", true)

	// Standard migration flow: backup target > drop target > create target (fixed)
	migrationConfig.BackupTarget = true
	migrationConfig.DropTarget = true
	migrationConfig.CreateTarget = true

	return migrationConfig, nil
}

// AddCommonMigrationFlags adds common migration flags for single database migration
func AddCommonMigrationFlags(cmd *cobra.Command) {
	// Source configuration options
	cmd.Flags().String("source-config", "", "source encrypted configuration file (.cnf.enc)")
	cmd.Flags().String("source-host", "", "source database host")
	cmd.Flags().Int("source-port", 0, "source database port")
	cmd.Flags().String("source-user", "", "source database user")
	cmd.Flags().String("source-password", "", "source database password")
	cmd.Flags().String("source-db", "", "source database name")

	// Target configuration options
	cmd.Flags().String("target-config", "", "target encrypted configuration file (.cnf.enc)")
	cmd.Flags().String("target-host", "", "target database host")
	cmd.Flags().Int("target-port", 0, "target database port")
	cmd.Flags().String("target-user", "", "target database user")
	cmd.Flags().String("target-password", "", "target database password")
	cmd.Flags().String("target-db", "", "target database name (defaults to source database name)")

	// Migration options
	cmd.Flags().Bool("migrate-users", true, "migrate database users and grants")
	cmd.Flags().Bool("migrate-data", true, "migrate database data")
	cmd.Flags().Bool("migrate-structure", true, "migrate database structure")
	cmd.Flags().Bool("verify-data", true, "verify data integrity after migration")
	cmd.Flags().Bool("backup-target", true, "backup target database before migration")
}
