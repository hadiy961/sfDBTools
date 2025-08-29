package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskYesNo prompts user for yes/no input with default value
func AskYesNo(question string, defaultValue bool) bool {
	if defaultValue {
		fmt.Printf("%s (Y/n): ", question)
	} else {
		fmt.Printf("%s (y/N): ", question)
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

// AskWithContext prompts user for yes/no input with additional context/help
func AskWithContext(question, help string, defaultValue bool) bool {
	if help != "" {
		PrintInfo(help)
	}
	return AskYesNo(question, defaultValue)
}
