package dbconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmSingleDeletion prompts for confirmation to delete a single file
func ConfirmSingleDeletion(filename string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete '%s'? This action cannot be undone. [y/N]: ", filename)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}

// ConfirmMultipleDeletion prompts for confirmation to delete multiple files
func ConfirmMultipleDeletion(count int) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete %d config file(s)? This action cannot be undone. [y/N]: ", count)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}

// ConfirmAllDeletion prompts for confirmation to delete all files
func ConfirmAllDeletion(count int) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Are you sure you want to delete ALL %d encrypted configuration files? This action cannot be undone. [y/N]: ", count)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	return confirmation == "y" || confirmation == "yes"
}

// ConfirmSaveChanges prompts for confirmation to save changes
func ConfirmSaveChanges() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nSave these changes? [Y/n]: ")
	confirmInput, _ := reader.ReadString('\n')
	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	return confirm == "" || confirm == "y" || confirm == "yes"
}

// ParseFileSelections parses comma-separated file selections from user input
func ParseFileSelections(choice string, encFiles []string) ([]string, error) {
	if choice == "" {
		return nil, fmt.Errorf("no selection made")
	}

	selections := strings.Split(choice, ",")
	var selectedFiles []string

	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}

		// Try to parse as number
		if index := parseIndex(sel, len(encFiles)); index >= 0 {
			selectedFiles = append(selectedFiles, encFiles[index])
		} else {
			return nil, fmt.Errorf("invalid selection: %s", sel)
		}
	}

	if len(selectedFiles) == 0 {
		return nil, fmt.Errorf("no valid selections made")
	}

	return selectedFiles, nil
}

// parseIndex parses a string index and returns the corresponding array index (-1 if invalid)
func parseIndex(s string, maxLen int) int {
	if idx := parseInt(s); idx >= 1 && idx <= maxLen {
		return idx - 1 // Convert to 0-based index
	}
	return -1
}

// parseInt safely parses an integer
func parseInt(s string) int {
	if i := 0; len(s) > 0 {
		for _, char := range s {
			if char < '0' || char > '9' {
				return -1
			}
			i = i*10 + int(char-'0')
		}
		return i
	}
	return -1
}

// PromptForSelection prompts user to select configuration files
func PromptForSelection(fileCount int) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect configuration file(s) to delete (1-%d, comma-separated, or 'all'): ", fileCount)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}
	return strings.TrimSpace(choice), nil
}
