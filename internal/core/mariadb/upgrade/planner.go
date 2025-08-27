package upgrade

import (
	"fmt"
	"time"

	"sfDBTools/internal/core/mariadb/check_version"
	"sfDBTools/internal/logger"
)

// PlannerService handles upgrade planning
type PlannerService struct {
	validationService *ValidationService
}

// NewPlannerService creates a new planner service
func NewPlannerService() *PlannerService {
	return &PlannerService{
		validationService: NewValidationService(),
	}
}

// CreateUpgradePlan creates a detailed upgrade plan
func (p *PlannerService) CreateUpgradePlan(config *UpgradeConfig) (*UpgradePlan, error) {
	lg, _ := logger.Get()

	lg.Info("Creating upgrade plan")

	// Get current installation
	current, err := p.validationService.GetCurrentInstallation()
	if err != nil {
		return nil, fmt.Errorf("failed to get current installation: %w", err)
	}

	if !current.IsInstalled {
		return nil, fmt.Errorf("MariaDB is not installed")
	}

	// Determine target version if not specified
	targetVersion := config.TargetVersion
	if targetVersion == "" {
		versionService := check_version.NewVersionService(check_version.DefaultCheckVersionConfig())
		availableVersions, err := versionService.FetchAvailableVersions()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch available versions: %w", err)
		}

		if len(availableVersions) > 0 {
			targetVersion = availableVersions[0].LatestVersion
		} else {
			return nil, fmt.Errorf("no versions available for upgrade")
		}
	}

	// Determine upgrade type
	upgradeType := p.determineUpgradeType(current.Version, targetVersion)

	// Create upgrade plan
	plan := &UpgradePlan{
		CurrentVersion: current.Version,
		TargetVersion:  targetVersion,
		UpgradeType:    upgradeType,
		Steps:          p.createUpgradeSteps(config, upgradeType),
		BackupPath:     p.determineBackupPath(config),
		EstimatedTime:  p.estimateUpgradeTime(upgradeType),
		Risks:          p.identifyRisks(upgradeType),
		Prerequisites:  p.listPrerequisites(upgradeType),
	}

	lg.Info("Upgrade plan created",
		logger.String("current_version", current.Version),
		logger.String("target_version", targetVersion),
		logger.String("upgrade_type", string(upgradeType)),
		logger.Int("steps", len(plan.Steps)))

	return plan, nil
}

// createUpgradeSteps creates the list of upgrade steps
func (p *PlannerService) createUpgradeSteps(config *UpgradeConfig, upgradeType UpgradeType) []UpgradeStep {
	steps := []UpgradeStep{}

	// Common steps for all upgrades
	steps = append(steps, UpgradeStep{
		Name:        "validate_system",
		Description: "Validate system requirements and prerequisites",
		Required:    true,
	})

	steps = append(steps, UpgradeStep{
		Name:        "detect_installation",
		Description: "Detect current MariaDB installation",
		Required:    true,
	})

	// Backup step (unless skipped)
	if !config.SkipBackup && config.BackupData {
		steps = append(steps, UpgradeStep{
			Name:        "backup_data",
			Description: "Create backup of current data",
			Required:    true,
		})
	}

	steps = append(steps, UpgradeStep{
		Name:        "stop_service",
		Description: "Stop MariaDB service",
		Required:    true,
	})

	steps = append(steps, UpgradeStep{
		Name:        "update_repository",
		Description: "Update MariaDB repository for target version",
		Required:    true,
	})

	steps = append(steps, UpgradeStep{
		Name:        "upgrade_packages",
		Description: "Upgrade MariaDB packages",
		Required:    true,
	})

	steps = append(steps, UpgradeStep{
		Name:        "start_service",
		Description: "Start MariaDB service with new version",
		Required:    true,
	})

	// Post-upgrade steps
	if !config.SkipPostUpgrade {
		steps = append(steps, UpgradeStep{
			Name:        "run_mysql_upgrade",
			Description: "Run mysql_upgrade to update system tables",
			Required:    upgradeType == UpgradeTypeMajor, // Required for major upgrades
		})
	}

	steps = append(steps, UpgradeStep{
		Name:        "verify_upgrade",
		Description: "Verify upgrade success and functionality",
		Required:    true,
	})

	return steps
}

// determineBackupPath determines where to store backup
func (p *PlannerService) determineBackupPath(config *UpgradeConfig) string {
	if config.BackupPath != "" {
		return config.BackupPath
	}

	// Use default backup location with timestamp
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("/root/mariadb_backups/upgrade_backup_%s", timestamp)
}

// estimateUpgradeTime estimates how long the upgrade will take
func (p *PlannerService) estimateUpgradeTime(upgradeType UpgradeType) string {
	switch upgradeType {
	case UpgradeTypeMajor:
		return "15-30 minutes"
	case UpgradeTypeMinor:
		return "5-15 minutes"
	case UpgradeTypePatch:
		return "2-5 minutes"
	default:
		return "Unknown"
	}
}

// identifyRisks identifies potential risks for the upgrade
func (p *PlannerService) identifyRisks(upgradeType UpgradeType) []string {
	risks := []string{
		"Service downtime during upgrade",
		"Potential data corruption if upgrade fails",
	}

	switch upgradeType {
	case UpgradeTypeMajor:
		risks = append(risks,
			"Major version changes may introduce incompatibilities",
			"Configuration syntax may have changed",
			"Query behavior might be different",
			"Plugins or features may be deprecated",
		)
	case UpgradeTypeMinor:
		risks = append(risks,
			"Minor configuration changes may be required",
			"Some features may behave differently",
		)
	}

	return risks
}

// listPrerequisites lists what should be done before upgrade
func (p *PlannerService) listPrerequisites(upgradeType UpgradeType) []string {
	prereqs := []string{
		"Ensure sufficient disk space for backup",
		"Verify no critical applications are currently using database",
		"Review MariaDB changelog for target version",
	}

	switch upgradeType {
	case UpgradeTypeMajor:
		prereqs = append(prereqs,
			"Test upgrade on a copy of production data",
			"Review application compatibility with new version",
			"Plan for potential rollback",
			"Notify stakeholders about planned downtime",
		)
	}

	return prereqs
}

// determineUpgradeType determines the type of upgrade (reused from validation)
func (p *PlannerService) determineUpgradeType(currentVersion, targetVersion string) UpgradeType {
	return p.validationService.determineUpgradeType(currentVersion, targetVersion, nil)
}
