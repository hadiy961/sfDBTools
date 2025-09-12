package configure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
)

// MariaDBConfigTemplate berisi template konfigurasi MariaDB
type MariaDBConfigTemplate struct {
	TemplatePath  string            `json:"template_path"`
	Content       string            `json:"content"`
	Placeholders  map[string]string `json:"placeholders"`
	DefaultValues map[string]string `json:"default_values"`
	CurrentConfig string            `json:"current_config"`
	CurrentPath   string            `json:"current_path"`
}

// loadConfigurationTemplateWithInstallation memuat template konfigurasi MariaDB
// menggunakan data installation yang sudah ada untuk menghindari duplikasi discovery
// Sesuai dengan Step 2-4 dalam flow implementasi
func loadConfigurationTemplateWithInstallation(ctx context.Context, installation *mariadb_utils.MariaDBInstallation) (*MariaDBConfigTemplate, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Loading MariaDB configuration template")

	template := &MariaDBConfigTemplate{
		Placeholders:  make(map[string]string),
		DefaultValues: make(map[string]string),
	}

	// Step 2: Cari template konfigurasi di /etc/sfDBTools/server.cnf
	templatePath := "/etc/sfDBTools/server.cnf"
	if err := loadTemplateFile(template, templatePath); err != nil {
		return nil, fmt.Errorf("failed to load template file: %w", err)
	}

	// Step 3: Gunakan hasil discovery yang sudah ada untuk mencari config file
	currentConfigPath, err := findCurrentConfigFileFromInstallation(installation)
	if err != nil {
		lg.Warn("Failed to find current config file, will use default", logger.Error(err))
		template.CurrentPath = "/etc/my.cnf.d/50-server.cnf" // default fallback
	} else {
		template.CurrentPath = currentConfigPath
		lg.Info("Found current config file", logger.String("path", currentConfigPath))
	}

	// Step 4: Baca konfigurasi saat ini jika ada
	if err := loadCurrentConfig(template); err != nil {
		lg.Warn("Failed to load current config, using template defaults", logger.Error(err))
	}

	// Set default values dari template
	parsePlaceholders(template)
	setDefaultValues(template)

	lg.Info("Configuration template loaded successfully",
		logger.String("template_path", template.TemplatePath),
		logger.String("current_config_path", template.CurrentPath),
		logger.Int("placeholders", len(template.Placeholders)),
	)

	return template, nil
}

// loadTemplateFile memuat file template dari disk
func loadTemplateFile(template *MariaDBConfigTemplate, templatePath string) error {
	// Cek apakah file template ada
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s. Please ensure template is installed", templatePath)
	}

	// Baca content template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	template.TemplatePath = templatePath
	template.Content = string(content)

	// Parse placeholders dari template
	if err := parsePlaceholders(template); err != nil {
		return fmt.Errorf("failed to parse template placeholders: %w", err)
	}

	return nil
}

// findCurrentConfigFileFromInstallation mencari file konfigurasi MariaDB dari data installation yang sudah ada
func findCurrentConfigFileFromInstallation(installation *mariadb_utils.MariaDBInstallation) (string, error) {
	// Jika discovery menemukan config files, ambil yang pertama
	if len(installation.ConfigPaths) > 0 {
		return installation.ConfigPaths[0], nil
	}

	// Fallback ke lokasi standar
	standardPaths := []string{
		"/etc/my.cnf.d/50-server.cnf",
		"/etc/my.cnf.d/server.cnf",
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no MariaDB configuration file found in standard locations")
}

// loadCurrentConfig membaca konfigurasi MariaDB saat ini
func loadCurrentConfig(template *MariaDBConfigTemplate) error {
	if template.CurrentPath == "" {
		return fmt.Errorf("no current config path specified")
	}

	// Cek apakah file ada
	if _, err := os.Stat(template.CurrentPath); os.IsNotExist(err) {
		return fmt.Errorf("current config file does not exist: %s", template.CurrentPath)
	}

	// Baca content
	content, err := os.ReadFile(template.CurrentPath)
	if err != nil {
		return fmt.Errorf("failed to read current config file %s: %w", template.CurrentPath, err)
	}

	template.CurrentConfig = string(content)
	return nil
}

// parsePlaceholders mengurai placeholder dari template
func parsePlaceholders(template *MariaDBConfigTemplate) error {
	lines := strings.Split(template.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip komentar dan baris kosong
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Cari format key = value
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Jika value berupa placeholder {{VALUE}}
				if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
					placeholder := strings.Trim(value, "{}")
					template.Placeholders[key] = placeholder
				} else {
					// Value biasa, simpan sebagai default
					template.DefaultValues[key] = value
				}
			}
		}
	}

	return nil
}

// setDefaultValues mengatur nilai default untuk template
func setDefaultValues(template *MariaDBConfigTemplate) {
	// Default values sesuai spesifikasi
	defaults := map[string]string{
		"server_id": "1",
		"file_key_management_encryption_algorithm": "AES_CTR",
		"file_key_management_filename":             "/var/lib/mysql/encryption/keyfile",
		"innodb-encrypt-tables":                    "ON",
		"log_bin":                                  "/var/lib/mysqlbinlogs/mysql-bin",
		"datadir":                                  "/var/lib/mysql",
		"socket":                                   "/var/lib/mysql/mysql.sock",
		"port":                                     "3306",
		"innodb_buffer_pool_size":                  "128M",
		"innodb_data_home_dir":                     "/var/lib/mysql",
		"innodb_log_group_home_dir":                "/var/lib/mysql",
		"log_error":                                "/var/lib/mysql/mysql_error.log",
		"slow_query_log_file":                      "/var/lib/mysql/mysql_slow.log",
		"innodb_buffer_pool_instances":             "8",
	}

	// Merge dengan default values yang sudah ada
	for key, value := range defaults {
		if _, exists := template.DefaultValues[key]; !exists {
			template.DefaultValues[key] = value
		}
	}
}

// GenerateConfigFromTemplate menghasilkan konfigurasi dari template dengan values
func (t *MariaDBConfigTemplate) GenerateConfigFromTemplate(values map[string]string) (string, error) {
	if t.Content == "" {
		return "", fmt.Errorf("template content is empty")
	}

	result := t.Content

	// Replace placeholders dengan values
	for key, placeholder := range t.Placeholders {
		placeholderPattern := fmt.Sprintf("{{%s}}", placeholder)

		// Cari value untuk key ini
		var value string
		if val, exists := values[key]; exists {
			value = val
		} else if val, exists := t.DefaultValues[key]; exists {
			value = val
		} else {
			return "", fmt.Errorf("no value provided for placeholder %s (key: %s)", placeholder, key)
		}

		result = strings.ReplaceAll(result, placeholderPattern, value)
	}

	return result, nil
}

// BackupCurrentConfig membuat backup konfigurasi saat ini
func (t *MariaDBConfigTemplate) BackupCurrentConfig(backupDir string) (string, error) {
	if t.CurrentPath == "" {
		return "", fmt.Errorf("no current config path to backup")
	}

	// Pastikan backup directory ada
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}

	// Generate backup filename dengan timestamp
	timestamp := generateTimestamp()
	backupFilename := fmt.Sprintf("mariadb-config-backup-%s.cnf", timestamp)
	backupPath := fmt.Sprintf("%s/%s", backupDir, backupFilename)

	// Copy file
	if err := copyFile(t.CurrentPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to backup config file: %w", err)
	}

	return backupPath, nil
}

// generateTimestamp menghasilkan timestamp untuk backup filename
func generateTimestamp() string {
	return time.Now().Format("20060102-150405")
}

// copyFile menyalin file dari source ke destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Sync to ensure data is written
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
