package connection

import (
	"fmt"

	"sfDBTools/internal/logger"
)

// ValidateUser checks if the user has sufficient privileges
func ValidateUser(config Config) error {
	lg, err := getLogger()
	if err != nil {
		return err
	}

	dsn := buildDSN(config, false) // Connect without selecting a database
	lg.Debug("Validating database user",
		logger.String("user", config.User))

	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Check user privileges
	rows, err := db.Query("SHOW GRANTS FOR CURRENT_USER()")
	if err != nil {
		lg.Error("Failed to retrieve user grants", logger.Error(err))
		return fmt.Errorf("failed to retrieve user grants: %w", err)
	}
	defer rows.Close()

	// Basic privileges needed for backup
	hasSelectPriv := false
	hasLockTablesPriv := false
	hasAllPrivileges := false

	lg.Debug("Checking user grants")

	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			lg.Error("Failed to scan grant row", logger.Error(err))
			continue
		}

		// Log the grant for debugging
		lg.Debug("Found grant", logger.String("grant", grant))

		// Check for ALL PRIVILEGES
		if contains(grant, "ALL PRIVILEGES") {
			hasAllPrivileges = true
			hasSelectPriv = true
			hasLockTablesPriv = true
			break
		}

		// Check for specific privileges
		if contains(grant, "SELECT") {
			hasSelectPriv = true
		}
		if contains(grant, "LOCK TABLES") {
			hasLockTablesPriv = true
		}
	}

	// Additional check for GRANT ALL without explicit "PRIVILEGES" keyword
	if !hasAllPrivileges && !hasSelectPriv {
		// Reconnect to check actual ability to SELECT
		testDb, _ := Get(config)
		if testDb != nil {
			// Try a simple SELECT query
			var result int
			err := testDb.QueryRow("SELECT 1").Scan(&result)
			if err == nil && result == 1 {
				lg.Debug("User can execute SELECT despite grants not showing it explicitly")
				hasSelectPriv = true
			}
			testDb.Close()
		}
	}

	if !hasSelectPriv {
		lg.Error("User lacks SELECT privilege required for backup")
		return fmt.Errorf("user '%s' lacks SELECT privilege required for backup", config.User)
	}

	if !hasLockTablesPriv && !hasAllPrivileges {
		lg.Warn("User lacks LOCK TABLES privilege which might affect backup consistency")
	}

	lg.Info("Database user has sufficient privileges",
		logger.String("user", config.User))
	return nil
}
