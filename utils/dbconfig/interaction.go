package dbconfig

import (
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/utils/terminal"
)

// ConfirmDeletion prompts user for deletion confirmation
func ConfirmDeletion(deletionType DeletionType, items []string) bool {
	var message string

	switch deletionType {
	case DeletionSingle:
		if len(items) > 0 {
			message = fmt.Sprintf("Are you sure you want to delete configuration '%s'?", items[0])
		} else {
			message = "Are you sure you want to delete this configuration?"
		}
	case DeletionMultiple:
		message = fmt.Sprintf("Are you sure you want to delete %d selected configurations?", len(items))
	case DeletionAll:
		message = "Are you sure you want to delete ALL configurations? This action cannot be undone!"
	default:
		message = "Are you sure you want to proceed with this deletion?"
	}

	terminal.PrintWarning(message)
	if deletionType == DeletionAll {
		terminal.PrintError("This will permanently delete all database configurations!")
	}

	return terminal.AskYesNo("Confirm deletion", false)
}

// SelectFilesForDeletion prompts user to select files for deletion
func SelectFilesForDeletion(files []*FileInfo) ([]string, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no configuration files available")
	}

	fm := NewFileManager()
	fm.DisplayFileListSummary(files)

	terminal.PrintInfo("\nEnter file numbers to delete (comma-separated, e.g., 1,3,5):")
	input := terminal.AskString("File numbers", "")

	if input == "" {
		return nil, fmt.Errorf("no files selected")
	}

	selectedPaths, err := ParseFileSelections(input, files)
	if err != nil {
		return nil, fmt.Errorf("invalid selection: %v", err)
	}

	return selectedPaths, nil
}

// ParseFileSelections parses user input and returns selected file paths
func ParseFileSelections(input string, files []*FileInfo) ([]string, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("no selection provided")
	}

	selections := strings.Split(input, ",")
	var selectedPaths []string

	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}

		index, err := strconv.Atoi(sel)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", sel)
		}

		if index < 1 || index > len(files) {
			return nil, fmt.Errorf("number %d is out of range (1-%d)", index, len(files))
		}

		selectedPaths = append(selectedPaths, files[index-1].Path)
	}

	if len(selectedPaths) == 0 {
		return nil, fmt.Errorf("no valid files selected")
	}

	return selectedPaths, nil
}

// PromptForConfigName prompts user to enter or select a configuration name
func PromptForConfigName(availableConfigs []*FileInfo, operation OperationType) (string, error) {
	if len(availableConfigs) == 0 {
		return "", fmt.Errorf("no configuration files available for %s", operation.String())
	}

	if len(availableConfigs) == 1 {
		configName := availableConfigs[0].Name
		message := fmt.Sprintf("Use configuration '%s' for %s?", configName, operation.String())
		if terminal.AskYesNo(message, true) {
			return configName, nil
		}
		return "", fmt.Errorf("operation cancelled")
	}

	terminal.PrintSubHeader(fmt.Sprintf("📋 Available configurations for %s:", operation.String()))

	data := [][]string{}
	for i, config := range availableConfigs {
		status := "✅"
		if !config.IsValid {
			status = "❌"
		}

		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			config.Name,
			config.GetFormattedSize(),
			status,
		})
	}

	headers := []string{"#", "Name", "Size", "Status"}
	terminal.FormatTable(headers, data)

	for {
		input := terminal.AskString("Enter configuration number or name", "")
		if input == "" {
			return "", fmt.Errorf("no configuration selected")
		}

		// Try to parse as number first
		if index, err := strconv.Atoi(input); err == nil {
			if index >= 1 && index <= len(availableConfigs) {
				return availableConfigs[index-1].Name, nil
			}
			terminal.PrintError(fmt.Sprintf("Invalid number. Please enter 1-%d", len(availableConfigs)))
			continue
		}

		// Try to find by name
		for _, config := range availableConfigs {
			if config.Name == input {
				return config.Name, nil
			}
		}

		terminal.PrintError("Configuration not found. Please try again.")
	}
}

// ConfirmOverwrite prompts user to confirm file overwrite
func ConfirmOverwrite(filename string) bool {
	message := fmt.Sprintf("Configuration '%s' already exists. Overwrite?", filename)
	return terminal.AskYesNo(message, false)
}

// PromptForFileSelection prompts user to select configuration files
func PromptForFileSelection(fileCount int, operation string) (string, error) {
	message := fmt.Sprintf("Select configuration file(s) to %s (1-%d, comma-separated, or 'all')", operation, fileCount)
	choice := terminal.AskString(message, "")

	if choice == "" {
		return "", fmt.Errorf("no selection made")
	}

	return choice, nil
}
