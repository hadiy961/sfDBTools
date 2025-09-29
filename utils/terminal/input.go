package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// AskYesNo prompts user for yes/no input with default value
func AskYesNo(question string, defaultValue bool) bool {
	// Show default in brackets like AskString
	if defaultValue {
		fmt.Printf("%s [Y/n]: ", question)
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response == "" {
		return defaultValue
	}

	return response == "y" || response == "yes"
}

// AskString prompts user for string input with default value
func AskString(question, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.TrimSpace(scanner.Text())

	if response == "" {
		return defaultValue
	}

	return response
}

// AskInt prompts user for integer input with default value and validation.
// If the user enters an empty string, the defaultValue is returned.
// If the user enters a non-integer value, the prompt repeats until a valid integer
// or empty input is provided.
func AskInt(question string, defaultValue int) int {
	defaultStr := ""
	if defaultValue != 0 {
		defaultStr = fmt.Sprintf("%d", defaultValue)
	}

	for {
		if defaultStr != "" {
			fmt.Printf("%s [%s]: ", question, defaultStr)
		} else {
			fmt.Printf("%s: ", question)
		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.TrimSpace(scanner.Text())

		if response == "" {
			return defaultValue
		}

		// try parse int
		var v int
		_, err := fmt.Sscanf(response, "%d", &v)
		if err == nil {
			return v
		}

		// invalid int, show an error and repeat
		fmt.Println("Invalid integer, please try again.")
	}
}

// AskIntWithContext prompts for integer with additional help text
func AskIntWithContext(question, help string, defaultValue int) int {
	if help != "" {
		PrintInfo(help)
	}
	return AskInt(question, defaultValue)
}

// AskWithContext prompts user for yes/no input with additional context/help
func AskWithContext(question, help string, defaultValue bool) bool {
	if help != "" {
		PrintInfo(help)
	}
	return AskYesNo(question, defaultValue)
}

// AskPassword prompts user for a password with input masking. If a non-empty
// defaultValue is provided and the user presses Enter without typing any
// characters, the defaultValue will be returned. In case reading the password
// with masking fails, it falls back to AskString (unmasked) to remain usable
// in environments where terminal masking is unsupported.
func AskPassword(question, defaultValue string) string {
	// Show hint that default exists but do not display the default itself
	if defaultValue != "" {
		fmt.Printf("%s [hidden]: ", question)
	} else {
		fmt.Printf("%s: ", question)
	}

	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		// Fallback to unmasked input if terminal masking isn't available
		return AskString(question, defaultValue)
	}

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		return defaultValue
	}
	return password
}
