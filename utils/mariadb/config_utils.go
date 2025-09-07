package mariadb

import (
	"bufio"
	"os"
	"strings"
)

// ConfigUtils provides utilities for working with MariaDB configuration files
type ConfigUtils struct {
	fileUtils *FileUtils
}

// NewConfigUtils creates a new configuration utilities instance
func NewConfigUtils() *ConfigUtils {
	return &ConfigUtils{
		fileUtils: NewFileUtils(),
	}
}

// ParseConfigFile reads and parses a MariaDB configuration file
func (cu *ConfigUtils) ParseConfigFile(configPath string) (map[string]string, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Split on '=' and extract key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(strings.ToLower(parts[0]))
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if any
			value = strings.Trim(value, "\"'")
			config[key] = value
		}
	}

	return config, scanner.Err()
}

// ExtractDataDir extracts the datadir value from a configuration file
func (cu *ConfigUtils) ExtractDataDir(configPath string) string {
	config, err := cu.ParseConfigFile(configPath)
	if err != nil {
		return ""
	}

	if datadir, exists := config["datadir"]; exists {
		return cu.fileUtils.CleanPath(datadir)
	}

	return ""
}

// FindConfigFiles finds all my.cnf files in common and custom locations
func (cu *ConfigUtils) FindConfigFiles() []string {
	var results []string

	// Check common locations
	commonLocations := []string{
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/usr/local/etc/my.cnf",
		"/opt/my.cnf",
	}

	for _, location := range commonLocations {
		if cu.fileUtils.Exists(location) {
			results = append(results, location)
		}
	}

	// Find additional config files in the filesystem
	customFiles := cu.fileUtils.FindFilesWithName("/", "my.cnf")
	results = append(results, customFiles...)

	// Remove duplicates
	return cu.fileUtils.DeduplicateStringSlice(results)
}

// GetStandardDirectories returns standard MariaDB directories that should be removed
func (cu *ConfigUtils) GetStandardDirectories() []string {
	return []string{
		"/var/lib/mysql",
		"/etc/mysql",
		"/etc/my.cnf",
		"/etc/mysql/mariadb.conf.d",
		"/usr/lib/systemd/system/mariadb.service",
		"/var/log/mysql",
		"/var/run/mysqld",
	}
}
