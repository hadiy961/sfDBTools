package installcmd

import (
	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// PromptOverwrite asks if user wants to overwrite existing installation
func PromptOverwrite() bool {
	lg, _ := logger.Get()
	confirmed, err := terminal.ConfirmAndClear("Existing MariaDB installation detected. Do you want to continue?")
	if err != nil {
		lg.Warn("promptOverwrite failed, falling back to no", logger.Error(err))
		return false
	}
	return confirmed
}

// PromptCustomConfiguration asks user whether to run custom configuration now
func PromptCustomConfiguration() bool {
	confirmed, err := terminal.ConfirmAndClear("Would you like to customize MariaDB configuration now?")
	if err != nil {
		return true
	}
	return confirmed
}

// PromptConfirmInstall displays summary and asks user to proceed
func PromptConfirmInstall(version string) bool {
	terminal.PrintHeader("Install Summary")
	terminal.PrintInfo("Version:      " + version)

	confirmed, err := terminal.ConfirmAndClear("Proceed with installation?")
	if err != nil {
		return false
	}
	return confirmed
}
