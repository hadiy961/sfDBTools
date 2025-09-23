package info

import (
	"database/sql"
	"fmt"
	"strconv"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"
	"sfDBTools/utils/terminal"
)

// DatabaseInfo represents information about a database
type DatabaseInfo struct {
	DatabaseName string  `json:"database_name"`
	SizeBytes    int64   `json:"size_bytes"`
	SizeMB       float64 `json:"size_mb"`
	SizeHuman    string  `json:"size_human"`
	TableCount   int     `json:"table_count"`
	ViewCount    int     `json:"view_count"`
	RoutineCount int     `json:"routine_count"`
	TriggerCount int     `json:"trigger_count"`
	UserCount    int     `json:"user_count"`
}

// GetDatabaseInfo retrieves comprehensive information about a database
func GetDatabaseInfo(config database.Config) (*DatabaseInfo, error) {
	lg, _ := logger.Get()

	db, err := database.GetDatabaseConnection(config)
	if err != nil {
		lg.Error("Failed to connect to database", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	info := &DatabaseInfo{
		DatabaseName: config.DBName,
	}

	// Use a single shared spinner for the whole metadata collection and
	// update its message between steps. Track if any step produced an
	// error so we can show a final warning message when finished.
	spinner := terminal.NewProgressSpinner("Collecting database metadata...")
	spinner.Start()
	defer func() {
		// Ensure spinner is stopped; if there were warnings we show a warning
		// message, otherwise show success.
		if spinner == nil {
			return
		}
	}()

	hadWarning := false

	// Get database size
	spinner.UpdateMessage("Calculating database size...")
	if size, err := getDatabaseSize(db, config.DBName); err == nil {
		info.SizeBytes = size
		info.SizeMB = float64(size) / (1024 * 1024)
		info.SizeHuman = common.FormatSize(size)
		spinner.UpdateMessage(fmt.Sprintf("Database size: %s", info.SizeHuman))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get database size")
		lg.Warn("Failed to get database size", logger.Error(err))
	}

	// Get table count
	spinner.UpdateMessage("Counting tables...")
	if count, err := getTableCount(db, config.DBName); err == nil {
		info.TableCount = count
		spinner.UpdateMessage(fmt.Sprintf("Tables: %d", info.TableCount))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get table count")
		lg.Warn("Failed to get table count", logger.Error(err))
	}

	// Get view count
	spinner.UpdateMessage("Counting views...")
	if count, err := getViewCount(db, config.DBName); err == nil {
		info.ViewCount = count
		spinner.UpdateMessage(fmt.Sprintf("Views: %d", info.ViewCount))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get view count")
		lg.Warn("Failed to get view count", logger.Error(err))
	}

	// Get routine count (stored procedures + functions)
	spinner.UpdateMessage("Counting routines (procs & funcs)...")
	if count, err := getRoutineCount(db, config.DBName); err == nil {
		info.RoutineCount = count
		spinner.UpdateMessage(fmt.Sprintf("Routines: %d", info.RoutineCount))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get routine count")
		lg.Warn("Failed to get routine count", logger.Error(err))
	}

	// Get trigger count
	spinner.UpdateMessage("Counting triggers...")
	if count, err := getTriggerCount(db, config.DBName); err == nil {
		info.TriggerCount = count
		spinner.UpdateMessage(fmt.Sprintf("Triggers: %d", info.TriggerCount))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get trigger count")
		lg.Warn("Failed to get trigger count", logger.Error(err))
	}

	// Get user count with grants to this database
	spinner.UpdateMessage("Counting users with grants...")
	if count, err := getUserCount(db, config.DBName); err == nil {
		info.UserCount = count
		spinner.UpdateMessage(fmt.Sprintf("Users with grants: %d", info.UserCount))
	} else {
		hadWarning = true
		spinner.UpdateMessage("Failed to get user count")
		lg.Warn("Failed to get user count", logger.Error(err))
	}

	// Finalize spinner with appropriate final status
	if hadWarning {
		spinner.StopWithWarning("Completed with warnings")
	} else {
		spinner.StopWithSuccess("Database information collected")
	}
	spinner.Stop()
	return info, nil
}

// getDatabaseSize calculates the total size of a database in bytes
// getDatabaseSize calculates the total size of a database in bytes using SHOW TABLE STATUS
func getDatabaseSize(db *sql.DB, dbName string) (int64, error) {
	// Use SHOW TABLE STATUS which is much faster than information_schema
	query := "SHOW TABLE STATUS FROM " + "`" + dbName + "`"

	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var totalSize int64 = 0

	// Get column names to handle different MySQL versions
	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	// Create a slice to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Find the indices of Data_length and Index_length columns
	var dataLengthIdx, indexLengthIdx int = -1, -1
	for i, col := range columns {
		if col == "Data_length" {
			dataLengthIdx = i
		} else if col == "Index_length" {
			indexLengthIdx = i
		}
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			continue // Skip problematic rows
		}

		// Extract Data_length and Index_length
		var dataLength, indexLength int64

		if dataLengthIdx >= 0 && values[dataLengthIdx] != nil {
			if val, ok := values[dataLengthIdx].([]byte); ok {
				if parsed, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					dataLength = parsed
				}
			} else if val, ok := values[dataLengthIdx].(int64); ok {
				dataLength = val
			}
		}

		if indexLengthIdx >= 0 && values[indexLengthIdx] != nil {
			if val, ok := values[indexLengthIdx].([]byte); ok {
				if parsed, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					indexLength = parsed
				}
			} else if val, ok := values[indexLengthIdx].(int64); ok {
				indexLength = val
			}
		}

		totalSize += dataLength + indexLength
	}

	return totalSize, nil
}

// getTableCount returns the number of tables in a database
// getTableCount returns the number of BASE TABLES only (excluding views)
func getTableCount(db *sql.DB, dbName string) (int, error) {
	// Use information_schema to distinguish BASE TABLE from VIEW
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE'"

	var count int
	err := db.QueryRow(query, dbName).Scan(&count)
	if err != nil {
		// Fallback to SHOW TABLES if information_schema fails
		return getTableCountFallback(db, dbName)
	}

	return count, nil
}

// Fallback method using SHOW FULL TABLES
func getTableCountFallback(db *sql.DB, dbName string) (int, error) {
	query := "SHOW FULL TABLES FROM " + "`" + dbName + "`" + " WHERE Table_type = 'BASE TABLE'"

	rows, err := db.Query(query)
	if err != nil {
		// Final fallback - assume all are tables
		return getTableCountSimple(db, dbName)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count, nil
}

func getTableCountSimple(db *sql.DB, dbName string) (int, error) {
	query := "SHOW TABLES FROM " + "`" + dbName + "`"

	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count, nil
}

// getViewCount returns the number of views in a database
func getViewCount(db *sql.DB, dbName string) (int, error) {
	// Try to use SHOW FULL TABLES to get views (faster than information_schema)
	query := "SHOW FULL TABLES FROM " + "`" + dbName + "`" + " WHERE Table_type = 'VIEW'"

	rows, err := db.Query(query)
	if err != nil {
		// Fallback: try without FULL keyword for older MySQL versions
		fallbackQuery := "SHOW TABLES FROM " + "`" + dbName + "`"
		fallbackRows, fallbackErr := db.Query(fallbackQuery)
		if fallbackErr != nil {
			return 0, fallbackErr
		}
		defer fallbackRows.Close()

		// For fallback, we can't distinguish views from tables easily
		// So we return 0 to avoid incorrect counts
		return 0, nil
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count, nil
}

// getRoutineCount returns the number of stored procedures and functions in a database
// It tries a fast query on mysql.proc first, then falls back to information_schema.
func getRoutineCount(db *sql.DB, dbName string) (int, error) {
	var count int

	// 1. Try the fast method first (querying mysql.proc)
	procQuery := "SELECT COUNT(*) FROM mysql.proc WHERE db = ?"
	err := db.QueryRow(procQuery, dbName).Scan(&count)
	if err == nil {
		return count, nil // Success! Return the count.
	}

	return count, nil
}

// getTriggerCount returns the number of triggers in a database
func getTriggerCount(db *sql.DB, dbName string) (int, error) {
	// Use SHOW TRIGGERS which is much faster than information_schema
	showQuery := "SHOW TRIGGERS FROM " + "`" + dbName + "`"

	rows, err := db.Query(showQuery)
	if err != nil {
		// If SHOW TRIGGERS fails, return 0 to avoid breaking backup
		return 0, nil
	}
	defer rows.Close()

	// Count the rows returned by SHOW TRIGGERS
	count := 0
	for rows.Next() {
		count++
	}

	return count, nil
}

// getUserCount returns the number of users with grants to a specific database
func getUserCount(db *sql.DB, dbName string) (int, error) {
	// This query gets users with specific database privileges
	// Note: This might not work perfectly in all MySQL/MariaDB versions
	// as it depends on the mysql.db table structure
	query := `
		SELECT COUNT(DISTINCT user) 
		FROM mysql.db 
		WHERE db = ? OR db = '*'
	`

	var count int
	err := db.QueryRow(query, dbName).Scan(&count)
	if err != nil {
		// Fallback: try to get global user count if database-specific fails
		fallbackQuery := `SELECT COUNT(*) FROM mysql.user`
		err = db.QueryRow(fallbackQuery).Scan(&count)
	}

	return count, err
}

// GetDetailedTableInfo returns detailed information about tables in the database
func GetDetailedTableInfo(config database.Config) ([]TableInfo, error) {
	lg, _ := logger.Get()

	db, err := database.GetDatabaseConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Use SHOW FULL TABLES to get both tables and views
	query := "SHOW FULL TABLES FROM " + "`" + config.DBName + "`"

	rows, err := db.Query(query)
	if err != nil {
		// Fallback to simple SHOW TABLES if SHOW FULL TABLES fails
		query = "SHOW TABLES FROM " + "`" + config.DBName + "`"
		rows, err = db.Query(query)
		if err != nil {
			lg.Error("Failed to get table information", logger.Error(err))
			return nil, fmt.Errorf("failed to get table information: %w", err)
		}
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var tableType sql.NullString

		// Try to scan with table type first (for SHOW FULL TABLES)
		err := rows.Scan(&table.TableName, &tableType)
		if err != nil {
			// If that fails, try scanning just the table name (for SHOW TABLES)
			err = rows.Scan(&table.TableName)
			if err != nil {
				lg.Warn("Failed to scan table row", logger.Error(err))
				continue
			}
			table.TableType = "BASE TABLE" // Default type
		} else {
			if tableType.Valid {
				table.TableType = tableType.String
			} else {
				table.TableType = "BASE TABLE"
			}
		}

		// For basic table info, we don't get size information with SHOW commands
		// This is a trade-off for performance - we get table names quickly
		// but lose detailed size information
		table.RowCount = 0
		table.DataSize = 0
		table.IndexSize = 0
		table.TotalSize = 0

		tables = append(tables, table)
	}

	lg.Info("Retrieved table information using SHOW commands",
		logger.String("database", config.DBName),
		logger.Int("table_count", len(tables)))

	return tables, nil
}

// TableInfo represents information about a single table
type TableInfo struct {
	TableName string `json:"table_name"`
	RowCount  int64  `json:"row_count"`
	DataSize  int64  `json:"data_size"`
	IndexSize int64  `json:"index_size"`
	TotalSize int64  `json:"total_size"`
	TableType string `json:"table_type"`
}

// collectDatabaseInfo retrieves database information and logs it
func CollectDatabaseInfo(config database.Config, lg *logger.Logger) *DatabaseInfo {
	lg.Debug("Collecting database information", logger.String("database", config.DBName))

	dbInfo, err := GetDatabaseInfo(config)
	if err != nil {
		lg.Warn("Failed to collect database information", logger.Error(err))
		return nil
	}

	lg.Info("Database information collected",
		logger.String("database", dbInfo.DatabaseName),
		// Log both raw bytes as integer and size in MB as float to avoid
		// scientific notation for very large sizes in logs.
		logger.String("size", dbInfo.SizeHuman),
		logger.Int("tables", dbInfo.TableCount),
		logger.Int("views", dbInfo.ViewCount),
		logger.Int("routines", dbInfo.RoutineCount),
		logger.Int("triggers", dbInfo.TriggerCount),
		logger.Int("users", dbInfo.UserCount))

	return dbInfo
}
