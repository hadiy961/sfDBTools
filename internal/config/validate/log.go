package validate

import (
	"errors"
	"fmt"
	"sfDBTools/internal/config/model"
)

var validLogLevels = []string{"trace", "debug", "info", "warn", "error", "fatal"}
var validLogFormats = []string{"text", "json"}

func Log(l model.LogConfig) error {
	if !InSlice(l.Level, validLogLevels) {
		return fmt.Errorf("log.level tidak valid: '%s'", l.Level)
	}

	if !InSlice(l.Format, validLogFormats) {
		return fmt.Errorf("log.format tidak valid: '%s'", l.Format)
	}

	if err := IsValidTimezone(l.Timezone); err != nil {
		return fmt.Errorf("timezone tidak valid: %w", err)
	}

	if l.Output.File.Enabled {
		if l.Output.File.Dir == "" {
			return errors.New("log.output.file.dir wajib diisi saat output.file.enabled = true")
		}
		if err := DirExistsAndWritable(l.Output.File.Dir); err != nil {
			return fmt.Errorf("log.output.file.dir tidak valid: %w", err)
		}
		if l.Output.File.Rotation.RetentionDays < 1 {
			return fmt.Errorf("log.output.file.rotation.retention_days harus >= 1, sekarang: %d", l.Output.File.Rotation.RetentionDays)
		}
	}
	return nil
}
