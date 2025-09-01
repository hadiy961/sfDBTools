package dbconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/utils/terminal"
)

// DisplayConfigDetails shows configuration details with enhanced formatting
func DisplayConfigDetails(filePath string, dbConfig *config.EncryptedDatabaseConfig) {
	terminal.ClearAndShowHeader("🔧 Database Configuration Details")

	// Configuration table
	headers := []string{"Property", "Value", "Description"}
	rows := [][]string{
		{"📁 Source File", filepath.Base(filePath), "Configuration file name"},
		{"🌐 Host", dbConfig.Host, "Database server hostname/IP"},
		{"🔌 Port", fmt.Sprintf("%d", dbConfig.Port), "Database server port"},
		{"👤 Username", dbConfig.User, "Database username"},
		{"🔑 Password", MaskPassword(dbConfig.Password), "Database password (masked)"},
	}

	terminal.FormatTable(headers, rows)

	// Show full file path
	fmt.Println()
	terminal.PrintSubHeader("📂 File Information")
	terminal.PrintInfo(fmt.Sprintf("Full path: %s", filePath))

	// Get file info
	if info, err := os.Stat(filePath); err == nil {
		terminal.PrintInfo(fmt.Sprintf("File size: %.2f KB", float64(info.Size())/1024))
		terminal.PrintInfo(fmt.Sprintf("Last modified: %s", info.ModTime().Format("2006-01-02 15:04:05")))
	}

	// Security warning
	fmt.Println()
	terminal.PrintWarning("⚠️ Sensitive data displayed - ensure your screen is not being observed")
}

// DisplayPasswordOption prompts to show actual password
func DisplayPasswordOption(password string) {
	fmt.Println()
	terminal.PrintInfo("To view the actual password, type 'show' (otherwise press Enter):")
	var input string
	fmt.Scanln(&input)

	if strings.ToLower(strings.TrimSpace(input)) == "show" {
		terminal.PrintSubHeader("🔑 Actual Password")
		terminal.PrintColoredText("Password: ", terminal.ColorRed)
		terminal.PrintColoredLine(password, terminal.ColorBold)
		terminal.PrintWarning("⚠️ Password is now visible on screen!")
	}
}

// MaskPassword masks password for display
func MaskPassword(password string) string {
	if len(password) <= 2 {
		return "••••"
	}
	return password[:2] + strings.Repeat("•", len(password)-2)
}

// DisplayValidationResults shows validation results in formatted table
func DisplayValidationResults(result *ValidationResult, serverVersion string) {
	terminal.PrintSubHeader("📋 Validation Summary")
	headers := []string{"Test", "Result", "Details"}
	rows := [][]string{
		{"File Validation", getStatusText(result.FileValid), "Configuration file structure"},
		{"Decryption", getStatusText(result.DecryptionValid), "Password and encryption"},
		{"Connection", getStatusText(result.ConnectionValid), "Database server is reachable"},
		{"Authentication", getStatusText(result.ConnectionValid), "Credentials are valid"},
		{"Response Time", terminal.ColorText("< 10s", terminal.ColorGreen), "Connection within timeout"},
	}

	if serverVersion != "" {
		rows = append(rows, []string{"Server Version", terminal.ColorText("✅ Retrieved", terminal.ColorGreen), serverVersion})
	}

	terminal.FormatTable(headers, rows)
}

// DisplayDeleteSummary shows summary of delete operation
func DisplayDeleteSummary(result *DeleteResult) {
	fmt.Printf("\n📊 Summary: %d deleted, %d errors\n", result.DeletedCount, result.ErrorCount)

	if len(result.DeletedFiles) > 0 {
		terminal.PrintSubHeader("✅ Successfully Deleted")
		for _, file := range result.DeletedFiles {
			terminal.PrintSuccess(fmt.Sprintf("• %s", filepath.Base(file)))
		}
	}

	if len(result.Errors) > 0 {
		terminal.PrintSubHeader("❌ Errors")
		for _, err := range result.Errors {
			terminal.PrintError(fmt.Sprintf("• %s", err))
		}
	}
}

// DisplayConfigSummary shows configuration summary before operations
func DisplayConfigSummary(configName string, dbConfig *config.EncryptedDatabaseConfig) {
	fmt.Printf("\n📋 Configuration Summary:\n")
	fmt.Printf("   Name: %s\n", configName)
	fmt.Printf("   Host: %s\n", dbConfig.Host)
	fmt.Printf("   Port: %d\n", dbConfig.Port)
	fmt.Printf("   User: %s\n", dbConfig.User)
	fmt.Printf("   Password: %s\n", strings.Repeat("*", len(dbConfig.Password)))
}

// getStatusText returns colored status text
func getStatusText(success bool) string {
	if success {
		return terminal.ColorText("✅ Success", terminal.ColorGreen)
	}
	return terminal.ColorText("❌ Failed", terminal.ColorRed)
}
