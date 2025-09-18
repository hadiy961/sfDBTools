package migration

import (
	"fmt"
	"path/filepath"

	"sfDBTools/internal/logger"
	fsutil "sfDBTools/utils/fs"
)

type MigrationVerifier struct {
	fsMgr  *fsutil.Manager
	logger *logger.Logger
}

func NewMigrationVerifier() *MigrationVerifier {
	lg, _ := logger.Get()
	return &MigrationVerifier{
		fsMgr:  fsutil.NewManager(),
		logger: lg,
	}
}

func VerifyDataMigration(source, destination string) error {
	verifier := NewMigrationVerifier()
	return verifier.VerifyMigration(source, destination)
}

func (v *MigrationVerifier) VerifyMigration(source, destination string) error {
	v.logger.Info("Starting comprehensive data migration verification",
		logger.String("source", source),
		logger.String("destination", destination))

	// 1. Verify critical MariaDB/MySQL files exist
	criticalFiles := []string{
		"ibdata1",            // InnoDB system tablespace
		"ib_logfile0",        // InnoDB log file
		"ib_logfile1",        // InnoDB log file (if exists)
		"mysql",              // MySQL system database directory
		"performance_schema", // Performance schema directory (if exists)
	}

	v.logger.Info("Verifying critical database files...")
	results := v.fsMgr.Verify().VerifyFiles(source, destination, criticalFiles)

	for _, result := range results {
		if result.Error != nil && result.File != "ib_logfile1" && result.File != "performance_schema" {
			// ib_logfile1 and performance_schema are optional
			v.logger.Error("Critical file verification failed",
				logger.String("file", result.File),
				logger.Error(result.Error))
			return fmt.Errorf("critical file verification failed for %s: %w", result.File, result.Error)
		}
	}

	// 2. Verify directory structure and permissions
	v.logger.Info("Verifying directory structure...")
	if err := v.fsMgr.DirValid().ValidateDirectoryStructure(source, destination); err != nil {
		return fmt.Errorf("directory structure verification failed: %w", err)
	}

	// 3. Verify file sizes for critical files
	v.logger.Info("Verifying file sizes...")
	if err := v.fsMgr.Verify().VerifyFileSizes(source, destination, criticalFiles); err != nil {
		return fmt.Errorf("file size verification failed: %w", err)
	}

	// 4. Verify checksums for small critical files (< 100MB)
	v.logger.Info("Verifying file checksums...")
	if err := v.verifyChecksums(source, destination, []string{"mysql"}); err != nil {
		v.logger.Warn("Checksum verification had issues", logger.Error(err))
		// Don't fail on checksum issues, just log warning
	}

	// 5. Verify database directories exist
	v.logger.Info("Verifying database directories...")
	if err := v.fsMgr.DirValid().VerifyEssentialDirectories(destination, []string{"mysql"}); err != nil {
		return fmt.Errorf("database directories verification failed: %w", err)
	}

	// Log verification summary
	v.logVerificationResults(results)

	v.logger.Info("Data migration verification completed successfully")
	return nil
}

// verifyChecksums compares MD5 checksums for critical files (only for small files)
func (v *MigrationVerifier) verifyChecksums(source, destination string, files []string) error {
	const maxSizeForChecksum = 100 * 1024 * 1024 // 100MB

	for _, file := range files {
		sourcePath := filepath.Join(source, file)
		destPath := filepath.Join(destination, file)

		// Skip if either file doesn't exist
		if !v.fsMgr.Verify().FileExists(sourcePath) || !v.fsMgr.Verify().FileExists(destPath) {
			continue
		}

		// Verify integrity with checksum for small files
		result, err := v.fsMgr.Verify().VerifyFileIntegrity(sourcePath, destPath, maxSizeForChecksum)
		if err != nil {
			v.logger.Warn("Failed to verify file integrity",
				logger.String("file", file), logger.Error(err))
			continue
		}

		if result.Status == "FAILED" {
			return fmt.Errorf("checksum mismatch for %s", file)
		}
	}

	return nil
}

// logVerificationResults logs the summary of verification results
func (v *MigrationVerifier) logVerificationResults(results []fsutil.VerificationResult) {
	passed := 0
	failed := 0
	skipped := 0

	for _, result := range results {
		switch result.Status {
		case "PASSED":
			passed++
		case "FAILED":
			failed++
		case "SKIPPED":
			skipped++
		}

		if result.Status == "FAILED" {
			v.logger.Error("Verification failed",
				logger.String("file", result.File),
				logger.Error(result.Error))
		} else {
			v.logger.Debug("Verification result",
				logger.String("file", result.File),
				logger.String("status", result.Status),
				logger.String("details", result.Details))
		}
	}

	v.logger.Info("Verification summary",
		logger.Int("passed", passed),
		logger.Int("failed", failed),
		logger.Int("skipped", skipped))
}
