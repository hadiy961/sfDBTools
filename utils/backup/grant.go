package backup_utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/compression"
	"sfDBTools/utils/database"
)

// GenerateGrantOutputPaths generates output paths for grant backup files
func GenerateGrantOutputPaths(options BackupOptions) (string, string) {
	timestamp := time.Now().Format("2006_01_02_150405")
	dateDir := time.Now().Format("2006_01_02")

	var targetName string
	if options.SystemUsers {
		targetName = "system_users"
	} else if options.DBName != "" {
		targetName = options.DBName + "_grants"
	} else {
		targetName = "all_grants"
	}

	// Create subdirectory for grants
	grantsDir := filepath.Join(options.OutputDir, dateDir, "grants")

	// Generate base filename
	baseFilename := fmt.Sprintf("%s_%s", targetName, timestamp)

	// Add appropriate extension based on compression
	var extension string
	if options.Compress {
		// Validate compression type and get extension
		compressionType, err := compression.ValidateCompressionType(options.Compression)
		if err != nil {
			// Default to gzip if invalid
			compressionType = compression.CompressionGzip
		}
		extension = ".sql" + compression.GetFileExtension(compressionType)
	} else {
		extension = ".sql"
	}

	// Add .enc extension if encryption is enabled
	if options.Encrypt {
		extension = extension + ".enc"
	}

	outputFile := filepath.Join(grantsDir, baseFilename+extension)
	metaFile := filepath.Join(grantsDir, baseFilename+".json")

	return outputFile, metaFile
}

// BackupDatabaseGrants backs up grants for a specific database
func BackupDatabaseGrants(db *sql.DB, options BackupOptions) (*BackupResult, error) {
	lg, _ := logger.Get()
	startTime := time.Now()

	lg.Info("Starting database grants backup",
		logger.String("database", options.DBName),
		logger.String("output_dir", options.OutputDir))

	// Validate database exists
	if err := database.ValidateDatabaseExists(db, options.DBName); err != nil {
		return nil, fmt.Errorf("database validation failed: %w", err)
	}

	// Generate output paths
	outputFile, metaFile := GenerateGrantOutputPaths(options)

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get database grantees
	grantees, err := database.GetDatabaseGrantees(db, options.DBName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database grantees: %w", err)
	}

	if len(grantees) == 0 {
		lg.Info("No users have grants for this database, skipping backup",
			logger.String("database", options.DBName))
		return nil, fmt.Errorf("no users found with grants for database '%s'", options.DBName)
	}

	// Generate grant statements
	var grantStatements []string
	grantStatements = append(grantStatements, fmt.Sprintf("-- Grants for database: %s", options.DBName))
	grantStatements = append(grantStatements, fmt.Sprintf("-- Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	grantStatements = append(grantStatements, "")

	validUsersProcessed := 0
	for _, grantee := range grantees {
		lg.Debug("Processing grantee", logger.String("grantee", grantee))

		// Parse grantee format 'username'@'hostname'
		// Example: 'sfnbc_vimut_admin'@'%'
		atIndex := strings.LastIndex(grantee, "@")
		if atIndex == -1 {
			lg.Warn("Invalid grantee format - no @ found", logger.String("grantee", grantee))
			continue
		}

		// Extract username (remove surrounding quotes)
		userPart := grantee[:atIndex]
		username := strings.Trim(userPart, "'")

		// Extract hostname (remove surrounding quotes)
		hostPart := grantee[atIndex+1:]
		hostname := strings.Trim(hostPart, "'")

		lg.Debug("Parsed grantee",
			logger.String("original", grantee),
			logger.String("username", username),
			logger.String("hostname", hostname))

		// Check if user exists
		if !database.UserExistsInMysql(db, username, hostname, lg) {
			lg.Info("User does not exist, skipping grants backup",
				logger.String("user", username),
				logger.String("host", hostname),
				logger.String("database", options.DBName))
			continue
		}

		// Get user grants filtered by database
		grants, err := database.GetUserGrantsForDatabase(db, username, hostname, options.DBName)
		if err != nil {
			lg.Warn("Failed to get grants for user",
				logger.String("user", username),
				logger.String("host", hostname),
				logger.Error(err))
			continue
		}

		if len(grants) > 0 {
			grantStatements = append(grantStatements, fmt.Sprintf("-- Grants for %s@%s", username, hostname))
			for _, grant := range grants {
				grantStatements = append(grantStatements, grant+";")
			}
			grantStatements = append(grantStatements, "")
			validUsersProcessed++
		} else {
			lg.Info("User has no grants for this database",
				logger.String("user", username),
				logger.String("host", hostname),
				logger.String("database", options.DBName))
		}
	}

	// Check if any valid users were processed
	if validUsersProcessed == 0 {
		lg.Info("No valid users with grants found for database, skipping backup",
			logger.String("database", options.DBName))
		return nil, fmt.Errorf("no valid users with grants found for database '%s'", options.DBName)
	}

	// Write grants to file
	content := strings.Join(grantStatements, "\n")
	result, err := writeGrantsToFile(outputFile, content, options)
	if err != nil {
		return nil, fmt.Errorf("failed to write grants: %w", err)
	}

	// Calculate duration and set metadata
	duration := time.Since(startTime)
	result.Duration = duration
	result.BackupMetaFile = metaFile

	lg.Info("Database grants backup completed",
		logger.String("database", options.DBName),
		logger.String("output_file", outputFile),
		logger.String("duration", duration.String()),
		logger.Int("total_grantees", len(grantees)),
		logger.Int("valid_users_processed", validUsersProcessed))

	return result, nil
}

// BackupSystemUserGrants backs up grants for system users
func BackupSystemUserGrants(db *sql.DB, options BackupOptions) (*BackupResult, error) {
	lg, _ := logger.Get()
	startTime := time.Now()

	lg.Info("Starting system users grants backup",
		logger.String("output_dir", options.OutputDir))

	// Generate output paths
	outputFile, metaFile := GenerateGrantOutputPaths(options)

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get system users
	systemUsers, err := database.GetSystemUsers(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get system users: %w", err)
	}

	if len(systemUsers) == 0 {
		lg.Info("No system users found, skipping backup")
		return nil, fmt.Errorf("no system users found")
	}

	// Generate grant statements
	var grantStatements []string
	grantStatements = append(grantStatements, "-- System Users Grants")
	grantStatements = append(grantStatements, fmt.Sprintf("-- Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	grantStatements = append(grantStatements, "")

	totalGrantsCount := 0
	validSystemUsersProcessed := 0
	for _, userInfo := range systemUsers {
		// Validate if system user exists before processing
		if !database.UserExistsInMysql(db, userInfo.Username, userInfo.Hostname, lg) {
			lg.Info("System user does not exist, skipping grants backup",
				logger.String("user", userInfo.Username),
				logger.String("host", userInfo.Hostname))
			continue
		}

		if len(userInfo.Grants) > 0 {
			grantStatements = append(grantStatements, fmt.Sprintf("-- Grants for %s@%s", userInfo.Username, userInfo.Hostname))
			for _, grant := range userInfo.Grants {
				grantStatements = append(grantStatements, grant+";")
				totalGrantsCount++
			}
			grantStatements = append(grantStatements, "")
			validSystemUsersProcessed++
		} else {
			lg.Info("System user has no grants to backup",
				logger.String("user", userInfo.Username),
				logger.String("host", userInfo.Hostname))
		}
	}

	// Check if any valid system users were processed
	if validSystemUsersProcessed == 0 {
		lg.Info("No valid system users with grants found, skipping backup")
		return nil, fmt.Errorf("no valid system users with grants found")
	}

	// Write grants to file
	content := strings.Join(grantStatements, "\n")
	result, err := writeGrantsToFile(outputFile, content, options)
	if err != nil {
		return nil, fmt.Errorf("failed to write grants: %w", err)
	}

	// Calculate duration and set metadata
	duration := time.Since(startTime)
	result.Duration = duration
	result.BackupMetaFile = metaFile

	lg.Info("System users grants backup completed",
		logger.String("output_file", outputFile),
		logger.String("duration", duration.String()),
		logger.Int("total_system_users", len(systemUsers)),
		logger.Int("valid_system_users_processed", validSystemUsersProcessed),
		logger.Int("total_grants_count", totalGrantsCount))

	return result, nil
}

// writeGrantsToFile writes grant content to file with optional compression and encryption
func writeGrantsToFile(outputFile, content string, options BackupOptions) (*BackupResult, error) {
	lg, _ := logger.Get()

	// Create file
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Use the same BuildWriterChain as other backup operations
	writer, closers, err := BuildWriterChain(file, options, lg)
	if err != nil {
		return nil, fmt.Errorf("failed to build writer chain: %w", err)
	}

	// Write content
	bytesWritten, err := writer.Write([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to write content: %w", err)
	}

	// Close writers in reverse order (inner to outer)
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].Close(); err != nil {
			lg.Warn("Failed to close writer", logger.Error(err))
		}
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		lg.Warn("Failed to get file info", logger.Error(err))
	}

	result := &BackupResult{
		Success:         true,
		OutputFile:      outputFile,
		OutputSize:      fileInfo.Size(),
		CompressionUsed: options.Compression,
		Encrypted:       options.Encrypt,
		IncludedData:    false,
	}

	lg.Debug("Grants written to file",
		logger.String("file", outputFile),
		logger.Int("bytes_written", bytesWritten),
		logger.Int64("file_size", fileInfo.Size()),
		logger.Bool("compressed", options.Compress),
		logger.Bool("encrypted", options.Encrypt))

	return result, nil
}
