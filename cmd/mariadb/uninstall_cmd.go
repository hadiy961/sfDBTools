package command_mariadb

import (
	"fmt"
	"os"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely uninstall MariaDB/MySQL from the system",
	Long: `Uninstall command provides a comprehensive solution for completely removing MariaDB/MySQL from your system.

This command supports:
- Cross-platform support: CentOS, RHEL, AlmaLinux, Rocky Linux, Ubuntu, Debian
- Complete package removal: All MariaDB/MySQL server, client, and common packages
- Data cleanup: Removes data directories, configuration files, and logs
- Service management: Stops and disables MariaDB services
- Verification: Confirms complete removal and reports any remaining components
- Safety prompts: Interactive confirmation with detailed warnings about data loss

⚠️  WARNING: This will permanently delete all databases, users, and configuration!
Always backup your data before uninstalling.

Examples:
  # Interactive uninstall with confirmation prompts
  sfDBTools mariadb uninstall

  # Force uninstall without confirmation (use with caution)
  sfDBTools mariadb uninstall --force

  # Keep data directories (remove only packages and configs)
  sfDBTools mariadb uninstall --keep-data

  # Keep configuration files (remove only packages and data)
  sfDBTools mariadb uninstall --keep-config`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeUninstall(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("Command failed", logger.Error(err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "uninstall",
		"category": "mariadb",
	},
}

func executeUninstall(cmd *cobra.Command) error {
	// Clear screen and show header
	terminal.ClearAndShowHeader("MariaDB/MySQL Uninstaller")

	lg, err := logger.Get()
	if err != nil {
		terminal.PrintError("Failed to initialize logger")
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Get flags
	force, _ := cmd.Flags().GetBool("force")
	keepData, _ := cmd.Flags().GetBool("keep-data")
	keepConfig, _ := cmd.Flags().GetBool("keep-config")
	backupFirst, _ := cmd.Flags().GetBool("backup-first")
	backupDir, _ := cmd.Flags().GetString("backup-dir")

	lg.Info("Starting MariaDB uninstall process",
		logger.Bool("force", force),
		logger.Bool("keep_data", keepData),
		logger.Bool("keep_config", keepConfig),
		logger.Bool("backup_first", backupFirst))

	// Show uninstall options summary
	terminal.PrintSubHeader("📋 Uninstall Configuration")
	showUninstallConfiguration(force, keepData, keepConfig, backupFirst, backupDir)

	// Show current system status
	terminal.PrintSubHeader("🔍 System Status Check")
	showSystemStatus()

	// Show warning and get confirmation (unless force mode)
	if !force {
		fmt.Println()
		terminal.PrintSeparator()
		showEnhancedWarning()

		confirmed, err := terminal.ConfirmAndClear("Do you want to proceed with the uninstall?")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirmed {
			lg.Info("Uninstall cancelled by user")
			terminal.PrintInfo("Uninstall cancelled by user.")
			return nil
		}
	}

	// Clear screen and start uninstall process
	terminal.ClearAndShowHeader("🗑️ MariaDB/MySQL Uninstall in Progress")

	// Prepare options first
	options := mariadb_utils.UninstallOptions{
		Force:       force,
		KeepData:    keepData,
		KeepConfig:  keepConfig,
		BackupFirst: backupFirst,
		BackupDir:   backupDir,
	}

	// Show what will be done
	terminal.PrintSubHeader("📋 Uninstall Steps")
	showUninstallSteps(options)

	// Create progress spinner for the uninstall process
	spinner := terminal.NewProgressSpinner("🔍 Initializing uninstall process...")
	spinner.Start()

	// Simulate different phases for better user feedback
	spinner.UpdateMessage("🛑 Stopping MariaDB services...")
	// Small delay to show the message (in real implementation, this would be actual work)

	spinner.UpdateMessage("📦 Removing packages...")

	spinner.UpdateMessage("🗂️ Cleaning up directories...")

	spinner.UpdateMessage("⚙️ Removing configuration files...")

	spinner.UpdateMessage("🔧 Finalizing cleanup...")

	// Execute uninstall
	result, err := mariadb.UninstallMariaDB(options)

	spinner.Stop()

	if err != nil {
		lg.Error("Uninstall failed", logger.Error(err))
		terminal.PrintError(fmt.Sprintf("❌ Uninstall failed: %v", err))

		if result != nil {
			displayEnhancedResults(result)
		}
		return fmt.Errorf("uninstall failed: %w", err)
	}

	// Display results with enhanced formatting
	displayEnhancedResults(result)

	// Log results
	if result.Success {
		lg.Info("MariaDB uninstall completed successfully",
			logger.String("duration", result.Duration.String()),
			logger.Int("packages_removed", result.PackagesRemoved),
			logger.Int("directories_removed", len(result.DirectoriesRemoved)))
	} else {
		lg.Warn("MariaDB uninstall completed with issues",
			logger.String("duration", result.Duration.String()),
			logger.Int("warnings", len(result.Warnings)),
			logger.Int("errors", len(result.Errors)))
	}

	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")
	return nil
}

// showUninstallSteps displays what steps will be performed
func showUninstallSteps(options mariadb_utils.UninstallOptions) {
	steps := []string{
		"🛑 Stop MariaDB/MySQL services",
		"📦 Remove MariaDB/MySQL packages",
	}

	if !options.KeepData {
		steps = append(steps, "🗂️ Remove data directories")
	}

	if !options.KeepConfig {
		steps = append(steps, "⚙️ Remove configuration files")
	}

	if options.BackupFirst {
		steps = append([]string{"💾 Create backup to " + options.BackupDir}, steps...)
	}

	steps = append(steps, "🧹 Clean up repositories and cache")
	steps = append(steps, "✅ Verify complete removal")

	for i, step := range steps {
		terminal.PrintColoredText(fmt.Sprintf("  %d. ", i+1), terminal.ColorCyan)
		fmt.Println(step)
	}
}

// showSystemStatus displays current MariaDB/MySQL system status
func showSystemStatus() {
	// Check if MariaDB service is running
	serviceStatus := checkMariaDBService()

	// Check for installed packages (simplified check)
	packagesInstalled := checkInstalledPackages()

	headers := []string{"Component", "Status", "Description"}
	rows := [][]string{
		{"MariaDB Service", serviceStatus, "Current service status"},
		{"Packages", packagesInstalled, "Estimated packages installed"},
	}

	terminal.FormatTable(headers, rows)
}

// checkMariaDBService checks if MariaDB service is running (simplified)
func checkMariaDBService() string {
	// This is a simplified check - in real implementation you'd use proper service checking
	return terminal.ColorText("⚠️ Unknown", terminal.ColorYellow)
}

// checkInstalledPackages provides an estimate of installed packages (simplified)
func checkInstalledPackages() string {
	// This is a simplified check - in real implementation you'd query package manager
	return terminal.ColorText("🔍 Detecting...", terminal.ColorBlue)
}

// showUninstallConfiguration displays the current uninstall configuration
func showUninstallConfiguration(force, keepData, keepConfig, backupFirst bool, backupDir string) {
	headers := []string{"Option", "Value", "Description"}
	rows := [][]string{
		{"Force Mode", formatBoolValue(force), "Skip confirmation prompts"},
		{"Keep Data", formatBoolValue(keepData), "Preserve data directories"},
		{"Keep Config", formatBoolValue(keepConfig), "Preserve configuration files"},
		{"Backup First", formatBoolValue(backupFirst), "Create backup before uninstall"},
	}

	if backupFirst {
		rows = append(rows, []string{"Backup Directory", backupDir, "Location for backup files"})
	}

	terminal.FormatTable(headers, rows)
}

// formatBoolValue formats boolean values with colors
func formatBoolValue(value bool) string {
	if value {
		return terminal.ColorText("✓ Yes", terminal.ColorGreen)
	}
	return terminal.ColorText("✗ No", terminal.ColorRed)
}

// showEnhancedWarning displays enhanced warning messages
func showEnhancedWarning() {
	terminal.PrintWarning("⚠️  CRITICAL WARNING - DATA LOSS IMMINENT!")
	fmt.Println()

	terminal.PrintColoredLine("This operation will:", terminal.ColorRed)
	terminal.PrintColoredText("  • ", terminal.ColorRed)
	fmt.Println("Stop all MariaDB/MySQL services")
	terminal.PrintColoredText("  • ", terminal.ColorRed)
	fmt.Println("Remove all MariaDB/MySQL packages")
	terminal.PrintColoredText("  • ", terminal.ColorRed)
	fmt.Println("Delete all databases and user data")
	terminal.PrintColoredText("  • ", terminal.ColorRed)
	fmt.Println("Remove configuration files")
	terminal.PrintColoredText("  • ", terminal.ColorRed)
	fmt.Println("Clean up log files and temporary data")

	fmt.Println()
	terminal.PrintError("🚨 ALL DATA WILL BE PERMANENTLY LOST!")
	terminal.PrintWarning("📋 Make sure you have backed up all important data!")
	fmt.Println()
}

// displayEnhancedResults displays uninstall results with enhanced formatting
func displayEnhancedResults(result *mariadb_utils.UninstallResult) {
	terminal.ClearAndShowHeader("Uninstall Results")

	// Overall status
	if result.Success {
		terminal.PrintSuccess("🎉 MariaDB/MySQL uninstall completed successfully!")
	} else {
		terminal.PrintWarning("⚠️ Uninstall completed with some issues")
	}

	fmt.Println()
	terminal.PrintSubHeader("Summary Statistics")

	// Create summary table
	headers := []string{"Metric", "Count", "Status"}
	rows := [][]string{
		{"Packages Removed", fmt.Sprintf("%d", result.PackagesRemoved), getStatusIcon(result.PackagesRemoved > 0)},
		{"Directories Cleaned", fmt.Sprintf("%d", len(result.DirectoriesRemoved)), getStatusIcon(len(result.DirectoriesRemoved) > 0)},
		{"Config Files Removed", fmt.Sprintf("%d", len(result.ConfigFilesRemoved)), getStatusIcon(len(result.ConfigFilesRemoved) > 0)},
		{"Repositories Cleaned", fmt.Sprintf("%d", len(result.RepositoriesRemoved)), getStatusIcon(len(result.RepositoriesRemoved) > 0)},
		{"Total Duration", result.Duration.String(), "✓"},
	}

	terminal.FormatTable(headers, rows)

	// Show detailed information if available
	if len(result.DirectoriesRemoved) > 0 {
		fmt.Println()
		terminal.PrintSubHeader("Directories Removed")
		for _, dir := range result.DirectoriesRemoved {
			terminal.PrintColoredText("  ✓ ", terminal.ColorGreen)
			fmt.Println(dir)
		}
	}

	if len(result.ConfigFilesRemoved) > 0 {
		fmt.Println()
		terminal.PrintSubHeader("Configuration Files Removed")
		for _, file := range result.ConfigFilesRemoved {
			terminal.PrintColoredText("  ✓ ", terminal.ColorGreen)
			fmt.Println(file)
		}
	}

	if len(result.RepositoriesRemoved) > 0 {
		fmt.Println()
		terminal.PrintSubHeader("Repositories Cleaned")
		for _, repo := range result.RepositoriesRemoved {
			terminal.PrintColoredText("  ✓ ", terminal.ColorGreen)
			fmt.Println(repo)
		}
	}

	// Show warnings if any
	if len(result.Warnings) > 0 {
		fmt.Println()
		terminal.PrintSubHeader("Warnings")
		for _, warning := range result.Warnings {
			terminal.PrintWarning(fmt.Sprintf("⚠️ %s", warning))
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Println()
		terminal.PrintSubHeader("Errors")
		for _, err := range result.Errors {
			terminal.PrintError(fmt.Sprintf("❌ %s", err))
		}
	}

	// Final status message
	fmt.Println()
	terminal.PrintDashedSeparator()
	if result.Success {
		terminal.PrintSuccess("✅ MariaDB/MySQL has been completely removed from your system.")
		terminal.PrintInfo("🔄 You can now safely install a fresh MariaDB/MySQL instance if needed.")
	} else {
		terminal.PrintWarning("⚠️ Uninstall completed but some issues were encountered.")
		terminal.PrintInfo("📋 Please review the warnings and errors above.")
		terminal.PrintInfo("🔧 You may need to manually clean up remaining components.")
	}
}

// getStatusIcon returns appropriate status icon based on count
func getStatusIcon(hasItems bool) string {
	if hasItems {
		return terminal.ColorText("✓", terminal.ColorGreen)
	}
	return terminal.ColorText("✗", terminal.ColorRed)
}

func init() {
	UninstallCmd.Flags().Bool("force", false, "Skip confirmation prompts (use with caution)")
	UninstallCmd.Flags().Bool("keep-data", false, "Keep data directories (only remove packages)")
	UninstallCmd.Flags().Bool("keep-config", false, "Keep configuration files")
	UninstallCmd.Flags().Bool("backup-first", false, "Create backup before uninstalling")
	UninstallCmd.Flags().String("backup-dir", "./mariadb_backup", "Directory for backup files")
}
