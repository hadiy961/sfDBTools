package fs

import (
	"os"
	"path/filepath"

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
	fileName := baseName(path)

	// Check extensions
	for _, ext := range LogExtensions {
		if hasSuffixCI(fileName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range LogPrefixes {
		if hasPrefixCI(fileName, prefix) {
			return true
		}
	}

	// Check exact names
	for _, name := range LogNames {
		if equalsCI(fileName, name) {
			return true
		}
	}

	return false
}

// IsConfigFile determines if a file is a configuration file
func (p *patternMatchingOperations) IsConfigFile(path string) bool {
	fileName := baseName(path)

	// Check extensions
	for _, ext := range ConfigExtensions {
		if hasSuffixCI(fileName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range ConfigPrefixes {
		if hasPrefixCI(fileName, prefix) {
			return true
		}
	}

	// Check exact names
	for _, name := range ConfigNames {
		if equalsCI(fileName, name) {
			return true
		}
	}

	return false
}

// IsDatabaseFile determines if a file is a database data file
func (p *patternMatchingOperations) IsDatabaseFile(path string) bool {
	fileName := baseName(path)

	for _, ext := range DBExtensions {
		if hasSuffixCI(fileName, ext) {
			return true
		}
	}

	// Special database files
	for _, dbFile := range DBFiles {
		if equalsCI(fileName, dbFile) {
			return true
		}
	}

	return false
}

// IsBackupFile determines if a file is a backup file
func (p *patternMatchingOperations) IsBackupFile(path string) bool {
	fileName := baseName(path)

	// Check extensions
	for _, ext := range BackupExtensions {
		if hasSuffixCI(fileName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range BackupPrefixes {
		if hasPrefixCI(fileName, prefix) {
			return true
		}
	}

	// Check suffixes (before extension)
	nameWithoutExt := filepath.Base(fileName)
	nameWithoutExt = nameWithoutExt[:len(nameWithoutExt)-len(filepath.Ext(nameWithoutExt))]
	for _, suffix := range BackupSuffixes {
		if hasSuffixCI(nameWithoutExt, suffix) {
			return true
		}
	}

	return false
}

// IsTemporaryFile determines if a file is a temporary file
func (p *patternMatchingOperations) IsTemporaryFile(path string) bool {
	fileName := baseName(path)

	// Check extensions
	for _, ext := range TempExtensions {
		if hasSuffixCI(fileName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range TempPrefixes {
		if hasPrefixCI(fileName, prefix) {
			return true
		}
	}

	// Check suffixes
	for _, suffix := range TempSuffixes {
		if hasSuffixCI(fileName, suffix) {
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
	dirName := baseName(path)
	for _, skipDir := range SkipDirs {
		if equalsCI(dirName, skipDir) {
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
	dirName := baseName(path)
	for _, sysDir := range SystemDirs {
		if equalsCI(dirName, sysDir) {
			return true
		}
	}
	return false
}

// IsLogDirectory determines if a directory is used for logs
func (p *patternMatchingOperations) IsLogDirectory(path string) bool {
	dirName := baseName(path)
	for _, logDir := range LogDirs {
		if equalsCI(dirName, logDir) {
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
	for _, prefix := range prefixes {
		if hasPrefixCI(filename, prefix) {
			return true
		}
	}
	return false
}

// MatchesSuffix checks if a filename ends with any of the given suffixes
func (p *patternMatchingOperations) MatchesSuffix(filename string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if hasSuffixCI(filename, suffix) {
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
