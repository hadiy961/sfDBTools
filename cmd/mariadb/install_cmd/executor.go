package installcmd

import (
	"fmt"
	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"

	"github.com/spf13/cobra"
)

// executeInstall orchestrates the installation flow. This was extracted from the
// original long function for clarity.
// Execute runs the install command flow.
func Execute(cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	// Get flags
	version, _ := cmd.Flags().GetString("version")
	force, _ := cmd.Flags().GetBool("force")

	// Confirm with user immediately after parsing flags (unless forced)
	if !force {
		if !PromptConfirmInstall(version) {
			lg.Info("Installation cancelled by user before any checks")
			terminal.PrintError("Installation cancelled.")
			return nil
		}
	}

	// Detect OS without additional connectivity checks (spinner)
	sp := terminal.NewProgressSpinner("Detecting operating system...")
	sp.Start()
	osInfo, detErr := mariadb_utils.DetectOS()
	sp.Stop()
	if detErr != nil {
		lg.Warn("Failed to detect OS", logger.Error(detErr))
	}

	// Step 2: Check internet connectivity for installation (spinner)
	lg.Info("Verifying internet connectivity for MariaDB installation")
	sp = terminal.NewProgressSpinner("Checking internet connectivity...")
	sp.Start()
	if err := common.RequireInternetForOperation("MariaDB installation"); err != nil {
		sp.Stop()
		return fmt.Errorf("internet connectivity is required for MariaDB installation: %w", err)
	}
	sp.Stop()

	// Validate version (interactive removed)
	{
		// Validate version first for non-interactive mode
		if !mariadb_utils.IsValidVersionWithConnectivityCheck(version, false) {
			lg.Error("Invalid MariaDB version specified", logger.String("version", version))

			// Show supported versions (without additional connectivity check)
			versions := mariadb_utils.GetSupportedVersionsWithConnectivityCheck(false)

			// Safely get first N versions without causing slice bounds error
			maxVersions := 10
			if len(versions.AllVersions) < maxVersions {
				maxVersions = len(versions.AllVersions)
			}

			lg.Info("Supported versions:", logger.Strings("stable", versions.AllVersions[:maxVersions]))

			return fmt.Errorf("unsupported MariaDB version: %s. Use 'sfDBTools mariadb versions' to see all supported versions", version)
		}
	}

	// Show recommendation if needed and OS was detected
	if osInfo == nil {
		lg.Warn("Failed to detect OS", logger.Error(detErr))
	} else {
		recommended := mariadb_utils.GetRecommendedVersion(osInfo)
		if version != recommended {
			lg.Info("Version compatibility info",
				logger.String("selected_version", version),
				logger.String("recommended_version", recommended),
				logger.String("os", osInfo.ID))
		}

		// Check version compatibility with OS
		{
			if err := mariadb_utils.ValidateVersionForOS(version, osInfo); err != nil {
				if !force {
					lg.Error("Version compatibility issue", logger.Error(err))
					return fmt.Errorf("version compatibility issue: %w. Use --force to override or choose recommended version: %s", err, recommended)
				} else {
					lg.Warn("Version compatibility warning (forced)", logger.Error(err))
				}
			}
		}
	}

	// Check for existing installation
	if !force {
		if serviceInfo, err := mariadb_utils.GetServiceInfo(); err == nil && serviceInfo.Status != "not-found" {
			lg.Warn("Existing MariaDB installation detected",
				logger.String("service", serviceInfo.Name),
				logger.String("status", serviceInfo.Status))

			if !PromptOverwrite() {
				lg.Info("Installation cancelled by user")
				terminal.PrintError("Installation cancelled.")
				return nil
			} else {
				// User chose to continue, set force to true
				force = true
				lg.Info("User confirmed to continue installation", logger.Bool("force", force))
			}
		}
	}

	lg.Info("Starting MariaDB install process",
		logger.String("version", version),
		logger.Bool("force", force))

	// Execute install (long-running) with spinner
	sp = terminal.NewProgressSpinner("Installing MariaDB...")
	sp.Start()
	result, err := mariadb.InstallMariaDB(mariadb_utils.InstallOptions{
		Version: version,
		Force:   force,
	})
	sp.Stop()
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// Display results
	mariadb_utils.DisplayInstallResults(result)

	// Print concise install summary to stdout
	PrintInstallSummary(result)

	// Log results
	lg.Info("MariaDB install completed successfully",
		logger.String("duration", result.Duration.String()),
		logger.String("version", result.Version),
		logger.String("os", result.OperatingSystem),
		logger.String("service_status", result.ServiceStatus))

	// Ask for custom configuration
	if !force && PromptCustomConfiguration() {
		lg.Info("Starting custom configuration process")

		sp = terminal.NewProgressSpinner("Applying custom configuration...")
		sp.Start()
		configResult, err := mariadb.ConfigureMariaDB()
		sp.Stop()

		if err != nil {
			lg.Error("Custom configuration failed", logger.Error(err))
			terminal.PrintError(fmt.Sprintf("Custom configuration failed: %v", err))
			terminal.PrintInfo("MariaDB installation completed, but configuration failed.")
			terminal.PrintInfo("You can run configuration manually later if needed.")
		} else {
			mariadb.DisplayConfigResult(configResult)
		}
	} else {
		terminal.PrintInfo("\nYou can customize MariaDB configuration later by running:")
		terminal.PrintInfo("   sfDBTools mariadb configure")
	}

	return nil
}
