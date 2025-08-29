package remove

import (
	"context"
	"fmt"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// ServiceStopStep stops MariaDB services
type ServiceStopStep struct {
	deps Dependencies
}

// Name returns the step name
func (s *ServiceStopStep) Name() string {
	return "Stop Services"
}

// Validate validates the step preconditions
func (s *ServiceStopStep) Validate(state *State) error {
	if state.Installation == nil {
		return fmt.Errorf("installation detection is required before stopping services")
	}
	return nil
}

// Execute stops the MariaDB services
func (s *ServiceStopStep) Execute(ctx context.Context, state *State) error {
	lg, _ := logger.Get()
	lg.Info("Stopping MariaDB services")

	installation := state.Installation
	if !installation.ServiceActive {
		terminal.PrintInfo("MariaDB service is not running, skipping stop")
		return nil
	}

	spinner := terminal.NewProgressSpinner(fmt.Sprintf("Stopping %s service...", installation.ServiceName))
	spinner.Start()
	defer spinner.Stop()

	// Store previous service state for rollback
	if state.RollbackData == nil {
		state.RollbackData = make(map[string]interface{})
	}
	state.RollbackData["serviceWasActive"] = installation.ServiceActive
	state.RollbackData["serviceName"] = installation.ServiceName

	// Stop the service
	if err := s.deps.ServiceManager.Stop(installation.ServiceName); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", installation.ServiceName, err)
	}

	terminal.PrintSuccess(fmt.Sprintf("Stopped %s service", installation.ServiceName))

	// Disable the service if it was enabled
	if installation.ServiceEnabled {
		state.RollbackData["serviceWasEnabled"] = true
		spinner.UpdateMessage(fmt.Sprintf("Disabling %s service...", installation.ServiceName))

		if err := s.deps.ServiceManager.Disable(installation.ServiceName); err != nil {
			lg.Warn("Failed to disable service, continuing anyway", logger.String("service", installation.ServiceName), logger.Error(err))
		} else {
			terminal.PrintSuccess(fmt.Sprintf("Disabled %s service", installation.ServiceName))
		}
	}

	return nil
}

// Rollback restarts the service if it was previously running
func (s *ServiceStopStep) Rollback(ctx context.Context, state *State) error {
	lg, _ := logger.Get()

	serviceName, ok := state.RollbackData["serviceName"].(string)
	if !ok {
		return nil // No service to rollback
	}

	wasActive, _ := state.RollbackData["serviceWasActive"].(bool)
	wasEnabled, _ := state.RollbackData["serviceWasEnabled"].(bool)

	lg.Info("Rolling back service state", logger.String("service", serviceName))

	// Re-enable service if it was enabled
	if wasEnabled {
		if err := s.deps.ServiceManager.Enable(serviceName); err != nil {
			lg.Error("Failed to re-enable service during rollback", logger.String("service", serviceName), logger.Error(err))
		}
	}

	// Restart service if it was active
	if wasActive {
		if err := s.deps.ServiceManager.Start(serviceName); err != nil {
			lg.Error("Failed to restart service during rollback", logger.String("service", serviceName), logger.Error(err))
			return err
		}
		lg.Info("Service restarted during rollback", logger.String("service", serviceName))
	}

	return nil
}
