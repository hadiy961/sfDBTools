package remove

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// ValidationStep validates the removal configuration and system state
type ValidationStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *ValidationStep) Name() string {
	return "Safety Validation"
}

// Validate validates the step preconditions
func (s *ValidationStep) Validate(state *State) error {
	if state.Installation == nil {
		return fmt.Errorf("installation detection is required before validation")
	}
	return nil
}

// Execute performs safety validations
func (s *ValidationStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Performing safety validation")

	installation := state.Installation
	config := state.Config

	spinner := terminal.NewProgressSpinner("Validating removal safety...")
	spinner.Start()
	defer spinner.Stop()

	// Validate data directory safety if removal is requested
	if config.RemoveData && installation.ActualDataDir != "" {
		if err := system.MariaDBDataValidator.Validate(installation.ActualDataDir); err != nil {
			return fmt.Errorf("data directory validation failed: %w", err)
		}
		lg.Info("Data directory validation passed", logger.String("path", installation.ActualDataDir))
	}

	// Validate that we're not removing critical system directories
	criticalPaths := []string{"/", "/var", "/etc", "/usr", "/home", "/root"}
	for _, critical := range criticalPaths {
		if installation.ActualDataDir == critical {
			return fmt.Errorf("refusing to remove critical system directory: %s", critical)
		}
	}

	// Check available disk space for backup if needed
	if config.BackupData && installation.DataDirectorySize > 0 {
		lg.Info("Data backup will be required",
			logger.String("dataDir", installation.ActualDataDir),
			logger.Int64("size", installation.DataDirectorySize))
	}

	terminal.PrintSuccess("Safety validation completed")
	return nil
}

// Rollback for validation (no-op)
func (s *ValidationStep) Rollback(ctx context.Context, state *State) error {
	// Validation has no rollback - it's read-only
	return nil
}
