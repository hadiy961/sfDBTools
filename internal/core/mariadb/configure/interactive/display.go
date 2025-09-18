package interactive

import (
	"fmt"

	"sfDBTools/utils/terminal"
)

// RequestUserConfirmation meminta konfirmasi user untuk menerapkan konfigurasi - Task 2: modular function
func RequestUserConfirmation() error {
	terminal.PrintWarning("The above configuration will be applied to your MariaDB server.")
	terminal.PrintWarning("This may require stopping the service and migrating data.")

	question := "Do you want to proceed?"
	confirmed := terminal.AskYesNo(question, false)

	if !confirmed {
		return fmt.Errorf("configuration cancelled by user")
	}

	return nil
}
