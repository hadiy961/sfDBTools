package health

import (
	"database/sql"
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// ReplicationInfo represents replication status information
type ReplicationInfo struct {
	Status       string `json:"status"`
	StatusText   string `json:"status_text"`
	HasError     bool   `json:"has_error"`
	Role         string `json:"role"`
	SlaveCount   int    `json:"slave_count"`
	LastIOError  string `json:"last_io_error"`
	LastSQLError string `json:"last_sql_error"`
	IsHealthy    bool   `json:"is_healthy"`
}

// GetReplicationInfo retrieves replication status information
func GetReplicationInfo(config database.Config) (*ReplicationInfo, error) {
	lg, _ := logger.Get()

	info := &ReplicationInfo{
		Status:     "Unknown",
		StatusText: "❓ Unknown",
		Role:       "Unknown",
		IsHealthy:  false,
	}

	// Get replication status
	role, err := getReplicationRole(config)
	if err != nil {
		lg.Warn("Failed to get replication role", logger.Error(err))
		info.Status = "Error"
		info.StatusText = fmt.Sprintf("❌ Error: %v", err)
		return info, nil
	}
	info.Role = role

	// Check if it's a master or slave
	if strings.ToLower(role) == "master" {
		// Get master replication info
		slaveCount, err := getSlaveCount(config)
		if err != nil {
			lg.Warn("Failed to get slave count", logger.Error(err))
			info.SlaveCount = 0
		} else {
			info.SlaveCount = slaveCount
		}

		// Master is healthy if no errors
		info.IsHealthy = true
		info.Status = "OK"
		info.StatusText = "✅ Replication running normally"
	} else if strings.ToLower(role) == "slave" {
		// Get slave replication status
		isRunning, lastIOError, lastSQLError, err := getSlaveStatus(config)
		if err != nil {
			lg.Warn("Failed to get slave status", logger.Error(err))
			info.Status = "Error"
			info.StatusText = fmt.Sprintf("❌ Error: %v", err)
			return info, nil
		}

		info.LastIOError = lastIOError
		info.LastSQLError = lastSQLError

		if !isRunning || lastIOError != "" || lastSQLError != "" {
			info.HasError = true
			info.IsHealthy = false
			info.Status = "Error"

			if lastIOError != "" {
				info.StatusText = fmt.Sprintf("❌ Replication error detected: %s", lastIOError)
			} else if lastSQLError != "" {
				info.StatusText = fmt.Sprintf("❌ Replication error detected: %s", lastSQLError)
			} else {
				info.StatusText = "❌ Replication stopped"
			}
		} else {
			info.IsHealthy = true
			info.Status = "OK"
			info.StatusText = "✅ Replication running normally"
		}
	} else {
		// Not configured for replication
		info.Status = "Not Configured"
		info.StatusText = "ℹ️ Replication not configured"
		info.IsHealthy = true // Not an error if not configured
	}

	return info, nil
}

// getReplicationRole determines if the server is a master, slave, or standalone
func getReplicationRole(config database.Config) (string, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Check if binary logging is enabled (required for master)
	var logBin string
	err = db.QueryRow("SELECT @@log_bin").Scan(&logBin)
	if err != nil {
		return "", fmt.Errorf("failed to check binary logging: %w", err)
	}

	// Check slave status
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return "", fmt.Errorf("failed to check slave status: %w", err)
	}
	defer rows.Close()

	hasSlaveConfig := false
	if rows.Next() {
		// If we have data in SHOW SLAVE STATUS, it's configured as slave
		hasSlaveConfig = true
	}

	if hasSlaveConfig {
		return "Slave", nil
	} else if logBin == "1" || strings.ToUpper(logBin) == "ON" {
		return "Master", nil
	} else {
		return "Standalone", nil
	}
}

// getSlaveCount gets the number of connected slaves (for master)
func getSlaveCount(config database.Config) (int, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.processlist WHERE command = 'Binlog Dump'").Scan(&count)
	if err != nil {
		// Fallback method - count connections with 'Binlog Dump' in state
		err = db.QueryRow("SELECT COUNT(*) FROM information_schema.processlist WHERE state LIKE '%Binlog%'").Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to count slaves: %w", err)
		}
	}

	return count, nil
}

// getSlaveStatus gets detailed slave status information
func getSlaveStatus(config database.Config) (bool, string, string, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return false, "", "", fmt.Errorf("failed to execute SHOW SLAVE STATUS: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		// No slave configuration
		return false, "", "", nil
	}

	// Get column names to find the right indices
	columns, err := rows.Columns()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get columns: %w", err)
	}

	// Create slice to hold values
	values := make([]sql.NullString, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to scan slave status: %w", err)
	}

	// Find indices for the columns we need
	var slaveIORunningIdx, slaveSQLRunningIdx, lastIOErrorIdx, lastSQLErrorIdx int = -1, -1, -1, -1

	for i, col := range columns {
		switch col {
		case "Slave_IO_Running":
			slaveIORunningIdx = i
		case "Slave_SQL_Running":
			slaveSQLRunningIdx = i
		case "Last_IO_Error":
			lastIOErrorIdx = i
		case "Last_SQL_Error":
			lastSQLErrorIdx = i
		}
	}

	// Extract values
	var slaveIORunning, slaveSQLRunning bool
	var lastIOError, lastSQLError string

	if slaveIORunningIdx >= 0 && values[slaveIORunningIdx].Valid {
		slaveIORunning = strings.ToUpper(values[slaveIORunningIdx].String) == "YES"
	}

	if slaveSQLRunningIdx >= 0 && values[slaveSQLRunningIdx].Valid {
		slaveSQLRunning = strings.ToUpper(values[slaveSQLRunningIdx].String) == "YES"
	}

	if lastIOErrorIdx >= 0 && values[lastIOErrorIdx].Valid {
		lastIOError = values[lastIOErrorIdx].String
	}

	if lastSQLErrorIdx >= 0 && values[lastSQLErrorIdx].Valid {
		lastSQLError = values[lastSQLErrorIdx].String
	}

	isRunning := slaveIORunning && slaveSQLRunning
	return isRunning, lastIOError, lastSQLError, nil
}

// FormatReplicationInfo formats replication info for display
func FormatReplicationInfo(info *ReplicationInfo) []string {
	var details []string

	details = append(details, fmt.Sprintf("- Status: %s", info.StatusText))
	details = append(details, fmt.Sprintf("- Role: %s", info.Role))

	if strings.ToLower(info.Role) == "master" && info.SlaveCount >= 0 {
		details = append(details, fmt.Sprintf("- Number of Slaves: %d", info.SlaveCount))
	}

	if info.LastIOError != "" {
		details = append(details, fmt.Sprintf("- Last IO Error: '%s'", info.LastIOError))
	}

	if info.LastSQLError != "" {
		details = append(details, fmt.Sprintf("- Last SQL Error: '%s'", info.LastSQLError))
	}

	return details
}

// ValidateReplicationInfo validates replication information for potential issues
func ValidateReplicationInfo(info *ReplicationInfo) []string {
	var issues []string

	if info.HasError {
		if info.LastIOError != "" {
			issues = append(issues, fmt.Sprintf("Replication IO Error: %s", info.LastIOError))
		}
		if info.LastSQLError != "" {
			issues = append(issues, fmt.Sprintf("Replication SQL Error: %s", info.LastSQLError))
		}
	}

	if strings.ToLower(info.Role) == "master" && info.SlaveCount == 0 {
		issues = append(issues, "Master has no connected slaves")
	}

	return issues
}
