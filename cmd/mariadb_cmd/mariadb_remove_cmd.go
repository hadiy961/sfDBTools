package mariadb_cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove MariaDB server",
	Long: `Remove MariaDB server including packages, services, and optionally data directories.

This command will:
1. Detect existing MariaDB installation
2. Optionally backup data before removal
3. Stop and disable MariaDB services
4. Remove MariaDB packages
5. Optionally remove data directories
6. Remove configuration and log files
7. Optionally remove MariaDB repositories
8. Clean up residual files

Examples:
  # Remove MariaDB but keep data
  sfdbtools mariadb remove

  # Remove MariaDB and all data (with backup)
  sfdbtools mariadb remove --remove-data --backup-data

  # Remove everything including repositories (auto-confirm)
  sfdbtools mariadb remove --remove-data --remove-repositories --auto-confirm

  # Remove with custom backup location
  sfdbtools mariadb remove --backup-data --backup-path /path/to/backup

Safety Features:
- Data is backed up by default before removal
- Confirmation is required unless --auto-confirm is used
- Data directories are preserved by default`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			terminal.PrintError("Failed to initialize logger")
			os.Exit(1)
		}

		lg.Info("MariaDB remove command started")

		// Get flags
		removeData, _ := cmd.Flags().GetBool("remove-data")
		backupData, _ := cmd.Flags().GetBool("backup-data")
		backupPath, _ := cmd.Flags().GetString("backup-path")
		removeRepositories, _ := cmd.Flags().GetBool("remove-repositories")
		autoConfirm, _ := cmd.Flags().GetBool("auto-confirm")

		// Check if user provided any flags - if not, use interactive mode
		flagsProvided := cmd.Flags().Changed("remove-data") ||
			cmd.Flags().Changed("backup-data") ||
			cmd.Flags().Changed("backup-path") ||
			cmd.Flags().Changed("remove-repositories") ||
			cmd.Flags().Changed("auto-confirm")

		// If no flags provided, use interactive mode
		if !flagsProvided {
			terminal.PrintInfo("MariaDB Removal Configuration")
			terminal.PrintInfo("Please choose your removal options:")

			// Ask about data removal
			removeData = promptYesNo("Remove data directories? (This will permanently delete all databases)", false)

			// Ask about backup (if not removing data, backup is recommended)
			if !removeData {
				backupData = promptYesNo("Create backup of data before removal?", true)
			} else {
				backupData = promptYesNo("Create backup of data before removal? (Recommended)", true)
			}

			// Ask about backup path if backup is enabled
			if backupData {
				backupPath = promptString("Enter custom backup path (leave empty for default ~/mariadb_backups)", "")
			}

			// Ask about repositories
			removeRepositories = promptYesNo("Remove MariaDB repositories?", false)

			terminal.PrintInfo("\nYour configuration:")
			terminal.PrintInfo(fmt.Sprintf("  Remove data: %v", removeData))
			terminal.PrintInfo(fmt.Sprintf("  Backup data: %v", backupData))
			if backupData && backupPath != "" {
				terminal.PrintInfo(fmt.Sprintf("  Backup path: %s", backupPath))
			}
			terminal.PrintInfo(fmt.Sprintf("  Remove repositories: %v", removeRepositories))
		}

		// Create removal configuration
		config := &remove.RemovalConfig{
			RemoveData:         removeData,
			BackupData:         backupData,
			BackupPath:         backupPath,
			RemoveRepositories: removeRepositories,
			AutoConfirm:        autoConfirm,
			// Don't hardcode directories - let detection service find actual paths
			DataDirectory:   "",
			ConfigDirectory: "",
			LogDirectory:    "",
		}

		// Create and run removal runner
		runner := remove.NewRemovalRunner(config)
		if err := runner.Run(); err != nil {
			terminal.PrintError("MariaDB removal failed: " + err.Error())
			lg.Error("MariaDB removal failed", logger.Error(err))
			os.Exit(1)
		}

		lg.Info("MariaDB removal completed successfully")
	},
	Annotations: map[string]string{
		"command":  "remove",
		"category": "mariadb",
	},
}

func init() {
	// Add flags for removal configuration
	RemoveCmd.Flags().Bool("remove-data", false, "Remove data directories (WARNING: This will permanently delete all databases)")
	RemoveCmd.Flags().Bool("backup-data", true, "Create backup of data before removal")
	RemoveCmd.Flags().String("backup-path", "", "Custom path for data backup (default: ~/mariadb_backups)")
	RemoveCmd.Flags().Bool("remove-repositories", false, "Remove MariaDB repositories")
	RemoveCmd.Flags().BoolP("auto-confirm", "y", false, "Automatically confirm all prompts")
}

// promptYesNo prompts user for yes/no input with default value
func promptYesNo(question string, defaultValue bool) bool {
	if defaultValue {
		fmt.Printf("%s (Y/n): ", question)
	} else {
		fmt.Printf("%s (y/N): ", question)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response == "" {
		return defaultValue
	}

	return response == "y" || response == "yes"
}

// promptString prompts user for string input with default value
func promptString(question, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.TrimSpace(scanner.Text())

	if response == "" {
		return defaultValue
	}

	return response
}
