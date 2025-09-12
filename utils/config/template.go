package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sfDBTools/internal/logger"
)

// TemplateConfig berisi data template konfigurasi
type TemplateConfig struct {
	FilePath  string            `json:"file_path"`
	Content   string            `json:"content"`
	Variables map[string]string `json:"variables"`
	Sections  map[string]string `json:"sections"`
}

// MariaDBTemplateVars berisi variabel template untuk MariaDB
type MariaDBTemplateVars struct {
	ServerID                             string `json:"server_id"`
	Port                                 string `json:"port"`
	DataDir                              string `json:"datadir"`
	Socket                               string `json:"socket"`
	LogBin                               string `json:"log_bin"`
	LogError                             string `json:"log_error"`
	SlowQueryLogFile                     string `json:"slow_query_log_file"`
	InnodbDataHomeDir                    string `json:"innodb_data_home_dir"`
	InnodbLogGroupHomeDir                string `json:"innodb_log_group_home_dir"`
	InnodbBufferPoolSize                 string `json:"innodb_buffer_pool_size"`
	InnodbBufferPoolInstances            string `json:"innodb_buffer_pool_instances"`
	InnodbEncryptTables                  string `json:"innodb_encrypt_tables"`
	FileKeyManagementEncryptionAlgorithm string `json:"file_key_management_encryption_algorithm"`
	FileKeyManagementEncryptionKeyFile   string `json:"file_key_management_encryption_key_file"`
}

// LoadTemplate memuat template dari file
func LoadTemplate(templatePath string) (*TemplateConfig, error) {
	lg, _ := logger.Get()
	lg.Debug("Loading template", logger.String("path", templatePath))

	// Cek apakah file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file tidak ditemukan: %s", templatePath)
	}

	// Baca content file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca template file: %w", err)
	}

	template := &TemplateConfig{
		FilePath:  templatePath,
		Content:   string(content),
		Variables: make(map[string]string),
		Sections:  make(map[string]string),
	}

	// Parse variables dari template
	if err := template.parseVariables(); err != nil {
		lg.Warn("Gagal parsing variables dari template", logger.Error(err))
	}

	lg.Info("Template berhasil dimuat",
		logger.String("path", templatePath),
		logger.Int("variables_count", len(template.Variables)))

	return template, nil
}

// parseVariables mengekstrak variabel dari template content
func (tc *TemplateConfig) parseVariables() error {
	lines := strings.Split(tc.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments dan empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Parse key = value
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				tc.Variables[key] = value
			}
		}
	}

	return nil
}

// ReplaceVariables mengganti placeholder dengan nilai yang diberikan
func (tc *TemplateConfig) ReplaceVariables(vars map[string]string) string {
	lg, _ := logger.Get()

	result := tc.Content
	replacedCount := 0

	for key, value := range vars {
		// Replace dengan format {{key}}
		placeholder := fmt.Sprintf("{{%s}}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, value)
			replacedCount++
			lg.Debug("Replaced placeholder",
				logger.String("placeholder", placeholder),
				logger.String("value", value))
		}

		// Replace dengan format ${key}
		placeholder = fmt.Sprintf("${%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, value)
			replacedCount++
			lg.Debug("Replaced placeholder",
				logger.String("placeholder", placeholder),
				logger.String("value", value))
		}

		// Replace direct key = value replacement
		pattern := fmt.Sprintf(`^%s\s*=.*$`, regexp.QuoteMeta(key))
		re := regexp.MustCompile(pattern)
		lines := strings.Split(result, "\n")
		for i, line := range lines {
			if re.MatchString(strings.TrimSpace(line)) {
				lines[i] = fmt.Sprintf("%s = %s", key, value)
				replacedCount++
				lg.Debug("Replaced config line",
					logger.String("key", key),
					logger.String("value", value))
				break
			}
		}
		result = strings.Join(lines, "\n")
	}

	lg.Info("Template variables replaced",
		logger.Int("replaced_count", replacedCount),
		logger.Int("total_vars", len(vars)))

	return result
}

// WriteToFile menulis template yang sudah diproses ke file
func (tc *TemplateConfig) WriteToFile(content, outputPath string) error {
	lg, _ := logger.Get()
	lg.Debug("Writing processed template to file", logger.String("path", outputPath))

	// Buat directory jika belum ada
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("gagal membuat directory %s: %w", dir, err)
	}

	// Tulis file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("gagal menulis file %s: %w", outputPath, err)
	}

	lg.Info("Template berhasil ditulis ke file", logger.String("path", outputPath))
	return nil
}

// CreateMariaDBTemplateVars membuat struct vars untuk MariaDB template
func CreateMariaDBTemplateVars(
	serverID int,
	port int,
	dataDir, logDir, binlogDir string,
	innodbBufferPoolSize string,
	innodbBufferPoolInstances int,
	innodbEncryptTables bool,
	encryptionKeyFile string,
) *MariaDBTemplateVars {

	// Format boolean untuk MariaDB config
	encryptTablesValue := "OFF"
	if innodbEncryptTables {
		encryptTablesValue = "ON"
	}

	return &MariaDBTemplateVars{
		ServerID:                             fmt.Sprintf("%d", serverID),
		Port:                                 fmt.Sprintf("%d", port),
		DataDir:                              dataDir,
		Socket:                               filepath.Join(dataDir, "mysql.sock"),
		LogBin:                               filepath.Join(binlogDir, "mysql-bin"),
		LogError:                             filepath.Join(logDir, "mysql_error.log"),
		SlowQueryLogFile:                     filepath.Join(logDir, "mysql_slow.log"),
		InnodbDataHomeDir:                    dataDir,
		InnodbLogGroupHomeDir:                dataDir,
		InnodbBufferPoolSize:                 innodbBufferPoolSize,
		InnodbBufferPoolInstances:            fmt.Sprintf("%d", innodbBufferPoolInstances),
		InnodbEncryptTables:                  encryptTablesValue,
		FileKeyManagementEncryptionAlgorithm: "AES_CTR",
		FileKeyManagementEncryptionKeyFile:   encryptionKeyFile,
	}
}

// ToMap mengkonversi MariaDBTemplateVars ke map[string]string
func (mtv *MariaDBTemplateVars) ToMap() map[string]string {
	return map[string]string{
		"server_id":                    mtv.ServerID,
		"port":                         mtv.Port,
		"datadir":                      mtv.DataDir,
		"socket":                       mtv.Socket,
		"log_bin":                      mtv.LogBin,
		"log_error":                    mtv.LogError,
		"slow_query_log_file":          mtv.SlowQueryLogFile,
		"innodb_data_home_dir":         mtv.InnodbDataHomeDir,
		"innodb_log_group_home_dir":    mtv.InnodbLogGroupHomeDir,
		"innodb_buffer_pool_size":      mtv.InnodbBufferPoolSize,
		"innodb_buffer_pool_instances": mtv.InnodbBufferPoolInstances,
		"innodb_encrypt_tables":        mtv.InnodbEncryptTables,
		"file_key_management_encryption_algorithm": mtv.FileKeyManagementEncryptionAlgorithm,
		"file_key_management_encryption_key_file":  mtv.FileKeyManagementEncryptionKeyFile,
	}
}

// ValidateTemplate memvalidasi template content
func ValidateTemplate(templatePath string) error {
	lg, _ := logger.Get()
	lg.Debug("Validating template", logger.String("path", templatePath))

	template, err := LoadTemplate(templatePath)
	if err != nil {
		return err
	}

	// Cek apakah template memiliki section [mysqld] atau [mariadb]
	content := strings.ToLower(template.Content)
	if !strings.Contains(content, "[mysqld]") && !strings.Contains(content, "[mariadb]") {
		return fmt.Errorf("template harus memiliki section [mysqld] atau [mariadb]")
	}

	// Cek apakah template memiliki beberapa variable penting
	requiredVars := []string{"server_id", "port", "datadir"}
	missingVars := []string{}

	for _, varName := range requiredVars {
		found := false
		for key := range template.Variables {
			if strings.EqualFold(key, varName) {
				found = true
				break
			}
		}
		if !found {
			// Cek apakah ada placeholder
			placeholder1 := fmt.Sprintf("{{%s}}", varName)
			placeholder2 := fmt.Sprintf("${%s}", varName)
			if !strings.Contains(template.Content, placeholder1) && !strings.Contains(template.Content, placeholder2) {
				missingVars = append(missingVars, varName)
			}
		}
	}

	if len(missingVars) > 0 {
		lg.Warn("Template missing some recommended variables",
			logger.Strings("missing_vars", missingVars))
		// Warning saja, bukan error karena mungkin template sudah memiliki nilai hardcoded
	}

	lg.Info("Template validation completed", logger.String("path", templatePath))
	return nil
}

// BackupExistingConfig mem-backup file konfigurasi yang ada
func BackupExistingConfig(configPath, backupDir string) (string, error) {
	lg, _ := logger.Get()

	// Cek apakah config file ada
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		lg.Info("Config file tidak ada, skip backup", logger.String("path", configPath))
		return "", nil
	}

	// Buat backup directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat backup directory: %w", err)
	}

	// Generate backup filename dengan timestamp
	filename := filepath.Base(configPath)
	backupName := fmt.Sprintf("%s.backup.%d", filename, os.Getpid())
	backupPath := filepath.Join(backupDir, backupName)

	// Copy file
	input, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca config file: %w", err)
	}

	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		return "", fmt.Errorf("gagal menulis backup file: %w", err)
	}

	lg.Info("Config file berhasil di-backup",
		logger.String("original", configPath),
		logger.String("backup", backupPath))

	return backupPath, nil
}
