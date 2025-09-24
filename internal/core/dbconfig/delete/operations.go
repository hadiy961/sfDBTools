package delete

import (
	"sfDBTools/utils/dbconfig"
	"sfDBTools/utils/terminal"
)

// ProcessDelete handles the core delete operation logic
func ProcessDelete(cfg *dbconfig.Config, args []string) error {
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	processor.LogOperation("database configuration deletion", "")
	terminal.Headers("Hapus Konfigurasi DB")
	// Route to appropriate handler based on parameters
	switch {
	case cfg.DeleteAll:
		return processor.processDeleteAll(cfg)
	case cfg.FilePath != "":
		return processor.processDeleteSpecific(cfg.FilePath, cfg.ForceDelete)
	case len(args) > 0:
		return processor.processDeleteMultiple(args, cfg.ForceDelete)
	default:
		return processor.processDeleteWithSelection(cfg.ForceDelete)
	}
}
