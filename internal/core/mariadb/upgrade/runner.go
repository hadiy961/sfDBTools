package upgrade

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// UpgradeRunner orchestrates the complete MariaDB upgrade process
type UpgradeRunner struct {
	config            *UpgradeConfig
	validationService *ValidationService
	plannerService    *PlannerService
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

	// Step 2: Create upgrade plan
	plan, err := r.createUpgradePlan()
	if err != nil {
		return fmt.Errorf("upgrade planning failed: %w", err)
	}

	// Step 3: Display upgrade plan and get confirmation
	if err := r.confirmUpgradePlan(plan); err != nil {
		return fmt.Errorf("upgrade confirmation failed: %w", err)
	}

	// Step 4: Execute upgrade
	result, err := r.executeUpgrade(plan)
	if err != nil {
		return fmt.Errorf("upgrade execution failed: %w", err)
	}

	// Step 5: Handle upgrade result
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

	terminal.PrintInfo("Validating upgrade prerequisites...")

	validation, err := r.validationService.ValidateUpgrade(r.config)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !validation.Valid {
		terminal.PrintError("Upgrade validation failed:")
		for _, error := range validation.Errors {
			terminal.PrintError(fmt.Sprintf("  ❌ %s", error))
		}
		return fmt.Errorf("upgrade validation failed")
	}

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

	terminal.PrintInfo("Creating upgrade plan...")

	plan, err := r.plannerService.CreateUpgradePlan(r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create upgrade plan: %w", err)
	}

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

	terminal.PrintInfo("Executing upgrade...")

	// Create executor
	executor := NewExecutorService(r.config, plan)

	// Execute upgrade
	result, err := executor.ExecuteUpgrade()
	if err != nil {
		lg.Error("Upgrade execution failed", logger.Error(err))
		return result, fmt.Errorf("upgrade execution failed: %w", err)
	}

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
