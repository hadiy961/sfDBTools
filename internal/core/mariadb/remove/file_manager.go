package remove

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sfDBTools/utils/terminal"
)

// FileManager handles file and directory operations during removal
type FileManager struct{}

// NewFileManager creates a new file manager for removal operations
func NewFileManager() *FileManager {
	return &FileManager{}
}

// RemoveDefaultDirectories removes default MariaDB directories and files
func (fm *FileManager) RemoveDefaultDirectories() {
	terminal.PrintInfo("Cleaning up standard MariaDB directories...")

	// Standard MariaDB/MySQL directories (expanded for modern MariaDB)
	defaultDirs := []string{
		// Data directories
		"/var/lib/mysql",
		"/var/lib/mariadb",

		// Configuration directories
		"/etc/mysql",
		"/etc/mariadb",
		"/etc/my.cnf",
		"/etc/my.cnf.d/",
		"/etc/mysql/my.cnf.d/",
		"/etc/mysql/conf.d/",
		"/etc/mysql/mariadb.conf.d/",
		"/etc/mariadb/conf.d/",
		"/etc/mariadb/mariadb.conf.d/",

		// Service files
		"/usr/lib/systemd/system/mariadb.service",
		"/usr/lib/systemd/system/mysql.service",
		"/etc/systemd/system/mariadb.service",
		"/etc/systemd/system/mysql.service",

		// Log directories
		"/var/log/mysql",
		"/var/log/mariadb",

		// Runtime directories
		"/var/run/mysqld",
		"/var/run/mariadb",
		"/run/mysqld",
		"/run/mariadb",

		// Share directories
		"/usr/share/mysql",
		"/usr/share/mariadb",

		// Binary directories (be careful with these)
		"/usr/bin/mysql*",
		"/usr/bin/maria*",
		"/usr/sbin/mysql*",
		"/usr/sbin/maria*",

		// Environment files
		"/etc/default/mysql",
		"/etc/default/mariadb",
		"/etc/sysconfig/mysql",
		"/etc/sysconfig/mariadb",

		// Additional directories
		"/var/cache/mysql",
		"/var/cache/mariadb",
	}

	for _, p := range defaultDirs {
		if strings.Contains(p, "*") {
			// Handle wildcard patterns
			fm.removeWildcardPath(p)
		} else if fm.exists(p) {
			terminal.PrintInfo(fmt.Sprintf("Removing %s", p))
			_ = os.RemoveAll(p)
		}
	}
}

// removeWildcardPath removes files/directories matching wildcard patterns
func (fm *FileManager) removeWildcardPath(pattern string) {
	dir := filepath.Dir(pattern)
	if !fm.exists(dir) {
		return
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, file := range files {
		if fm.exists(file) {
			terminal.PrintInfo(fmt.Sprintf("Removing %s", file))
			_ = os.RemoveAll(file)
		}
	}
}

// RemoveCustomDirectories removes all custom directories found in configuration files
func (fm *FileManager) RemoveCustomDirectories(configFiles []string, skipConfirm bool) {
	if len(configFiles) == 0 {
		return
	}

	parser := NewConfigParser()
	customDirs := parser.ExtractAllCustomDirectories(configFiles)

	if len(customDirs) == 0 {
		terminal.PrintInfo("No custom directories found in configuration files")
		return
	}

	terminal.PrintWarning(fmt.Sprintf("Found %d custom directories in configuration files:", len(customDirs)))
	for _, dir := range customDirs {
		terminal.PrintInfo(fmt.Sprintf("  • %s", dir.Path))
		terminal.PrintInfo(fmt.Sprintf("    └─ %s (from %s)", dir.DirectiveType, dir.ConfigFile))
	}

	shouldRemove := skipConfirm
	if !skipConfirm {
		terminal.PrintWarning("\n⚠️  WARNING: This will permanently delete all data in these directories!")
		var ans string
		fmt.Print("Remove all discovered custom directories? (y/n): ")
		fmt.Scanln(&ans)
		shouldRemove = (ans == "y" || ans == "Y")
	}

	if shouldRemove {
		for _, dir := range customDirs {
			if fm.exists(dir.Path) {
				terminal.PrintInfo(fmt.Sprintf("Removing custom directory: %s", dir.Path))
				terminal.PrintInfo(fmt.Sprintf("  └─ Source: %s", dir.DirectiveType))
				if err := os.RemoveAll(dir.Path); err != nil {
					terminal.PrintWarning(fmt.Sprintf("Failed to remove %s: %v", dir.Path, err))
				}
			} else {
				terminal.PrintInfo(fmt.Sprintf("Directory does not exist (skipped): %s", dir.Path))
			}
		}
	} else {
		terminal.PrintInfo("Skipping removal of custom directories")
	}
}

// RemoveCustomDataDirs removes custom data directories found in configuration files (legacy method)
func (fm *FileManager) RemoveCustomDataDirs(configFiles []string, skipConfirm bool) {
	// Use the new comprehensive method
	fm.RemoveCustomDirectories(configFiles, skipConfirm)
}

// exists checks if a file or directory exists
func (fm *FileManager) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// RemoveUserAndGroup removes MySQL/MariaDB system user and group
func (fm *FileManager) RemoveUserAndGroup() {
	terminal.PrintInfo("Cleaning up system user and group...")

	// List of users/groups to remove
	users := []string{"mysql", "mariadb"}
	groups := []string{"mysql", "mariadb"}

	// Remove users
	for _, user := range users {
		if fm.userExists(user) {
			terminal.PrintInfo(fmt.Sprintf("Removing user: %s", user))
			if err := fm.executeCommand("userdel", "-r", user); err != nil {
				terminal.PrintWarning(fmt.Sprintf("Failed to remove user %s: %v", user, err))
			}
		}
	}

	// Remove groups (if they still exist)
	for _, group := range groups {
		if fm.groupExists(group) {
			terminal.PrintInfo(fmt.Sprintf("Removing group: %s", group))
			if err := fm.executeCommand("groupdel", group); err != nil {
				terminal.PrintWarning(fmt.Sprintf("Failed to remove group %s: %v", group, err))
			}
		}
	}
}

// userExists checks if a system user exists
func (fm *FileManager) userExists(username string) bool {
	err := fm.executeCommand("id", username)
	return err == nil
}

// groupExists checks if a system group exists
func (fm *FileManager) groupExists(groupname string) bool {
	err := fm.executeCommand("getent", "group", groupname)
	return err == nil
}

// executeCommand executes a system command
func (fm *FileManager) executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// parseDataDirFromConfig extracts datadir from my.cnf configuration file (legacy method)
func (fm *FileManager) parseDataDirFromConfig(configPath string) string {
	// This functionality is moved to a dedicated config parser
	parser := NewConfigParser()
	return parser.ExtractDataDir(configPath)
}
