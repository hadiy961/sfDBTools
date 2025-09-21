package dbconfig

import (
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/utils/terminal"
)

// InputConfig represents configuration input data
type InputConfig struct {
	Name     string
	Host     string
	Port     int
	User     string
	Password string
}

// PromptConfigName prompts for configuration name with validation
func PromptConfigName(defaultName string) (string, error) {
	if defaultName == "" {
		defaultName = "database"
	}

	terminal.PrintSubHeader("📁 Configuration File Name")
	name := terminal.AskString("Enter configuration name (without extension)", defaultName)

	// Validate and sanitize filename
	name = sanitizeFileName(name)

	terminal.PrintInfo(fmt.Sprintf("Configuration will be saved as: %s.cnf.enc", name))
	return name, nil
}

// PromptDatabaseConfig prompts for complete database configuration
func PromptDatabaseConfig() (*InputConfig, error) {
	terminal.PrintSubHeader("Database Configuration")
	terminal.PrintInfo("Please provide database connection details:")

	config := &InputConfig{}

	// Prompt for host
	config.Host = terminal.AskString("Enter database host", "localhost")

	// Prompt for port with validation
	for {
		portStr := terminal.AskString("Enter database port", "3306")
		if port, err := strconv.Atoi(portStr); err == nil && port >= 1 && port <= 65535 {
			config.Port = port
			break
		} else {
			terminal.PrintError("Invalid port number. Please enter a value between 1 and 65535.")
		}
	}

	// Prompt for username
	for {
		config.User = terminal.AskString("Enter database username", "")
		if config.User != "" {
			break
		}
		terminal.PrintError("Username cannot be empty.")
	}

	return config, nil
}

// PromptConfigEdit prompts for editing existing configuration
func PromptConfigEdit(current *InputConfig, currentName string) (*InputConfig, string, bool, error) {
	terminal.PrintSubHeader("✏️ Edit Configuration")
	terminal.PrintInfo("Press Enter to keep current value, or type new value:")

	// Edit configuration name
	newName := terminal.AskString(fmt.Sprintf("Configuration name [%s]", currentName), currentName)

	// Edit other fields
	newConfig := &InputConfig{
		Host:     terminal.AskString(fmt.Sprintf("Host [%s]", current.Host), current.Host),
		User:     terminal.AskString(fmt.Sprintf("User [%s]", current.User), current.User),
		Password: current.Password, // Password will be handled separately
	}

	// Edit port with validation
	for {
		portStr := terminal.AskString(fmt.Sprintf("Port [%d]", current.Port), fmt.Sprintf("%d", current.Port))
		if port, err := strconv.Atoi(portStr); err == nil && port >= 1 && port <= 65535 {
			newConfig.Port = port
			break
		} else {
			terminal.PrintError("Invalid port number. Please enter a value between 1 and 65535.")
		}
	}

	// Prompt for password (optional change)
	changePassword := terminal.AskYesNo("Change password?", false)
	if changePassword {
		newConfig.Password = terminal.AskString("Enter new password", "")
	}

	// Check if any changes were made
	hasChanges := newName != currentName ||
		newConfig.Host != current.Host ||
		newConfig.Port != current.Port ||
		newConfig.User != current.User ||
		(changePassword && newConfig.Password != current.Password)

	return newConfig, newName, hasChanges, nil
}

// DisplayChangesSummary shows what will be changed
func DisplayChangesSummary(current *InputConfig, new *InputConfig, currentName, newName string) {
	terminal.PrintSubHeader("📋 Changes Summary")

	changes := [][]string{}

	if newName != currentName {
		changes = append(changes, []string{"Name", currentName, newName})
	}
	if new.Host != current.Host {
		changes = append(changes, []string{"Host", current.Host, new.Host})
	}
	if new.Port != current.Port {
		changes = append(changes, []string{"Port", fmt.Sprintf("%d", current.Port), fmt.Sprintf("%d", new.Port)})
	}
	if new.User != current.User {
		changes = append(changes, []string{"User", current.User, new.User})
	}
	if new.Password != current.Password {
		changes = append(changes, []string{"Password", strings.Repeat("*", len(current.Password)), strings.Repeat("*", len(new.Password))})
	}

	if len(changes) == 0 {
		terminal.PrintInfo("No changes made.")
		return
	}

	headers := []string{"Property", "Current", "New"}
	terminal.FormatTable(headers, changes)
}

// sanitizeFileName removes invalid characters from filename
func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "database"
	}

	// Replace invalid characters
	replacements := map[string]string{
		" ":  "_",
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	for old, new := range replacements {
		name = strings.ReplaceAll(name, old, new)
	}

	return name
}
