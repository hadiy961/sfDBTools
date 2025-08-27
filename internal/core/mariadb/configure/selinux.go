package configure

import (
	"fmt"
	"os/exec"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// SELinuxManager handles SELinux configuration for MariaDB
type SELinuxManager struct {
	dataDir string
}

// NewSELinuxManager creates a new SELinux manager
func NewSELinuxManager(dataDir string) *SELinuxManager {
	return &SELinuxManager{
		dataDir: dataDir,
	}
}

// ConfigureSELinux configures SELinux contexts for MariaDB data directory
func (s *SELinuxManager) ConfigureSELinux() error {
	lg, _ := logger.Get()

	// Check if SELinux is enabled
	if !s.isSELinuxEnabled() {
		lg.Info("SELinux is not enabled, skipping SELinux configuration")
		terminal.PrintInfo("SELinux is not enabled, skipping SELinux configuration")
		return nil
	}

	terminal.PrintInfo("Configuring SELinux contexts for MariaDB...")

	// Install policycoreutils-python if needed
	if err := s.installPolicyCoreUtils(); err != nil {
		lg.Warn("Failed to install policycoreutils-python", logger.Error(err))
		// Continue anyway, might already be installed
	}

	// Set SELinux context
	if err := s.setSELinuxContext(); err != nil {
		return fmt.Errorf("failed to set SELinux context: %w", err)
	}

	// Restore SELinux contexts
	if err := s.restoreContexts(); err != nil {
		return fmt.Errorf("failed to restore SELinux contexts: %w", err)
	}

	lg.Info("SELinux configured successfully", logger.String("data_dir", s.dataDir))
	terminal.PrintSuccess("SELinux contexts configured successfully")
	return nil
}

// isSELinuxEnabled checks if SELinux is enabled
func (s *SELinuxManager) isSELinuxEnabled() bool {
	cmd := exec.Command("getenforce")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return string(output) != "Disabled\n"
}

// installPolicyCoreUtils installs policycoreutils-python package
func (s *SELinuxManager) installPolicyCoreUtils() error {
	lg, _ := logger.Get()

	cmd := exec.Command("yum", "install", "-y", "policycoreutils-python")
	if err := cmd.Run(); err != nil {
		lg.Error("Failed to install policycoreutils-python", logger.Error(err))
		return err
	}

	lg.Info("policycoreutils-python installed successfully")
	return nil
}

// setSELinuxContext sets the SELinux context for the data directory
func (s *SELinuxManager) setSELinuxContext() error {
	lg, _ := logger.Get()

	contextPath := fmt.Sprintf("%s(/.*)?", s.dataDir)
	cmd := exec.Command("semanage", "fcontext", "-a", "-t", "mysqld_db_t", contextPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to set SELinux context",
			logger.String("path", contextPath),
			logger.Error(err),
			logger.String("output", string(output)))
		return err
	}

	lg.Info("SELinux context set successfully", logger.String("path", contextPath))
	return nil
}

// restoreContexts restores SELinux contexts recursively
func (s *SELinuxManager) restoreContexts() error {
	lg, _ := logger.Get()

	cmd := exec.Command("restorecon", "-Rv", s.dataDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Failed to restore SELinux contexts",
			logger.String("path", s.dataDir),
			logger.Error(err),
			logger.String("output", string(output)))
		return err
	}

	lg.Info("SELinux contexts restored successfully",
		logger.String("path", s.dataDir),
		logger.String("output", string(output)))
	return nil
}
