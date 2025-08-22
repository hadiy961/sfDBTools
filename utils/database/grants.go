package database

import (
	"database/sql"
	"fmt"
	"strings"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
)

// GetSystemUsersFromConfig gets the list of system users from configuration
func GetSystemUsersFromConfig() []string {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Fallback to default system users if config fails
		return []string{
			"sst_user",
			"papp",
			"sysadmin",
			"backup_user",
			"dbaDO",
			"maxscale",
		}
	}

	if len(cfg.SystemUsers.Users) == 0 {
		// Fallback to default if no users configured
		return []string{
			"sst_user",
			"papp",
			"sysadmin",
			"backup_user",
			"dbaDO",
			"maxscale",
		}
	}

	return cfg.SystemUsers.Users
}

// SystemUsers defines the list of system users (deprecated: use GetSystemUsersFromConfig)
var SystemUsers = []string{
	"sst_user",
	"papp",
	"sysadmin",
	"backup_user",
	"dbaDO",
	"maxscale",
}

// GetDatabaseGrantees gets all users with privileges on a database
func GetDatabaseGrantees(db *sql.DB, dbName string) ([]string, error) {
	query := `
		SELECT DISTINCT GRANTEE 
		FROM information_schema.schema_privileges 
		WHERE TABLE_SCHEMA = ? 
		GROUP BY GRANTEE
	`

	rows, err := db.Query(query, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to query database privileges: %w", err)
	}
	defer rows.Close()

	var grantees []string
	for rows.Next() {
		var grantee string
		if err := rows.Scan(&grantee); err != nil {
			continue
		}
		grantees = append(grantees, grantee)
	}

	return grantees, nil
}

// GetUserGrants retrieves all grants for a specific user@host
func GetUserGrants(db *sql.DB, username, hostname string) ([]string, error) {
	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, hostname)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			continue
		}
		grants = append(grants, grant)
	}

	return grants, nil
}

// GetUserGrantsForDatabase retrieves grants for a specific user@host filtered by database
func GetUserGrantsForDatabase(db *sql.DB, username, hostname, dbName string) ([]string, error) {
	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, hostname)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			continue
		}

		// Filter grants to include only:
		// 1. USAGE grants (always needed for user creation)
		// 2. Grants specifically for the target database
		// 3. Global grants that might be needed (like GRANT OPTION)
		if isRelevantGrant(grant, dbName) {
			grants = append(grants, grant)
		}
	}

	return grants, nil
}

// isRelevantGrant checks if a grant statement is relevant for the specific database
func isRelevantGrant(grant string, dbName string) bool {
	grantUpper := strings.ToUpper(grant)

	// Always include USAGE grants (needed for user creation)
	if strings.Contains(grantUpper, "GRANT USAGE ON *.*") {
		return true
	}

	// Include grants specifically for the target database
	targetDBPattern := fmt.Sprintf("`%s`.*", dbName)
	if strings.Contains(grant, targetDBPattern) {
		return true
	}

	// Also check without backticks for different MySQL versions
	targetDBPatternNoBackticks := fmt.Sprintf("%s.*", dbName)
	return strings.Contains(grant, targetDBPatternNoBackticks)
}

// GetSystemUserHosts gets all hosts for a system user
func GetSystemUserHosts(db *sql.DB, systemUser string, lg *logger.Logger) []string {
	hostsQuery := "SELECT DISTINCT Host FROM mysql.user WHERE User = ?"
	rows, err := db.Query(hostsQuery, systemUser)
	if err != nil {
		lg.Warn("Failed to query hosts for system user",
			logger.String("user", systemUser),
			logger.Error(err))
		return nil
	}
	defer rows.Close()

	var hosts []string
	for rows.Next() {
		var host string
		if err := rows.Scan(&host); err != nil {
			continue
		}
		hosts = append(hosts, host)
	}

	return hosts
}

// UserExistsInMysql checks if user exists in mysql.user table
func UserExistsInMysql(db *sql.DB, username, hostname string, lg *logger.Logger) bool {
	// Use string formatting instead of prepared statement to avoid Error 1615
	userExistsQuery := fmt.Sprintf("SELECT COUNT(*) FROM mysql.user WHERE User = '%s' AND Host = '%s'",
		escapeStringLiteral(username), escapeStringLiteral(hostname))

	var count int
	err := db.QueryRow(userExistsQuery).Scan(&count)
	if err != nil {
		lg.Warn("Failed to query mysql.user table",
			logger.String("user", username),
			logger.String("host", hostname),
			logger.Error(err))
		return false
	}

	if count == 0 {
		lg.Debug("User not found in mysql.user table",
			logger.String("user", username),
			logger.String("host", hostname))
		return false
	}

	lg.Debug("User found in mysql.user table",
		logger.String("user", username),
		logger.String("host", hostname))
	return true
}

// escapeStringLiteral escapes single quotes in SQL string literals
func escapeStringLiteral(s string) string {
	// Replace single quotes with double single quotes for SQL escaping
	return strings.ReplaceAll(s, "'", "''")
}

// IsSystemUser checks if a username is a system user
func IsSystemUser(username string) bool {
	systemUsers := GetSystemUsersFromConfig()
	for _, systemUser := range systemUsers {
		if username == systemUser {
			return true
		}
	}
	return false
}

// ValidateDatabaseExists checks if a database exists
func ValidateDatabaseExists(db *sql.DB, dbName string) error {
	var exists bool
	query := "SELECT 1 FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"
	err := db.QueryRow(query, dbName).Scan(&exists)

	if err == sql.ErrNoRows {
		return fmt.Errorf("database %s does not exist", dbName)
	}
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	return nil
}

// GrantInfo represents a user grant
type GrantInfo struct {
	Username string
	Hostname string
	Database string
	Grants   []string
}

// GetSystemUsers retrieves system/admin users (root, mysql.sys, etc.)
func GetSystemUsers(db *sql.DB) ([]GrantInfo, error) {
	var users []GrantInfo
	userMap := make(map[string]*GrantInfo)

	// Get configured system users
	configSystemUsers := GetSystemUsersFromConfig()

	// Build the query to include both privilege-based and configured system users
	query := `
		SELECT DISTINCT User, Host 
		FROM mysql.user 
		WHERE (User LIKE 'root%' 
			OR User LIKE 'mysql%' 
			OR User LIKE 'admin%'
			OR User LIKE 'replication%'
			OR Super_priv = 'Y'
			OR Process_priv = 'Y'
			OR Reload_priv = 'Y'
			OR Shutdown_priv = 'Y'
			OR File_priv = 'Y'
			OR Grant_priv = 'Y'`

	// Add configured system users to the query
	if len(configSystemUsers) > 0 {
		for i := range configSystemUsers {
			if i == 0 {
				query += " OR User = ?"
			} else {
				query += " OR User = ?"
			}
		}
	}

	query += `
		)
		AND User != ''
		ORDER BY User, Host
	`

	// Prepare arguments for the query
	args := make([]interface{}, len(configSystemUsers))
	for i, user := range configSystemUsers {
		args[i] = user
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query system users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username, hostname string
		if err := rows.Scan(&username, &hostname); err != nil {
			return nil, fmt.Errorf("failed to scan system user: %w", err)
		}

		key := fmt.Sprintf("%s@%s", username, hostname)
		userMap[key] = &GrantInfo{
			Username: username,
			Hostname: hostname,
			Database: "*", // System users typically have global privileges
			Grants:   []string{},
		}
	}

	// Get grants for each system user
	for _, userInfo := range userMap {
		grants, err := GetUserGrants(db, userInfo.Username, userInfo.Hostname)
		if err != nil {
			// Skip users we can't get grants for (might not exist or permission issues)
			continue
		}

		userInfo.Grants = grants
		if len(userInfo.Grants) > 0 {
			users = append(users, *userInfo)
		}
	}

	return users, nil
}
