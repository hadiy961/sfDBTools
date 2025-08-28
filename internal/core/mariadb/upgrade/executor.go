package upgrade

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sfDBTools/internal/core/mariadb/install"
	"sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// ExecutorService handles the actual upgrade execution
type ExecutorService struct {
	config            *UpgradeConfig
	plan              *UpgradePlan
	validationService *ValidationService
	osInfo            *common.OSInfo
	backupService     *remove.BackupService
}

// NewExecutorService creates a new executor service
func NewExecutorService(config *UpgradeConfig, plan *UpgradePlan) *ExecutorService {
	validationService := NewValidationService()

	// Get OS info
	detector := common.NewOSDetector()
	osInfo, _ := detector.DetectOS()

	// Create backup service
	backupService := remove.NewBackupService(osInfo)

	return &ExecutorService{
		config:            config,
		plan:              plan,
		validationService: validationService,
		osInfo:            osInfo,
		backupService:     backupService,
	}
}

// ExecuteUpgrade performs the actual upgrade
func (e *ExecutorService) ExecuteUpgrade() (*UpgradeResult, error) {
	lg, _ := logger.Get()

	startTime := time.Now()
	result := &UpgradeResult{
		PreviousVersion: e.plan.CurrentVersion,
		NewVersion:      e.plan.TargetVersion,
		StepsTotal:      len(e.plan.Steps),
	}

	lg.Info("Starting MariaDB upgrade execution",
		logger.String("from_version", e.plan.CurrentVersion),
		logger.String("to_version", e.plan.TargetVersion))

	terminal.PrintHeader("MariaDB Upgrade")
	terminal.PrintInfo(fmt.Sprintf("Upgrading MariaDB from %s to %s",
		e.plan.CurrentVersion, e.plan.TargetVersion))

	// Execute each step
	for i, step := range e.plan.Steps {
		lg.Info("Executing upgrade step",
			logger.String("step", step.Name),
			logger.String("description", step.Description))

		if err := e.executeStep(step); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("step '%s' failed: %w", step.Name, err)
			result.StepsCompleted = i
			break
		}

		result.StepsCompleted++
	}

	// Calculate duration
	result.Duration = time.Since(startTime).String()

	// If all steps completed successfully
	if result.StepsCompleted == result.StepsTotal {
		result.Success = true

		// Verify final version
		current, err := e.validationService.GetCurrentInstallation()
		if err == nil {
			result.NewVersion = current.Version
		}

		lg.Info("MariaDB upgrade completed successfully",
			logger.String("new_version", result.NewVersion),
			logger.String("duration", result.Duration))

		terminal.PrintSuccess("MariaDB upgrade completed successfully!")
	} else {
		lg.Error("MariaDB upgrade failed",
			logger.Error(result.Error),
			logger.Int("completed_steps", result.StepsCompleted),
			logger.Int("total_steps", result.StepsTotal))

		terminal.PrintError("MariaDB upgrade failed")

		// Provide rollback information
		result.RollbackInfo = &RollbackInfo{
			AvailableBackup: e.plan.BackupPath,
			PreviousVersion: e.plan.CurrentVersion,
			RollbackSteps: []string{
				"1. Stop MariaDB service",
				"2. Restore data from backup: " + e.plan.BackupPath,
				"3. Downgrade packages to previous version",
				"4. Start MariaDB service",
			},
		}
	}

	return result, nil
}

// executeStep executes a single upgrade step
func (e *ExecutorService) executeStep(step UpgradeStep) error {
	lg, _ := logger.Get()

	switch step.Name {
	case "validate_system":
		return e.validateSystem()
	case "detect_installation":
		return e.detectInstallation()
	case "backup_data":
		return e.backupData()
	case "stop_service":
		return e.stopService()
	case "update_repository":
		return e.updateRepository()
	case "upgrade_packages":
		return e.upgradePackages()
	case "start_service":
		return e.startService()
	case "run_mysql_upgrade":
		return e.runMysqlUpgrade()
	case "verify_upgrade":
		return e.verifyUpgrade()
	default:
		lg.Warn("Unknown upgrade step", logger.String("step", step.Name))
		return fmt.Errorf("unknown upgrade step: %s", step.Name)
	}
}

// validateSystem validates system requirements
func (e *ExecutorService) validateSystem() error {
	spinner := terminal.NewProgressSpinner("Validating system requirements...")
	spinner.Start()

	// Use validation service
	validation, err := e.validationService.ValidateUpgrade(e.config)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("validation failed: %w", err)
	}

	if !validation.Valid {
		spinner.Stop()
		return fmt.Errorf("system validation failed: %v", validation.Errors)
	}

	spinner.Stop()
	terminal.PrintSuccess("System validation completed")
	return nil
}

// detectInstallation detects current installation
func (e *ExecutorService) detectInstallation() error {
	spinner := terminal.NewProgressSpinner("Detecting current installation...")
	spinner.Start()

	current, err := e.validationService.GetCurrentInstallation()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("installation detection failed: %w", err)
	}

	if !current.IsInstalled {
		spinner.Stop()
		return fmt.Errorf("MariaDB is not installed")
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Current installation detected: %s", current.Version))
	return nil
}

// backupData creates backup of current data
func (e *ExecutorService) backupData() error {
	spinner := terminal.NewProgressSpinner("Creating data backup...")
	spinner.Start()

	// Get detected installation for backup service
	installation, err := e.validationService.detectionService.DetectInstallation()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to detect installation: %w", err)
	}

	// Use backup service from remove module
	if err := e.backupService.BackupData(installation, e.plan.BackupPath); err != nil {
		spinner.Stop()
		return fmt.Errorf("backup failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Data backup completed")
	return nil
}

// stopService stops MariaDB service
func (e *ExecutorService) stopService() error {
	spinner := terminal.NewProgressSpinner("Stopping MariaDB service...")
	spinner.Start()

	services := []string{"mariadb", "mysql", "mysqld"}

	var lastErr error
	for _, service := range services {
		cmd := exec.Command("systemctl", "stop", service)
		if err := cmd.Run(); err == nil {
			spinner.Stop()
			terminal.PrintSuccess("MariaDB service stopped")
			return nil
		} else {
			lastErr = err
		}
	}

	spinner.Stop()
	return fmt.Errorf("failed to stop MariaDB service: %w", lastErr)
}

// updateRepository updates MariaDB repository for target version
func (e *ExecutorService) updateRepository() error {
	spinner := terminal.NewProgressSpinner("Updating MariaDB repository...")
	spinner.Start()

	// Use repository setup from install module
	repoManager := install.NewRepoSetupManager(e.osInfo)

	// Extract major.minor version from target version
	targetMajorMinor := e.extractMajorMinorVersion(e.plan.TargetVersion)

	if err := repoManager.SetupRepository(targetMajorMinor); err != nil {
		spinner.Stop()
		return fmt.Errorf("repository update failed: %w", err)
	}

	// Update package cache
	if err := repoManager.UpdatePackageCache(); err != nil {
		spinner.Stop()
		return fmt.Errorf("package cache update failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Repository updated")
	return nil
}

// upgradePackages upgrades MariaDB packages
func (e *ExecutorService) upgradePackages() error {
	spinner := terminal.NewProgressSpinner("Upgrading MariaDB packages...")
	spinner.Start()

	// Upgrade packages using yum directly
	cmd := exec.Command("yum", "update", "-y", "MariaDB-server", "MariaDB-client")
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("package upgrade failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Packages upgraded")
	return nil
}

// startService starts MariaDB service
func (e *ExecutorService) startService() error {
	spinner := terminal.NewProgressSpinner("Starting MariaDB service...")
	spinner.Start()

	services := []string{"mariadb", "mysql"}

	var lastErr error
	for _, service := range services {
		cmd := exec.Command("systemctl", "start", service)
		if err := cmd.Run(); err == nil {
			spinner.Stop()
			terminal.PrintSuccess("MariaDB service started")
			return nil
		} else {
			lastErr = err
		}
	}

	spinner.Stop()
	return fmt.Errorf("failed to start MariaDB service: %w", lastErr)
}

// runMysqlUpgrade runs mysql_upgrade utility
func (e *ExecutorService) runMysqlUpgrade() error {
	spinner := terminal.NewProgressSpinner("Running mysql_upgrade...")
	spinner.Start()

	cmd := exec.Command("mysql_upgrade", "--force")
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("mysql_upgrade failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("mysql_upgrade completed")
	return nil
}

// verifyUpgrade verifies that upgrade was successful
func (e *ExecutorService) verifyUpgrade() error {
	spinner := terminal.NewProgressSpinner("Verifying upgrade...")
	spinner.Start()

	// Check if service is running
	cmd := exec.Command("systemctl", "is-active", "mariadb")
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("MariaDB service is not running")
	}

	// Check version
	current, err := e.validationService.GetCurrentInstallation()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to verify installation: %w", err)
	}

	// Simple connectivity test
	cmd = exec.Command("mysql", "-e", "SELECT VERSION();")
	if err := cmd.Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("database connectivity test failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Upgrade verified - Running MariaDB %s", current.Version))
	return nil
}

// extractMajorMinorVersion extracts major.minor version (e.g., "11.4.8" -> "11.4")
func (e *ExecutorService) extractMajorMinorVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s.%s", parts[0], parts[1])
	}
	return version
}
