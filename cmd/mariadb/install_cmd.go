package command_mariadb

import (
	"bufio"
	"fmt"
	"os"
	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	mariadb_utils "sfDBTools/utils/mariadb"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB with custom configuration",
	Long: `Install command provides a comprehensive solution for installing MariaDB with custom configurations.

This command supports:
- Cross-platform support: CentOS, RHEL, AlmaLinux, Rocky Linux, Ubuntu, Debian
- Version management: Support for MariaDB versions 10.6.22 to 12.1.11
- Interactive version selection: Use --interactive flag for guided version selection
- Custom configuration: Uses config/server.cnf template
- Upgrade detection: Automatically detects existing installations
- Custom paths: Supports custom data, log, and binlog directories
- Encryption support: Optional encryption key file configuration

Examples:
  # Install MariaDB with interactive version selection
  sfDBTools mariadb install --interactive

  # Install MariaDB with default settings
  sfDBTools mariadb install --version 10.6.22

  # Install with custom port and directories
  sfDBTools mariadb install --version 11.4.2 --port 3307 --data-dir /var/lib/mariadb/data

  # Force installation without prompts (automation)
  sfDBTools mariadb install --version 11.4.2 --force`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeInstall(cmd); err != nil {
			lg, _ := logger.Get()
			lg.Error("MariaDB install failed", logger.Error(err))
			os.Exit(1)
		}
	},
	Annotations: map[string]string{
		"command":  "install",
		"category": "mariadb",
	},
}

func executeInstall(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Get flags
	version, _ := cmd.Flags().GetString("version")
	port, _ := cmd.Flags().GetInt("port")
	dataDir, _ := cmd.Flags().GetString("data-dir")
	logDir, _ := cmd.Flags().GetString("log-dir")
	binlogDir, _ := cmd.Flags().GetString("binlog-dir")
	keyFile, _ := cmd.Flags().GetString("key-file")
	force, _ := cmd.Flags().GetBool("force")
	customPaths, _ := cmd.Flags().GetBool("custom-paths")
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Detect OS without additional connectivity checks
	osInfo, err := mariadb_utils.DetectOS()
	if err != nil {
		lg.Warn("Failed to detect OS", logger.Error(err))
	}

	// Step 2: Check internet connectivity for installation
	lg.Info("Verifying internet connectivity for MariaDB installation")
	if err := common.RequireInternetForOperation("MariaDB installation"); err != nil {
		return fmt.Errorf("internet connectivity is required for MariaDB installation: %w", err)
	}

	// Interactive version selection if requested
	if interactive && osInfo != nil {
		selectedVersion, err := selectVersionInteractive(osInfo)
		if err != nil {
			return fmt.Errorf("interactive version selection failed: %w", err)
		}
		version = selectedVersion
		lg.Info("Version selected interactively", logger.String("version", version))
	} else {
		// Validate version first for non-interactive mode
		if !mariadb_utils.IsValidVersionWithConnectivityCheck(version, false) {
			lg.Error("Invalid MariaDB version specified", logger.String("version", version))

			// Show supported versions (without additional connectivity check)
			versions := mariadb_utils.GetSupportedVersionsWithConnectivityCheck(false)

			// Safely get first N versions without causing slice bounds error
			maxVersions := 10
			if len(versions.AllVersions) < maxVersions {
				maxVersions = len(versions.AllVersions)
			}

			lg.Info("Supported versions:", logger.Strings("stable", versions.AllVersions[:maxVersions]))

			return fmt.Errorf("unsupported MariaDB version: %s. Use 'sfDBTools mariadb versions' to see all supported versions", version)
		}
	}

	// Show recommendation if needed and OS was detected
	if osInfo == nil {
		lg.Warn("Failed to detect OS", logger.Error(err))
	} else {
		recommended := mariadb_utils.GetRecommendedVersion(osInfo)
		if version != recommended && !interactive {
			lg.Info("Version compatibility info",
				logger.String("selected_version", version),
				logger.String("recommended_version", recommended),
				logger.String("os", osInfo.ID))
		}

		// Check version compatibility with OS (skip if interactive since it was already handled)
		if !interactive {
			if err := mariadb_utils.ValidateVersionForOS(version, osInfo); err != nil {
				if !force {
					lg.Error("Version compatibility issue", logger.Error(err))
					return fmt.Errorf("version compatibility issue: %w. Use --force to override or choose recommended version: %s", err, recommended)
				} else {
					lg.Warn("Version compatibility warning (forced)", logger.Error(err))
				}
			}
		}
	}

	// Check for existing installation
	if !force {
		if serviceInfo, err := mariadb_utils.GetServiceInfo(); err == nil && serviceInfo.Status != "not-found" {
			lg.Warn("Existing MariaDB installation detected",
				logger.String("service", serviceInfo.Name),
				logger.String("status", serviceInfo.Status))

			if !promptOverwrite() {
				lg.Info("Installation cancelled by user")
				fmt.Println("❌ Installation cancelled.")
				return nil
			} else {
				// User chose to continue, set force to true
				force = true
				lg.Info("User confirmed to continue installation", logger.Bool("force", force))
			}
		}
	}

	lg.Info("Starting MariaDB install process",
		logger.String("version", version),
		logger.Int("port", port),
		logger.Bool("force", force))

	// Execute install
	result, err := mariadb.InstallMariaDB(mariadb_utils.InstallOptions{
		Version:     version,
		Port:        port,
		DataDir:     dataDir,
		LogDir:      logDir,
		BinlogDir:   binlogDir,
		KeyFile:     keyFile,
		Force:       force,
		CustomPaths: customPaths,
	})

	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// Display results
	mariadb_utils.DisplayInstallResults(result)

	// Log results
	lg.Info("MariaDB install completed successfully",
		logger.String("duration", result.Duration.String()),
		logger.String("version", result.Version),
		logger.String("os", result.OperatingSystem),
		logger.String("service_status", result.ServiceStatus))

	// Ask for custom configuration
	if !force && promptCustomConfiguration() {
		lg.Info("Starting custom configuration process")

		configResult, err := mariadb.ConfigureMariaDB()
		if err != nil {
			lg.Error("Custom configuration failed", logger.Error(err))
			fmt.Printf("❌ Custom configuration failed: %v\n", err)
			fmt.Println("MariaDB installation completed, but configuration failed.")
			fmt.Println("You can run configuration manually later if needed.")
		} else {
			mariadb.DisplayConfigResult(configResult)
		}
	} else {
		fmt.Println("\n💡 You can customize MariaDB configuration later by running:")
		fmt.Println("   sfDBTools mariadb configure")
	}

	return nil
}

func init() {
	// Use a static default version to avoid network calls during app startup
	defaultVersion := "10.6.23"

	InstallCmd.Flags().String("version", defaultVersion, "MariaDB version to install")
	InstallCmd.Flags().Int("port", 3306, "MariaDB port number")
	InstallCmd.Flags().String("data-dir", "/var/lib/mysql", "MariaDB data directory")
	InstallCmd.Flags().String("log-dir", "/var/lib/mysql", "MariaDB log directory")
	InstallCmd.Flags().String("binlog-dir", "/var/lib/mysqlbinlog", "MariaDB binary log directory")
	InstallCmd.Flags().String("key-file", "", "Path to encryption key file")
	InstallCmd.Flags().Bool("force", false, "Skip confirmation prompts")
	InstallCmd.Flags().Bool("custom-paths", false, "Use custom directory structure")
	InstallCmd.Flags().Bool("interactive", false, "Interactive version selection")
}

// promptOverwrite prompts user for confirmation to overwrite existing installation
func promptOverwrite() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Existing MariaDB installation detected. Do you want to continue? (y/n): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// promptCustomConfiguration prompts user for custom configuration
func promptCustomConfiguration() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n🔧 MariaDB Installation Complete!")
	fmt.Print("Would you like to customize MariaDB configuration now? (Y/n): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return true // Default to yes on error
	}

	response = strings.TrimSpace(strings.ToLower(response))
	// Default is yes, so only return false if explicitly no
	return !(response == "n" || response == "no")
}

// selectVersionInteractive allows user to select MariaDB version interactively
func selectVersionInteractive(osInfo *mariadb_utils.OSInfo) (string, error) {
	// Get versions without additional connectivity check (already verified)
	versions := mariadb_utils.GetSupportedVersionsWithConnectivityCheck(false)

	fmt.Println("\n🔧 MariaDB Version Selection")
	fmt.Println("==============================")
	fmt.Printf("Detected OS: %s %s\n", osInfo.Name, osInfo.Version)

	recommended := mariadb_utils.GetRecommendedVersion(osInfo)
	fmt.Printf("Recommended: %s\n\n", recommended)

	// Create indexed list of all versions
	var allVersionsList []string

	// Add stable versions first
	for _, versionList := range versions.StableVersions {
		allVersionsList = append(allVersionsList, versionList...)
	}

	// Add other versions
	for _, versionList := range versions.OtherVersions {
		allVersionsList = append(allVersionsList, versionList...)
	}

	// Display versions by category
	fmt.Println("✅ Stable Versions (Recommended):")
	idx := 1
	for series, versionList := range versions.StableVersions {
		fmt.Printf("  %s series: ", series)
		for i, version := range versionList {
			marker := ""
			if version == recommended {
				marker = " (recommended)"
			}
			fmt.Printf("[%d] %s%s", idx, version, marker)
			if i < len(versionList)-1 {
				fmt.Print(", ")
			}
			idx++
		}
		fmt.Println()
	}

	fmt.Println("\n🔄 Other Versions:")
	for series, versionList := range versions.OtherVersions {
		fmt.Printf("  %s series: ", series)
		for i, version := range versionList {
			fmt.Printf("[%d] %s", idx, version)
			if i < len(versionList)-1 {
				fmt.Print(", ")
			}
			idx++
		}
		fmt.Println()
	}

	fmt.Printf("\nTotal %d versions available\n\n", len(allVersionsList))

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Enter version number (1-%d), or 'r' for recommended [%s]: ", len(allVersionsList), recommended)

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// If empty, use recommended
		if input == "" || input == "r" || input == "R" {
			fmt.Printf("Selected: %s (recommended)\n", recommended)
			return recommended, nil
		}

		// Try to parse as number
		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= len(allVersionsList) {
				selected := allVersionsList[num-1]
				fmt.Printf("Selected: %s\n", selected)

				// Warn if not recommended for this OS
				if err := mariadb_utils.ValidateVersionForOS(selected, osInfo); err != nil {
					fmt.Printf("⚠️  Warning: %v\n", err)
					fmt.Print("Continue anyway? (y/n): ")

					confirm, err := reader.ReadString('\n')
					if err != nil {
						return "", fmt.Errorf("failed to read confirmation: %w", err)
					}

					confirm = strings.TrimSpace(strings.ToLower(confirm))
					if confirm != "y" && confirm != "yes" {
						continue
					}
				}

				return selected, nil
			}
		}

		fmt.Printf("Invalid selection. Please enter a number between 1 and %d, or 'r' for recommended.\n", len(allVersionsList))
	}
}
