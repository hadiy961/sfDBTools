package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"sfDBTools/internal/logger"
	fileutils "sfDBTools/utils/fs/file"
)

func validateEncryptionKeyFile(keyFile string) error {
	lg, _ := logger.Get()
	lg.Debug("Validating encryption key file", logger.String("path", keyFile))

	if keyFile == "" {
		return fmt.Errorf("encryption key file path is required when encryption is enabled")
	}

	if !filepath.IsAbs(keyFile) {
		return fmt.Errorf("encryption key file must be absolute path: %s", keyFile)
	}

	// Ensure parent directory exists using file helper (delegates to MkdirAll)
	if err := fileutils.EnsureParentDir(keyFile); err != nil {
		return fmt.Errorf("failed to ensure encryption key directory exists: %w", err)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		lg.Info("Encryption key file does not exist, will be created during configuration")
		testFile := keyFile + ".test"
		if err := fileutils.TestWrite(testFile, 0600); err != nil {
			return fmt.Errorf("cannot create encryption key file at %s: %w", keyFile, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check encryption key file: %w", err)
	} else {
		if _, err := os.ReadFile(keyFile); err != nil {
			return fmt.Errorf("encryption key file is not readable: %w", err)
		}
	}

	return nil
}
