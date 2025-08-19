package mariadb

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"sfDBTools/internal/logger"
)

// CleanupDirectories removes MariaDB data and configuration directories
func CleanupDirectories(keepData, keepConfig bool) ([]string, error) {
	lg, _ := logger.Get()
	var removedDirs []string

	// Define directories to remove
	dataDirs := []string{
		"/var/lib/mysql",
		"/var/lib/mariadb",
	}

	configDirs := []string{
		"/etc/mysql",
		"/etc/mariadb",
		"/etc/my.cnf.d",
	}

	configFiles := []string{
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/root/.my.cnf",
	}

	logDirs := []string{
		"/var/log/mysql",
		"/var/log/mariadb",
	}

	runtimeDirs := []string{
		"/run/mysqld",
		"/var/run/mysqld",
		"/run/mariadb",
		"/var/run/mariadb",
	}

	// Remove data directories (unless keepData is true)
	if !keepData {
		for _, dir := range dataDirs {
			if removed := removeDirectory(lg, dir); removed {
				removedDirs = append(removedDirs, dir)
			}
		}
	} else {
		lg.Info("Keeping data directories", logger.Bool("keep_data", keepData))
	}

	// Remove configuration directories and files (unless keepConfig is true)
	if !keepConfig {
		for _, dir := range configDirs {
			if removed := removeDirectory(lg, dir); removed {
				removedDirs = append(removedDirs, dir)
			}
		}

		for _, file := range configFiles {
			if removed := removeFile(lg, file); removed {
				removedDirs = append(removedDirs, file)
			}
		}
	} else {
		lg.Info("Keeping configuration directories", logger.Bool("keep_config", keepConfig))
	}

	// Always remove log and runtime directories (they can be recreated)
	for _, dir := range logDirs {
		if removed := removeDirectory(lg, dir); removed {
			removedDirs = append(removedDirs, dir)
		}
	}

	for _, dir := range runtimeDirs {
		if removed := removeDirectory(lg, dir); removed {
			removedDirs = append(removedDirs, dir)
		}
	}

	// Remove systemd service files
	if removed := cleanupSystemdFiles(lg); len(removed) > 0 {
		removedDirs = append(removedDirs, removed...)
	}

	lg.Info("Directory cleanup completed",
		logger.Int("removed_count", len(removedDirs)))

	return removedDirs, nil
}

// cleanupSystemdFiles removes MariaDB/MySQL systemd service files
func cleanupSystemdFiles(lg *logger.Logger) []string {
	var removedFiles []string

	// Common systemd service file locations
	systemdDirs := []string{
		"/etc/systemd/system",
		"/usr/lib/systemd/system",
		"/lib/systemd/system",
	}

	serviceFiles := []string{
		"mariadb.service",
		"mysql.service",
		"mysqld.service",
		"mariadb.service.d",
		"mysql.service.d",
		"mysqld.service.d",
	}

	for _, systemdDir := range systemdDirs {
		for _, serviceFile := range serviceFiles {
			fullPath := fmt.Sprintf("%s/%s", systemdDir, serviceFile)

			// Check if it's a directory or file
			if info, err := os.Stat(fullPath); err == nil {
				if info.IsDir() {
					if removed := removeDirectory(lg, fullPath); removed {
						removedFiles = append(removedFiles, fullPath)
					}
				} else {
					if removed := removeFile(lg, fullPath); removed {
						removedFiles = append(removedFiles, fullPath)
					}
				}
			}
		}
	}

	// Also check for masked service links
	maskedLinks := []string{
		"/etc/systemd/system/mariadb.service",
		"/etc/systemd/system/mysql.service",
		"/etc/systemd/system/mysqld.service",
	}

	for _, link := range maskedLinks {
		if info, err := os.Lstat(link); err == nil && info.Mode()&os.ModeSymlink != 0 {
			// Check if it's a masked service (points to /dev/null)
			if target, err := os.Readlink(link); err == nil && target == "/dev/null" {
				if removed := removeFile(lg, link); removed {
					removedFiles = append(removedFiles, link)
				}
			}
		}
	}

	return removedFiles
}

// removeDirectory removes a directory and logs the result
func removeDirectory(lg *logger.Logger, dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory doesn't exist, skip
		return false
	}

	lg.Debug("Removing directory", logger.String("directory", dir))

	err := os.RemoveAll(dir)
	if err != nil {
		lg.Warn("Failed to remove directory",
			logger.String("directory", dir),
			logger.Error(err))
		return false
	}

	lg.Info("Directory removed", logger.String("directory", dir))
	return true
}

// removeFile removes a file and logs the result
func removeFile(lg *logger.Logger, file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// File doesn't exist, skip
		return false
	}

	lg.Debug("Removing file", logger.String("file", file))

	err := os.Remove(file)
	if err != nil {
		lg.Warn("Failed to remove file",
			logger.String("file", file),
			logger.Error(err))
		return false
	}

	lg.Info("File removed", logger.String("file", file))
	return true
}

// GetRemainingFiles checks for remaining MariaDB files after cleanup
func GetRemainingFiles() ([]string, error) {
	var remainingFiles []string

	// Check for remaining directories and files
	checkPaths := []string{
		"/var/lib/mysql",
		"/var/lib/mariadb",
		"/etc/mysql",
		"/etc/mariadb",
		"/etc/my.cnf",
		"/etc/my.cnf.d",
		"/var/log/mysql",
		"/var/log/mariadb",
		"/run/mysqld",
		"/var/run/mysqld",
	}

	for _, path := range checkPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			remainingFiles = append(remainingFiles, path)
		}
	}

	return remainingFiles, nil
}

// VerifyUninstall verifies that MariaDB has been completely uninstalled
func VerifyUninstall(osInfo *OSInfo) (bool, []string, error) {
	lg, _ := logger.Get()
	var warnings []string

	// Check for running processes
	processes, err := GetRunningProcesses()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to check running processes: %v", err))
	} else if len(processes) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found running MariaDB processes: %v", processes))
	}

	// Check for remaining packages
	var remainingPackages []string
	if IsRHELBased(osInfo.ID) {
		packages, err := getRHELInstalledPackages()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to check remaining packages: %v", err))
		} else {
			remainingPackages = packages
		}
	} else if IsDebianBased(osInfo.ID) {
		packages, err := getDebianInstalledPackages()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to check remaining packages: %v", err))
		} else {
			remainingPackages = packages
		}
	}

	if len(remainingPackages) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found remaining packages: %v", remainingPackages))
	}

	// Check for remaining files
	remainingFiles, err := GetRemainingFiles()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to check remaining files: %v", err))
	} else if len(remainingFiles) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found remaining files: %v", remainingFiles))
	}

	// Check service status
	isRunning, service := IsServiceRunning()
	if isRunning {
		warnings = append(warnings, fmt.Sprintf("Service still running: %s", service))
	}

	success := len(warnings) == 0
	lg.Info("Uninstall verification completed",
		logger.Bool("success", success),
		logger.Int("warnings", len(warnings)))

	return success, warnings, nil
}

// CleanupRepositories removes MariaDB repositories based on the OS
func CleanupRepositories(osInfo *OSInfo) ([]string, error) {
	lg, _ := logger.Get()
	var removedRepos []string

	lg.Info("Starting repository cleanup", logger.String("os", osInfo.ID))

	if IsRHELBased(osInfo.ID) {
		// RHEL/CentOS/Fedora - remove yum/dnf repositories
		repoFiles := []string{
			"/etc/yum.repos.d/mariadb.repo",
			"/etc/yum.repos.d/MariaDB.repo", 
			"/etc/yum.repos.d/mysql.repo",
		}

		for _, repoFile := range repoFiles {
			if removed := removeFile(lg, repoFile); removed {
				removedRepos = append(removedRepos, repoFile)
			}
		}

		// Clean repository cache
		cleanRepoCache(lg, osInfo.ID)

	} else if IsDebianBased(osInfo.ID) {
		// Ubuntu/Debian - remove APT repositories and keys
		
		// Remove repository files
		repoFiles := []string{
			"/etc/apt/sources.list.d/mariadb.list",
			"/etc/apt/sources.list.d/mysql.list",
		}

		for _, repoFile := range repoFiles {
			if removed := removeFile(lg, repoFile); removed {
				removedRepos = append(removedRepos, repoFile)
			}
		}

		// Remove GPG keys
		if err := removeMariaDBKeys(lg); err != nil {
			lg.Warn("Failed to remove some GPG keys", logger.Error(err))
		}

		// Update APT cache
		if err := updateAPTCache(lg); err != nil {
			lg.Warn("Failed to update APT cache", logger.Error(err))
		}
	}

	lg.Info("Repository cleanup completed", 
		logger.Int("removed_repos", len(removedRepos)))

	return removedRepos, nil
}

// cleanRepoCache cleans repository cache for RHEL-based systems
func cleanRepoCache(lg *logger.Logger, osID string) {
	var cmd *exec.Cmd
	
	// Use dnf for Fedora and newer RHEL/CentOS versions
	if osID == "fedora" || strings.Contains(osID, "rhel") || strings.Contains(osID, "centos") {
		if _, err := exec.LookPath("dnf"); err == nil {
			cmd = exec.Command("dnf", "clean", "all")
		} else {
			cmd = exec.Command("yum", "clean", "all")
		}
	} else {
		cmd = exec.Command("yum", "clean", "all")
	}

	lg.Info("Cleaning repository cache", logger.String("command", cmd.String()))
	
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to clean repository cache", logger.Error(err))
	} else {
		lg.Info("Repository cache cleaned successfully")
	}
}

// removeMariaDBKeys removes MariaDB GPG keys from APT keyring
func removeMariaDBKeys(lg *logger.Logger) error {
	lg.Info("Removing MariaDB GPG keys")

	// Common MariaDB key fingerprints and key IDs
	mariadbKeys := []string{
		"0xF1656F24C74CD1D8",      // MariaDB Package Repository Signing Key
		"0x177F4010FE56CA3336300305F1656F24C74CD1D8", // Full fingerprint
		"1BB943DB",                // Short key ID
		"C74CD1D8",                // Short key ID
	}

	for _, keyID := range mariadbKeys {
		// Try with apt-key (older systems)
		cmd := exec.Command("apt-key", "del", keyID)
		if err := cmd.Run(); err != nil {
			lg.Debug("Failed to remove key with apt-key (normal if using newer APT)", 
				logger.String("key", keyID), 
				logger.Error(err))
		} else {
			lg.Info("Removed GPG key with apt-key", logger.String("key", keyID))
		}

		// Try with newer gpg method for trusted.gpg.d
		gpgFiles := []string{
			"/etc/apt/trusted.gpg.d/mariadb.gpg",
			"/etc/apt/trusted.gpg.d/mariadb-keyring.gpg",
			"/usr/share/keyrings/mariadb-keyring.gpg",
		}

		for _, gpgFile := range gpgFiles {
			if removed := removeFile(lg, gpgFile); removed {
				lg.Info("Removed GPG keyring file", logger.String("file", gpgFile))
			}
		}
	}

	return nil
}

// updateAPTCache updates the APT package cache
func updateAPTCache(lg *logger.Logger) error {
	lg.Info("Updating APT cache after repository removal")
	
	cmd := exec.Command("apt-get", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update APT cache: %w", err)
	}

	lg.Info("APT cache updated successfully")
	return nil
}
