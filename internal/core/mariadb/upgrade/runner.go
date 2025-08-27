package upgrade

import (
	"fmt"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// UpgradeRunner orchestrates the complete MariaDB upgrade process
type UpgradeRunner struct {
	config            *UpgradeConfig
	validationService *ValidationService
	plannerService    *PlannerService
	versionService    *check_version.VersionService
	versionSelector   *install.VersionSelector
	selectedVersion   *install.SelectableVersion
}

// NewUpgradeRunner creates a new upgrade runner
func NewUpgradeRunner(config *UpgradeConfig) *UpgradeRunner {
	if config == nil {
		config = DefaultUpgradeConfig()
	}

	return &UpgradeRunner{
		config:            config,
		validationService: NewValidationService(),
		plannerService:    NewPlannerService(),
		versionService:    check_version.NewVersionService(check_version.DefaultCheckVersionConfig()),
	}
}

// Run executes the complete MariaDB upgrade process
func (r *UpgradeRunner) Run() error {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB upgrade process")
	terminal.PrintHeader("MariaDB Upgrade")

	// Step 1: Validate upgrade prerequisites
	if err := r.validateUpgrade(); err != nil {
		return fmt.Errorf("upgrade validation failed: %w", err)
	}

	// Step 2: Select target version (if not specified)
	if r.config.TargetVersion == "" {
		if err := r.selectTargetVersion(); err != nil {
			return fmt.Errorf("target version selection failed: %w", err)
		}
	} else {
		terminal.PrintInfo(fmt.Sprintf("Using specified target version: %s", r.config.TargetVersion))
	}

	// Step 3: Create upgrade plan
	plan, err := r.createUpgradePlan()
	if err != nil {
		return fmt.Errorf("upgrade planning failed: %w", err)
	}

	// Step 4: Display upgrade plan and get confirmation
	if err := r.confirmUpgradePlan(plan); err != nil {
		return fmt.Errorf("upgrade confirmation failed: %w", err)
	}

	// Step 5: Execute upgrade
	result, err := r.executeUpgrade(plan)
	if err != nil {
		return fmt.Errorf("upgrade execution failed: %w", err)
	}

	// Step 6: Handle upgrade result
	if err := r.handleUpgradeResult(result); err != nil {
		return fmt.Errorf("upgrade result handling failed: %w", err)
	}

	terminal.PrintSuccess("MariaDB upgrade completed successfully!")
	lg.Info("MariaDB upgrade process completed successfully")

	return nil
}

// validateUpgrade validates if upgrade is possible
func (r *UpgradeRunner) validateUpgrade() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Validating upgrade prerequisites...")
	spinner.Start()

	validation, err := r.validationService.ValidateUpgrade(r.config)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("validation error: %w", err)
	}

	if !validation.Valid {
		spinner.Stop()
		terminal.PrintError("Upgrade validation failed:")
		for _, error := range validation.Errors {
			terminal.PrintError(fmt.Sprintf("  ❌ %s", error))
		}
		return fmt.Errorf("upgrade validation failed")
	}

	spinner.Stop()
	terminal.PrintSuccess("Upgrade prerequisites validated")

	// Display warnings if any
	if len(validation.Warnings) > 0 {
		terminal.PrintWarning("Upgrade warnings:")
		for _, warning := range validation.Warnings {
			terminal.PrintWarning(fmt.Sprintf("  ⚠️  %s", warning))
		}
	}

	// Display suggestions if any
	if len(validation.Suggestions) > 0 {
		terminal.PrintInfo("Suggestions:")
		for _, suggestion := range validation.Suggestions {
			terminal.PrintInfo(fmt.Sprintf("  💡 %s", suggestion))
		}
	}

	terminal.PrintSuccess("Upgrade validation completed")
	lg.Info("Upgrade validation completed successfully")

	return nil
}

// createUpgradePlan creates the upgrade execution plan
func (r *UpgradeRunner) createUpgradePlan() (*UpgradePlan, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Creating upgrade plan...")
	spinner.Start()

	plan, err := r.plannerService.CreateUpgradePlan(r.config)
	if err != nil {
		spinner.Stop()
		return nil, fmt.Errorf("failed to create upgrade plan: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Upgrade plan created")

	lg.Info("Upgrade plan created",
		logger.String("current_version", plan.CurrentVersion),
		logger.String("target_version", plan.TargetVersion),
		logger.String("upgrade_type", string(plan.UpgradeType)))

	return plan, nil
}

// confirmUpgradePlan displays the plan and gets user confirmation
func (r *UpgradeRunner) confirmUpgradePlan(plan *UpgradePlan) error {
	lg, _ := logger.Get()

	// Display upgrade plan
	r.displayUpgradePlan(plan)

	// Skip confirmation if auto-confirm is enabled
	if r.config.AutoConfirm {
		terminal.PrintInfo("Auto-confirm enabled, proceeding with upgrade...")
		lg.Info("Upgrade auto-confirmed")
		return nil
	}

	// Get user confirmation
	if !r.confirmUpgrade() {
		return fmt.Errorf("upgrade cancelled by user")
	}

	lg.Info("Upgrade confirmed by user")
	return nil
}

// displayUpgradePlan displays the upgrade plan to user
func (r *UpgradeRunner) displayUpgradePlan(plan *UpgradePlan) {
	terminal.PrintInfo("📋 Upgrade Plan:")
	terminal.PrintInfo(fmt.Sprintf("  Current Version: %s", plan.CurrentVersion))
	terminal.PrintInfo(fmt.Sprintf("  Target Version:  %s", plan.TargetVersion))
	terminal.PrintInfo(fmt.Sprintf("  Upgrade Type:    %s", plan.UpgradeType))
	terminal.PrintInfo(fmt.Sprintf("  Estimated Time:  %s", plan.EstimatedTime))

	if plan.BackupPath != "" {
		terminal.PrintInfo(fmt.Sprintf("  Backup Location: %s", plan.BackupPath))
	}

	// Display steps
	terminal.PrintInfo("\n📝 Upgrade Steps:")
	for i, step := range plan.Steps {
		required := ""
		if step.Required {
			required = " (required)"
		}
		terminal.PrintInfo(fmt.Sprintf("  %d. %s%s", i+1, step.Description, required))
	}

	// Display risks
	if len(plan.Risks) > 0 {
		terminal.PrintWarning("\n⚠️  Risks:")
		for _, risk := range plan.Risks {
			terminal.PrintWarning(fmt.Sprintf("  • %s", risk))
		}
	}

	// Display prerequisites
	if len(plan.Prerequisites) > 0 {
		terminal.PrintInfo("\n✅ Prerequisites:")
		for _, prereq := range plan.Prerequisites {
			terminal.PrintInfo(fmt.Sprintf("  • %s", prereq))
		}
	}
}

// confirmUpgrade asks user for confirmation
func (r *UpgradeRunner) confirmUpgrade() bool {
	terminal.PrintWarning("\n⚠️  This operation will:")
	terminal.PrintWarning("   • Stop MariaDB service temporarily")
	terminal.PrintWarning("   • Upgrade MariaDB packages")
	terminal.PrintWarning("   • Modify system configuration")

	confirmed, err := terminal.ConfirmAndClear("Do you want to proceed with the upgrade?")
	if err != nil {
		return false
	}
	return confirmed
}

// executeUpgrade executes the upgrade plan
func (r *UpgradeRunner) executeUpgrade(plan *UpgradePlan) (*UpgradeResult, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Executing MariaDB upgrade...")
	spinner.Start()

	// Create executor
	executor := NewExecutorService(r.config, plan)

	// Execute upgrade
	result, err := executor.ExecuteUpgrade()
	if err != nil {
		spinner.Stop()
		lg.Error("Upgrade execution failed", logger.Error(err))
		return result, fmt.Errorf("upgrade execution failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("MariaDB upgrade executed successfully")

	lg.Info("Upgrade execution completed",
		logger.Bool("success", result.Success),
		logger.String("duration", result.Duration))

	return result, nil
}

// handleUpgradeResult handles the upgrade result
func (r *UpgradeRunner) handleUpgradeResult(result *UpgradeResult) error {
	if result.Success {
		// Success case
		terminal.PrintSuccess("✅ Upgrade completed successfully!")
		terminal.PrintInfo(fmt.Sprintf("   Previous Version: %s", result.PreviousVersion))
		terminal.PrintInfo(fmt.Sprintf("   New Version:      %s", result.NewVersion))
		terminal.PrintInfo(fmt.Sprintf("   Duration:         %s", result.Duration))
		terminal.PrintInfo(fmt.Sprintf("   Steps Completed:  %d/%d", result.StepsCompleted, result.StepsTotal))

		if result.BackupPath != "" {
			terminal.PrintInfo(fmt.Sprintf("   Backup Location:  %s", result.BackupPath))
		}

		// Post-upgrade recommendations
		terminal.PrintInfo("\n📋 Post-Upgrade Recommendations:")
		terminal.PrintInfo("   • Test your applications with the new version")
		terminal.PrintInfo("   • Review MariaDB changelog for new features")
		terminal.PrintInfo("   • Consider running mysql_secure_installation if not done before")

		return nil
	} else {
		// Failure case
		terminal.PrintError("❌ Upgrade failed!")
		terminal.PrintError(fmt.Sprintf("   Error: %s", result.Error))
		terminal.PrintError(fmt.Sprintf("   Steps Completed: %d/%d", result.StepsCompleted, result.StepsTotal))

		if result.RollbackInfo != nil {
			terminal.PrintWarning("\n🔄 Rollback Information:")
			terminal.PrintWarning(fmt.Sprintf("   Backup Available: %s", result.RollbackInfo.AvailableBackup))
			terminal.PrintWarning(fmt.Sprintf("   Previous Version: %s", result.RollbackInfo.PreviousVersion))
			terminal.PrintWarning("\n   Rollback Steps:")
			for i, step := range result.RollbackInfo.RollbackSteps {
				terminal.PrintWarning(fmt.Sprintf("   %d. %s", i+1, step))
			}
		}

		return fmt.Errorf("upgrade failed: %w", result.Error)
	}
}

// selectTargetVersion handles interactive target version selection
func (r *UpgradeRunner) selectTargetVersion() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Fetching available MariaDB versions...")
	spinner.Start()

	// Fetch available versions
	versions, err := r.versionService.FetchAvailableVersions()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to fetch available versions: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Available versions fetched")

	// Convert to SelectableVersion format using the converter
	selectableVersions := install.ConvertVersionInfo(versions)

	// Create version selector
	r.versionSelector = install.NewVersionSelector(selectableVersions)

	terminal.PrintInfo("Please select a target MariaDB version for upgrade:")

	// Select version
	selectedVersion, err := r.versionSelector.SelectVersion(r.config.AutoConfirm, r.config.TargetVersion)
	if err != nil {
		return fmt.Errorf("version selection failed: %w", err)
	}

	r.selectedVersion = selectedVersion
	r.config.TargetVersion = selectedVersion.Version

	terminal.PrintSuccess(fmt.Sprintf("Selected target version: %s (%s)",
		selectedVersion.Version, selectedVersion.LatestVersion))

	lg.Info("Target version selected",
		logger.String("target_version", selectedVersion.Version),
		logger.String("latest_version", selectedVersion.LatestVersion))

	return nil
}
