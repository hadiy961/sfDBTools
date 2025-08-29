package remove

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// PackageRemovalStep removes MariaDB packages
type PackageRemovalStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *PackageRemovalStep) Name() string {
	return "Remove Packages"
}

// Validate validates the step preconditions
func (s *PackageRemovalStep) Validate(state *State) error {
	if state.Installation == nil {
		return fmt.Errorf("installation detection is required before removing packages")
	}
	return nil
}

// Execute removes MariaDB packages
func (s *PackageRemovalStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Removing MariaDB packages")

	// Get list of installed MariaDB packages
	packages, err := s.deps.PackageManager.GetInstalledPackages()
	if err != nil {
		return fmt.Errorf("failed to get installed packages: %w", err)
	}

	if len(packages) == 0 {
		terminal.PrintInfo("No MariaDB packages found to remove")
		return nil
	}

	// Store packages for rollback (though package rollback is typically not possible)
	if state.RollbackData == nil {
		state.RollbackData = make(map[string]interface{})
	}
	state.RollbackData["removedPackages"] = packages

	terminal.PrintInfo(fmt.Sprintf("Found %d MariaDB packages to remove", len(packages)))
	for _, pkg := range packages {
		terminal.PrintInfo(fmt.Sprintf("  - %s", pkg))
	}

	// Confirm removal unless auto-confirm is enabled
	if !state.Config.AutoConfirm {
		if !terminal.AskYesNo("Do you want to remove these packages?", true) {
			return fmt.Errorf("package removal cancelled by user")
		}
	}

	spinner := terminal.NewProgressSpinner("Removing MariaDB packages...")
	spinner.Start()
	defer spinner.Stop()

	// Remove packages
	if err := s.deps.PackageManager.Remove(packages); err != nil {
		return fmt.Errorf("failed to remove packages: %w", err)
	}

	terminal.PrintSuccess(fmt.Sprintf("Successfully removed %d MariaDB packages", len(packages)))

	return nil
}

// Rollback for package removal (typically not possible, just log)
func (s *PackageRemovalStep) Rollback(ctx context.Context, state *State) error {
	lg, _ := logger.Get()

	removedPackages, ok := state.RollbackData["removedPackages"].([]string)
	if !ok || len(removedPackages) == 0 {
		return nil
	}

	lg.Warn("Package removal rollback requested but packages cannot be automatically reinstalled",
		logger.Strings("packages", removedPackages))

	terminal.PrintWarning("Note: Removed packages cannot be automatically reinstalled during rollback")
	terminal.PrintInfo("To reinstall MariaDB, use: sfdbtools mariadb install")

	return nil
}
