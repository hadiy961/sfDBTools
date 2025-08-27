package configure

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"
)

// UserPrompt handles user input for MariaDB configuration
type UserPrompt struct {
	defaults *MariaDBSettings
}

// NewUserPrompt creates a new user prompt handler
func NewUserPrompt(defaults *MariaDBSettings) *UserPrompt {
	return &UserPrompt{
		defaults: defaults,
	}
}

// PromptForSettings prompts user for MariaDB configuration settings
func (p *UserPrompt) PromptForSettings(autoConfirm bool) (*MariaDBSettings, error) {
	lg, _ := logger.Get()

	if autoConfirm {
		lg.Info("Using default MariaDB settings (auto-confirm mode)")
		terminal.PrintInfo("Using default MariaDB configuration settings")
		return p.defaults, nil
	}

	terminal.PrintInfo("MariaDB Configuration Setup")
	terminal.PrintInfo("Please provide the following configuration values:")
	fmt.Println()

	settings := &MariaDBSettings{}
	reader := bufio.NewReader(os.Stdin)

	// Server ID
	serverID, err := p.promptForString(reader, "Server ID", p.defaults.ServerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get server ID: %w", err)
	}
	settings.ServerID = serverID

	// Data Directory (must not be /var/lib/mysql)
	for {
		dataDir, err := p.promptForString(reader, "Data Directory", p.defaults.DataDir, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get data directory: %w", err)
		}

		if dataDir == "/var/lib/mysql" {
			terminal.PrintError("Data directory cannot be /var/lib/mysql. Please choose a different directory.")
			continue
		}

		settings.DataDir = dataDir
		break
	}

	// Binlog Directory
	binlogDir, err := p.promptForString(reader, "Binlog Directory", p.defaults.BinlogDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get binlog directory: %w", err)
	}
	settings.BinlogDir = binlogDir

	// Log Directory
	logDir, err := p.promptForString(reader, "Log Directory", p.defaults.LogDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get log directory: %w", err)
	}
	settings.LogDir = logDir

	// Port
	portStr, err := p.promptForString(reader, "Port", strconv.Itoa(p.defaults.Port), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}
	settings.Port = port

	// Encryption settings
	enableEncryption, err := p.promptForBool(reader, "Enable encryption", p.defaults.EncryptionEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption setting: %w", err)
	}
	settings.EncryptionEnabled = enableEncryption

	if enableEncryption {
		keyFile, err := p.promptForString(reader, "Encryption key file", p.defaults.FileKeyManagementFile, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get encryption key file: %w", err)
		}
		settings.FileKeyManagementFile = keyFile
	}

	lg.Info("User configuration collected",
		logger.String("server_id", settings.ServerID),
		logger.String("data_dir", settings.DataDir),
		logger.String("binlog_dir", settings.BinlogDir),
		logger.String("log_dir", settings.LogDir),
		logger.Int("port", settings.Port))

	return settings, nil
}

// promptForString prompts for a string value with default
func (p *UserPrompt) promptForString(reader *bufio.Reader, prompt, defaultValue string, allowEmpty bool) (string, error) {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			if defaultValue != "" {
				return defaultValue, nil
			}
			if !allowEmpty {
				terminal.PrintError("This field is required. Please provide a value.")
				continue
			}
		}

		return input, nil
	}
}

// promptForBool prompts for a boolean value with default
func (p *UserPrompt) promptForBool(reader *bufio.Reader, prompt string, defaultValue bool) (bool, error) {
	defaultStr := "y"
	if !defaultValue {
		defaultStr = "n"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultValue, nil
		}

		switch input {
		case "y", "yes", "true":
			return true, nil
		case "n", "no", "false":
			return false, nil
		default:
			terminal.PrintError("Please enter 'y' for yes or 'n' for no.")
		}
	}
}
