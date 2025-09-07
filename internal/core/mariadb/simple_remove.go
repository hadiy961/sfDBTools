package mariadb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// RemoveMariaDB completely removes MariaDB installation
func RemoveMariaDB(skipConfirm bool) error {
	lg, _ := logger.Get()

	terminal.ClearScreen()
	fmt.Println("🗑️  MariaDB Complete Removal")
	fmt.Println("========================================")

	// Root permission check
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required for MariaDB removal")
	}

	// Confirmation unless skipped
	if !skipConfirm {
		fmt.Println("⚠️  WARNING: This will completely remove MariaDB including:")
		fmt.Println("   • MariaDB packages")
		fmt.Println("   • All data directories")
		fmt.Println("   • Configuration files")
		fmt.Println("   • System user and group")
		fmt.Print("\nContinue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("❌ Removal cancelled")
			return nil
		}
	}

	lg.Info("Starting MariaDB complete removal")

	// Step 1: Stop services
	fmt.Println("🛑 Stopping MariaDB services...")
	stopMariaDBServices()

	// Step 2: Remove packages
	fmt.Println("📦 Removing MariaDB packages...")
	if err := removeMariaDBPackages(); err != nil {
		lg.Warn("Package removal failed", logger.Error(err))
	}

	// Step 3: Remove directories
	fmt.Println("📁 Removing data and configuration directories...")
	removeMariaDBDirectories()

	// Step 4: Remove user/group
	fmt.Println("👤 Removing MariaDB user and group...")
	removeMariaDBUser()

	// Step 5: Clean package repos
	fmt.Println("🧹 Cleaning package repositories...")
	cleanMariaDBRepos()

	fmt.Println("✅ MariaDB removal completed!")
	lg.Info("MariaDB removal completed successfully")

	return nil
}

// stopMariaDBServices stops all MariaDB related services
func stopMariaDBServices() {
	serviceMgr := system.NewServiceManager()
	services := []string{"mariadb", "mysql", "mysqld"}

	for _, service := range services {
		// Stop service (ignore errors if service doesn't exist)
		serviceMgr.Stop(service)
		serviceMgr.Disable(service)
	}
} // removeMariaDBPackages removes MariaDB packages using system package manager
func removeMariaDBPackages() error {
	packageMgr := system.NewPackageManager()

	// Common MariaDB/MySQL packages to remove
	packages := []string{
		"mariadb-server", "mariadb-client", "mariadb-common",
		"mysql-server", "mysql-client", "mysql-common",
		"MariaDB-server", "MariaDB-client", "MariaDB-common",
	}

	return packageMgr.Remove(packages)
}

// removeMariaDBDirectories removes all MariaDB related directories
func removeMariaDBDirectories() {
	// Standard directories
	standardDirs := []string{
		"/var/lib/mysql",
		"/var/lib/mariadb",
		"/etc/mysql",
		"/etc/mariadb",
		"/etc/my.cnf.d",
		"/var/log/mysql",
		"/var/log/mariadb",
		"/run/mysqld",
		"/run/mariadb",
		"/tmp/mysql.sock",
		"/usr/share/mysql",
		"/usr/share/mariadb",
		"/etc/systemd/system/mariadb.service.d", // Systemd service override directory
		"/etc/systemd/system/mysql.service.d",   // Systemd service override directory
	}

	// Custom directories from config files
	customDirs := findCustomMariaDBDirectories()
	allDirs := append(standardDirs, customDirs...)

	for _, dir := range allDirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("   Removing: %s\n", dir)
			os.RemoveAll(dir)
		}
	}

	// Also remove any systemd service files
	removeSystemdServiceFiles()
} // findCustomMariaDBDirectories finds custom directories from config files
func findCustomMariaDBDirectories() []string {
	var customDirs []string

	configPaths := []string{
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/etc/mariadb/my.cnf",
		"/usr/etc/my.cnf",
	}

	for _, configPath := range configPaths {
		if dirs := parseConfigForDirectories(configPath); len(dirs) > 0 {
			customDirs = append(customDirs, dirs...)
		}
	}

	return customDirs
}

// parseConfigForDirectories extracts directory paths from config files
func parseConfigForDirectories(configPath string) []string {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var dirs []string
	lines := strings.Split(string(content), "\n")

	directives := []string{"datadir", "innodb_data_home_dir", "innodb_log_group_home_dir",
		"log-bin", "relay-log", "slow_query_log_file", "general_log_file"}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}

		for _, directive := range directives {
			if strings.HasPrefix(line, directive) && strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					path := strings.TrimSpace(parts[1])
					path = strings.Trim(path, `"'`)
					if filepath.IsAbs(path) {
						dirs = append(dirs, filepath.Dir(path))
					}
				}
			}
		}
	}

	return dirs
}

// removeMariaDBUser removes MariaDB system user and group
func removeMariaDBUser() {
	processMgr := system.NewProcessManager()

	// Remove users (ignore errors as they might not exist)
	processMgr.Execute("userdel", []string{"-r", "mysql"})
	processMgr.Execute("userdel", []string{"-r", "mariadb"})

	// Remove groups (ignore errors as they might not exist)
	processMgr.Execute("groupdel", []string{"mysql"})
	processMgr.Execute("groupdel", []string{"mariadb"})
}

// removeSystemdServiceFiles removes any remaining systemd service files
func removeSystemdServiceFiles() {
	serviceFiles := []string{
		"/etc/systemd/system/mariadb.service",
		"/etc/systemd/system/mysql.service",
		"/etc/systemd/system/mysqld.service",
		"/usr/lib/systemd/system/mariadb.service",
		"/usr/lib/systemd/system/mysql.service",
		"/usr/lib/systemd/system/mysqld.service",
	}

	for _, file := range serviceFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("   Removing service file: %s\n", file)
			os.Remove(file)
		}
	}

	// Reload systemd daemon to refresh service list
	processMgr := system.NewProcessManager()
	processMgr.Execute("systemctl", []string{"daemon-reload"})
}

// cleanMariaDBRepos removes MariaDB repositories
func cleanMariaDBRepos() {
	repoFiles := []string{
		"/etc/yum.repos.d/mariadb.repo",
		"/etc/yum.repos.d/MariaDB.repo",
		"/etc/apt/sources.list.d/mariadb.list",
		"/etc/apt/sources.list.d/MariaDB.list",
	}

	for _, file := range repoFiles {
		os.Remove(file)
	}
}
