package generate

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/utils/crypto"
)

// promptDatabaseConfig prompts the user for database configuration
func PromptDatabaseConfig() (*EncryptedDatabaseConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	config := &EncryptedDatabaseConfig{}

	fmt.Println("\n📋 Database Configuration")
	fmt.Println("========================")

	// Prompt for host
	fmt.Print("Enter database host [localhost]: ")
	host, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read host: %w", err)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	config.Host = host

	// Prompt for port
	fmt.Print("Enter database port [3306]: ")
	portStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read port: %w", err)
	}
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		config.Port = 3306
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %w", err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port number must be between 1 and 65535")
		}
		config.Port = port
	}

	// Prompt for username
	fmt.Print("Enter database username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	config.User = username

	// Get database password from environment variable or prompt
	password, err := crypto.GetDatabasePassword("Enter database password: ")
	if err != nil {
		return nil, fmt.Errorf("failed to get database password: %w", err)
	}
	config.Password = password

	// Confirm configuration
	fmt.Printf("\n📋 Configuration Summary:\n")
	fmt.Printf("   Host: %s\n", config.Host)
	fmt.Printf("   Port: %d\n", config.Port)
	fmt.Printf("   User: %s\n", config.User)
	fmt.Printf("   Password: %s\n", strings.Repeat("*", len(config.Password)))

	fmt.Print("\nSave this configuration? [Y/n]: ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "" && confirm != "y" && confirm != "yes" {
		return nil, fmt.Errorf("configuration generation cancelled")
	}

	return config, nil
}

// PromptConfigName prompts the user for configuration name
func PromptConfigName() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n📁 Configuration File Name")
	fmt.Println("==========================")
	fmt.Print("Enter configuration name (without extension) [database]: ")

	name, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read configuration name: %w", err)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "database"
	}

	// Validate filename (remove invalid characters)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	fmt.Printf("Configuration will be saved as: %s.cnf.enc\n", name)

	return name, nil
}
