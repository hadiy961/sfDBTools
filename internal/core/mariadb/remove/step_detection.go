package remove

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
)

// DetectionStep detects existing MariaDB installation
type DetectionStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *DetectionStep) Name() string {
	return "Installation Detection"
}

// Validate validates the step preconditions
func (s *DetectionStep) Validate(state *State) error {
	if state.Config == nil {
		return fmt.Errorf("removal config is required")
	}
	return nil
}

// Execute detects the MariaDB installation
func (s *DetectionStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Starting MariaDB installation detection")

	spinner := terminal.NewProgressSpinner("Detecting MariaDB installation...")
	spinner.Start()
	defer spinner.Stop()

	// Get OS info first
	detector := common.NewOSDetector()
	osInfo, err := detector.DetectOS()
	if err != nil {
		return fmt.Errorf("failed to get OS information: %w", err)
	}

	// Create detection service with OS info
	detectionService := NewDetectionService(osInfo)

	// Detect installation
	installation, err := detectionService.DetectInstallation()
	if err != nil {
		return fmt.Errorf("failed to detect MariaDB installation: %w", err)
	}

	if !installation.IsInstalled {
		terminal.PrintWarning("No MariaDB installation detected")
		return fmt.Errorf("no MariaDB installation found")
	}

	// Store installation info in state
	state.Installation = installation

	terminal.PrintSuccess("MariaDB installation detected")
	terminal.PrintInfo(fmt.Sprintf("Service: %s", installation.ServiceName))
	terminal.PrintInfo(fmt.Sprintf("Data directory: %s", installation.ActualDataDir))
	terminal.PrintInfo(fmt.Sprintf("Config files: %v", installation.ConfigFiles))

	return nil
}

// Rollback rolls back the detection step (no-op)
func (s *DetectionStep) Rollback(ctx context.Context, state *State) error {
	// Detection has no rollback - it's read-only
	return nil
}
