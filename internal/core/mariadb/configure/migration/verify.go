package migration

import (
	"fmt"
	"path/filepath"

	"os"
	"sfDBTools/internal/logger"
)

func VerifyDataMigration(source, destination string) error {
	lg, _ := logger.Get()
	lg.Info("Verifying data migration integrity")

	criticalFiles := []string{"ibdata1", "ib_logfile0", "mysql"}

	for _, file := range criticalFiles {
		sourcePath := filepath.Join(source, file)
		destPath := filepath.Join(destination, file)

		if _, err := os.Stat(sourcePath); err == nil {
			if _, err := os.Stat(destPath); err != nil {
				return fmt.Errorf("critical file missing after migration: %s", file)
			}
		}
	}

	lg.Info("Data migration verification passed")
	return nil
}
