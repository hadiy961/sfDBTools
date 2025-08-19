package mariadb

import (
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
)

// DisplayUninstallResults displays the results of MariaDB uninstall
func DisplayUninstallResults(result *UninstallResult) {
	lg, _ := logger.Get()

	lg.Info("Displaying MariaDB uninstall results")

	fmt.Println()
	fmt.Println("📋 MARIADB UNINSTALL RESULTS")
	fmt.Println("=============================")

	// Basic information
	fmt.Printf("Operating System: %s\n", result.OperatingSystem)
	if result.Distribution != "" {
		fmt.Printf("Distribution: %s\n", result.Distribution)
	}

	if result.Success {
		fmt.Printf("Success: ✅ Yes\n")
	} else {
		fmt.Printf("Success: ❌ No\n")
	}

	fmt.Println()

	// Process steps
	fmt.Println("📝 Process Steps:")
	fmt.Println("  ✅ Detecting operating system and distribution")
	fmt.Println("  ✅ Checking MariaDB service status")
	fmt.Println("  ✅ Stopping MariaDB service")
	fmt.Printf("  ✅ Removing MariaDB packages (%d removed)\n", result.PackagesRemoved)
	fmt.Printf("  ✅ Cleaning up data and configuration directories (%d removed)\n", len(result.DirectoriesRemoved))
	fmt.Printf("  ✅ Removing MariaDB repositories (%d removed)\n", len(result.RepositoriesRemoved))
	fmt.Println("  ✅ Verifying MariaDB uninstall")

	fmt.Println()

	// Details
	if result.PackagesRemoved > 0 {
		fmt.Printf("📦 Packages Removed: %d\n", result.PackagesRemoved)
	}

	if len(result.DirectoriesRemoved) > 0 {
		fmt.Printf("📁 Directories/Files Removed: %d\n", len(result.DirectoriesRemoved))
		for _, dir := range result.DirectoriesRemoved {
			fmt.Printf("  - %s\n", dir)
		}
	}

	if len(result.ConfigFilesRemoved) > 0 {
		fmt.Printf("📄 Config Files Removed: %d\n", len(result.ConfigFilesRemoved))
		for _, file := range result.ConfigFilesRemoved {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(result.RepositoriesRemoved) > 0 {
		fmt.Printf("🗂️  Repositories Removed: %d\n", len(result.RepositoriesRemoved))
		for _, repo := range result.RepositoriesRemoved {
			fmt.Printf("  - %s\n", repo)
		}
	}

	fmt.Println()

	// Final status
	fmt.Println("🔍 Final Status:")
	fmt.Printf("  Service Status: %s\n", result.ServiceStatus)
	fmt.Printf("  Duration: %s\n", result.Duration)

	// Backup information
	if result.BackupCreated {
		fmt.Printf("  📁 Backup Created: %s\n", result.BackupLocation)
	}

	// Warnings
	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println("❌ Errors:")
		for _, error := range result.Errors {
			fmt.Printf("  - %s\n", error)
		}
	}

	fmt.Println()

	// Final message
	if result.Success {
		if len(result.Warnings) == 0 {
			fmt.Println("✅ MariaDB has been successfully uninstalled from the system.")
		} else {
			fmt.Println("✅ MariaDB has been uninstalled with some warnings (see above).")
		}
	} else {
		fmt.Println("❌ MariaDB uninstall completed with errors (see above).")
	}

	// Post-uninstall recommendations
	fmt.Println()
	fmt.Println("📋 Post-uninstall verification commands:")
	fmt.Println("  # Check for running processes:")
	fmt.Println("  ps aux | grep -i maria")
	fmt.Println()

	if IsRHELBased(result.OperatingSystem) {
		fmt.Println("  # Check for remaining packages:")
		fmt.Println("  rpm -qa | grep -i maria")
	} else if IsDebianBased(result.OperatingSystem) {
		fmt.Println("  # Check for remaining packages:")
		fmt.Println("  dpkg -l | grep -i maria")
	}

	fmt.Println()
	fmt.Println("  # Check for remaining directories:")
	fmt.Println("  ls -la /var/lib/mysql /etc/mysql 2>/dev/null")
	fmt.Println()
	fmt.Println("  # Check systemd services:")
	fmt.Println("  systemctl list-units | grep -i maria")

	fmt.Println()
}

// ShowUninstallWarning shows warning message before uninstall
func ShowUninstallWarning() {
	fmt.Println()
	fmt.Println("⚠️  WARNING: MARIADB UNINSTALL")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("🔥 This will completely remove MariaDB from your system!")
	fmt.Println()
	fmt.Println("📊 Data that will be PERMANENTLY DELETED:")
	fmt.Println("  • All databases and tables")
	fmt.Println("  • User accounts and permissions")
	fmt.Println("  • Configuration files")
	fmt.Println("  • Log files")
	fmt.Println("  • MariaDB packages")
	fmt.Println()
	fmt.Println("💡 RECOMMENDED: Backup your databases first!")
	fmt.Println()
	fmt.Println("📝 Backup commands:")
	fmt.Println("  # Backup all databases:")
	fmt.Println("  sfDBTools backup all --output-dir ./backup")
	fmt.Println()
	fmt.Println("  # Backup with encryption:")
	fmt.Println("  sfDBTools backup all --output-dir ./backup --encrypt")
	fmt.Println()
	fmt.Println("  # Backup specific database:")
	fmt.Println("  sfDBTools backup single --target_db mydb --output-dir ./backup")
	fmt.Println()
}

// PromptConfirmation prompts user for confirmation
func PromptConfirmation() bool {
	fmt.Print("❓ Are you sure you want to completely uninstall MariaDB? (yes/no): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}

// ShowUninstallSummary shows summary before starting uninstall
func ShowUninstallSummary(options UninstallOptions) {
	fmt.Println()
	fmt.Println("📋 UNINSTALL SUMMARY")
	fmt.Println("====================")
	fmt.Printf("Force mode: %t\n", options.Force)
	fmt.Printf("Keep data: %t\n", options.KeepData)
	fmt.Printf("Keep config: %t\n", options.KeepConfig)
	fmt.Printf("Backup first: %t\n", options.BackupFirst)
	if options.BackupDir != "" {
		fmt.Printf("Backup directory: %s\n", options.BackupDir)
	}
	fmt.Println()
}

// FormatDuration formats duration in a readable format
func FormatDuration(d float64) string {
	if d < 1.0 {
		return fmt.Sprintf("%.2fs", d)
	} else if d < 60.0 {
		return fmt.Sprintf("%.1fs", d)
	} else {
		minutes := int(d / 60)
		seconds := d - float64(minutes*60)
		return fmt.Sprintf("%dm%.1fs", minutes, seconds)
	}
}

// DisplayInstallResults displays the results of MariaDB installation
func DisplayInstallResults(result *InstallResult) {
	lg, _ := logger.Get()

	lg.Info("Displaying MariaDB installation results")

	fmt.Println()
	fmt.Println("📋 MARIADB INSTALLATION RESULTS")
	fmt.Println("================================")

	// Basic information
	fmt.Printf("Operating System: %s\n", result.OperatingSystem)
	if result.Distribution != "" {
		fmt.Printf("Distribution: %s\n", result.Distribution)
	}

	if result.Success {
		fmt.Printf("Success: ✅ Yes\n")
	} else {
		fmt.Printf("Success: ❌ No\n")
	}

	fmt.Println()

	// Installation details
	fmt.Println("📝 Installation Details:")
	fmt.Printf("  Version: %s\n", result.Version)
	fmt.Printf("  Port: %d\n", result.Port)
	fmt.Printf("  Data Directory: %s\n", result.DataDir)
	fmt.Printf("  Log Directory: %s\n", result.LogDir)
	fmt.Printf("  Binary Log Directory: %s\n", result.BinlogDir)

	fmt.Println()

	// Service status
	fmt.Println("🔍 Service Status:")
	fmt.Printf("  Status: %s\n", result.ServiceStatus)
	fmt.Printf("  Duration: %s\n", result.Duration)

	fmt.Println()

	// Final message
	if result.Success {
		fmt.Println("🎉 MariaDB has been successfully installed!")
		fmt.Println()
		fmt.Println("📋 Next steps:")
		fmt.Println("  # Check service status:")
		fmt.Printf("  systemctl status mariadb\n")
		fmt.Println()
		fmt.Println("  # Connect to MariaDB:")
		fmt.Printf("  mysql -u root -p\n")
		fmt.Println()
		fmt.Println("  # Run health check:")
		fmt.Println("  sfDBTools mariadb check")
	} else {
		fmt.Println("❌ MariaDB installation failed!")
		fmt.Println("Please check the logs for more details.")
	}

	fmt.Println()
}
