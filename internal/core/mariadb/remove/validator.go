package remove

import (
	"fmt"
	"os/exec"
	"time"

	"sfDBTools/utils/terminal"
)

// Validator handles validation and confirmation operations
type Validator struct{}

// NewValidator creates a new validator for removal operations
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMariaDBServices checks if MariaDB services exist on the system
func (v *Validator) ValidateMariaDBServices() (bool, error) {
	terminal.PrintInfo("Checking MariaDB services...")
	services := []string{"mariadb", "mysql"}
	servicesFound := false

	for _, svcName := range services {
		if v.serviceExists(svcName) {
			terminal.PrintInfo(fmt.Sprintf("✅ Found %s service", svcName))
			servicesFound = true
		}
	}

	if !servicesFound {
		terminal.PrintWarning("⚠️  No MariaDB services found on this system")
		terminal.PrintInfo("Nothing to remove")
		return false, nil
	}

	return true, nil
}

// ConfirmRemoval displays warning and gets user confirmation
func (v *Validator) ConfirmRemoval(skipConfirm bool) (bool, error) {
	terminal.PrintWarning("⚠️  This will remove MariaDB packages, data directories and configuration. This action is irreversible.")

	if skipConfirm {
		return true, nil
	}

	var confirm string
	fmt.Print("Are you sure you want to continue? (y/n): ")
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		terminal.PrintInfo("Operation cancelled by user")
		return false, nil
	}

	return true, nil
}

// CreateResult creates a removal result with the specified parameters
func (v *Validator) CreateResult(success bool, message string) *RemoveResult {
	return &RemoveResult{
		Success:   success,
		Message:   message,
		RemovedAt: time.Now(),
	}
}

// serviceExists checks if a service unit file exists in the system
func (v *Validator) serviceExists(serviceName string) bool {
	cmd := exec.Command("systemctl", "cat", serviceName)
	err := cmd.Run()
	return err == nil
}
