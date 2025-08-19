package health

import (
	"database/sql"
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// DatabasesInfo represents databases statistics information
type DatabasesInfo struct {
	TotalDatabases int64       `json:"total_databases"`
	TotalUsers     int64       `json:"total_users"`
	TotalSizeGB    float64     `json:"total_size_gb"`
	LargestDB      *DBSizeInfo `json:"largest_db"`
	SmallestDB     *DBSizeInfo `json:"smallest_db"`
}

// DBSizeInfo represents individual database size information
type DBSizeInfo struct {
	Name   string  `json:"name"`
	SizeGB float64 `json:"size_gb"`
	SizeMB float64 `json:"size_mb"`
}

// GetDatabasesInfo retrieves databases statistics information
func GetDatabasesInfo(config database.Config) (*DatabasesInfo, error) {
	lg, _ := logger.Get()

	info := &DatabasesInfo{}

	// Get total databases count
	totalDatabases, err := getTotalDatabases(config)
	if err != nil {
		lg.Warn("Failed to get total databases count", logger.Error(err))
		info.TotalDatabases = 0
	} else {
		info.TotalDatabases = totalDatabases
	}

	// Get total users count
	totalUsers, err := getTotalUsers(config)
	if err != nil {
		lg.Warn("Failed to get total users count", logger.Error(err))
		info.TotalUsers = 0
	} else {
		info.TotalUsers = totalUsers
	}

	// Get database sizes information
	totalSize, largestDB, smallestDB, err := getDatabasesSizeInfo(config)
	if err != nil {
		lg.Warn("Failed to get databases size information", logger.Error(err))
		info.TotalSizeGB = 0
	} else {
		info.TotalSizeGB = totalSize
		info.LargestDB = largestDB
		info.SmallestDB = smallestDB
	}

	return info, nil
}

// getTotalDatabases gets the total number of user databases (excluding system databases)
func getTotalDatabases(config database.Config) (int64, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := `
		SELECT COUNT(*) 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN (
			'information_schema', 'performance_schema', 'mysql', 'sys'
		)
	`

	var count int64
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get databases count: %w", err)
	}

	return count, nil
}

// getTotalUsers gets the total number of users
func getTotalUsers(config database.Config) (int64, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM mysql.user"

	var count int64
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get users count: %w", err)
	}

	return count, nil
}

// getDatabasesSizeInfo gets total size and largest/smallest database information
func getDatabasesSizeInfo(config database.Config) (float64, *DBSizeInfo, *DBSizeInfo, error) {
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := `
		SELECT 
			table_schema,
			ROUND(SUM(data_length + index_length) / 1024 / 1024 / 1024, 2) as size_gb
		FROM information_schema.tables 
		WHERE table_schema NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		GROUP BY table_schema 
		HAVING size_gb > 0
		ORDER BY size_gb DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		// Fallback to alternative method
		return getDatabasesSizeInfoFallback(db)
	}
	defer rows.Close()

	var totalSize float64
	var largestDB, smallestDB *DBSizeInfo

	for rows.Next() {
		var dbName string
		var sizeGB float64

		err := rows.Scan(&dbName, &sizeGB)
		if err != nil {
			continue
		}

		totalSize += sizeGB

		// Set largest database
		if largestDB == nil || sizeGB > largestDB.SizeGB {
			largestDB = &DBSizeInfo{
				Name:   dbName,
				SizeGB: sizeGB,
				SizeMB: sizeGB * 1024,
			}
		}

		// Set smallest database
		if smallestDB == nil || sizeGB < smallestDB.SizeGB {
			smallestDB = &DBSizeInfo{
				Name:   dbName,
				SizeGB: sizeGB,
				SizeMB: sizeGB * 1024,
			}
		}
	}

	return totalSize, largestDB, smallestDB, nil
}

// getDatabasesSizeInfoFallback uses an alternative method to get database sizes
func getDatabasesSizeInfoFallback(db *sql.DB) (float64, *DBSizeInfo, *DBSizeInfo, error) {
	// Get list of databases first
	databases, err := getUserDatabasesList(db)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get databases list: %w", err)
	}

	var totalSize float64
	var largestDB, smallestDB *DBSizeInfo

	for _, dbName := range databases {
		size, err := getDatabaseSize(db, dbName)
		if err != nil {
			continue
		}

		sizeGB := float64(size) / (1024 * 1024 * 1024)
		totalSize += sizeGB

		// Set largest database
		if largestDB == nil || sizeGB > largestDB.SizeGB {
			largestDB = &DBSizeInfo{
				Name:   dbName,
				SizeGB: sizeGB,
				SizeMB: sizeGB * 1024,
			}
		}

		// Set smallest database
		if smallestDB == nil || sizeGB < smallestDB.SizeGB {
			smallestDB = &DBSizeInfo{
				Name:   dbName,
				SizeGB: sizeGB,
				SizeMB: sizeGB * 1024,
			}
		}
	}

	return totalSize, largestDB, smallestDB, nil
}

// getUserDatabasesList gets the list of user databases
func getUserDatabasesList(db *sql.DB) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN (
			'information_schema', 'performance_schema', 'mysql', 'sys'
		)
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		databases = append(databases, dbName)
	}

	return databases, nil
}

// getDatabaseSize gets the size of a specific database in bytes
func getDatabaseSize(db *sql.DB, dbName string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(data_length + index_length), 0) as size_bytes
		FROM information_schema.tables 
		WHERE table_schema = ?
	`

	var sizeBytes int64
	err := db.QueryRow(query, dbName).Scan(&sizeBytes)
	if err != nil {
		return 0, err
	}

	return sizeBytes, nil
}

// FormatDatabasesInfo formats databases info for display
func FormatDatabasesInfo(info *DatabasesInfo) []string {
	var details []string

	details = append(details, fmt.Sprintf("- Total Databases: %d", info.TotalDatabases))
	details = append(details, fmt.Sprintf("- Total Users: %d", info.TotalUsers))
	details = append(details, fmt.Sprintf("- Total Logical Size: %.1f GB", info.TotalSizeGB))

	if info.LargestDB != nil {
		if info.LargestDB.SizeGB >= 1 {
			details = append(details, fmt.Sprintf("- Largest Database: '%s' (%.1f GB)", info.LargestDB.Name, info.LargestDB.SizeGB))
		} else {
			details = append(details, fmt.Sprintf("- Largest Database: '%s' (%.1f MB)", info.LargestDB.Name, info.LargestDB.SizeMB))
		}
	}

	if info.SmallestDB != nil {
		if info.SmallestDB.SizeGB >= 1 {
			details = append(details, fmt.Sprintf("- Smallest Database: '%s' (%.1f GB)", info.SmallestDB.Name, info.SmallestDB.SizeGB))
		} else {
			details = append(details, fmt.Sprintf("- Smallest Database: '%s' (%.1f MB)", info.SmallestDB.Name, info.SmallestDB.SizeMB))
		}
	}

	return details
}

// ValidateDatabasesInfo validates databases information for potential issues
func ValidateDatabasesInfo(info *DatabasesInfo) []string {
	var issues []string

	if info.TotalDatabases == 0 {
		issues = append(issues, "No user databases found")
	}

	if info.TotalUsers == 0 {
		issues = append(issues, "No users found in mysql.user table")
	}

	if info.TotalSizeGB > 100 {
		issues = append(issues, fmt.Sprintf("Large total database size: %.1f GB", info.TotalSizeGB))
	}

	if info.LargestDB != nil && info.LargestDB.SizeGB > 50 {
		issues = append(issues, fmt.Sprintf("Large database detected: '%s' (%.1f GB)", info.LargestDB.Name, info.LargestDB.SizeGB))
	}

	return issues
}
