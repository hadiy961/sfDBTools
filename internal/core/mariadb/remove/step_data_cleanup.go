package remove

import (
	"context"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// DataCleanupStep handles data directory cleanup
type DataCleanupStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *DataCleanupStep) Name() string {
	return "Data Cleanup"
}

// Validate validates the step preconditions
func (s *DataCleanupStep) Validate(state *State) error {
	if state.Installation == nil {
		return fmt.Errorf("installation detection is required before data cleanup")
	}
	return nil
}

// Execute removes data directories
func (s *DataCleanupStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Starting data cleanup")

	installation := state.Installation

	// Skip if not removing data
	if !state.Config.RemoveData {
		terminal.PrintInfo("Skipping data removal (--remove-data not specified)")
		return nil
	}

	// Confirm data removal unless auto-confirm is enabled
	if !state.Config.AutoConfirm {
		terminal.PrintWarning("⚠️  WARNING: This will permanently delete all databases!")
		terminal.PrintInfo(fmt.Sprintf("Data directory: %s", installation.ActualDataDir))
		if installation.DataDirectorySize > 0 {
			terminal.PrintInfo(fmt.Sprintf("Data size: %.2f MB", float64(installation.DataDirectorySize)/(1024*1024)))
		}

		if !terminal.AskYesNo("Are you absolutely sure you want to delete all data?", false) {
			return fmt.Errorf("data removal cancelled by user")
		}
	}

	// Store backup information
	if state.RollbackData == nil {
		state.RollbackData = make(map[string]interface{})
	}

	// Remove data directory
	if installation.ActualDataDir != "" && s.deps.FileSystem.Exists(installation.ActualDataDir) {
		spinner := terminal.NewProgressSpinner(fmt.Sprintf("Removing data directory %s...", installation.ActualDataDir))
		spinner.Start()

		// Use MariaDB data validator for safety
		if err := s.deps.FileSystem.SafeRemove(installation.ActualDataDir, system.MariaDBDataValidator); err != nil {
			spinner.Stop()
			return fmt.Errorf("failed to remove data directory %s: %w", installation.ActualDataDir, err)
		}

		spinner.Stop()
		terminal.PrintSuccess(fmt.Sprintf("Removed data directory: %s", installation.ActualDataDir))
		state.RollbackData["dataDirectoryRemoved"] = installation.ActualDataDir
	}

	// Remove binlog directory if different from data directory
	if installation.ActualBinlogDir != "" &&
		installation.ActualBinlogDir != installation.ActualDataDir &&
		s.deps.FileSystem.Exists(installation.ActualBinlogDir) {

		spinner := terminal.NewProgressSpinner(fmt.Sprintf("Removing binlog directory %s...", installation.ActualBinlogDir))
		spinner.Start()

		if err := s.deps.FileSystem.SafeRemove(installation.ActualBinlogDir); err != nil {
			spinner.Stop()
			lg.Warn("Failed to remove binlog directory", logger.String("path", installation.ActualBinlogDir), logger.Error(err))
		} else {
			spinner.Stop()
			terminal.PrintSuccess(fmt.Sprintf("Removed binlog directory: %s", installation.ActualBinlogDir))
			state.RollbackData["binlogDirectoryRemoved"] = installation.ActualBinlogDir
		}
	}

	// Remove log directory if different from data directory
	if installation.ActualLogDir != "" &&
		installation.ActualLogDir != installation.ActualDataDir &&
		s.deps.FileSystem.Exists(installation.ActualLogDir) {

		spinner := terminal.NewProgressSpinner(fmt.Sprintf("Removing log directory %s...", installation.ActualLogDir))
		spinner.Start()

		if err := s.deps.FileSystem.SafeRemove(installation.ActualLogDir); err != nil {
			spinner.Stop()
			lg.Warn("Failed to remove log directory", logger.String("path", installation.ActualLogDir), logger.Error(err))
		} else {
			spinner.Stop()
			terminal.PrintSuccess(fmt.Sprintf("Removed log directory: %s", installation.ActualLogDir))
			state.RollbackData["logDirectoryRemoved"] = installation.ActualLogDir
		}
	}

	return nil
}

// Rollback attempts to restore data from backup
func (s *DataCleanupStep) Rollback(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Attempting to rollback data cleanup")

	// Check if we have a backup path
	backupPath, exists := state.RollbackData["backupPath"].(string)
	if !exists || backupPath == "" {
		lg.Error("No backup path available for data rollback")
		return fmt.Errorf("cannot rollback data: no backup available")
	}

	// Get removed paths
	dataDir, _ := state.RollbackData["dataDirectoryRemoved"].(string)

	terminal.PrintInfo("Attempting to restore data from backup...")

	// Restore data directory
	if dataDir != "" {
		dataBackupPath := filepath.Join(backupPath, "data")
		if s.deps.FileSystem.Exists(dataBackupPath) {
			spinner := terminal.NewProgressSpinner(fmt.Sprintf("Restoring data directory to %s...", dataDir))
			spinner.Start()

			if err := s.deps.FileSystem.CreateBackup(dataBackupPath, dataDir); err != nil {
				spinner.Stop()
				lg.Error("Failed to restore data directory", logger.String("from", dataBackupPath), logger.String("to", dataDir), logger.Error(err))
			} else {
				spinner.Stop()
				terminal.PrintSuccess(fmt.Sprintf("Restored data directory: %s", dataDir))
			}
		}
	}

	return nil
}
