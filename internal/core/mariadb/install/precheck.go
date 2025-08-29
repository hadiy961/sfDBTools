package install

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// PrecheckManager handles pre-installation checks
type PrecheckManager struct {
	osInfo *common.OSInfo
}

// NewPrecheckManager creates a new precheck manager
func NewPrecheckManager() *PrecheckManager {
	return &PrecheckManager{}
}

// CheckOSCompatibility checks if the OS is supported
func (p *PrecheckManager) CheckOSCompatibility() (*common.OSInfo, error) {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Detecting operating system...")
	spinner.Start()

	// Detect OS using common utility
	detector := common.NewOSDetector()
	osInfo, err := detector.DetectOS()
	if err != nil {
		spinner.Stop()
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}

	p.osInfo = osInfo

	// Check OS compatibility using MariaDB supported OS list
	supportedOS := common.MariaDBSupportedOS()
	if err := common.ValidateOperatingSystem(supportedOS); err != nil {
		spinner.Stop()
		return nil, fmt.Errorf("OS compatibility check failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess(fmt.Sprintf("Operating system detected: %s %s (%s)",
		osInfo.ID, osInfo.Version, osInfo.Architecture))

	lg.Info("OS compatibility check passed",
		logger.String("os", osInfo.ID),
		logger.String("version", osInfo.Version),
		logger.String("arch", osInfo.Architecture))

	return osInfo, nil
}

// CheckInternetConnectivity verifies internet connection
func (p *PrecheckManager) CheckInternetConnectivity() error {
	lg, _ := logger.Get()

	spinner := terminal.NewProgressSpinner("Checking internet connectivity...")
	spinner.Start()

	if err := common.CheckInternetConnectivity(); err != nil {
		spinner.Stop()
		return fmt.Errorf("internet connectivity check failed: %w", err)
	}

	spinner.Stop()
	terminal.PrintSuccess("Internet connectivity verified")

	lg.Info("Internet connectivity check passed")
	return nil
}
