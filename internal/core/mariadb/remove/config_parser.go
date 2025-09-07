package remove

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ConfigParser handles MariaDB configuration file operations
type ConfigParser struct{}

// CustomDirectory represents a custom directory found in MariaDB configuration
type CustomDirectory struct {
	Path          string
	ConfigFile    string
	DirectiveType string // datadir, innodb_data_home_dir, etc.
}

// NewConfigParser creates a new configuration parser
func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// FindCustomConfigFiles finds all MariaDB configuration files in standard and custom locations
func (cp *ConfigParser) FindCustomConfigFiles() []string {
	var results []string
	lg, _ := logger.Get()

	// MariaDB modern configuration file locations (in priority order)
	standardLocations := []string{
		// Global config files
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/etc/mysql/mariadb.cnf",
		"/etc/mariadb/my.cnf",

		// Directory-based configs
		"/etc/my.cnf.d/",
		"/etc/mysql/my.cnf.d/",
		"/etc/mysql/conf.d/",
		"/etc/mysql/mariadb.conf.d/",
		"/etc/mariadb/conf.d/",
		"/etc/mariadb/mariadb.conf.d/",

		// Alternative locations
		"/usr/local/etc/my.cnf",
		"/usr/local/etc/mysql/my.cnf",
		"/usr/local/mysql/etc/my.cnf",
		"/opt/mysql/my.cnf",
		"/opt/mariadb/my.cnf",

		// User-specific configs
		"~/.my.cnf",
	}

	// Check standard locations
	for _, location := range standardLocations {
		if strings.HasSuffix(location, "/") {
			// Directory - check for .cnf files inside
			cp.addConfigFilesFromDir(location, &results)
		} else {
			// Single file
			if cp.exists(location) {
				results = append(results, location)
			}
		}
	}

	// Use MariaDB/MySQL to detect active configuration files
	activeConfigs := cp.getActiveConfigFiles()
	results = append(results, activeConfigs...)

	// Scan for additional my.cnf files in common directories
	searchDirs := []string{"/etc", "/usr", "/opt", "/var"}
	for _, dir := range searchDirs {
		cp.scanForConfigFiles(dir, &results)
	}

	// Deduplicate and validate
	results = cp.deduplicateSlice(results)

	terminal.PrintInfo(fmt.Sprintf("Found %d configuration files", len(results)))
	for _, config := range results {
		lg.Debug("Found config file", logger.String("path", config))
	}

	return results
}

// addConfigFilesFromDir adds all .cnf files from a directory
func (cp *ConfigParser) addConfigFilesFromDir(dirPath string, results *[]string) {
	if !cp.exists(dirPath) {
		return
	}

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".cnf") {
			*results = append(*results, path)
		}
		return nil
	})
}

// getActiveConfigFiles uses MariaDB/MySQL to get currently active config files
func (cp *ConfigParser) getActiveConfigFiles() []string {
	var results []string

	// Try multiple commands to get active config
	commands := [][]string{
		{"mysql", "--help", "--verbose"},
		{"mysqld", "--help", "--verbose"},
		{"mariadbd", "--help", "--verbose"},
	}

	for _, cmd := range commands {
		if output, err := exec.Command(cmd[0], cmd[1:]...).Output(); err == nil {
			configs := cp.parseConfigFromHelp(string(output))
			results = append(results, configs...)
		}
	}

	return results
}

// parseConfigFromHelp parses configuration file paths from mysqld --help output
func (cp *ConfigParser) parseConfigFromHelp(output string) []string {
	var results []string

	// Look for "Default options are read from the following files in the given order:"
	lines := strings.Split(output, "\n")
	foundConfigSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Default options are read from") {
			foundConfigSection = true
			continue
		}

		if foundConfigSection {
			// Stop at empty line or next section
			if line == "" || strings.Contains(line, "Variables") {
				break
			}

			// Extract file paths
			if strings.HasPrefix(line, "/") && (strings.Contains(line, ".cnf") || strings.Contains(line, "my.cnf")) {
				// Clean up the path
				path := strings.Fields(line)[0]
				if cp.exists(path) {
					results = append(results, path)
				}
			}
		}
	}

	return results
}

// scanForConfigFiles recursively scans a directory for .cnf files
func (cp *ConfigParser) scanForConfigFiles(baseDir string, results *[]string) {
	if !cp.exists(baseDir) {
		return
	}

	// Define patterns for config files
	configPatterns := []string{
		"my.cnf",
		"mariadb.cnf",
		"mysql.cnf",
		"*.cnf",
	}

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip certain directories to avoid unnecessary scanning
			dirName := info.Name()
			if strings.HasPrefix(dirName, ".") ||
				dirName == "proc" || dirName == "sys" || dirName == "dev" ||
				dirName == "tmp" || dirName == "run" {
				return filepath.SkipDir
			}
			return nil
		}

		fileName := info.Name()
		for _, pattern := range configPatterns {
			if matched, _ := regexp.MatchString(strings.Replace(pattern, "*", ".*", -1), fileName); matched {
				*results = append(*results, path)
				break
			}
		}
		return nil
	})
}

// ExtractAllCustomDirectories extracts all custom directories from MariaDB configuration files
func (cp *ConfigParser) ExtractAllCustomDirectories(configFiles []string) []CustomDirectory {
	var customDirs []CustomDirectory
	lg, _ := logger.Get()

	// Directory directives to look for in MariaDB configuration
	directivePatterns := map[string]string{
		"datadir":                    "Main data directory",
		"innodb_data_home_dir":       "InnoDB data home directory",
		"innodb_log_group_home_dir":  "InnoDB log group home directory",
		"innodb_undo_directory":      "InnoDB undo log directory",
		"innodb_temp_data_file_path": "InnoDB temporary data file path",
		"log_bin":                    "Binary log directory",
		"log_bin_index":              "Binary log index directory",
		"relay_log":                  "Relay log directory",
		"relay_log_index":            "Relay log index directory",
		"slow_query_log_file":        "Slow query log file",
		"general_log_file":           "General log file",
		"log_error":                  "Error log file",
		"pid_file":                   "PID file",
		"socket":                     "Socket file",
		"secure_file_priv":           "Secure file privileges directory",
		"tmpdir":                     "Temporary directory",
		"plugin_dir":                 "Plugin directory",
		"character_sets_dir":         "Character sets directory",
		"lc_messages_dir":            "Locale messages directory",
		"ssl_ca":                     "SSL CA file",
		"ssl_cert":                   "SSL certificate file",
		"ssl_key":                    "SSL key file",
		"keyring_file_data":          "Keyring file data",
	}

	for _, configFile := range configFiles {
		if !cp.exists(configFile) {
			continue
		}

		lg.Debug("Parsing config file", logger.String("file", configFile))

		for directive, description := range directivePatterns {
			if value := cp.extractDirectiveValue(configFile, directive); value != "" {
				// Convert file paths to directory paths where needed
				dirPath := cp.getDirectoryFromPath(value, directive)

				if dirPath != "" && cp.exists(dirPath) {
					customDir := CustomDirectory{
						Path:          dirPath,
						ConfigFile:    configFile,
						DirectiveType: fmt.Sprintf("%s (%s)", directive, description),
					}
					customDirs = append(customDirs, customDir)
					lg.Info("Found custom directory",
						logger.String("directive", directive),
						logger.String("path", dirPath),
						logger.String("config", configFile))
				}
			}
		}
	}

	return cp.deduplicateCustomDirs(customDirs)
}

// extractDirectiveValue extracts the value of a specific directive from config file
func (cp *ConfigParser) extractDirectiveValue(configPath, directive string) string {
	f, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Track sections
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Skip if not in relevant sections
		if currentSection != "" &&
			currentSection != "mysqld" &&
			currentSection != "mariadb" &&
			currentSection != "mysql" &&
			currentSection != "server" {
			continue
		}

		// Look for the directive (case insensitive)
		lower := strings.ToLower(line)
		directiveLower := strings.ToLower(directive)

		if strings.HasPrefix(lower, directiveLower) {
			// Check if it's followed by = or whitespace
			if len(lower) > len(directiveLower) {
				nextChar := lower[len(directiveLower)]
				if nextChar == '=' || nextChar == ' ' || nextChar == '\t' {
					// Extract value
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						value := strings.TrimSpace(parts[1])
						// Remove surrounding quotes
						value = strings.Trim(value, "\"'")
						// Expand variables if any
						value = cp.expandVariables(value)
						return filepath.Clean(value)
					}
				}
			}
		}
	}
	return ""
}

// getDirectoryFromPath converts file paths to directory paths where appropriate
func (cp *ConfigParser) getDirectoryFromPath(value, directive string) string {
	// For file-based directives, extract the directory
	fileDirectives := []string{
		"log_bin", "log_bin_index", "relay_log", "relay_log_index",
		"slow_query_log_file", "general_log_file", "log_error",
		"pid_file", "socket", "ssl_ca", "ssl_cert", "ssl_key",
		"keyring_file_data",
	}

	for _, fileDir := range fileDirectives {
		if strings.EqualFold(directive, fileDir) {
			return filepath.Dir(value)
		}
	}

	// For innodb_temp_data_file_path, extract directory from file path
	if strings.EqualFold(directive, "innodb_temp_data_file_path") {
		// Format: path:size, extract just the path part
		if strings.Contains(value, ":") {
			value = strings.Split(value, ":")[0]
		}
		return filepath.Dir(value)
	}

	// For directory directives, return as-is
	return value
}

// expandVariables expands common MariaDB variables in paths
func (cp *ConfigParser) expandVariables(value string) string {
	// Common MariaDB variables
	variables := map[string]string{
		"@datadir@": "/var/lib/mysql",
		"@prefix@":  "/usr",
		"@tmpdir@":  "/tmp",
	}

	for variable, replacement := range variables {
		value = strings.ReplaceAll(value, variable, replacement)
	}

	return value
}

// deduplicateCustomDirs removes duplicate custom directories
func (cp *ConfigParser) deduplicateCustomDirs(dirs []CustomDirectory) []CustomDirectory {
	seen := make(map[string]bool)
	var result []CustomDirectory

	for _, dir := range dirs {
		key := fmt.Sprintf("%s|%s", dir.Path, dir.DirectiveType)
		if !seen[key] {
			seen[key] = true
			result = append(result, dir)
		}
	}

	return result
}

// ExtractDataDir extracts datadir from my.cnf configuration file (legacy method)
func (cp *ConfigParser) ExtractDataDir(configPath string) string {
	return cp.extractDirectiveValue(configPath, "datadir")
}

// exists checks if a file or directory exists
func (cp *ConfigParser) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// deduplicateSlice removes duplicate entries from a string slice
func (cp *ConfigParser) deduplicateSlice(slice []string) []string {
	seen := map[string]struct{}{}
	result := []string{}

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
