package dbconfig

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/crypto"
	"sfDBTools/utils/terminal"
)

// BaseProcessor provides common functionality for all dbconfig processors
type BaseProcessor struct {
	logger *logger.Logger
}

// NewBaseProcessor creates a new base processor with logger
func NewBaseProcessor() (*BaseProcessor, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	return &BaseProcessor{
		logger: lg,
	}, nil
}

// LogOperation logs the start of an operation
func (bp *BaseProcessor) LogOperation(operation, details string) {
	bp.logger.Info(fmt.Sprintf("Starting %s", operation))
}

// GetEncryptionPassword prompts for encryption password with consistent messaging
func (bp *BaseProcessor) GetEncryptionPassword(purpose string) (string, error) {
	terminal.PrintSubHeader("Authentication Required")

	encryptionPassword, source, err := crypto.GetEncryptionPasswordWithSource("🔑 Encryption password: ")
	if err != nil {
		return "", fmt.Errorf("failed to get encryption password: %w", err)
	}

	if source == "env" {
		bp.logger.Warn(fmt.Sprintf("Encryption password for %s obtained from environment variable %s", purpose, crypto.ENV_ENCRYPTION_PASSWORD))
	} else {
		bp.logger.Info(fmt.Sprintf("Encryption password for %s obtained via user prompt", purpose))
	}

	return encryptionPassword, nil
}

// ConfirmEncryptionPassword prompts for confirmation of the encryption password
func (bp *BaseProcessor) ConfirmEncryptionPassword(purpose string) (string, error) {
	terminal.PrintSubHeader("Confirm Encryption Password")

	confirmPassword := terminal.AskPassword("Please re-enter the encryption password to confirm: ", "")

	return confirmPassword, nil
}

// HandleOperationResult displays operation result and handles common patterns
func (bp *BaseProcessor) HandleOperationResult(operation string, err error) error {
	if err != nil {
		bp.logger.Error(fmt.Sprintf("%s failed", operation))
		return err
	}

	terminal.PrintSuccess(fmt.Sprintf("%s completed successfully!", operation))
	return nil
}

// WaitForUserContinue prompts user to continue with consistent messaging
func (bp *BaseProcessor) WaitForUserContinue() {
	terminal.WaitForEnterWithMessage("\nPress Enter to continue...")
}
