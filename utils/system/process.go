package system

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ProcessManager interface provides abstraction for process execution
type ProcessManager interface {
	ExecuteWithTimeout(command string, args []string, timeout time.Duration) error
	Execute(command string, args []string) error
	ExecuteWithOutput(command string, args []string) (string, error)
}

// processManager implements ProcessManager interface
type processManager struct{}

// NewProcessManager creates a new process manager
func NewProcessManager() ProcessManager {
	return &processManager{}
}

// ExecuteWithTimeout executes a command with a timeout
func (pm *processManager) ExecuteWithTimeout(command string, args []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("command %s timed out after %v", command, timeout)
	}

	if err != nil {
		return fmt.Errorf("command %s failed: %w\nOutput: %s", command, err, string(output))
	}

	return nil
}

// Execute executes a command without timeout
func (pm *processManager) Execute(command string, args []string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %s failed: %w\nOutput: %s", command, err, string(output))
	}
	return nil
}

// ExecuteWithOutput executes a command and returns its output
func (pm *processManager) ExecuteWithOutput(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w\nOutput: %s", command, err, string(output))
	}
	return string(output), nil
}
