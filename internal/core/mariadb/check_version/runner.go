package check_version

import (
	"fmt"

	"sfDBTools/internal/core/mariadb"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// CheckVersionRunner orchestrates the MariaDB version checking process
type CheckVersionRunner struct {
	versionService *VersionService
}

// NewCheckVersionRunner creates a new version check runner
func NewCheckVersionRunner(config *CheckVersionConfig) *CheckVersionRunner {
	return &CheckVersionRunner{
		versionService: NewVersionService(config),
	}
}

// Run executes the complete version checking flow
func (r *CheckVersionRunner) Run() error {
	lg, _ := logger.Get()
	lg.Info("Starting MariaDB version check")

	// Step 1: Check internet connectivity
	if err := r.checkInternetConnectivity(); err != nil {
		return fmt.Errorf("internet connectivity check failed: %w", err)
	}

	// Step 2: Check operating system
	if err := r.checkOperatingSystem(); err != nil {
		return fmt.Errorf("operating system check failed: %w", err)
	}

	// Step 3: Fetch and display MariaDB versions
	if err := r.fetchAndDisplayVersions(); err != nil {
		return fmt.Errorf("failed to fetch MariaDB versions: %w", err)
	}

	lg.Info("MariaDB version check completed successfully")
	return nil
}

// checkInternetConnectivity validates internet connectivity with spinner
func (r *CheckVersionRunner) checkInternetConnectivity() error {
	spinner := terminal.NewProgressSpinner("Checking internet connectivity...")
	spinner.Start()
	defer spinner.Stop()

	if err := common.RequireInternetForOperation("MariaDB version check"); err != nil {
		return err
	}

	terminal.PrintSuccess("Internet connectivity verified")
	return nil
}

// checkOperatingSystem validates OS compatibility with spinner
func (r *CheckVersionRunner) checkOperatingSystem() error {
	spinner := terminal.NewProgressSpinner("Checking operating system compatibility...")
	spinner.Start()
	defer spinner.Stop()

	if err := mariadb.ValidateOperatingSystem(); err != nil {
		return err
	}

	terminal.PrintSuccess("Operating system is supported")
	return nil
}

// fetchAndDisplayVersions retrieves and displays version information with spinner
func (r *CheckVersionRunner) fetchAndDisplayVersions() error {
	spinner := terminal.NewProgressSpinner("Fetching MariaDB version information...")
	spinner.Start()

	versions, err := r.versionService.FetchAvailableVersions()
	if err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	terminal.PrintSuccess("Version information retrieved successfully")

	// Display results
	r.displayVersionTable(versions)
	return nil
}

// displayVersionTable displays the version information in a table format
func (r *CheckVersionRunner) displayVersionTable(versions []VersionInfo) {
	if len(versions) == 0 {
		terminal.PrintWarning("No MariaDB versions found matching the criteria")
		return
	}

	terminal.PrintInfo("Available MariaDB Versions:")
	fmt.Println()

	headers := []string{"No", "Version", "EOL", "Latest Version"}
	var rows [][]string

	for i, version := range versions {
		row := []string{
			fmt.Sprintf("%d", i+1),
			version.Version,
			version.EOL,
			version.LatestVersion,
		}
		rows = append(rows, row)
	}

	terminal.FormatTable(headers, rows)
}
