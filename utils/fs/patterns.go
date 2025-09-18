package fs

import (
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
)

// PatternMatchingOperations provides file and directory pattern matching operations
type PatternMatchingOperations interface {
	// File pattern matching
	IsLogFile(path string) bool
	IsConfigFile(path string) bool
	IsDatabaseFile(path string) bool
	IsBackupFile(path string) bool
	IsTemporaryFile(path string) bool

	// Directory pattern matching
	IsDataDirectory(path, sourceRoot string) bool
	IsSystemDirectory(path string) bool
	IsLogDirectory(path string) bool

	// Custom pattern matching
	MatchesPattern(filename string, patterns []string) bool
	MatchesExtension(filename string, extensions []string) bool
	MatchesPrefix(filename string, prefixes []string) bool
	MatchesSuffix(filename string, suffixes []string) bool

	// Batch pattern operations
	FilterFilesByPattern(files []string, pattern string) []string
	GroupFilesByType(files []string) map[string][]string
	FindFilesByPattern(dir, pattern string, recursive bool) ([]string, error)
}

type patternMatchingOperations struct {
	logger *logger.Logger
}

func newPatternMatchingOperations(logger *logger.Logger) PatternMatchingOperations {
	return &patternMatchingOperations{
		logger: logger,
	}
}

// IsLogFile determines if a file is a log-related file
func (p *patternMatchingOperations) IsLogFile(path string) bool {
	fileName := filepath.Base(path)

	// Common log file patterns
	logExtensions := []string{".log", ".err", ".pid", ".out"}
	logPrefixes := []string{"mysql-bin.", "mysql-relay-bin.", "slow", "error", "general", "access", "audit"}
	logNames := []string{
		"mysql.log", "mysqld.log", "error.log", "slow.log",
		"general.log", "relay.log", "mysqld.pid", "access.log",
		"audit.log", "binary.log", "update.log",
	}

	// Check extensions
	for _, ext := range logExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range logPrefixes {
		if strings.HasPrefix(strings.ToLower(fileName), prefix) {
			return true
		}
	}

	// Check exact names
	for _, name := range logNames {
		if strings.ToLower(fileName) == strings.ToLower(name) {
			return true
		}
	}

	return false
}

// IsConfigFile determines if a file is a configuration file
func (p *patternMatchingOperations) IsConfigFile(path string) bool {
	fileName := filepath.Base(path)

	configExtensions := []string{".cnf", ".conf", ".cfg", ".ini", ".config", ".yaml", ".yml", ".json", ".toml"}
	configPrefixes := []string{"my.", "mysql.", "mariadb.", "config.", "settings."}
	configNames := []string{
		"my.cnf", "mysql.cnf", "mariadb.cnf", "server.cnf",
		"client.cnf", "mysqld.cnf", "mysql.conf", "config",
	}

	// Check extensions
	for _, ext := range configExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range configPrefixes {
		if strings.HasPrefix(strings.ToLower(fileName), prefix) {
			return true
		}
	}

	// Check exact names
	for _, name := range configNames {
		if strings.ToLower(fileName) == strings.ToLower(name) {
			return true
		}
	}

	return false
}

// IsDatabaseFile determines if a file is a database data file
func (p *patternMatchingOperations) IsDatabaseFile(path string) bool {
	fileName := filepath.Base(path)

	dbExtensions := []string{
		".frm", ".ibd", ".MYD", ".MYI", ".opt", ".ARZ", ".ARM",
		".CSM", ".CSV", ".db", ".sqlite", ".sqlite3",
	}

	for _, ext := range dbExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	// Special database files
	dbFiles := []string{"ibdata1", "ib_logfile0", "ib_logfile1", "auto.cnf"}
	for _, dbFile := range dbFiles {
		if strings.ToLower(fileName) == strings.ToLower(dbFile) {
			return true
		}
	}

	return false
}

// IsBackupFile determines if a file is a backup file
func (p *patternMatchingOperations) IsBackupFile(path string) bool {
	fileName := filepath.Base(path)

	backupExtensions := []string{".bak", ".backup", ".dump", ".sql", ".gz", ".tar", ".zip", ".7z"}
	backupPrefixes := []string{"backup", "dump", "export", "mysqldump"}
	backupSuffixes := []string{"backup", "dump", "export", "old", "orig"}

	// Check extensions
	for _, ext := range backupExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range backupPrefixes {
		if strings.HasPrefix(strings.ToLower(fileName), prefix) {
			return true
		}
	}

	// Check suffixes (before extension)
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	for _, suffix := range backupSuffixes {
		if strings.HasSuffix(strings.ToLower(nameWithoutExt), suffix) {
			return true
		}
	}

	return false
}

// IsTemporaryFile determines if a file is a temporary file
func (p *patternMatchingOperations) IsTemporaryFile(path string) bool {
	fileName := filepath.Base(path)

	tempExtensions := []string{".tmp", ".temp", ".swp", ".~", ".bak"}
	tempPrefixes := []string{"tmp", "temp", ".", "#"}
	tempSuffixes := []string{"~", ".tmp", ".temp"}

	// Check extensions
	for _, ext := range tempExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range tempPrefixes {
		if strings.HasPrefix(fileName, prefix) {
			return true
		}
	}

	// Check suffixes
	for _, suffix := range tempSuffixes {
		if strings.HasSuffix(strings.ToLower(fileName), suffix) {
			return true
		}
	}

	return false
}

// IsDataDirectory determines if a directory contains database data
func (p *patternMatchingOperations) IsDataDirectory(path, sourceRoot string) bool {
	rel, err := filepath.Rel(sourceRoot, path)
	if err != nil {
		return false
	}

	// Skip database directories and system schemas
	dirName := filepath.Base(path)
	skipDirs := []string{"mysql", "performance_schema", "information_schema", "sys", "test"}

	for _, skipDir := range skipDirs {
		if strings.ToLower(dirName) == strings.ToLower(skipDir) {
			return true
		}
	}

	// Skip any directory that contains database files (but allow if it's the root directory)
	if rel != "." {
		entries, err := os.ReadDir(path)
		if err != nil {
			return false
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if p.IsDatabaseFile(entry.Name()) {
				return true
			}
		}
	}

	return false
}

// IsSystemDirectory determines if a directory is a system directory
func (p *patternMatchingOperations) IsSystemDirectory(path string) bool {
	dirName := strings.ToLower(filepath.Base(path))

	systemDirs := []string{
		"mysql", "performance_schema", "information_schema", "sys",
		"proc", "dev", "tmp", "var", "etc", "bin", "sbin", "usr",
		"lib", "lib64", "boot", "root", "home",
	}

	for _, sysDir := range systemDirs {
		if dirName == sysDir {
			return true
		}
	}

	return false
}

// IsLogDirectory determines if a directory is used for logs
func (p *patternMatchingOperations) IsLogDirectory(path string) bool {
	dirName := strings.ToLower(filepath.Base(path))

	logDirs := []string{"logs", "log", "var", "tmp", "spool"}

	for _, logDir := range logDirs {
		if dirName == logDir {
			return true
		}
	}

	return false
}

// MatchesPattern checks if a filename matches any of the given patterns
func (p *patternMatchingOperations) MatchesPattern(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// MatchesExtension checks if a filename has any of the given extensions
func (p *patternMatchingOperations) MatchesExtension(filename string, extensions []string) bool {
	return p.MatchesSuffix(filename, extensions)
}

// MatchesPrefix checks if a filename starts with any of the given prefixes
func (p *patternMatchingOperations) MatchesPrefix(filename string, prefixes []string) bool {
	lowerName := strings.ToLower(filename)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowerName, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// MatchesSuffix checks if a filename ends with any of the given suffixes
func (p *patternMatchingOperations) MatchesSuffix(filename string, suffixes []string) bool {
	lowerName := strings.ToLower(filename)
	for _, suffix := range suffixes {
		if strings.HasSuffix(lowerName, strings.ToLower(suffix)) {
			return true
		}
	}
	return false
}

// FilterFilesByPattern filters a list of files by a pattern
func (p *patternMatchingOperations) FilterFilesByPattern(files []string, pattern string) []string {
	var filtered []string
	for _, file := range files {
		matched, err := filepath.Match(pattern, filepath.Base(file))
		if err == nil && matched {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// GroupFilesByType groups files by their type (log, config, database, backup, temporary, other)
func (p *patternMatchingOperations) GroupFilesByType(files []string) map[string][]string {
	groups := map[string][]string{
		"log":       {},
		"config":    {},
		"database":  {},
		"backup":    {},
		"temporary": {},
		"other":     {},
	}

	for _, file := range files {
		switch {
		case p.IsLogFile(file):
			groups["log"] = append(groups["log"], file)
		case p.IsConfigFile(file):
			groups["config"] = append(groups["config"], file)
		case p.IsDatabaseFile(file):
			groups["database"] = append(groups["database"], file)
		case p.IsBackupFile(file):
			groups["backup"] = append(groups["backup"], file)
		case p.IsTemporaryFile(file):
			groups["temporary"] = append(groups["temporary"], file)
		default:
			groups["other"] = append(groups["other"], file)
		}
	}

	return groups
}

// FindFilesByPattern finds files in a directory matching a pattern
func (p *patternMatchingOperations) FindFilesByPattern(dir, pattern string, recursive bool) ([]string, error) {
	var matches []string

	if recursive {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				matched, matchErr := filepath.Match(pattern, filepath.Base(path))
				if matchErr == nil && matched {
					matches = append(matches, path)
				}
			}
			return nil
		})
		return matches, err
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				matched, matchErr := filepath.Match(pattern, entry.Name())
				if matchErr == nil && matched {
					matches = append(matches, filepath.Join(dir, entry.Name()))
				}
			}
		}
		return matches, nil
	}
}
