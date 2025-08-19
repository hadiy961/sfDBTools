package health

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// LogsConfig represents MariaDB/MySQL logs configuration information
type LogsConfig struct {
	BinaryLog  *LogInfo `json:"binary_log"`
	SlowLog    *LogInfo `json:"slow_log"`
	ErrorLog   *LogInfo `json:"error_log"`
	GeneralLog *LogInfo `json:"general_log"`
}

// LogInfo represents individual log information
type LogInfo struct {
	Enabled     bool   `json:"enabled"`
	EnabledText string `json:"enabled_text"`
	Path        string `json:"path"`
	Status      string `json:"status"`
}

// GetLogsConfig retrieves MariaDB/MySQL logs configuration information
func GetLogsConfig(config database.Config) (*LogsConfig, error) {
	lg, _ := logger.Get()

	info := &LogsConfig{}

	// Get binary log info
	binaryLog, err := getBinaryLogInfo(config)
	if err != nil {
		lg.Warn("Failed to get binary log info", logger.Error(err))
		info.BinaryLog = &LogInfo{
			Enabled:     false,
			EnabledText: "Unknown",
			Path:        "Unknown",
			Status:      "❌ Unknown",
		}
	} else {
		info.BinaryLog = binaryLog
	}

	// Get slow query log info
	slowLog, err := getSlowLogInfo(config)
	if err != nil {
		lg.Warn("Failed to get slow log info", logger.Error(err))
		info.SlowLog = &LogInfo{
			Enabled:     false,
			EnabledText: "Unknown",
			Path:        "Unknown",
			Status:      "❌ Unknown",
		}
	} else {
		info.SlowLog = slowLog
	}

	// Get error log info
	errorLog, err := getErrorLogInfo(config)
	if err != nil {
		lg.Warn("Failed to get error log info", logger.Error(err))
		info.ErrorLog = &LogInfo{
			Enabled:     false,
			EnabledText: "Unknown",
			Path:        "Unknown",
			Status:      "❌ Unknown",
		}
	} else {
		info.ErrorLog = errorLog
	}

	// Get general log info
	generalLog, err := getGeneralLogInfo(config)
	if err != nil {
		lg.Warn("Failed to get general log info", logger.Error(err))
		info.GeneralLog = &LogInfo{
			Enabled:     false,
			EnabledText: "Unknown",
			Path:        "Unknown",
			Status:      "❌ Unknown",
		}
	} else {
		info.GeneralLog = generalLog
	}

	return info, nil
}

// getBinaryLogInfo retrieves binary log information
func getBinaryLogInfo(config database.Config) (*LogInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	logInfo := &LogInfo{}

	// Check if binary logging is enabled
	enabled, err := getGlobalVariableBool(db, "log_bin")
	if err != nil {
		return nil, fmt.Errorf("failed to check binary log status: %w", err)
	}

	logInfo.Enabled = enabled
	if enabled {
		logInfo.EnabledText = "✅ Enabled"
		logInfo.Status = "✅ Enabled"
	} else {
		logInfo.EnabledText = "❌ Disabled"
		logInfo.Status = "❌ Disabled"
	}

	// Get binary log path/basename
	if enabled {
		basename, err := getGlobalVariable(db, "log_bin_basename")
		if err != nil {
			// Fallback to datadir + mysql-bin
			datadir, err2 := getGlobalVariable(db, "datadir")
			if err2 == nil {
				logInfo.Path = strings.TrimRight(datadir, "/") + "/mysql-bin"
			} else {
				logInfo.Path = "/var/log/mysql/mysql-bin.log"
			}
		} else {
			logInfo.Path = basename
		}
	} else {
		logInfo.Path = "Not configured"
	}

	return logInfo, nil
}

// getSlowLogInfo retrieves slow query log information
func getSlowLogInfo(config database.Config) (*LogInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	logInfo := &LogInfo{}

	// Check if slow query log is enabled
	enabled, err := getGlobalVariableBool(db, "slow_query_log")
	if err != nil {
		return nil, fmt.Errorf("failed to check slow query log status: %w", err)
	}

	logInfo.Enabled = enabled
	if enabled {
		logInfo.EnabledText = "✅ Enabled"
		logInfo.Status = "✅ Enabled"
	} else {
		logInfo.EnabledText = "❌ Disabled"
		logInfo.Status = "❌ Disabled"
	}

	// Get slow query log file path
	if enabled {
		logFile, err := getGlobalVariable(db, "slow_query_log_file")
		if err != nil {
			logInfo.Path = "/var/log/mysql/mariadb-slow.log"
		} else {
			logInfo.Path = logFile
		}
	} else {
		logInfo.Path = "Not configured"
	}

	return logInfo, nil
}

// getErrorLogInfo retrieves error log information
func getErrorLogInfo(config database.Config) (*LogInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	logInfo := &LogInfo{}

	// Error log is always enabled, get the path
	logFile, err := getGlobalVariable(db, "log_error")
	if err != nil {
		logInfo.Path = "/var/log/mysql/error.log"
	} else {
		if logFile == "" {
			// If empty, it usually goes to stderr or default location
			logInfo.Path = "/var/log/mysql/error.log"
		} else {
			logInfo.Path = logFile
		}
	}

	// Check if the log file exists
	if _, err := os.Stat(logInfo.Path); err == nil {
		logInfo.Enabled = true
		logInfo.EnabledText = "✅ Enabled"
		logInfo.Status = "✅ Enabled"
	} else {
		logInfo.Enabled = true
		logInfo.EnabledText = "✅ Enabled"
		logInfo.Status = "⚠️ Enabled (file not accessible)"
	}

	return logInfo, nil
}

// getGeneralLogInfo retrieves general query log information
func getGeneralLogInfo(config database.Config) (*LogInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	logInfo := &LogInfo{}

	// Check if general log is enabled
	enabled, err := getGlobalVariableBool(db, "general_log")
	if err != nil {
		return nil, fmt.Errorf("failed to check general log status: %w", err)
	}

	logInfo.Enabled = enabled
	if enabled {
		logInfo.EnabledText = "✅ Enabled"
		logInfo.Status = "✅ Enabled"
	} else {
		logInfo.EnabledText = "❌ Disabled"
		logInfo.Status = "❌ Disabled"
	}

	// Get general log file path
	if enabled {
		logFile, err := getGlobalVariable(db, "general_log_file")
		if err != nil {
			logInfo.Path = "/var/log/mysql/mysql.log"
		} else {
			logInfo.Path = logFile
		}
	} else {
		logInfo.Path = "Not configured"
	}

	return logInfo, nil
}

// getGlobalVariableBool retrieves a global variable value as boolean
func getGlobalVariableBool(db *sql.DB, variable string) (bool, error) {
	value, err := getGlobalVariable(db, variable)
	if err != nil {
		return false, err
	}

	upperValue := strings.ToUpper(value)
	return upperValue == "ON" || upperValue == "1" || upperValue == "TRUE", nil
}

// FormatLogsConfig formats logs configuration info for display
func FormatLogsConfig(info *LogsConfig) []string {
	var details []string

	// Binary Log
	if info.BinaryLog != nil {
		details = append(details, fmt.Sprintf("- Binary Log (Binlog) Path: %s", info.BinaryLog.Path))
		details = append(details, fmt.Sprintf("  - Status: %s", info.BinaryLog.Status))
	}

	// Slow Query Log
	if info.SlowLog != nil {
		details = append(details, "- Slow Query Log:")
		details = append(details, fmt.Sprintf("  - Status: %s", info.SlowLog.Status))
		if info.SlowLog.Enabled {
			details = append(details, fmt.Sprintf("  - Path: %s", info.SlowLog.Path))
		}
	}

	// Error Log
	if info.ErrorLog != nil {
		details = append(details, "- Error Log:")
		details = append(details, fmt.Sprintf("  - Status: %s", info.ErrorLog.Status))
		details = append(details, fmt.Sprintf("  - Path: %s", info.ErrorLog.Path))
	}

	return details
}

// ValidateLogPaths validates that log file paths are accessible
func ValidateLogPaths(info *LogsConfig) []string {
	var issues []string

	// Check binary log path
	if info.BinaryLog != nil && info.BinaryLog.Enabled {
		if _, err := os.Stat(info.BinaryLog.Path); err != nil {
			issues = append(issues, fmt.Sprintf("Binary log path not accessible: %s", info.BinaryLog.Path))
		}
	}

	// Check slow log path
	if info.SlowLog != nil && info.SlowLog.Enabled {
		if _, err := os.Stat(info.SlowLog.Path); err != nil {
			issues = append(issues, fmt.Sprintf("Slow log path not accessible: %s", info.SlowLog.Path))
		}
	}

	// Check error log path
	if info.ErrorLog != nil && info.ErrorLog.Enabled {
		if _, err := os.Stat(info.ErrorLog.Path); err != nil {
			issues = append(issues, fmt.Sprintf("Error log path not accessible: %s", info.ErrorLog.Path))
		}
	}

	return issues
}
