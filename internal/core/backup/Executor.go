package backup

import (
	"context"
	"fmt"
	execute_all_db "sfDBTools/internal/core/backup/execute_all_database"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/structs"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// BackupExecutor mendefinisikan struktur yang menampung dependencies untuk proses backup.
type BackupExecutor struct {
	Cmd    *cobra.Command
	Logger *logger.Logger
}

// NewBackupExecutor membuat instance baru dari BackupExecutor.
func NewBackupExecutor(lg *logger.Logger, Cmd *cobra.Command) *BackupExecutor {
	return &BackupExecutor{
		Logger: lg,
		Cmd:    Cmd,
	}
}

func (e *BackupExecutor) AllDB(ctx context.Context) (*structs.BackupAllDBResult, error) {
	// Placeholder for actual backup logic
	terminal.Headers("Backup All Databases")
	e.Logger.Info("Executing All Databases Backup command", logger.String("command", e.Cmd.Name()))
	result, err := execute_all_db.BackupAllDB(e.Cmd, e.Logger)
	if err != nil {
		e.Logger.Error("All Databases Backup failed", logger.Error(err))
		return nil, err
	}
	e.Logger.Info("All Databases Backup completed successfully.")
	fmt.Println("Backup result:", result) // Asumsi fungsi ini ada
	return result, nil
}
