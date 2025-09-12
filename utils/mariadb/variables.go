package mariadb

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// MariaDBVariable berisi informasi variabel MariaDB
type MariaDBVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MariaDBVariables berisi koleksi variabel MariaDB
type MariaDBVariables struct {
	Variables map[string]string `json:"variables"`
}

// GetMariaDBVariables mengambil variabel dari database MariaDB yang sedang berjalan
func GetMariaDBVariables(dbConfig database.Config) (*MariaDBVariables, error) {
	lg, _ := logger.Get()
	lg.Debug("Mengambil variabel MariaDB")

	// Buat koneksi ke database
	db, err := database.GetDatabaseConnection(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer db.Close()

	// Execute SHOW VARIABLES
	rows, err := db.Query("SHOW VARIABLES")
	if err != nil {
		return nil, fmt.Errorf("gagal menjalankan SHOW VARIABLES: %w", err)
	}
	defer rows.Close()

	variables := &MariaDBVariables{
		Variables: make(map[string]string),
	}

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			lg.Warn("Gagal scan variable row", logger.Error(err))
			continue
		}
		variables.Variables[name] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error saat membaca rows: %w", err)
	}

	lg.Info("Berhasil mengambil variabel MariaDB",
		logger.Int("count", len(variables.Variables)))

	return variables, nil
}

// GetSpecificVariables mengambil variabel spesifik dari database
func GetSpecificVariables(dbConfig database.Config, varNames []string) (map[string]string, error) {
	lg, _ := logger.Get()
	lg.Debug("Mengambil variabel spesifik", logger.Strings("variables", varNames))

	// Buat koneksi ke database
	db, err := database.GetDatabaseConnection(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer db.Close()

	result := make(map[string]string)

	for _, varName := range varNames {
		query := "SHOW VARIABLES LIKE ?"
		var name, value string
		err := db.QueryRow(query, varName).Scan(&name, &value)
		if err != nil {
			if err == sql.ErrNoRows {
				lg.Debug("Variabel tidak ditemukan", logger.String("variable", varName))
				result[varName] = ""
				continue
			}
			return nil, fmt.Errorf("gagal query variabel %s: %w", varName, err)
		}
		result[varName] = value
	}

	lg.Info("Berhasil mengambil variabel spesifik",
		logger.Int("requested", len(varNames)),
		logger.Int("found", len(result)))

	return result, nil
}

// ValidateConfigurationApplied memvalidasi apakah konfigurasi sudah diterapkan
func ValidateConfigurationApplied(dbConfig database.Config, expectedConfig map[string]string) (map[string]bool, error) {
	lg, _ := logger.Get()
	lg.Debug("Validasi konfigurasi yang diterapkan")

	// Ambil variabel yang perlu dicek
	varNames := make([]string, 0, len(expectedConfig))
	for varName := range expectedConfig {
		varNames = append(varNames, varName)
	}

	actualValues, err := GetSpecificVariables(dbConfig, varNames)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil variabel aktual: %w", err)
	}

	// Bandingkan nilai
	results := make(map[string]bool)
	for varName, expectedValue := range expectedConfig {
		actualValue, exists := actualValues[varName]
		if !exists {
			results[varName] = false
			lg.Warn("Variabel tidak ditemukan", logger.String("variable", varName))
			continue
		}

		// Normalisasi nilai untuk perbandingan
		isMatch := normalizeAndCompare(expectedValue, actualValue)
		results[varName] = isMatch

		if isMatch {
			lg.Debug("Variabel sesuai",
				logger.String("variable", varName),
				logger.String("expected", expectedValue),
				logger.String("actual", actualValue))
		} else {
			lg.Warn("Variabel tidak sesuai",
				logger.String("variable", varName),
				logger.String("expected", expectedValue),
				logger.String("actual", actualValue))
		}
	}

	return results, nil
}

// normalizeAndCompare normalisasi dan bandingkan nilai variabel
func normalizeAndCompare(expected, actual string) bool {
	// Trim whitespace
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	// Case insensitive comparison untuk boolean values
	expectedLower := strings.ToLower(expected)
	actualLower := strings.ToLower(actual)

	// Handle boolean values
	if expectedLower == "on" || expectedLower == "1" || expectedLower == "true" {
		return actualLower == "on" || actualLower == "1" || actualLower == "true"
	}
	if expectedLower == "off" || expectedLower == "0" || expectedLower == "false" {
		return actualLower == "off" || actualLower == "0" || actualLower == "false"
	}

	// Handle numeric values
	if expectedNum, err := strconv.ParseInt(expected, 10, 64); err == nil {
		if actualNum, err := strconv.ParseInt(actual, 10, 64); err == nil {
			return expectedNum == actualNum
		}
	}

	// Handle size values (1G, 1024M, etc.)
	if expectedBytes := parseSize(expected); expectedBytes > 0 {
		if actualBytes := parseSize(actual); actualBytes > 0 {
			return expectedBytes == actualBytes
		}
	}

	// Direct string comparison
	return expected == actual
}

// parseSize parsing size string ke bytes (mendukung K, M, G, T suffixes)
func parseSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if len(sizeStr) == 0 {
		return 0
	}

	// Extract number dan suffix
	var number int64
	var suffix string

	for i, char := range sizeStr {
		if char < '0' || char > '9' {
			var err error
			number, err = strconv.ParseInt(sizeStr[:i], 10, 64)
			if err != nil {
				return 0
			}
			suffix = sizeStr[i:]
			break
		}
	}

	// Jika tidak ada suffix, anggap sudah dalam bytes
	if suffix == "" {
		var err error
		number, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0
		}
		return number
	}

	// Convert berdasarkan suffix
	switch suffix {
	case "K", "KB":
		return number * 1024
	case "M", "MB":
		return number * 1024 * 1024
	case "G", "GB":
		return number * 1024 * 1024 * 1024
	case "T", "TB":
		return number * 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
}

// GetKeyVariables mengambil variabel kunci untuk konfigurasi MariaDB
func GetKeyVariables(dbConfig database.Config) (map[string]string, error) {
	keyVars := []string{
		"server_id",
		"port",
		"datadir",
		"socket",
		"log_bin",
		"log_error",
		"slow_query_log_file",
		"innodb_data_home_dir",
		"innodb_log_group_home_dir",
		"innodb_buffer_pool_size",
		"innodb_buffer_pool_instances",
		"innodb_encrypt_tables",
		"file_key_management_encryption_algorithm",
		"file_key_management_filename",
	}

	return GetSpecificVariables(dbConfig, keyVars)
}

// FormatVariablesForDisplay memformat variabel untuk ditampilkan
func FormatVariablesForDisplay(variables map[string]string) [][]string {
	// Convert ke slice untuk display dalam tabel
	rows := make([][]string, 0, len(variables))

	for name, value := range variables {
		// Truncate value yang terlalu panjang
		displayValue := value
		if len(displayValue) > 50 {
			displayValue = displayValue[:47] + "..."
		}

		rows = append(rows, []string{name, displayValue})
	}

	return rows
}

// Get mengambil nilai variabel tertentu
func (mv *MariaDBVariables) Get(name string) (string, bool) {
	value, exists := mv.Variables[name]
	return value, exists
}

// GetInt mengambil nilai variabel sebagai integer
func (mv *MariaDBVariables) GetInt(name string) (int, error) {
	value, exists := mv.Variables[name]
	if !exists {
		return 0, fmt.Errorf("variabel %s tidak ditemukan", name)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("gagal konversi %s ke integer: %w", name, err)
	}

	return intValue, nil
}

// GetBool mengambil nilai variabel sebagai boolean
func (mv *MariaDBVariables) GetBool(name string) (bool, error) {
	value, exists := mv.Variables[name]
	if !exists {
		return false, fmt.Errorf("variabel %s tidak ditemukan", name)
	}

	valueLower := strings.ToLower(strings.TrimSpace(value))

	switch valueLower {
	case "on", "1", "true", "yes":
		return true, nil
	case "off", "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("nilai boolean tidak valid untuk %s: %s", name, value)
	}
}
