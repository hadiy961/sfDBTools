package dbconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/utils/terminal"
)

// DisplayConfigDetails shows configuration file details
func DisplayConfigDetails(configName, configPath string) error {
	terminal.PrintHeader(fmt.Sprintf("📋 Configuration Details: %s", configName))

	// File information
	if stat, err := os.Stat(configPath); err == nil {
		data := [][]string{
			{"Name", configName},
			{"File Path", configPath},
			{"Size", formatFileSize(stat.Size())},
			{"Last Modified", stat.ModTime().Format("2006-01-02 15:04:05")},
			{"Permissions", stat.Mode().String()},
		}

		headers := []string{"Property", "Value"}
		terminal.FormatTable(headers, data)
	} else {
		terminal.PrintError(fmt.Sprintf("Error reading file info: %v", err))
		return err
	}

	return nil
}

// DisplayConfigSummary shows a summary table of configurations
func DisplayConfigSummary(configs []*ConfigInfo) {
	if len(configs) == 0 {
		terminal.PrintWarning("No configuration files found.")
		return
	}

	terminal.PrintHeader(fmt.Sprintf("📁 Database Configurations (%d found)", len(configs)))

	data := [][]string{}
	for i, config := range configs {
		status := "✅ Valid"
		if !config.IsValid {
			status = "❌ Invalid"
		}

		passwordStatus := "❌ No"
		if config.HasPassword {
			passwordStatus = "✅ Yes"
		}

		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			config.Name,
			fmt.Sprintf("%s:%d", config.Host, config.Port),
			config.User,
			passwordStatus,
			config.FileSize,
			config.LastModified.Format("2006-01-02"),
			status,
		})
	}

	headers := []string{"#", "Name", "Host:Port", "User", "Password", "Size", "Modified", "Status"}
	terminal.FormatTable(headers, data)
}

// DisplayPasswordOption prompts for password handling option
func DisplayPasswordOption() (string, error) {
	terminal.PrintSubHeader("🔑 Password Configuration")

	options := []string{
		"1. Enter password now",
		"2. Use environment variable",
		"3. Skip password (enter manually when needed)",
	}

	for _, option := range options {
		terminal.PrintInfo(option)
	}

	choice := terminal.AskString("Select password option (1-3)", "1")

	switch choice {
	case "1":
		return "manual", nil
	case "2":
		return "env", nil
	case "3":
		return "skip", nil
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

// DisplayValidationResults shows the result of configuration validation
func DisplayValidationResults(result *ValidationResult, serverVersion string) {
	terminal.PrintSubHeader("📋 Validation Summary")

	if result.IsValid {
		terminal.PrintSuccess("✅ Configuration is valid")
	} else {
		terminal.PrintError("❌ Configuration has errors")
	}

	// Display test results
	if len(result.TestResults) > 0 {
		headers := []string{"Test", "Result", "Details"}
		rows := [][]string{}

		for test, passed := range result.TestResults {
			status := getStatusText(passed)
			description := getTestDescription(test)
			rows = append(rows, []string{formatTestName(test), status, description})
		}

		terminal.FormatTable(headers, rows)
	}

	// Display errors if any
	if len(result.Errors) > 0 {
		terminal.PrintError("\n🚨 Errors:")
		for i, err := range result.Errors {
			terminal.PrintError(fmt.Sprintf("  %d. %s", i+1, err))
		}
	}

	// Display warnings if any
	if len(result.Warnings) > 0 {
		terminal.PrintWarning("\n⚠️ Warnings:")
		for i, warning := range result.Warnings {
			terminal.PrintWarning(fmt.Sprintf("  %d. %s", i+1, warning))
		}
	}

	if serverVersion != "" {
		terminal.PrintInfo(fmt.Sprintf("\n🗃️ Server Version: %s", serverVersion))
	}
}

// DisplayDeleteSummary shows summary of delete operation
func DisplayDeleteSummary(result *DeleteResult) {
	terminal.PrintSubHeader("📊 Delete Operation Summary")

	if result.DeletedCount > 0 {
		terminal.PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d configuration(s)", result.DeletedCount))

		if len(result.DeletedFiles) > 0 {
			terminal.PrintInfo("\nDeleted files:")
			for _, file := range result.DeletedFiles {
				terminal.PrintInfo(fmt.Sprintf("  • %s", file))
			}
		}
	}

	if result.ErrorCount > 0 {
		terminal.PrintError(fmt.Sprintf("\n❌ %d error(s) occurred", result.ErrorCount))

		if len(result.Errors) > 0 {
			terminal.PrintError("\nErrors:")
			for _, err := range result.Errors {
				terminal.PrintError(fmt.Sprintf("  • %s", err))
			}
		}
	}

	if result.DeletedCount == 0 && result.ErrorCount == 0 {
		terminal.PrintWarning("No files were processed")
	}
}

// DisplayFileList shows a numbered list of files
func DisplayFileList(files []string) {
	if len(files) == 0 {
		terminal.PrintWarning("No files available.")
		return
	}

	terminal.PrintSubHeader(fmt.Sprintf("📁 Available Files (%d)", len(files)))

	for i, file := range files {
		filename := filepath.Base(file)
		filename = strings.TrimSuffix(filename, ".cnf.enc")
		terminal.PrintInfo(fmt.Sprintf("%d. %s", i+1, filename))
	}
}

// formatFileSize returns human-readable file size
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

// getStatusText returns colored status text
func getStatusText(success bool) string {
	if success {
		return terminal.ColorText("✅ Passed", terminal.ColorGreen)
	}
	return terminal.ColorText("❌ Failed", terminal.ColorRed)
}

// getTestDescription returns description for test type
func getTestDescription(testKey string) string {
	descriptions := map[string]string{
		"file_exists":         "Configuration file structure",
		"correct_extension":   "File naming convention",
		"secure_permissions":  "File security",
		"encrypted":           "Data encryption",
		"reasonable_size":     "File size validation",
		"has_required_fields": "Required configuration fields",
		"has_port":            "Database port configuration",
		"has_socket":          "Socket configuration",
	}

	if desc, exists := descriptions[testKey]; exists {
		return desc
	}
	return "Validation check"
}
