package mariadb

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// SimpleInstaller handles basic MariaDB installation
type SimpleInstaller struct {
	// Simple installer doesn't need state
}

// InstallMariaDB installs MariaDB with version selection
func InstallMariaDB() error {
	lg, _ := logger.Get()
	lg.Info("Starting MariaDB installation")

	terminal.ClearAndShowHeader("MariaDB Installation")

	// Step 1: Check if MariaDB service already exists
	if err := checkExistingMariaDBService(); err != nil {
		return err
	}

	// Step 2: Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("installation requires root privileges. Please run with sudo")
	}

	// Step 3: Detect OS
	osInfo, err := system.DetectOS()
	if err != nil {
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	// Step 4: Get available versions
	fmt.Println("📦 Fetching available MariaDB versions...")
	versions, err := GetAvailableVersions()
	if err != nil {
		return fmt.Errorf("failed to get available versions: %w", err)
	}

	// Step 5: Let user select version
	selectedVersion, err := selectVersion(versions)
	if err != nil {
		return fmt.Errorf("version selection failed: %w", err)
	}

	// Step 5: Confirm installation
	if !confirmInstallation(selectedVersion) {
		fmt.Println("❌ Installation cancelled by user")
		return nil
	}

	// Step 6: Setup repository using utils
	fmt.Printf("🚀 Installing MariaDB %s...\n", selectedVersion)
	if err := setupRepositoryWithUtils(selectedVersion, osInfo); err != nil {
		return fmt.Errorf("repository setup failed: %w", err)
	}

	// Step 7: Install packages using utils
	if err := installPackagesWithUtils(osInfo); err != nil {
		return fmt.Errorf("package installation failed: %w", err)
	}

	// Step 8: Start and enable service using utils
	fmt.Println("⚙️  Starting MariaDB service...")
	if err := startServiceWithUtils(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Println("✅ MariaDB installation completed successfully!")
	fmt.Printf("📋 Version: %s\n", selectedVersion)
	fmt.Println("🔧 Run 'sudo mysql_secure_installation' to secure your installation")

	return nil
}

// selectVersion allows user to select MariaDB version
func selectVersion(versions []SimpleVersionInfo) (string, error) {
	fmt.Println("\n📋 Available MariaDB versions:")

	stableVersions := make([]SimpleVersionInfo, 0)
	for _, v := range versions {
		if v.Status == "Stable" {
			stableVersions = append(stableVersions, v)
		}
	}

	if len(stableVersions) == 0 {
		return "", fmt.Errorf("no stable versions available")
	}

	// Show stable versions only
	for i, version := range stableVersions {
		supportInfo := ""
		if version.SupportType == "Long Term Support" {
			supportInfo = " (LTS)"
		}
		fmt.Printf("  %d) %s%s\n", i+1, version.Version, supportInfo)
	}

	// Get user choice
	fmt.Printf("\nSelect version (1-%d): ", len(stableVersions))
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(stableVersions) {
		return "", fmt.Errorf("invalid selection")
	}

	selected := stableVersions[choice-1]
	return selected.Version, nil
}

// confirmInstallation asks user to confirm installation
func confirmInstallation(version string) bool {
	fmt.Printf("\n⚠️  This will install MariaDB %s on your system.\n", version)
	fmt.Printf("Continue? (y/N): ")

	var response string
	fmt.Scanf("%s", &response)

	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

// setupRepositoryWithUtils sets up MariaDB repository using utils
func setupRepositoryWithUtils(version string, osInfo *system.OSInfo) error {
	fmt.Println("📥 Setting up MariaDB repository...")

	repoManager := repository.NewManager(osInfo)

	// Clean existing repositories first
	if err := repoManager.Clean(); err != nil {
		lg, _ := logger.Get()
		lg.Warn("Failed to clean existing repositories", logger.Error(err))
	}

	// Setup official repository
	if err := repoManager.SetupOfficial(version); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Update package cache
	if err := repoManager.UpdateCache(); err != nil {
		return fmt.Errorf("failed to update package cache: %w", err)
	}

	return nil
}

// installPackagesWithUtils installs MariaDB packages using utils
func installPackagesWithUtils(osInfo *system.OSInfo) error {
	fmt.Println("📦 Installing MariaDB packages...")

	pkgManager := system.NewPackageManager()

	// Determine packages based on OS
	var packages []string
	switch osInfo.PackageType {
	case "deb":
		packages = []string{"mariadb-server", "mariadb-client"}
	case "rpm":
		packages = []string{"MariaDB-server", "MariaDB-client"}
	default:
		packages = []string{"mariadb-server", "mariadb-client"}
	}

	// Install packages
	if err := pkgManager.Install(packages); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

// startServiceWithUtils starts and enables MariaDB service using utils
func startServiceWithUtils() error {
	svcManager := system.NewServiceManager()
	serviceName := "mariadb"

	// Start service
	if err := svcManager.Start(serviceName); err != nil {
		return fmt.Errorf("failed to start mariadb service: %w", err)
	}

	// Enable service
	if err := svcManager.Enable(serviceName); err != nil {
		return fmt.Errorf("failed to enable mariadb service: %w", err)
	}

	return nil
}

// checkExistingMariaDBService checks if MariaDB service already exists and is active
func checkExistingMariaDBService() error {
	lg, _ := logger.Get()
	svcManager := system.NewServiceManager()

	// Services to check (common MariaDB/MySQL service names)
	services := []string{"mariadb", "mysql", "mysqld"}

	for _, serviceName := range services {
		lg.Debug("Checking service status", logger.String("service", serviceName))

		// Check if service is active
		if svcManager.IsActive(serviceName) {
			fmt.Printf("🚨 MariaDB/MySQL service '%s' is already running!\n", serviceName)
			fmt.Println("\nTo avoid conflicts, please:")
			fmt.Printf("  1. Stop the service: sudo systemctl stop %s\n", serviceName)
			fmt.Printf("  2. Remove existing installation: sudo ./sfdbtools mariadb remove\n")
			fmt.Println("  3. Then run this installation again")
			fmt.Println("\nOr check the running service with:")
			fmt.Printf("  sudo systemctl status %s\n", serviceName)

			lg.Info("Installation blocked - existing MariaDB service found",
				logger.String("active_service", serviceName))
			return fmt.Errorf("MariaDB/MySQL service '%s' is already active", serviceName)
		}

		// Check if service exists (enabled or disabled) by checking if it's known to systemd
		if svcManager.IsEnabled(serviceName) {
			fmt.Printf("⚠️  MariaDB/MySQL service '%s' exists (enabled but not running)\n", serviceName)
			fmt.Println("\nRecommendation:")
			fmt.Printf("  Remove existing installation: sudo ./sfdbtools mariadb remove\n")
			fmt.Println("  Then run this installation again for a clean setup")

			lg.Info("Installation blocked - existing MariaDB service found (enabled)",
				logger.String("enabled_service", serviceName))
			return fmt.Errorf("MariaDB/MySQL service '%s' already exists and is enabled", serviceName)
		}

		// Additional check using systemctl status to see if service unit file exists
		processMgr := system.NewProcessManager()
		if _, err := processMgr.ExecuteWithOutput("systemctl", []string{"status", serviceName}); err == nil {
			// If systemctl status succeeds, service unit exists (even if disabled)
			fmt.Printf("⚠️  MariaDB/MySQL service '%s' exists (disabled)\n", serviceName)
			fmt.Println("\nRecommendation:")
			fmt.Printf("  Remove existing installation: sudo ./sfdbtools mariadb remove\n")
			fmt.Println("  Then run this installation again for a clean setup")

			lg.Info("Installation blocked - existing MariaDB service found (disabled)",
				logger.String("disabled_service", serviceName))
			return fmt.Errorf("MariaDB/MySQL service '%s' already exists", serviceName)
		}
	}

	lg.Debug("No existing MariaDB services found - proceeding with installation")
	return nil
}
