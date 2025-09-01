package dbconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/utils/terminal"
)

// ValidateConfigFile validates configuration file format and structure
func ValidateConfigFile(filePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		ConfigName:  filepath.Base(filePath),
		TestResults: make(map[string]bool),
		Errors:      []string{},
		Warnings:    []string{},
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, "Configuration file does not exist")
		return result, nil
	}
	result.TestResults["file_exists"] = true

	// Check file extension
	if !strings.HasSuffix(filePath, ".cnf.enc") {
		result.Warnings = append(result.Warnings, "Configuration file should have .cnf.enc extension")
	} else {
		result.TestResults["correct_extension"] = true
	}

	// Check file permissions
	if info, err := os.Stat(filePath); err == nil {
		mode := info.Mode()
		if mode.Perm() > 0o600 {
			result.Warnings = append(result.Warnings, "Configuration file has loose permissions, consider setting to 600")
		} else {
			result.TestResults["secure_permissions"] = true
		}
	}

	// Check if file is encrypted (simple check - files should have minimum size for encrypted content)
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() > 32 { // Minimum size for encrypted file (nonce + tag)
			result.TestResults["encrypted"] = true
		} else {
			result.Warnings = append(result.Warnings, "Configuration file too small to be properly encrypted")
		}
	}

	// Check file size
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() == 0 {
			result.Errors = append(result.Errors, "Configuration file is empty")
		} else if info.Size() > 10*1024*1024 { // 10MB
			result.Warnings = append(result.Warnings, "Configuration file is unusually large")
		} else {
			result.TestResults["reasonable_size"] = true
		}
	}

	result.IsValid = len(result.Errors) == 0
	return result, nil
}

// ValidateConfigStructure validates the structure of decrypted config
func ValidateConfigStructure(content string) (*ValidationResult, error) {
	result := &ValidationResult{
		TestResults: make(map[string]bool),
		Errors:      []string{},
		Warnings:    []string{},
	}

	lines := strings.Split(content, "\n")
	hasRequiredFields := false
	requiredFields := map[string]bool{
		"host":     false,
		"user":     false,
		"password": false,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Check required fields
				if _, required := requiredFields[key]; required {
					if value != "" {
						requiredFields[key] = true
					}
				}

				// Validate specific fields
				switch key {
				case "port":
					if value != "" {
						result.TestResults["has_port"] = true
					}
				case "socket":
					if value != "" {
						result.TestResults["has_socket"] = true
					}
				}
			}
		}
	}

	// Check required fields
	for field, found := range requiredFields {
		if !found {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing required field: %s", field))
		} else {
			hasRequiredFields = true
		}
	}

	if hasRequiredFields {
		result.TestResults["has_required_fields"] = true
	}

	result.IsValid = len(result.Errors) == 0
	return result, nil
}

// DisplayValidationResult shows validation result in formatted way
func DisplayValidationResult(result *ValidationResult) {
	terminal.PrintSubHeader(fmt.Sprintf("🔍 Validation Results: %s", result.ConfigName))

	if result.IsValid {
		terminal.PrintSuccess("✅ Configuration is valid")
	} else {
		terminal.PrintError("❌ Configuration has errors")
	}

	// Display test results
	if len(result.TestResults) > 0 {
		terminal.PrintInfo("\n📋 Test Results:")

		testData := [][]string{}
		for test, passed := range result.TestResults {
			status := "❌ Failed"
			if passed {
				status = "✅ Passed"
			}
			testData = append(testData, []string{formatTestName(test), status})
		}

		headers := []string{"Test", "Status"}
		terminal.FormatTable(headers, testData)
	}

	// Display errors
	if len(result.Errors) > 0 {
		terminal.PrintError("\n🚨 Errors:")
		for i, err := range result.Errors {
			terminal.PrintError(fmt.Sprintf("  %d. %s", i+1, err))
		}
	}

	// Display warnings
	if len(result.Warnings) > 0 {
		terminal.PrintWarning("\n⚠️ Warnings:")
		for i, warning := range result.Warnings {
			terminal.PrintWarning(fmt.Sprintf("  %d. %s", i+1, warning))
		}
	}
}

// formatTestName converts test key to human readable format
func formatTestName(testKey string) string {
	replacements := map[string]string{
		"file_exists":         "File Exists",
		"correct_extension":   "Correct Extension",
		"secure_permissions":  "Secure Permissions",
		"encrypted":           "File Encrypted",
		"reasonable_size":     "Reasonable Size",
		"has_required_fields": "Required Fields",
		"has_port":            "Port Specified",
		"has_socket":          "Socket Specified",
	}

	if formatted, exists := replacements[testKey]; exists {
		return formatted
	}
	return strings.Title(strings.ReplaceAll(testKey, "_", " "))
}

// ValidateMultipleConfigs validates multiple configuration files
func ValidateMultipleConfigs(filePaths []string) ([]*ValidationResult, error) {
	results := make([]*ValidationResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		result, err := ValidateConfigFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error validating %s: %v", filePath, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// DisplayMultipleValidationResults shows summary of multiple validations
func DisplayMultipleValidationResults(results []*ValidationResult) {
	terminal.PrintHeader("📊 Validation Summary")

	validCount := 0
	invalidCount := 0

	summaryData := [][]string{}

	for _, result := range results {
		status := "❌ Invalid"
		if result.IsValid {
			status = "✅ Valid"
			validCount++
		} else {
			invalidCount++
		}

		errorCount := len(result.Errors)
		warningCount := len(result.Warnings)

		summaryData = append(summaryData, []string{
			result.ConfigName,
			status,
			fmt.Sprintf("%d", errorCount),
			fmt.Sprintf("%d", warningCount),
		})
	}

	headers := []string{"Configuration", "Status", "Errors", "Warnings"}
	terminal.FormatTable(headers, summaryData)

	// Overall summary
	total := len(results)
	terminal.PrintInfo(fmt.Sprintf("\n📈 Overall: %d/%d valid configurations", validCount, total))

	if invalidCount > 0 {
		terminal.PrintWarning(fmt.Sprintf("❗ %d configurations need attention", invalidCount))
	}
}
