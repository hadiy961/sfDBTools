package configure

import (
	"fmt"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// SystemdManager handles systemd service configuration
type SystemdManager struct {
	serviceName string
}

// NewSystemdManager creates a new systemd manager
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{
		serviceName: "mariadb",
	}
}

// ConfigureService configures systemd service for MariaDB
func (s *SystemdManager) ConfigureService() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Configuring systemd service...")

	// Update LimitNOFILE settings
	if err := s.updateLimitNOFILE(); err != nil {
		return fmt.Errorf("failed to update LimitNOFILE: %w", err)
	}

	// Add TimeoutStartSec
	if err := s.addTimeoutStartSec(); err != nil {
		return fmt.Errorf("failed to add TimeoutStartSec: %w", err)
	}

	// Reload systemd daemon
	if err := s.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	lg.Info("Systemd service configured successfully")
	terminal.PrintSuccess("Systemd service configured successfully")
	return nil
}

// updateLimitNOFILE updates the LimitNOFILE setting in mariadb.service
func (s *SystemdManager) updateLimitNOFILE() error {
	lg, _ := logger.Get()

	serviceFile := "/usr/lib/systemd/system/mariadb.service"

	// Update LimitNOFILE=16364 to LimitNOFILE=655360
	cmd1 := exec.Command("sed", "-i", "s/LimitNOFILE=16364/LimitNOFILE=655360/", serviceFile)
	if err := cmd1.Run(); err != nil {
		lg.Error("Failed to update LimitNOFILE=16364", logger.Error(err))
		// Continue anyway, this might not exist
	}

	// Update LimitNOFILE=32768 to LimitNOFILE=655360
	cmd2 := exec.Command("sed", "-i", "s/LimitNOFILE=32768/LimitNOFILE=655360/", serviceFile)
	if err := cmd2.Run(); err != nil {
		lg.Error("Failed to update LimitNOFILE=32768", logger.Error(err))
		// Continue anyway, this might not exist
	}

	lg.Info("LimitNOFILE settings updated")
	return nil
}

// addTimeoutStartSec adds TimeoutStartSec setting to mariadb.service
func (s *SystemdManager) addTimeoutStartSec() error {
	lg, _ := logger.Get()

	serviceFile := "/usr/lib/systemd/system/mariadb.service"

	// Add TimeoutStartSec setting
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 'TimeoutStartSec=7200s' >> %s", serviceFile))
	if err := cmd.Run(); err != nil {
		lg.Error("Failed to add TimeoutStartSec", logger.Error(err))
		return err
	}

	lg.Info("TimeoutStartSec setting added")
	return nil
}

// reloadSystemd reloads systemd daemon
func (s *SystemdManager) reloadSystemd() error {
	lg, _ := logger.Get()

	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		lg.Error("Failed to reload systemd daemon", logger.Error(err))
		return err
	}

	lg.Info("Systemd daemon reloaded")
	return nil
}

// FirewallManager handles firewall configuration
type FirewallManager struct {
	port int
}

// NewFirewallManager creates a new firewall manager
func NewFirewallManager(port int) *FirewallManager {
	return &FirewallManager{
		port: port,
	}
}

// ConfigureFirewall configures firewall for MariaDB port
func (f *FirewallManager) ConfigureFirewall() error {
	lg, _ := logger.Get()

	terminal.PrintInfo(fmt.Sprintf("Configuring firewall for port %d...", f.port))

	// Add port to firewall
	if err := f.addPort(); err != nil {
		return fmt.Errorf("failed to add port to firewall: %w", err)
	}

	// Reload firewall
	if err := f.reloadFirewall(); err != nil {
		return fmt.Errorf("failed to reload firewall: %w", err)
	}

	// List ports for verification
	if err := f.listPorts(); err != nil {
		lg.Warn("Failed to list firewall ports", logger.Error(err))
		// Non-critical error, continue
	}

	lg.Info("Firewall configured successfully", logger.Int("port", f.port))
	terminal.PrintSuccess(fmt.Sprintf("Firewall configured for port %d", f.port))
	return nil
}

// addPort adds the MariaDB port to firewall
func (f *FirewallManager) addPort() error {
	lg, _ := logger.Get()

	cmd := exec.Command("firewall-cmd", "--zone=public", fmt.Sprintf("--add-port=%d/tcp", f.port), "--permanent")
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to add port to firewall",
			logger.Int("port", f.port),
			logger.Error(err),
			logger.String("output", string(output)))
		return err
	}

	lg.Info("Port added to firewall", logger.Int("port", f.port))
	return nil
}

// reloadFirewall reloads the firewall configuration
func (f *FirewallManager) reloadFirewall() error {
	lg, _ := logger.Get()

	cmd := exec.Command("firewall-cmd", "--complete-reload")
	if err := cmd.Run(); err != nil {
		lg.Error("Failed to reload firewall", logger.Error(err))
		return err
	}

	lg.Info("Firewall reloaded")
	return nil
}

// listPorts lists current firewall ports for verification
func (f *FirewallManager) listPorts() error {
	lg, _ := logger.Get()

	cmd := exec.Command("firewall-cmd", "--list-ports")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lg.Info("Current firewall ports", logger.String("ports", string(output)))
	terminal.PrintInfo(fmt.Sprintf("Current firewall ports: %s", string(output)))
	return nil
}
