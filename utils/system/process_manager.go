package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ProcessManager interface provides abstraction for process execution
type ProcessManager interface {
	ExecuteWithTimeout(command string, args []string, timeout time.Duration) error
	Execute(command string, args []string) error
	ExecuteWithOutput(command string, args []string) (string, error)
	// ExecuteInteractiveWithTimeout runs a command connected to the current process's
	// stdin/stdout/stderr so the user can interact with it. The command is killed when
	// the timeout expires.
	ExecuteInteractiveWithTimeout(command string, args []string, timeout time.Duration) error
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

// ExecuteInteractiveWithTimeout runs a command with stdin/stdout/stderr attached to the
// current process. This allows interactive tools (prompts) to be used. The command
// is run with a context that times out after the provided duration.
func (pm *processManager) ExecuteInteractiveWithTimeout(command string, args []string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command %s timed out after %v", command, timeout)
		}
		return fmt.Errorf("command %s failed: %w", command, err)
	}
	return nil
}
