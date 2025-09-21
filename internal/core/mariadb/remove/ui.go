package remove

import (
	"fmt"

	"sfDBTools/utils/terminal"
)

// ui helpers centralize common formatted messages to reduce duplication
func info(msg string) {
	terminal.SafePrintln("   " + msg)
}

func infof(format string, a ...interface{}) {
	terminal.SafePrintln("   " + fmt.Sprintf(format, a...))
}

func warn(msg string) {
	terminal.SafePrintln("⚠️  " + msg)
}

func success(msg string) {
	terminal.SafePrintln("✓ " + msg)
}

func listHeader(title string) {
	terminal.SafePrintln(title)
}

// stepWithSpinner runs a step function while showing a spinner and prints concise result
// fn should return an error if the step failed
func stepWithSpinner(message string, fn func() error) error {
	spinner := terminal.NewProgressSpinner(message)
	spinner.Start()
	err := fn()
	spinner.Stop()

	if err != nil {
		warn(message + " gagal: " + err.Error())
		return err
	}

	success(message + " selesai")
	return nil
}
