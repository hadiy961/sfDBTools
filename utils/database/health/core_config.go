package health

import (
	"database/sql"
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// CoreConfig represents MariaDB/MySQL core configuration information
type CoreConfig struct {
	MariaDBVersion string          `json:"mariadb_version"`
	ReadOnlyStatus string          `json:"readonly_status"`
	IsReadOnly     bool            `json:"is_readonly"`
	GTIDStatus     *GTIDStatusInfo `json:"gtid_status"`
	InnoDBConfig   *InnoDBConfig   `json:"innodb_config"`
}

// GTIDStatusInfo represents GTID status information
type GTIDStatusInfo struct {
	Enabled        bool   `json:"enabled"`
	EnabledText    string `json:"enabled_text"`
	GTIDCurrentPos string `json:"gtid_current_pos"`
	GTIDPastPos    string `json:"gtid_past_pos"`
}

// InnoDBConfig represents InnoDB configuration
type InnoDBConfig struct {
	BufferPoolSize      string `json:"buffer_pool_size"`
	BufferPoolInstances string `json:"buffer_pool_instances"`
}

// GetCoreConfig retrieves MariaDB/MySQL core configuration information
func GetCoreConfig(config database.Config) (*CoreConfig, error) {
	lg, _ := logger.Get()

	info := &CoreConfig{}

	// Get MariaDB version
	version, err := database.GetMySQLVersion(config)
	if err != nil {
		lg.Warn("Failed to get MariaDB version", logger.Error(err))
		info.MariaDBVersion = "Unknown"
	} else {
		info.MariaDBVersion = version
	}

	// Get ReadOnly status
	readOnly, err := getReadOnlyStatus(config)
	if err != nil {
		lg.Warn("Failed to get read-only status", logger.Error(err))
		info.ReadOnlyStatus = "Unknown"
		info.IsReadOnly = false
	} else {
		info.IsReadOnly = readOnly
		if readOnly {
			info.ReadOnlyStatus = "ON"
		} else {
			info.ReadOnlyStatus = "OFF"
		}
	}

	// Get GTID status
	gtidStatus, err := getGTIDStatus(config)
	if err != nil {
		lg.Warn("Failed to get GTID status", logger.Error(err))
		info.GTIDStatus = &GTIDStatusInfo{
			Enabled:     false,
			EnabledText: "Unknown",
		}
	} else {
		info.GTIDStatus = gtidStatus
	}

	// Get InnoDB configuration
	innodbConfig, err := getInnoDBConfig(config)
	if err != nil {
		lg.Warn("Failed to get InnoDB configuration", logger.Error(err))
		info.InnoDBConfig = &InnoDBConfig{
			BufferPoolSize:      "Unknown",
			BufferPoolInstances: "Unknown",
		}
	} else {
		info.InnoDBConfig = innodbConfig
	}

	return info, nil
}

// getReadOnlyStatus retrieves the read_only status
func getReadOnlyStatus(config database.Config) (bool, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return false, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var variableName, value string
	query := "SHOW GLOBAL VARIABLES LIKE 'read_only'"

	err = db.QueryRow(query).Scan(&variableName, &value)
	if err != nil {
		return false, fmt.Errorf("failed to query read_only status: %w", err)
	}

	return strings.ToUpper(value) == "ON", nil
}

// getGTIDStatus retrieves GTID status information
func getGTIDStatus(config database.Config) (*GTIDStatusInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	gtidInfo := &GTIDStatusInfo{
		Enabled:     false,
		EnabledText: "❌ No",
	}

	// Check GTID mode
	var variableName, value string
	query := "SHOW GLOBAL VARIABLES LIKE 'gtid_mode'"

	err = db.QueryRow(query).Scan(&variableName, &value)
	if err == nil && strings.ToUpper(value) == "ON" {
		gtidInfo.Enabled = true
		gtidInfo.EnabledText = "✅ Yes"

		// Get GTID positions if enabled
		gtidInfo.GTIDCurrentPos, _ = getGTIDVariable(db, "gtid_executed")
		gtidInfo.GTIDPastPos, _ = getGTIDVariable(db, "gtid_purged")
	}

	return gtidInfo, nil
}

// getGTIDVariable retrieves a GTID-related global variable
func getGTIDVariable(db *sql.DB, variable string) (string, error) {
	var value string
	query := fmt.Sprintf("SELECT @@GLOBAL.%s", variable)

	err := db.QueryRow(query).Scan(&value)
	if err != nil {
		return "", err
	}

	if value == "" {
		return "Empty", nil
	}

	return value, nil
}

// getInnoDBConfig retrieves InnoDB configuration
func getInnoDBConfig(config database.Config) (*InnoDBConfig, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	innodbConfig := &InnoDBConfig{}

	// Get buffer pool size
	bufferPoolSize, err := getGlobalVariable(db, "innodb_buffer_pool_size")
	if err != nil {
		innodbConfig.BufferPoolSize = "Unknown"
	} else {
		innodbConfig.BufferPoolSize = formatBufferPoolSize(bufferPoolSize)
	}

	// Get buffer pool instances
	bufferPoolInstances, err := getGlobalVariable(db, "innodb_buffer_pool_instances")
	if err != nil {
		innodbConfig.BufferPoolInstances = "Unknown"
	} else {
		innodbConfig.BufferPoolInstances = bufferPoolInstances
	}

	return innodbConfig, nil
}

// getGlobalVariable retrieves a global variable value
func getGlobalVariable(db *sql.DB, variable string) (string, error) {
	var variableName, value string
	query := fmt.Sprintf("SHOW GLOBAL VARIABLES LIKE '%s'", variable)

	err := db.QueryRow(query).Scan(&variableName, &value)
	if err != nil {
		// If variable doesn't exist, try to get it as a system variable
		var sysValue sql.NullString
		sysQuery := fmt.Sprintf("SELECT @@GLOBAL.%s", variable)
		if sysErr := db.QueryRow(sysQuery).Scan(&sysValue); sysErr == nil && sysValue.Valid {
			return sysValue.String, nil
		}
		return "", err
	}

	return value, nil
}

// formatBufferPoolSize formats buffer pool size from bytes to readable format
func formatBufferPoolSize(sizeStr string) string {
	if sizeStr == "" || sizeStr == "0" {
		return "0"
	}

	// Try to parse as integer (bytes)
	var sizeBytes int64
	_, err := fmt.Sscanf(sizeStr, "%d", &sizeBytes)
	if err != nil {
		return sizeStr // Return as-is if can't parse
	}

	// Convert to readable format
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if sizeBytes >= GB {
		return fmt.Sprintf("%.0fG", float64(sizeBytes)/GB)
	} else if sizeBytes >= MB {
		return fmt.Sprintf("%.0fM", float64(sizeBytes)/MB)
	} else if sizeBytes >= KB {
		return fmt.Sprintf("%.0fK", float64(sizeBytes)/KB)
	}

	return fmt.Sprintf("%d", sizeBytes)
}

// FormatCoreConfig formats core configuration info for display
func FormatCoreConfig(info *CoreConfig) []string {
	var details []string

	details = append(details, fmt.Sprintf("- MariaDB Version: %s", info.MariaDBVersion))
	details = append(details, fmt.Sprintf("- ReadOnly Status: %s", info.ReadOnlyStatus))

	// GTID Status
	if info.GTIDStatus != nil {
		details = append(details, "- GTID Status:")
		details = append(details, fmt.Sprintf("  - Enabled: %s", info.GTIDStatus.EnabledText))

		if info.GTIDStatus.Enabled {
			if info.GTIDStatus.GTIDCurrentPos != "" && info.GTIDStatus.GTIDCurrentPos != "Empty" {
				// Truncate long GTID positions for readability
				currentPos := info.GTIDStatus.GTIDCurrentPos
				if len(currentPos) > 50 {
					currentPos = currentPos[:50] + "..."
				}
				details = append(details, fmt.Sprintf("  - Gtid_executed: %s", currentPos))
			}

			if info.GTIDStatus.GTIDPastPos != "" && info.GTIDStatus.GTIDPastPos != "Empty" {
				pastPos := info.GTIDStatus.GTIDPastPos
				if len(pastPos) > 50 {
					pastPos = pastPos[:50] + "..."
				}
				details = append(details, fmt.Sprintf("  - Gtid_purged: %s", pastPos))
			}
		}
	}

	// InnoDB Configuration
	if info.InnoDBConfig != nil {
		details = append(details, "- Innodb Configuration:")
		details = append(details, fmt.Sprintf("  - innodb_buffer_pool_size: %s", info.InnoDBConfig.BufferPoolSize))
		details = append(details, fmt.Sprintf("  - innodb_buffer_pool_instances: %s", info.InnoDBConfig.BufferPoolInstances))
	}

	return details
}
