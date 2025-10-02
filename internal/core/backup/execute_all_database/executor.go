package execute_all_db

import (
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/parsing"
	"sfDBTools/utils/common/structs"
	"time"

	"github.com/spf13/cobra"
)

func BackupAllDB(cmd *cobra.Command, lg *logger.Logger) (BackupAllDBResult *structs.BackupAllDBResult, err error) {
	opts, err := parsing.ParseBackupAllDBFlags(cmd)
	if err != nil {
		lg.Error("Error parsing flags", logger.Error(err))
		return
	}

	// 1. Inisialisasi Hasil & Timer
	startTime := time.Now()
	lg.Info("Starting all databases backup", logger.Time("start_time", startTime))
	result := &structs.BackupAllDBResult{
		BackupOperationResult: structs.BackupOperationResult{
			Success: false,
		},
		ReplicationMeta: &structs.ReplicationMeta{},
	}

	// 2. Validasi Opsi
	if opts.OutputDir == "" {
		err = fmt.Errorf("output directory is required")
		lg.Error("Invalid options", logger.Error(err))
		return
	}
	lg.Info("Backup options", logger.String("options", fmt.Sprintf("%+v", opts)))
	return result, nil
}
